package deps

import (
	"context"
	"testing"

	"github.com/txie/verdiff/internal/verdiff"
)

func TestGoModParser(t *testing.T) {
	patch := `--- a/go.mod
+++ b/go.mod
-	github.com/foo/bar v1.2.3
-	github.com/old/dep v0.1.0
+	github.com/foo/bar v1.3.0
+	github.com/new/dep v2.0.0
`
	p := &GoModParser{}
	changes := p.ExtractChanges("go.mod", patch)

	if len(changes) == 0 {
		t.Fatal("expected version changes, got 0")
	}

	found := map[string]verdiff.VersionChange{}
	for _, c := range changes {
		found[c.Name] = c
	}

	if c, ok := found["github.com/foo/bar"]; ok {
		if c.OldVersion != "v1.2.3" || c.NewVersion != "v1.3.0" {
			t.Errorf("foo/bar: got %s→%s, want v1.2.3→v1.3.0", c.OldVersion, c.NewVersion)
		}
		if c.BumpType != verdiff.BumpMinor {
			t.Errorf("foo/bar bump = %s, want minor", c.BumpType)
		}
	} else {
		t.Error("missing foo/bar change")
	}

	if c, ok := found["github.com/old/dep"]; ok {
		if c.Direction != verdiff.DirRemoved {
			t.Errorf("old/dep direction = %s, want removed", c.Direction)
		}
	} else {
		t.Error("missing old/dep change")
	}

	if c, ok := found["github.com/new/dep"]; ok {
		if c.Direction != verdiff.DirAdded {
			t.Errorf("new/dep direction = %s, want added", c.Direction)
		}
	} else {
		t.Error("missing new/dep change")
	}
}

func TestPackageJSONParser(t *testing.T) {
	patch := `--- a/package.json
+++ b/package.json
-    "react": "^17.0.2",
+    "react": "^18.2.0",
+    "axios": "^1.4.0",
`
	p := &PackageJSONParser{}
	changes := p.ExtractChanges("package.json", patch)

	found := map[string]verdiff.VersionChange{}
	for _, c := range changes {
		found[c.Name] = c
	}

	if c, ok := found["react"]; ok {
		if c.BumpType != verdiff.BumpMajor {
			t.Errorf("react bump = %s, want major", c.BumpType)
		}
	} else {
		t.Error("missing react change")
	}

	if _, ok := found["axios"]; !ok {
		t.Error("missing axios (added)")
	}
}

func TestGemfileLockParser(t *testing.T) {
	patch := `--- a/Gemfile.lock
+++ b/Gemfile.lock
-    chef (17.10.0)
+    chef (18.4.0)
+    sysdig (0.1.0)
`
	p := &GemfileLockParser{}
	changes := p.ExtractChanges("Gemfile.lock", patch)

	if len(changes) == 0 {
		t.Fatal("expected changes")
	}

	found := map[string]verdiff.VersionChange{}
	for _, c := range changes {
		found[c.Name] = c
	}

	if c, ok := found["chef"]; ok {
		if c.OldVersion != "17.10.0" || c.NewVersion != "18.4.0" {
			t.Errorf("chef: %s→%s", c.OldVersion, c.NewVersion)
		}
	} else {
		t.Error("missing chef")
	}
}

func TestPolicyfileParser(t *testing.T) {
	patch := `--- a/Policyfile.rb
+++ b/Policyfile.rb
-cookbook 'rblx_docker', '1.4.3'
+cookbook 'rblx_docker', '1.5.0'
+cookbook 'rblx_sysdig', '0.1.0'
`
	p := &PolicyfileParser{}
	changes := p.ExtractChanges("Policyfile.rb", patch)

	found := map[string]verdiff.VersionChange{}
	for _, c := range changes {
		found[c.Name] = c
	}

	if c, ok := found["rblx_docker"]; ok {
		if c.OldVersion != "1.4.3" || c.NewVersion != "1.5.0" {
			t.Errorf("rblx_docker: %s→%s", c.OldVersion, c.NewVersion)
		}
	} else {
		t.Error("missing rblx_docker")
	}

	if c, ok := found["rblx_sysdig"]; ok {
		if c.Direction != verdiff.DirAdded {
			t.Errorf("rblx_sysdig direction = %s, want added", c.Direction)
		}
	} else {
		t.Error("missing rblx_sysdig")
	}
}

func TestSemverClassification(t *testing.T) {
	tests := []struct {
		old, new string
		want     verdiff.SemverBump
	}{
		{"1.0.0", "2.0.0", verdiff.BumpMajor},
		{"1.2.3", "1.3.0", verdiff.BumpMinor},
		{"1.2.3", "1.2.4", verdiff.BumpPatch},
		{"v1.0.0", "v2.0.0", verdiff.BumpMajor},
		{"abc", "def", verdiff.BumpUnknown},
	}
	for _, tt := range tests {
		got := classifyBump(tt.old, tt.new)
		if got != tt.want {
			t.Errorf("classifyBump(%s, %s) = %s, want %s", tt.old, tt.new, got, tt.want)
		}
	}
}

func TestTrackerMajorBumpFinding(t *testing.T) {
	tracker := NewTracker()
	input := &verdiff.AnalysisInput{
		Diff: verdiff.DiffResult{
			Files: []verdiff.FileDiff{
				{
					Path:       "go.mod",
					ChangeType: verdiff.ChangeModified,
					Patch: `-	github.com/lib/pq v1.10.0
+	github.com/lib/pq v2.0.0
`,
				},
			},
		},
		Config: verdiff.DefaultConfig(),
	}

	out, err := tracker.Analyze(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}

	if len(out.Findings) == 0 {
		t.Error("expected a finding for major version bump")
	}
}
