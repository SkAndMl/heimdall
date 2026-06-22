package categories

import (
	"path/filepath"
	"strings"
)

type ID string

const (
	Unknown                  ID = "unknown"
	HuggingFaceCache         ID = "huggingface_cache"
	NodeModules              ID = "node_modules"
	Installer                ID = "installer"
	Archive                  ID = "archive"
	PythonVirtualEnvironment ID = "python_virtual_environment"
	PythonCache              ID = "python_cache"
)

type Info struct {
	Label  string
	Group  string
	Risk   string
	Why    string
	Action string
}

type DetailKind int

const (
	DetailCount DetailKind = iota
	DetailExtensions
)

type CleanupGroup struct {
	Section        string
	Label          string
	Action         string
	Order          int
	Categories     []ID
	DetailKind     DetailKind
	SingularDetail string
	PluralDetail   string
}

const (
	RiskUsuallySafe       = "Usually safe"
	RiskReviewRecommended = "Review recommended"
	ActionMoveToTrash     = "move to Trash"
	ActionSelectManually  = "select manually"
)

var infoByID = map[ID]Info{
	HuggingFaceCache: {
		Label:  "Hugging Face cache",
		Group:  "huggingface_cache",
		Risk:   RiskReviewRecommended,
		Why:    "Detected Hugging Face model and dataset cache.",
		Action: "Review old models and datasets before deleting.",
	},
	NodeModules: {
		Label:  "Node dependencies",
		Group:  "node_dependencies",
		Risk:   RiskReviewRecommended,
		Why:    "Detected node_modules directories.",
		Action: "Can usually be restored with npm/pnpm/yarn install.",
	},
	Installer: {
		Label:  "Installers and archives",
		Group:  "installers_and_archives",
		Risk:   RiskUsuallySafe,
		Why:    "Detected old DMG, PKG, ZIP, and TAR files.",
		Action: "Review installers that are no longer needed.",
	},
	Archive: {
		Label:  "Installers and archives",
		Group:  "installers_and_archives",
		Risk:   RiskUsuallySafe,
		Why:    "Detected old DMG, PKG, ZIP, and TAR files.",
		Action: "Review installers that are no longer needed.",
	},
	PythonVirtualEnvironment: {
		Label:  "Python virtual environments",
		Group:  "python_virtual_environments",
		Risk:   RiskReviewRecommended,
		Why:    "Detected Python virtual environment directories.",
		Action: "Can usually be recreated from dependency files, but review before deleting.",
	},
	PythonCache: {
		Label:  "Python cache",
		Group:  "python_cache",
		Risk:   RiskUsuallySafe,
		Why:    "Detected Python __pycache__ directories.",
		Action: "Can usually be deleted; Python will recreate cache files as needed.",
	},
}

var cleanupGroups = []CleanupGroup{
	{
		Section:        "Safe candidates",
		Label:          "Python bytecode cache",
		Action:         ActionMoveToTrash,
		Order:          0,
		Categories:     []ID{PythonCache},
		DetailKind:     DetailCount,
		SingularDetail: "__pycache__ directory",
		PluralDetail:   "__pycache__ directories",
	},
	{
		Section:    "Usually safe",
		Label:      "Installers and archives",
		Action:     ActionMoveToTrash,
		Order:      1,
		Categories: []ID{Installer, Archive},
		DetailKind: DetailExtensions,
	},
	{
		Section:        "Review recommended",
		Label:          "Python virtual environments",
		Action:         ActionSelectManually,
		Order:          2,
		Categories:     []ID{PythonVirtualEnvironment},
		DetailKind:     DetailCount,
		SingularDetail: "environment detected",
		PluralDetail:   "environments detected",
	},
	{
		Section:        "Review recommended",
		Label:          "Node dependencies",
		Action:         ActionSelectManually,
		Order:          3,
		Categories:     []ID{NodeModules},
		DetailKind:     DetailCount,
		SingularDetail: "node_modules directory",
		PluralDetail:   "node_modules directories",
	},
	{
		Section:        "Review recommended",
		Label:          "Hugging Face cache",
		Action:         ActionSelectManually,
		Order:          4,
		Categories:     []ID{HuggingFaceCache},
		DetailKind:     DetailCount,
		SingularDetail: "cache path detected",
		PluralDetail:   "cache paths detected",
	},
}

func Lookup(id ID) (Info, bool) {
	info, ok := infoByID[id]
	return info, ok
}

func CleanupGroups() []CleanupGroup {
	groups := make([]CleanupGroup, len(cleanupGroups))
	copy(groups, cleanupGroups)
	for i := range groups {
		groups[i].Categories = append([]ID(nil), cleanupGroups[i].Categories...)
	}
	return groups
}

func CleanupSections() []string {
	sectionsByName := make(map[string]bool)
	sections := make([]string, 0)
	for _, group := range cleanupGroups {
		if sectionsByName[group.Section] {
			continue
		}
		sectionsByName[group.Section] = true
		sections = append(sections, group.Section)
	}
	return sections
}

func ClassifyFile(path string) ID {
	name := strings.ToLower(filepath.Base(path))
	switch {
	case hasAnySuffix(name, installerSuffixes):
		return Installer
	case hasAnySuffix(name, archiveSuffixes):
		return Archive
	}
	return Unknown
}

func ArchiveExtension(path string) string {
	extension := strings.TrimPrefix(strings.ToUpper(filepath.Ext(path)), ".")
	if extension == "GZ" && strings.HasSuffix(strings.ToLower(path), ".tar.gz") {
		return "TAR.GZ"
	}
	if extension == "" {
		return "file"
	}
	return extension
}

func hasAnySuffix(name string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}

var installerSuffixes = []string{
	".dmg",
	".pkg",
	".mpkg",
	".msi",
	".exe",
	".deb",
	".rpm",
	".appimage",
}

var archiveSuffixes = []string{
	".tar.gz",
	".tar.bz2",
	".tgz",
	".tar.xz",
	".zip",
	".7z",
	".rar",
	".tar",
	".gz",
	".bz2",
	".xz",
}
