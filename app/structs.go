package app

type Package struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Source  struct {
		Type      string `json:"type"`
		URL       string `json:"url"`
		Reference string `json:"reference"`
	} `json:"source"`
	Homepage string `json:"homepage,omitempty"`
}

// ComposerLock struct
type ComposerLock struct {
	Checksum    string
	Packages    []Package `json:"packages"`
	PackagesDev []Package `json:"packages-dev"`
}

// ComposerDiffPackage struct
type ComposerDiffPackage struct {
	Name        string
	PreVersion  string
	PostVersion string //
	URL         string // url
	CompareURL  string // url
}

// ComposerDiff struct
type ComposerDiff struct {
	Checksum    string
	Packages    []ComposerDiffPackage
	Description string
}
