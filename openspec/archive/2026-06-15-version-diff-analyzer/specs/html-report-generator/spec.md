## ADDED Requirements

### Requirement: Self-contained HTML report
The system SHALL generate a single HTML file with all CSS and JavaScript inlined, requiring no external dependencies or network access to view.

#### Scenario: Open report offline
- **WHEN** user opens the generated HTML file in a browser without internet
- **THEN** the report renders fully with all styles, interactions, and data

### Requirement: Report sections
The report SHALL include the following sections:
1. Header with version A → B, repo name, generation timestamp
2. Change overview dashboard (total files changed, lines added/deleted, hotspot count)
3. Directory-level change tree with expandable/collapsible nodes
4. Dependency version change matrix
5. Breaking change candidates list
6. File-level detail view with inline diff (collapsed by default)

#### Scenario: All sections present
- **WHEN** analysis produces results across all dimensions
- **THEN** all six sections are present in the HTML report with data

#### Scenario: Empty section
- **WHEN** no breaking changes are detected
- **THEN** the breaking changes section shows a "none detected" message instead of being hidden

### Requirement: Interactive navigation
The report SHALL support client-side filtering and navigation: search/filter files by path, expand/collapse directory nodes, toggle between overview and detail views.

#### Scenario: File search
- **WHEN** user types a path fragment in the search box
- **THEN** the file list filters to show only matching files in real time

### Requirement: Theme support
The report SHALL support light and dark color themes, defaulting to the user's system preference with a toggle to override.

#### Scenario: Dark mode system preference
- **WHEN** user's OS is set to dark mode
- **THEN** report renders in dark theme by default

#### Scenario: Manual theme toggle
- **WHEN** user clicks the theme toggle button
- **THEN** report switches between light and dark themes

### Requirement: Large diff handling
The report SHALL handle large diffs gracefully: file diffs exceeding a configurable threshold (default 500 lines) SHALL be collapsed with a "show full diff" toggle. Overall report size SHALL be capped with a warning if truncation occurs.

#### Scenario: File with 2000-line diff
- **WHEN** a single file has more than 500 lines of diff
- **THEN** the diff is collapsed by default with a summary line and a "show full diff" button

### Requirement: Text summary output
The CLI SHALL support a `--format text` flag that outputs a plain-text summary to stdout instead of generating an HTML file, suitable for terminal and CI usage.

#### Scenario: Text output in CI
- **WHEN** user runs `verdiff v1 v2 --format text`
- **THEN** a structured plain-text summary is printed to stdout (no HTML file created)
