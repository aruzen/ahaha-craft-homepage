package service

import (
	"crypto/sha256"
	"encoding/hex"
	"path"
	"regexp"
	"strings"
)

var nonSlugChar = regexp.MustCompile(`[^a-z0-9]+`)

func SafeDocSlug(value string) string {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	trimmed = strings.ReplaceAll(trimmed, "\\", "/")
	parts := strings.Split(trimmed, "/")
	safeParts := make([]string, 0, len(parts))
	for _, part := range parts {
		part = nonSlugChar.ReplaceAllString(part, "-")
		part = strings.Trim(part, "-")
		if part != "" {
			safeParts = append(safeParts, part)
		}
	}
	if len(safeParts) == 0 {
		return "docs"
	}
	return strings.Join(safeParts, "-")
}

func docSlugFromPath(relPath string) string {
	ext := path.Ext(relPath)
	stem := strings.TrimSuffix(strings.ReplaceAll(relPath, "\\", "/"), ext)
	base := SafeDocSlug(stem)
	if base == "docs" {
		base = "note"
	}
	if len(base) > 180 {
		base = strings.Trim(base[:180], "-")
	}
	sum := sha256.Sum256([]byte(strings.ToLower(stem)))
	return base + "-" + hex.EncodeToString(sum[:6])
}

func tagSlug(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(value, "#")))
	base := SafeDocSlug(normalized)
	if base == normalized {
		return base
	}
	sum := sha256.Sum256([]byte(normalized))
	return base + "-" + hex.EncodeToString(sum[:4])
}
