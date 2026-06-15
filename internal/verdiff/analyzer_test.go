package verdiff

import (
	"context"
	"testing"
)

type mockAnalyzer struct {
	name     string
	findings []Finding
	versions []VersionChange
}

func (m *mockAnalyzer) Name() string { return m.name }
func (m *mockAnalyzer) Analyze(_ context.Context, input *AnalysisInput) (*AnalysisOutput, error) {
	return &AnalysisOutput{
		Findings:       m.findings,
		VersionChanges: m.versions,
	}, nil
}

func TestRegistryRunAll(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&mockAnalyzer{
		name:     "first",
		versions: []VersionChange{{Name: "dep-a", OldVersion: "1.0", NewVersion: "2.0"}},
	})
	reg.Register(&mockAnalyzer{
		name:     "second",
		findings: []Finding{{Title: "something", Category: "test"}},
	})

	diff := DiffResult{RepoName: "test", VersionA: "v1", VersionB: "v2"}
	result, err := reg.RunAll(context.Background(), diff, DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}

	if len(result.VersionChanges) != 1 {
		t.Errorf("expected 1 version change, got %d", len(result.VersionChanges))
	}
	if len(result.Findings) != 1 {
		t.Errorf("expected 1 finding, got %d", len(result.Findings))
	}
}

func TestRegistryPassesPriorResults(t *testing.T) {
	var receivedPrior []VersionChange

	reg := NewRegistry()
	reg.Register(&mockAnalyzer{
		name:     "producer",
		versions: []VersionChange{{Name: "lib", OldVersion: "1.0", NewVersion: "2.0"}},
	})

	checker := &checkingAnalyzer{callback: func(input *AnalysisInput) {
		receivedPrior = input.PriorVersions
	}}
	reg.Register(checker)

	diff := DiffResult{}
	_, err := reg.RunAll(context.Background(), diff, DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}

	if len(receivedPrior) != 1 || receivedPrior[0].Name != "lib" {
		t.Errorf("second analyzer should receive prior version changes, got %v", receivedPrior)
	}
}

type checkingAnalyzer struct {
	callback func(*AnalysisInput)
}

func (c *checkingAnalyzer) Name() string { return "checker" }
func (c *checkingAnalyzer) Analyze(_ context.Context, input *AnalysisInput) (*AnalysisOutput, error) {
	c.callback(input)
	return &AnalysisOutput{}, nil
}
