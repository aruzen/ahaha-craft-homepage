package service

import (
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
	return SafeDocSlug(strings.TrimSuffix(relPath, ext))
}

func tagSlug(value string) string {
	return SafeDocSlug(strings.TrimPrefix(value, "#"))
}
