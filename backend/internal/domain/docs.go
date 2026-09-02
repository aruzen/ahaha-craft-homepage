package domain

import "time"

type DocVault struct {
	Slug         string
	Title        string
	Branch       string
	LocalPath    string
	Status       string
	LastSyncedAt *time.Time
}

type DocNote struct {
	VaultSlug   string
	Slug        string
	Title       string
	Summary     string
	SourcePath  string
	ContentType string
	Published   bool
	Order       int
	Group       string
	Tags        []DocTag
	Metadata    DocNoteMetadata
	UpdatedAt   time.Time
}

type DocTag struct {
	Slug string
	Name string
}

type DocAsset struct {
	VaultSlug   string
	AssetPath   string
	ContentType string
	SizeBytes   int64
	UpdatedAt   time.Time
}

type DocNoteMetadata struct {
	Links  []DocReference `json:"links"`
	Embeds []DocReference `json:"embeds"`
}

type DocReference struct {
	Raw        string `json:"raw"`
	Label      string `json:"label,omitempty"`
	TargetSlug string `json:"target_slug,omitempty"`
	AssetPath  string `json:"asset_path,omitempty"`
}
