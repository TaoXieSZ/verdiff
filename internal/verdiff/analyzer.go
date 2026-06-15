package verdiff

import (
	"context"
	"fmt"
)

// Analyzer is the interface that all analysis plugins must implement.
type Analyzer interface {
	// Name returns a human-readable identifier for this analyzer.
	Name() string
	// Analyze inspects the diff result (and optionally prior findings)
	// and returns its own findings.
	Analyze(ctx context.Context, input *AnalysisInput) (*AnalysisOutput, error)
}

// AnalysisInput is passed to each analyzer with the diff and accumulated results.
type AnalysisInput struct {
	Diff          DiffResult
	Config        Config
	PriorFindings []Finding
	PriorVersions []VersionChange
}

// AnalysisOutput is what an analyzer returns.
type AnalysisOutput struct {
	Findings       []Finding
	VersionChanges []VersionChange
}

// Registry holds registered analyzers in execution order.
type Registry struct {
	analyzers []Analyzer
}

// NewRegistry creates an empty analyzer registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Register adds an analyzer to the registry. Analyzers execute in registration order.
func (r *Registry) Register(a Analyzer) {
	r.analyzers = append(r.analyzers, a)
}

// RunAll executes all registered analyzers in order, passing accumulated
// results from prior analyzers to each subsequent one.
func (r *Registry) RunAll(ctx context.Context, diff DiffResult, cfg Config) (*AnalysisResult, error) {
	result := &AnalysisResult{Diff: diff}

	for _, a := range r.analyzers {
		input := &AnalysisInput{
			Diff:          diff,
			Config:        cfg,
			PriorFindings: result.Findings,
			PriorVersions: result.VersionChanges,
		}
		out, err := a.Analyze(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("analyzer %s: %w", a.Name(), err)
		}
		if out != nil {
			result.Findings = append(result.Findings, out.Findings...)
			result.VersionChanges = append(result.VersionChanges, out.VersionChanges...)
		}
	}

	return result, nil
}
