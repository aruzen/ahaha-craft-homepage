package service

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrDocConflict = errors.New("docs: content exists")

const (
	maxUploadBytes   = 50 << 20
	maxExpandedBytes = 100 << 20
	maxArchiveFiles  = 1000
)

type DocUpload struct {
	Kind      string
	Slug      string
	Filename  string
	Data      []byte
	Overwrite bool
}

func (s *DocService) AdminUpload(ctx context.Context, session domain.SessionData, upload DocUpload) error {
	if err := s.requireAdmin(ctx, session); err != nil {
		return err
	}
	if !validSlug(upload.Slug) || len(upload.Data) == 0 || len(upload.Data) > maxUploadBytes {
		return ErrDocInvalidInput
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	vault, err := s.ensureUploadSource(ctx)
	if err != nil {
		return err
	}

	var target, staged string
	stageRoot := filepath.Join(vault.LocalPath, ".staging-"+uuid.NewString())
	defer os.RemoveAll(stageRoot)
	if err := os.MkdirAll(stageRoot, 0o755); err != nil {
		return err
	}
	switch upload.Kind {
	case "note":
		ext := strings.ToLower(filepath.Ext(upload.Filename))
		if ext != ".md" && ext != ".markdown" && ext != ".html" && ext != ".htm" {
			return ErrDocInvalidInput
		}
		target = filepath.Join(vault.LocalPath, "notes", upload.Slug+ext)
		staged = filepath.Join(stageRoot, upload.Slug+ext)
		if err := os.WriteFile(staged, upload.Data, 0o644); err != nil {
			return err
		}
	case "book":
		if strings.ToLower(filepath.Ext(upload.Filename)) != ".zip" {
			return ErrDocInvalidInput
		}
		target = filepath.Join(vault.LocalPath, "books", upload.Slug)
		staged = filepath.Join(stageRoot, upload.Slug)
		if err := extractDocArchive(upload.Data, staged); err != nil {
			return err
		}
	default:
		return ErrDocInvalidInput
	}
	if _, err := os.Stat(target); err == nil && !upload.Overwrite {
		return ErrDocConflict
	}
	return s.replaceLocalContent(ctx, vault, target, staged, "upload "+upload.Kind+" "+upload.Slug)
}

func (s *DocService) AdminTrashLocal(ctx context.Context, session domain.SessionData, noteSlug string) error {
	if err := s.requireAdmin(ctx, session); err != nil {
		return err
	}
	if !validSlug(noteSlug) {
		return ErrDocInvalidInput
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	vault, err := s.repo.GetVault(ctx, "uploads")
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrDocNotFound
	}
	if err != nil {
		return err
	}
	note, err := s.repo.GetNote(ctx, vault.Slug, noteSlug, false)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrDocNotFound
	}
	if err != nil {
		return err
	}
	target, err := safeJoin(vault.LocalPath, note.SourcePath)
	if err != nil {
		return err
	}
	parts := strings.Split(filepath.ToSlash(note.SourcePath), "/")
	if len(parts) >= 3 && parts[0] == "books" {
		target = filepath.Join(vault.LocalPath, "books", parts[1])
	}
	trash := filepath.Join(vault.LocalPath, ".trash", time.Now().UTC().Format("20060102T150405Z"), filepath.Base(target))
	if err := os.MkdirAll(filepath.Dir(trash), 0o755); err != nil {
		return err
	}
	if err := os.Rename(target, trash); err != nil {
		return err
	}
	if err := s.RescanVault(ctx, vault); err != nil {
		_ = os.Rename(trash, target)
		return err
	}
	if err := s.backupUploads(ctx, vault.LocalPath, "trash "+noteSlug); err != nil {
		_ = os.Rename(trash, target)
		_ = s.RescanVault(ctx, vault)
		return err
	}
	return nil
}

