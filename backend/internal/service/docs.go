package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"backend/internal/domain"
	"backend/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrDocNotFound      = errors.New("docs: not found")
	ErrDocForbidden     = errors.New("docs: forbidden")
	ErrDocInvalidInput  = errors.New("docs: invalid input")
	ErrDocNotConfigured = errors.New("docs: not configured")
)

type DocService struct {
	repo           *repository.DocRepository
	sessionRepo    *repository.LoginSessionRepository
	userRepo       *repository.UserRepository
	logger         *log.Logger
	repoPath       string
	contentRoot    string
	cache          *docContentCache
	backupRepoPath string
	writeMu        sync.Mutex
}

type DocServiceConfig struct {
	RepoPath       string
	ContentRoot    string
	BackupRepoPath string
}

func NewDocService(docRepo *repository.DocRepository, sessionRepo *repository.LoginSessionRepository, userRepo *repository.UserRepository, logger *log.Logger, cfg DocServiceConfig) *DocService {
	if logger == nil {
		logger = log.Default()
	}
	contentRoot := strings.TrimSpace(cfg.ContentRoot)
	if contentRoot == "" {
		contentRoot = filepath.Join("data", "docs", "vaults")
	}
	return &DocService{
		repo:           docRepo,
		sessionRepo:    sessionRepo,
		userRepo:       userRepo,
		logger:         logger,
		repoPath:       strings.TrimSpace(cfg.RepoPath),
		contentRoot:    contentRoot,
		cache:          newDocContentCache(),
		backupRepoPath: strings.TrimSpace(cfg.BackupRepoPath),
	}
}

func (s *DocService) ListVaults(ctx context.Context) ([]domain.DocVault, error) {
	return s.repo.ListVaults(ctx, true)
}

func (s *DocService) ListToys(ctx context.Context) ([]domain.DocToy, error) {
	return s.repo.ListPublishedToys(ctx)
}

func (s *DocService) ListNotes(ctx context.Context, vaultSlug, tag, group string) ([]domain.DocNote, error) {
	if !validSlug(vaultSlug) {
		return nil, ErrDocInvalidInput
	}
	return s.repo.ListNotes(ctx, vaultSlug, tagSlug(tag), strings.TrimSpace(group), true)
}

func (s *DocService) GetNote(ctx context.Context, vaultSlug, noteSlug string) (domain.DocNote, error) {
	if !validSlug(vaultSlug) || !validSlug(noteSlug) {
		return domain.DocNote{}, ErrDocInvalidInput
	}
	note, err := s.repo.GetNote(ctx, vaultSlug, noteSlug, true)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DocNote{}, ErrDocNotFound
	}
	return note, err
}

func (s *DocService) ReadNoteContent(ctx context.Context, vaultSlug, noteSlug string) ([]byte, string, error) {
	note, err := s.GetNote(ctx, vaultSlug, noteSlug)
	if err != nil {
		return nil, "", err
	}
	vault, err := s.repo.GetVault(ctx, vaultSlug)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", ErrDocNotFound
	}
	if err != nil {
		return nil, "", err
	}
	fullPath, err := safeJoin(vault.LocalPath, note.SourcePath)
	if err != nil {
		return nil, "", err
	}
	content, err := s.cache.Read(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", ErrDocNotFound
		}
		return nil, "", err
	}
	content = []byte(stripFrontmatter(string(content)))
	return content, contentMime(note.SourcePath), nil
}

func (s *DocService) ReadAsset(ctx context.Context, vaultSlug, assetPath string) ([]byte, string, error) {
	if !validSlug(vaultSlug) || !isAllowedAsset(assetPath) {
		return nil, "", ErrDocInvalidInput
	}
	vault, err := s.repo.GetVault(ctx, vaultSlug)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", ErrDocNotFound
	}
	if err != nil {
		return nil, "", err
	}
	fullPath, err := safeJoin(vault.LocalPath, assetPath)
	if err != nil {
		return nil, "", err
	}
	content, err := s.cache.Read(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", ErrDocNotFound
		}
		return nil, "", err
	}
	return content, contentMime(assetPath), nil
}

