package verdiff

// ChangeType classifies how a file changed between two versions.
type ChangeType string

const (
	ChangeAdded    ChangeType = "added"
	ChangeModified ChangeType = "modified"
	ChangeDeleted  ChangeType = "deleted"
	ChangeRenamed  ChangeType = "renamed"
)

// Severity indicates the importance level of a finding.
type Severity string

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityDanger  Severity = "danger"
)

// FileDiff represents the diff of a single file between two versions.
type FileDiff struct {
	Path         string     `json:"path"`
	OldPath      string     `json:"oldPath,omitempty"` // set when renamed
	ChangeType   ChangeType `json:"changeType"`
	LinesAdded   int        `json:"linesAdded"`
	LinesDeleted int        `json:"linesDeleted"`
	IsBinary     bool       `json:"isBinary"`
	Patch        string     `json:"-"` // raw patch text, omitted from JSON by default
}

// TotalChange returns the total number of changed lines.
func (f *FileDiff) TotalChange() int {
	return f.LinesAdded + f.LinesDeleted
}

// DiffResult is the top-level output of the git diff engine.
type DiffResult struct {
	RepoName string     `json:"repoName"`
	VersionA string     `json:"versionA"`
	VersionB string     `json:"versionB"`
	Files    []FileDiff `json:"files"`
	Tree     *DirNode   `json:"tree"`
	Hotspots []FileDiff `json:"hotspots"`
}

// DirNode represents a node in the directory change tree.
type DirNode struct {
	Name         string     `json:"name"`
	Path         string     `json:"path"`
	LinesAdded   int        `json:"linesAdded"`
	LinesDeleted int        `json:"linesDeleted"`
	FileCount    int        `json:"fileCount"`
	Children     []*DirNode `json:"children,omitempty"`
	Files        []FileDiff `json:"files,omitempty"`
}

// TotalChange returns the total number of changed lines in this directory subtree.
func (d *DirNode) TotalChange() int {
	return d.LinesAdded + d.LinesDeleted
}

// VersionChange tracks a version transition for a single dependency.
type VersionChange struct {
	Name       string          `json:"name"`
	Source     string          `json:"source"`     // e.g. "go.mod", "Gemfile.lock"
	OldVersion string          `json:"oldVersion"` // empty if newly added
	NewVersion string          `json:"newVersion"` // empty if removed
	BumpType   SemverBump      `json:"bumpType,omitempty"`
	Direction  ChangeDirection `json:"direction"`
}

// SemverBump indicates which part of a semver changed.
type SemverBump string

const (
	BumpMajor   SemverBump = "major"
	BumpMinor   SemverBump = "minor"
	BumpPatch   SemverBump = "patch"
	BumpUnknown SemverBump = "unknown"
)

// ChangeDirection indicates whether a dependency was added, removed, or changed.
type ChangeDirection string

const (
	DirUpgrade   ChangeDirection = "upgrade"
	DirDowngrade ChangeDirection = "downgrade"
	DirAdded     ChangeDirection = "added"
	DirRemoved   ChangeDirection = "removed"
)

// Finding is a single item reported by an analyzer.
type Finding struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Severity    Severity `json:"severity"`
	Category    string   `json:"category"`
	FilePaths   []string `json:"filePaths,omitempty"`
	OldValue    string   `json:"oldValue,omitempty"`
	NewValue    string   `json:"newValue,omitempty"`
}

// AnalysisResult is the combined output of all analyzers.
type AnalysisResult struct {
	Diff           DiffResult      `json:"diff"`
	VersionChanges []VersionChange `json:"versionChanges"`
	Findings       []Finding       `json:"findings"`
}