func (s *DocService) AdminListTrash(ctx context.Context, session domain.SessionData) ([]string, error) {
	if err := s.requireAdmin(ctx, session); err != nil {
		return nil, err
	}
	vault, err := s.repo.GetVault(ctx, "uploads")
	if errors.Is(err, pgx.ErrNoRows) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	root := filepath.Join(vault.LocalPath, ".trash")
	generations, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	var items []string
	for _, generation := range generations {
		entries, readErr := os.ReadDir(filepath.Join(root, generation.Name()))
		if readErr != nil {
			return nil, readErr
		}
		for _, entry := range entries {
			items = append(items, generation.Name()+"/"+entry.Name())
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(items)))
	return items, nil
}

func (s *DocService) AdminRestoreLocal(ctx context.Context, session domain.SessionData, trashPath string) error {
	if err := s.requireAdmin(ctx, session); err != nil {
		return err
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	vault, err := s.repo.GetVault(ctx, "uploads")
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrDocNotFound
	}
	if err != nil {
		return err
	}
	source, err := safeJoin(filepath.Join(vault.LocalPath, ".trash"), trashPath)
	if err != nil {
		return err
	}
	info, err := os.Stat(source)
	if os.IsNotExist(err) {
		return ErrDocNotFound
	}
	if err != nil {
		return err
	}
	var target string
	if info.IsDir() {
		target = filepath.Join(vault.LocalPath, "books", filepath.Base(source))
	} else {
		target = filepath.Join(vault.LocalPath, "notes", filepath.Base(source))
	}
	if _, err := os.Stat(target); err == nil {
		return ErrDocConflict
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	if err := os.Rename(source, target); err != nil {
		return err
	}
	if err := s.RescanVault(ctx, vault); err != nil {
		_ = os.Rename(target, source)
		return err
	}
	if err := s.backupUploads(ctx, vault.LocalPath, "restore "+trashPath); err != nil {
		_ = os.Rename(target, source)
		_ = s.RescanVault(ctx, vault)
		return err
	}
	return nil
}

func (s *DocService) ensureUploadSource(ctx context.Context) (domain.DocVault, error) {
	if vault, err := s.repo.GetVault(ctx, "uploads"); err == nil {
		return vault, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return domain.DocVault{}, err
	}
	localPath, err := safeJoin(s.contentRoot, "uploads")
	if err != nil {
		return domain.DocVault{}, err
	}
	vault := domain.DocVault{Slug: "uploads", Title: "Uploads", LocalPath: localPath, Status: "active", SourceType: domain.DocSourceLocalUpload}
	if err := os.MkdirAll(localPath, 0o755); err != nil {
		return domain.DocVault{}, err
	}
	if err := s.repo.UpsertVault(ctx, vault); err != nil {
		return domain.DocVault{}, err
	}
	return vault, nil
}

func (s *DocService) replaceLocalContent(ctx context.Context, vault domain.DocVault, target, staged, message string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	backup := filepath.Join(vault.LocalPath, ".rollback-"+uuid.NewString())
	existed := false
	if _, err := os.Stat(target); err == nil {
		existed = true
		if err := os.Rename(target, backup); err != nil {
			return err
		}
	}
	if err := os.Rename(staged, target); err != nil {
		if existed {
			_ = os.Rename(backup, target)
		}
		return err
	}
	rollback := func() {
		_ = os.RemoveAll(target)
		if existed {
			_ = os.Rename(backup, target)
		}
		_ = s.RescanVault(ctx, vault)
	}
	if err := s.RescanVault(ctx, vault); err != nil {
		rollback()
		return err
	}
	if err := s.backupUploads(ctx, vault.LocalPath, message); err != nil {
		rollback()
		return err
	}
	_ = os.RemoveAll(backup)
	return nil
}

func (s *DocService) updateLocalNoteMetadata(ctx context.Context, vault domain.DocVault, noteSlug string, override domain.DocNoteOverride) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	note, err := s.repo.GetNote(ctx, vault.Slug, noteSlug, false)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrDocNotFound
	}
	if err != nil {
		return err
	}
	path, err := safeJoin(vault.LocalPath, note.SourcePath)
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	meta, body := parseFrontmatter(string(raw))
	if override.Title != nil {
		meta["title"] = *override.Title
	}
	if override.Summary != nil {
		meta["summary"] = *override.Summary
	}
	if override.Published != nil {
		meta["published"] = strconv.FormatBool(*override.Published)
	}
	if override.Order != nil {
		meta["order"] = strconv.Itoa(*override.Order)
	}
	if override.Group != nil {
		meta["group"] = *override.Group
	}
	if override.Tags != nil {
		names := make([]string, len(*override.Tags))
		for i, tag := range *override.Tags {
			names[i] = tag.Name
		}
		meta["tags"] = "[" + strings.Join(names, ", ") + "]"
	}
	updated := renderFrontmatter(meta, body)
	return s.replaceLocalContent(ctx, vault, path, writeStagedFile(vault.LocalPath, filepath.Ext(path), []byte(updated)), "update metadata "+noteSlug)
}

