package service

import (
	"archive/zip"
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"backend/internal/domain"
)

func TestExtractDocArchive(t *testing.T) {
	var data bytes.Buffer
	zw := zip.NewWriter(&data)
	file, _ := zw.Create("chapters/01.md")
	_, _ = file.Write([]byte("---\npublished: true\n---\nchapter"))
	image, _ := zw.Create("images/cover.png")
	_, _ = image.Write([]byte("png"))
	_ = zw.Close()
	target := filepath.Join(t.TempDir(), "book")
	if err := extractDocArchive(data.Bytes(), target); err != nil {
		t.Fatalf("extract: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "chapters", "01.md")); err != nil {
		t.Fatalf("chapter missing: %v", err)
	}
}

func TestExtractDocArchiveRejectsTraversal(t *testing.T) {
	var data bytes.Buffer
	zw := zip.NewWriter(&data)
	file, _ := zw.Create("../escape.md")
	_, _ = file.Write([]byte("escape"))
	_ = zw.Close()
	if err := extractDocArchive(data.Bytes(), t.TempDir()); err == nil {
		t.Fatal("expected traversal rejection")
	}
}

func TestScanDocVaultFrontmatterAssetsAndObsidianReferences(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "Guide.md"), `---
published: true
title: Getting Started
tags: [go, docs]
summary: Intro
order: 2
group: handbook
---
# Body
[[Other Note|other]]
![[image.png]]
`)
	writeTestFile(t, filepath.Join(root, "Other Note.md"), `---
published: false
title: Other Note
---
hidden
`)
	writeTestFile(t, filepath.Join(root, "assets", "image.png"), "fake")

	vault := domain.DocVault{Slug: "main", LocalPath: root}
	notes, assets, err := scanDocVault(vault)
	if err != nil {
		t.Fatalf("scanDocVault returned error: %v", err)
	}
	resolveNoteReferences(notes, assets)

	if len(notes) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(notes))
	}
	if len(assets) != 1 || assets[0].AssetPath != "assets/image.png" {
		t.Fatalf("unexpected assets: %+v", assets)
	}

	var guide domain.DocNote
	for _, note := range notes {
		if note.Slug == "guide" {
			guide = note
		}
	}
	if guide.Title != "Getting Started" || !guide.Published || guide.Summary != "Intro" || guide.Order != 2 || guide.Group != "handbook" {
		t.Fatalf("unexpected guide metadata: %+v", guide)
	}
	if len(guide.Tags) != 2 || guide.Tags[0].Slug != "go" || guide.Tags[1].Slug != "docs" {
		t.Fatalf("unexpected tags: %+v", guide.Tags)
	}
	if len(guide.Metadata.Links) != 1 || guide.Metadata.Links[0].TargetSlug != "other-note" {
		t.Fatalf("unexpected links: %+v", guide.Metadata.Links)
	}
	if len(guide.Metadata.Embeds) != 1 || guide.Metadata.Embeds[0].AssetPath != "assets/image.png" {
		t.Fatalf("unexpected embeds: %+v", guide.Metadata.Embeds)
	}
}

func TestSafeJoinRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	if _, err := safeJoin(root, "../secret.md"); err == nil {
		t.Fatalf("expected traversal to be rejected")
	}
	if _, err := safeJoin(root, "/tmp/secret.md"); err == nil {
		t.Fatalf("expected absolute path to be rejected")
	}
	if _, err := safeJoin(root, "docs/note.md"); err != nil {
		t.Fatalf("expected safe path, got %v", err)
	}
}

func TestScanLocalUploadsSeparatesNotesAndBooks(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "notes", "single.md"), "---\npublished: true\n---\nsingle")
	writeTestFile(t, filepath.Join(root, "books", "guide", "chapters", "one.md"), "---\npublished: true\n---\none")
	vault := domain.DocVault{Slug: "uploads", LocalPath: root, SourceType: domain.DocSourceLocalUpload}
	notes, _, err := scanDocVault(vault)
	if err != nil {
		t.Fatal(err)
	}
	groups := map[string]string{}
	for _, note := range notes {
		groups[note.Title] = note.Group
	}
	if groups["single"] != "" {
		t.Fatalf("single note grouped as %q", groups["single"])
	}
	if groups["one"] != "guide" {
		t.Fatalf("book chapter group = %q", groups["one"])
	}
}

func TestExportBranchCopiesLocalGitBranch(t *testing.T) {
	ctx := context.Background()
	repo := t.TempDir()
	runGit(t, repo, "init")
	runGit(t, repo, "checkout", "-b", "vault/docs")
	writeTestFile(t, filepath.Join(repo, "Note.md"), "---\npublished: true\n---\nhello\n")
	runGit(t, repo, "add", "Note.md")
	runGit(t, repo, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-m", "add note")

	if err := validateBranchName(ctx, repo, "vault/docs"); err != nil {
		t.Fatalf("expected valid branch: %v", err)
	}
	if err := validateBranchName(ctx, repo, "../bad"); err == nil {
		t.Fatalf("expected invalid branch to be rejected")
	}

	target := filepath.Join(t.TempDir(), "vault")
	svc := NewDocService(nil, nil, nil, nil, DocServiceConfig{RepoPath: repo})
	if err := svc.exportBranch(ctx, "vault/docs", target); err != nil {
		t.Fatalf("exportBranch returned error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "Note.md")); err != nil {
		t.Fatalf("exported note missing: %v", err)
	}
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

func runGit(t *testing.T, repo string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(out))
	}
}