func (s *DocService) AdminListBranches(ctx context.Context, session domain.SessionData) ([]string, error) {
	if err := s.requireAdmin(ctx, session); err != nil {
		return nil, err
	}
	if s.repoPath == "" {
		return nil, ErrDocNotConfigured
	}
	out, err := exec.CommandContext(ctx, "git", "-C", s.repoPath, "for-each-ref", "refs/heads", "--format=%(refname:short)").Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	branches := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			branches = append(branches, line)
		}
	}
	sort.Strings(branches)
	return branches, nil
}

func (s *DocService) AdminListVaults(ctx context.Context, session domain.SessionData) ([]domain.DocVault, error) {
	if err := s.requireAdmin(ctx, session); err != nil {
		return nil, err
	}
	return s.repo.ListVaults(ctx, false)
}

func (s *DocService) AdminListNotes(ctx context.Context, session domain.SessionData, vaultSlug string) ([]domain.DocNote, error) {
	if err := s.requireAdmin(ctx, session); err != nil {
		return nil, err
	}
	if !validSlug(vaultSlug) {
		return nil, ErrDocInvalidInput
	}
	return s.repo.ListNotes(ctx, vaultSlug, "", "", false)
}

func (s *DocService) AdminSetDefaultPublished(ctx context.Context, session domain.SessionData, vaultSlug string, value bool) error {
	if err := s.requireAdmin(ctx, session); err != nil {
		return err
	}
	if !validSlug(vaultSlug) {
		return ErrDocInvalidInput
	}
	if err := s.repo.SetVaultDefaultPublished(ctx, vaultSlug, value); errors.Is(err, pgx.ErrNoRows) {
		return ErrDocNotFound
	} else if err != nil {
		return err
	}
	vault, err := s.repo.GetVault(ctx, vaultSlug)
	if err != nil {
		return err
	}
	return s.RescanVault(ctx, vault)
}

func (s *DocService) AdminUpdateNoteMetadata(ctx context.Context, session domain.SessionData, vaultSlug, noteSlug string, override domain.DocNoteOverride) error {
	if err := s.requireAdmin(ctx, session); err != nil {
		return err
	}
	if !validSlug(vaultSlug) || !validSlug(noteSlug) {
		return ErrDocInvalidInput
	}
	vault, err := s.repo.GetVault(ctx, vaultSlug)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrDocNotFound
	}
	if err != nil {
		return err
	}
	if vault.SourceType == domain.DocSourceLocalUpload {
		return s.updateLocalNoteMetadata(ctx, vault, noteSlug, override)
	}
	if err := s.repo.UpsertNoteOverride(ctx, vaultSlug, noteSlug, override); err != nil {
		return err
	}
	return s.RescanVault(ctx, vault)
}

func (s *DocService) AdminRegisterVault(ctx context.Context, session domain.SessionData, branch, slug, title string) (domain.DocVault, error) {
	if err := s.requireAdmin(ctx, session); err != nil {
		return domain.DocVault{}, err
	}
	if err := validateBranchName(ctx, s.repoPath, branch); err != nil {
		return domain.DocVault{}, err
	}
	if slug = strings.TrimSpace(slug); slug == "" {
		slug = SafeDocSlug(branch)
	}
	if !validSlug(slug) {
		return domain.DocVault{}, ErrDocInvalidInput
	}
	if title = strings.TrimSpace(title); title == "" {
		title = branch
	}
	localPath, err := safeJoin(s.contentRoot, slug)
	if err != nil {
		return domain.DocVault{}, err
	}
	vault := domain.DocVault{
		Slug:       slug,
		Title:      title,
		Branch:     branch,
		LocalPath:  localPath,
		Status:     "active",
		SourceType: domain.DocSourceGitVault,
	}
	if err := s.repo.UpsertVault(ctx, vault); err != nil {
		return domain.DocVault{}, err
	}
	if err := s.syncVault(ctx, vault); err != nil {
		return domain.DocVault{}, err
	}
	return vault, nil
}

