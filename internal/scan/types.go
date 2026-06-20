package scan

type ScanWarning struct {
	Path    string
	Type    string
	Message string
}

type Finding struct {
	Path string
	Size int64
}

type CategoryExplanation struct {
	Label  string
	Group  string
	Risk   string
	Why    string
	Action string
}

var CategoryExplanations = map[string]CategoryExplanation{
	"huggingface_cache": {
		Label:  "Hugging Face cache",
		Group:  "huggingface_cache",
		Risk:   "Review recommended",
		Why:    "Detected Hugging Face model and dataset cache.",
		Action: "Review old models and datasets before deleting.",
	},
	"node_modules": {
		Label:  "Node dependencies",
		Group:  "node_dependencies",
		Risk:   "Review recommended",
		Why:    "Detected node_modules directories.",
		Action: "Can usually be restored with npm/pnpm/yarn install.",
	},
	"installer": {
		Label:  "Installers and archives",
		Group:  "installers_and_archives",
		Risk:   "Usually safe",
		Why:    "Detected old DMG, PKG, ZIP, and TAR files.",
		Action: "Review installers that are no longer needed.",
	},
	"archive": {
		Label:  "Installers and archives",
		Group:  "installers_and_archives",
		Risk:   "Usually safe",
		Why:    "Detected old DMG, PKG, ZIP, and TAR files.",
		Action: "Review installers that are no longer needed.",
	},
	"python_virtual_environment": {
		Label:  "Python virtual environments",
		Group:  "python_virtual_environments",
		Risk:   "Review recommended",
		Why:    "Detected Python virtual environment directories.",
		Action: "Can usually be recreated from dependency files, but review before deleting.",
	},
	"python_cache": {
		Label:  "Python cache",
		Group:  "python_cache",
		Risk:   "Usually safe",
		Why:    "Detected Python __pycache__ directories.",
		Action: "Can usually be deleted; Python will recreate cache files as needed.",
	},
}

type Scanner struct {
	RootPath       string
	RootNode       *Node
	NumFiles       int
	NumDirectories int
	MaxDepth       int
	Warnings       []ScanWarning
	Categories     map[string][]Finding
}

type Node struct {
	Path     string
	TotSize  int64
	Type     string
	Children []*Node
	Depth    int
}

type Options struct {
	Path          string
	JSONReport    bool
	ExplainReport bool
	MaxDepth      int
	Limit         int
}