func writeStagedFile(root, ext string, data []byte) string {
	path := filepath.Join(root, ".staging-"+uuid.NewString()+ext)
	_ = os.WriteFile(path, data, 0o644)
	return path
}

func renderFrontmatter(meta frontmatter, body string) string {
	keys := make([]string, 0, len(meta))
	for key := range meta {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString("---\n")
	for _, key := range keys {
		fmt.Fprintf(&b, "%s: %s\n", key, meta[key])
	}
	b.WriteString("---\n")
	b.WriteString(body)
	return b.String()
}

func extractDocArchive(data []byte, target string) error {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return ErrDocInvalidInput
	}
	if len(zr.File) == 0 || len(zr.File) > maxArchiveFiles {
		return ErrDocInvalidInput
	}
	var expanded int64
	for _, file := range zr.File {
		if file.Mode()&os.ModeSymlink != 0 || filepath.IsAbs(file.Name) {
			return ErrDocInvalidInput
		}
		name := filepath.Clean(filepath.FromSlash(file.Name))
		if name == "." || strings.HasPrefix(name, ".."+string(filepath.Separator)) {
			return ErrDocInvalidInput
		}
		path := filepath.Join(target, name)
		if !strings.HasPrefix(path, target+string(filepath.Separator)) {
			return ErrDocInvalidInput
		}
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(path, 0o755); err != nil {
				return err
			}
			continue
		}
		expanded += int64(file.UncompressedSize64)
		if expanded > maxExpandedBytes || file.UncompressedSize64 > 20<<20 {
			return ErrDocInvalidInput
		}
		if !isNoteFile(name) && !isAllowedAsset(name) {
			return ErrDocInvalidInput
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		rc, err := file.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err != nil {
			rc.Close()
			return ErrDocInvalidInput
		}
		_, copyErr := io.Copy(out, rc)
		closeErr := out.Close()
		_ = rc.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}

func (s *DocService) backupUploads(ctx context.Context, root, message string) error {
	bare := s.backupRepoPath
	if bare == "" {
		bare = "/srv/git/toyspace-uploads.git"
	}
	if _, err := os.Stat(bare); os.IsNotExist(err) {
		if out, err := exec.CommandContext(ctx, "git", "init", "--bare", bare).CombinedOutput(); err != nil {
			return fmt.Errorf("init backup git: %w: %s", err, out)
		}
	}
	if _, err := os.Stat(filepath.Join(root, ".git")); os.IsNotExist(err) {
		if out, err := exec.CommandContext(ctx, "git", "-C", root, "init").CombinedOutput(); err != nil {
			return fmt.Errorf("init uploads: %w: %s", err, out)
		}
	}
	commands := [][]string{{"config", "user.name", "ToySpace Admin"}, {"config", "user.email", "toyspace@localhost"}, {"remote", "remove", "backup"}, {"remote", "add", "backup", bare}, {"add", "--all"}}
	for _, args := range commands {
		cmd := exec.CommandContext(ctx, "git", append([]string{"-C", root}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil && !(args[0] == "remote" && args[1] == "remove") {
			return fmt.Errorf("git %s: %w: %s", args[0], err, out)
		}
	}
	cmd := exec.CommandContext(ctx, "git", "-C", root, "commit", "--allow-empty", "-m", message)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit: %w: %s", err, out)
	}
	cmd = exec.CommandContext(ctx, "git", "-C", root, "push", "-u", "backup", "HEAD:master")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git push: %w: %s", err, out)
	}
	return nil
}