func (s *DocService) AdminDisableVault(ctx context.Context, session domain.SessionData, vaultSlug string) error {
	if err := s.requireAdmin(ctx, session); err != nil {
		return err
	}
	if !validSlug(vaultSlug) {
		return ErrDocInvalidInput
	}
	s.cache.InvalidateVault(vaultSlug)
	return s.repo.DisableVault(ctx, vaultSlug)
}

func (s *DocService) AdminSyncVault(ctx context.Context, session domain.SessionData, vaultSlug string) error {
	if err := s.requireAdmin(ctx, session); err != nil {
		return err
	}
	vault, err := s.repo.GetVault(ctx, vaultSlug)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrDocNotFound
	}
	if err != nil {
		return err
	}
	return s.syncVault(ctx, vault)
}

func (s *DocService) SyncRegisteredVaults(ctx context.Context) error {
	vaults, err := s.repo.ListVaults(ctx, false)
	if err != nil {
		return err
	}
	for _, vault := range vaults {
		if vault.Status != "active" || vault.SourceType != domain.DocSourceGitVault {
			continue
		}
		if err := s.syncVault(ctx, vault); err != nil {
			return fmt.Errorf("sync %s: %w", vault.Slug, err)
		}
	}
	return nil
}

func (s *DocService) syncVault(ctx context.Context, vault domain.DocVault) error {
	if err := validateBranchName(ctx, s.repoPath, vault.Branch); err != nil {
		return err
	}
	if err := s.exportBranch(ctx, vault.Branch, vault.LocalPath); err != nil {
		return err
	}
	s.cache.InvalidateVault(vault.Slug)
	if err := s.RescanVault(ctx, vault); err != nil {
		return err
	}
	return s.repo.MarkVaultSynced(ctx, vault.Slug, time.Now().UTC())
}

func (s *DocService) AdminRescanVault(ctx context.Context, session domain.SessionData, vaultSlug string) error {
	if err := s.requireAdmin(ctx, session); err != nil {
		return err
	}
	vault, err := s.repo.GetVault(ctx, vaultSlug)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrDocNotFound
	}
	if err != nil {
		return err
	}
	s.cache.InvalidateVault(vaultSlug)
	return s.RescanVault(ctx, vault)
}

func (s *DocService) AdminOverrideNotePublished(ctx context.Context, session domain.SessionData, vaultSlug, noteSlug string, published bool) error {
	return s.AdminUpdateNoteMetadata(ctx, session, vaultSlug, noteSlug, domain.DocNoteOverride{Published: &published})
}

func (s *DocService) RescanVault(ctx context.Context, vault domain.DocVault) error {
	s.cache.InvalidateVault(vault.Slug)
	notes, assets, err := scanDocVault(vault)
	if err != nil {
		return err
	}
	resolveNoteReferences(notes, assets)
	return s.repo.ReplaceScan(ctx, vault.Slug, notes, assets)
}

