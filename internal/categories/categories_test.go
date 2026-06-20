package categories

import "testing"

func TestClassifyFile(t *testing.T) {
	tests := []struct {
		name string
		path string
		want ID
	}{
		{name: "installer", path: "Downloads/tool.DMG", want: Installer},
		{name: "archive", path: "Downloads/project.tar.gz", want: Archive},
		{name: "unknown", path: "Documents/notes.txt", want: Unknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyFile(tt.path)
			if got != tt.want {
				t.Fatalf("ClassifyFile(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestLookupOwnsCategoryPresentation(t *testing.T) {
	info, ok := Lookup(NodeModules)
	if !ok {
		t.Fatal("Lookup(NodeModules) ok = false, want true")
	}
	if info.Risk != RiskReviewRecommended {
		t.Fatalf("NodeModules risk = %q, want %q", info.Risk, RiskReviewRecommended)
	}
	if info.Group != "node_dependencies" {
		t.Fatalf("NodeModules group = %q, want node_dependencies", info.Group)
	}
}

func TestCleanupGroupsReturnsCatalogOrder(t *testing.T) {
	groups := CleanupGroups()
	if len(groups) == 0 {
		t.Fatal("CleanupGroups() returned no groups")
	}
	if groups[0].Categories[0] != PythonCache {
		t.Fatalf("first cleanup category = %q, want %q", groups[0].Categories[0], PythonCache)
	}
	if groups[0].Section != "Safe candidates" {
		t.Fatalf("first cleanup section = %q, want Safe candidates", groups[0].Section)
	}
}