func (s *DocService) exportBranch(ctx context.Context, branch, targetDir string) error {
	tmpDir := fmt.Sprintf("%s.tmp-%s", targetDir, uuid.NewString())
	if err := os.MkdirAll(filepath.Dir(targetDir), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	gitCmd := exec.CommandContext(ctx, "git", "-C", s.repoPath, "archive", "--format=tar", branch)
	tarCmd := exec.CommandContext(ctx, "tar", "-x", "-C", tmpDir)

	reader, writer := io.Pipe()
	gitCmd.Stdout = writer
	tarCmd.Stdin = reader
	var gitErr bytes.Buffer
	var tarErr bytes.Buffer
	gitCmd.Stderr = &gitErr
	tarCmd.Stderr = &tarErr

	if err := tarCmd.Start(); err != nil {
		return err
	}
	if err := gitCmd.Start(); err != nil {
		_ = writer.Close()
		_ = tarCmd.Wait()
		return err
	}
	gitWaitErr := gitCmd.Wait()
	_ = writer.Close()
	tarWaitErr := tarCmd.Wait()
	_ = reader.Close()
	if gitWaitErr != nil {
		return fmt.Errorf("git archive failed: %w: %s", gitWaitErr, gitErr.String())
	}
	if tarWaitErr != nil {
		return fmt.Errorf("tar extract failed: %w: %s", tarWaitErr, tarErr.String())
	}

	backupDir := fmt.Sprintf("%s.old-%s", targetDir, uuid.NewString())
	if _, err := os.Stat(targetDir); err == nil {
		if err := os.Rename(targetDir, backupDir); err != nil {
			return err
		}
		defer func() { _ = os.RemoveAll(backupDir) }()
	}
	if err := os.Rename(tmpDir, targetDir); err != nil {
		if backupDir != "" {
			_ = os.Rename(backupDir, targetDir)
		}
		return err
	}
	return nil
}

func (s *DocService) requireAdmin(ctx context.Context, session domain.SessionData) error {
	loginSession, err := s.sessionRepo.Find(ctx, session.UserID(), session.Token())
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrDocForbidden
	}
	if err != nil {
		return err
	}
	if loginSession.IsExpired(time.Now()) {
		_ = s.sessionRepo.DeleteByID(ctx, loginSession.ID())
		return ErrDocForbidden
	}
	user, err := s.userRepo.FindByID(ctx, session.UserID())
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrDocForbidden
	}
	if err != nil {
		return err
	}
	if user.Role() != domain.UserRoleAdmin {
		return ErrDocForbidden
	}
	return nil
}

func scanDocVault(vault domain.DocVault) ([]domain.DocNote, []domain.DocAsset, error) {
	var notes []domain.DocNote
	var assets []domain.DocAsset
	err := filepath.WalkDir(vault.LocalPath, func(fullPath string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			name := entry.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(vault.LocalPath, fullPath)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		switch {
		case isNoteFile(rel):
			note, err := scanDocNoteFile(vault, rel, fullPath, info.ModTime())
			if err != nil {
				return err
			}
			notes = append(notes, note)
		case isAllowedAsset(rel):
			assets = append(assets, domain.DocAsset{
				VaultSlug:   vault.Slug,
				AssetPath:   rel,
				ContentType: contentMime(rel),
				SizeBytes:   info.Size(),
				UpdatedAt:   info.ModTime().UTC(),
			})
		}
		return nil
	})
	return notes, assets, err
}

func scanDocNoteFile(vault domain.DocVault, relPath, fullPath string, updatedAt time.Time) (domain.DocNote, error) {
	raw, err := os.ReadFile(fullPath)
	if err != nil {
		return domain.DocNote{}, err
	}
	meta, body := parseFrontmatter(string(raw))
	title := meta.String("title")
	if title == "" {
		title = strings.TrimSuffix(filepath.Base(relPath), filepath.Ext(relPath))
	}
	note := domain.DocNote{
		VaultSlug:   vault.Slug,
		Slug:        docSlugFromPath(relPath),
		Title:       title,
		Summary:     meta.String("summary"),
		SourcePath:  relPath,
		ContentType: noteContentType(relPath),
		Published:   vault.DefaultPublished,
		Order:       meta.Int("order"),
		Tags:        meta.Tags(),
		Metadata:    parseObsidianReferences(body),
		UpdatedAt:   updatedAt.UTC(),
	}
	if meta.Has("published") {
		note.Published = meta.Bool("published")
	}
	note.Group, note.ChapterPath = deriveDocLocation(vault.SourceType, relPath, meta.String("group"))
	return note, nil
}

func deriveDocLocation(sourceType, relPath, explicitGroup string) (string, string) {
	dir := filepath.ToSlash(filepath.Dir(relPath))
	if dir == "." {
		return explicitGroup, ""
	}
	parts := strings.Split(dir, "/")
	if sourceType == domain.DocSourceLocalUpload {
		if len(parts) < 2 || parts[0] != "books" {
			return explicitGroup, ""
		}
		group := parts[1]
		if explicitGroup != "" {
			group = explicitGroup
		}
		return group, strings.Join(parts[2:], "/")
	}
	group := parts[0]
	if explicitGroup != "" {
		group = explicitGroup
	}
	return group, strings.Join(parts[1:], "/")
}

func resolveNoteReferences(notes []domain.DocNote, assets []domain.DocAsset) {
	notesByName := map[string]string{}
	assetsByBase := map[string]string{}
	assetsByPath := map[string]string{}
	for _, note := range notes {
		notesByName[strings.ToLower(note.Title)] = note.Slug
		base := strings.TrimSuffix(filepath.Base(note.SourcePath), filepath.Ext(note.SourcePath))
		notesByName[strings.ToLower(base)] = note.Slug
	}
	for _, asset := range assets {
		assetsByPath[strings.ToLower(asset.AssetPath)] = asset.AssetPath
		assetsByBase[strings.ToLower(filepath.Base(asset.AssetPath))] = asset.AssetPath
	}
	for i := range notes {
		for j := range notes[i].Metadata.Links {
			key := strings.ToLower(notes[i].Metadata.Links[j].Raw)
			notes[i].Metadata.Links[j].TargetSlug = notesByName[key]
		}
		for j := range notes[i].Metadata.Embeds {
			key := strings.ToLower(notes[i].Metadata.Embeds[j].Raw)
			if asset, ok := assetsByPath[key]; ok {
				notes[i].Metadata.Embeds[j].AssetPath = asset
			} else {
				notes[i].Metadata.Embeds[j].AssetPath = assetsByBase[key]
			}
		}
	}
}

type frontmatter map[string]string

func parseFrontmatter(raw string) (frontmatter, string) {
	meta := frontmatter{}
	body := raw
	if strings.HasPrefix(raw, "---\n") || strings.HasPrefix(raw, "---\r\n") {
		normalized := strings.ReplaceAll(raw, "\r\n", "\n")
		rest := normalized[strings.Index(normalized, "\n")+1:]
		if end := strings.Index(rest, "\n---\n"); end >= 0 {
			header := rest[:end]
			body = rest[end+6:]
			for _, line := range strings.Split(header, "\n") {
				key, value, ok := strings.Cut(line, ":")
				if !ok {
					continue
				}
				key = strings.TrimSpace(strings.ToLower(key))
				value = strings.Trim(strings.TrimSpace(value), `"'`)
				if key != "" {
					meta[key] = value
				}
			}
		}
	}
	return meta, body
}

func stripFrontmatter(raw string) string {
	_, body := parseFrontmatter(raw)
	return body
}

func (m frontmatter) String(key string) string {
	return strings.TrimSpace(m[strings.ToLower(key)])
}

func (m frontmatter) Has(key string) bool { _, ok := m[strings.ToLower(key)]; return ok }

func (m frontmatter) Bool(key string) bool {
	value := strings.ToLower(m.String(key))
	return value == "true" || value == "yes" || value == "1"
}

func (m frontmatter) Int(key string) int {
	n, _ := strconv.Atoi(m.String(key))
	return n
}

func (m frontmatter) Tags() []domain.DocTag {
	raw := m.String("tags")
	raw = strings.Trim(raw, "[]")
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t'
	})
	tags := make([]domain.DocTag, 0, len(fields))
	seen := map[string]bool{}
	for _, field := range fields {
		name := strings.Trim(strings.TrimSpace(field), `"'#`)
		if name == "" {
			continue
		}
		slug := tagSlug(name)
		if slug == "" || seen[slug] {
			continue
		}
		seen[slug] = true
		tags = append(tags, domain.DocTag{Slug: slug, Name: name})
	}
	return tags
}

var obsidianPattern = regexp.MustCompile(`!?\[\[([^\]|#]+)(?:#[^\]|]+)?(?:\|([^\]]+))?\]\]`)

func parseObsidianReferences(body string) domain.DocNoteMetadata {
	var metadata domain.DocNoteMetadata
	matches := obsidianPattern.FindAllStringSubmatchIndex(body, -1)
	for _, match := range matches {
		full := body[match[0]:match[1]]
		raw := strings.TrimSpace(body[match[2]:match[3]])
		label := raw
		if match[4] >= 0 {
			label = strings.TrimSpace(body[match[4]:match[5]])
		}
		ref := domain.DocReference{Raw: raw, Label: label}
		if strings.HasPrefix(full, "![[") {
			metadata.Embeds = append(metadata.Embeds, ref)
		} else {
			metadata.Links = append(metadata.Links, ref)
		}
	}
	return metadata
}

func validateBranchName(ctx context.Context, repoPath, branch string) error {
	branch = strings.TrimSpace(branch)
	if repoPath == "" {
		return ErrDocNotConfigured
	}
	if branch == "" || strings.HasPrefix(branch, "-") || strings.Contains(branch, "..") || strings.Contains(branch, "\\") || strings.Contains(branch, "@{") {
		return ErrDocInvalidInput
	}
	if strings.ContainsAny(branch, "\x00 \t\r\n") {
		return ErrDocInvalidInput
	}
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "check-ref-format", "--branch", branch)
	if err := cmd.Run(); err != nil {
		return ErrDocInvalidInput
	}
	cmd = exec.CommandContext(ctx, "git", "-C", repoPath, "rev-parse", "--verify", "--quiet", branch)
	if err := cmd.Run(); err != nil {
		return ErrDocNotFound
	}
	return nil
}

func validSlug(value string) bool {
	if value == "" || value != SafeDocSlug(value) {
		return false
	}
	return !strings.Contains(value, "..") && !strings.Contains(value, "/")
}

func safeJoin(root, rel string) (string, error) {
	if strings.TrimSpace(root) == "" || strings.TrimSpace(rel) == "" {
		return "", ErrDocInvalidInput
	}
	if filepath.IsAbs(rel) || strings.Contains(rel, "\x00") {
		return "", ErrDocInvalidInput
	}
	cleanRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	fullPath := filepath.Join(cleanRoot, filepath.FromSlash(rel))
	cleanFull, err := filepath.Abs(fullPath)
	if err != nil {
		return "", err
	}
	if cleanFull != cleanRoot && !strings.HasPrefix(cleanFull, cleanRoot+string(os.PathSeparator)) {
		return "", ErrDocForbidden
	}
	return cleanFull, nil
}

func isNoteFile(rel string) bool {
	ext := strings.ToLower(filepath.Ext(rel))
	return ext == ".md" || ext == ".markdown" || ext == ".html" || ext == ".htm"
}

func noteContentType(rel string) string {
	ext := strings.ToLower(filepath.Ext(rel))
	if ext == ".html" || ext == ".htm" {
		return "html"
	}
	return "markdown"
}

func isAllowedAsset(rel string) bool {
	if strings.Contains(rel, "\x00") || strings.Contains(rel, "..") || strings.HasPrefix(rel, "/") {
		return false
	}
	switch strings.ToLower(filepath.Ext(rel)) {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg", ".avif", ".pdf":
		return true
	default:
		return false
	}
}

func contentMime(rel string) string {
	if contentType := mime.TypeByExtension(strings.ToLower(filepath.Ext(rel))); contentType != "" {
		return contentType
	}
	switch noteContentType(rel) {
	case "markdown":
		return "text/markdown; charset=utf-8"
	case "html":
		return "text/html; charset=utf-8"
	default:
		return http.DetectContentType([]byte(rel))
	}
}

type docContentCache struct {
	mu    sync.Mutex
	items map[string]docCacheItem
}

type docCacheItem struct {
	modTime time.Time
	size    int64
	data    []byte
}

func newDocContentCache() *docContentCache {
	return &docContentCache{items: map[string]docCacheItem{}}
}

func (c *docContentCache) Read(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	if item, ok := c.items[path]; ok && item.modTime.Equal(info.ModTime()) && item.size == info.Size() {
		data := append([]byte(nil), item.data...)
		c.mu.Unlock()
		return data, nil
	}
	c.mu.Unlock()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.items[path] = docCacheItem{modTime: info.ModTime(), size: info.Size(), data: append([]byte(nil), data...)}
	c.mu.Unlock()
	return data, nil
}

func (c *docContentCache) InvalidateVault(vaultSlug string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	marker := string(os.PathSeparator) + vaultSlug + string(os.PathSeparator)
	for path := range c.items {
		if strings.Contains(path, marker) {
			delete(c.items, path)
		}
	}
}
