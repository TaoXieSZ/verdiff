package report

import (
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"

	"github.com/txie/verdiff/internal/verdiff"
)

// WriteHTML generates a self-contained HTML report file.
func WriteHTML(path string, result *verdiff.AnalysisResult, cfg verdiff.Config) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create report file: %w", err)
	}
	defer f.Close()

	data := buildTemplateData(result, cfg)

	tmpl, err := template.New("report").Funcs(templateFuncs()).Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	return tmpl.Execute(f, data)
}

type templateData struct {
	RepoName        string
	VersionA        string
	VersionB        string
	GeneratedAt     string
	TotalFiles      int
	TotalAdded      int
	TotalDeleted    int
	HotspotCount    int
	Files           []verdiff.FileDiff
	Hotspots        []verdiff.FileDiff
	Tree            *verdiff.DirNode
	VersionChanges  []verdiff.VersionChange
	BreakingChanges []verdiff.Finding
	OtherFindings   []verdiff.Finding
	DiffLineLimit   int
	DefaultTheme    string
}

func buildTemplateData(result *verdiff.AnalysisResult, cfg verdiff.Config) templateData {
	d := result.Diff
	totalAdded, totalDeleted := 0, 0
	for _, f := range d.Files {
		totalAdded += f.LinesAdded
		totalDeleted += f.LinesDeleted
	}

	return templateData{
		RepoName:        d.RepoName,
		VersionA:        d.VersionA,
		VersionB:        d.VersionB,
		GeneratedAt:     time.Now().Format("2006-01-02 15:04:05"),
		TotalFiles:      len(d.Files),
		TotalAdded:      totalAdded,
		TotalDeleted:    totalDeleted,
		HotspotCount:    len(d.Hotspots),
		Files:           d.Files,
		Hotspots:        d.Hotspots,
		Tree:            d.Tree,
		VersionChanges:  result.VersionChanges,
		BreakingChanges: filterByCategory(result.Findings, "breaking-change"),
		OtherFindings:   filterNotCategory(result.Findings, "breaking-change"),
		DiffLineLimit:   cfg.Report.DiffLineLimit,
		DefaultTheme:    cfg.Report.DefaultTheme,
	}
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"severity_class": func(s verdiff.Severity) string {
			switch s {
			case verdiff.SeverityDanger:
				return "danger"
			case verdiff.SeverityWarning:
				return "warning"
			default:
				return "info"
			}
		},
		"change_class": func(ct verdiff.ChangeType) string {
			switch ct {
			case verdiff.ChangeAdded:
				return "added"
			case verdiff.ChangeDeleted:
				return "deleted"
			case verdiff.ChangeRenamed:
				return "renamed"
			default:
				return "modified"
			}
		},
		"direction_symbol": func(d verdiff.ChangeDirection) string {
			switch d {
			case verdiff.DirUpgrade:
				return "↑"
			case verdiff.DirDowngrade:
				return "↓"
			case verdiff.DirAdded:
				return "+"
			case verdiff.DirRemoved:
				return "−"
			default:
				return "~"
			}
		},
		"bump_class": func(b verdiff.SemverBump) string {
			switch b {
			case verdiff.BumpMajor:
				return "major"
			case verdiff.BumpMinor:
				return "minor"
			default:
				return "patch"
			}
		},
		"render_diff": func(patch string) template.HTML {
			var b strings.Builder
			lines := strings.Split(patch, "\n")
			shown := 0
			limit := 500
			for _, line := range lines {
				if shown >= limit {
					b.WriteString(`<div class="line" style="color:var(--muted)">... truncated ...</div>`)
					break
				}
				cls := ""
				if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
					cls = "add"
				} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
					cls = "del"
				}
				escaped := template.HTMLEscapeString(line)
				if cls != "" {
					b.WriteString(`<div class="line ` + cls + `">` + escaped + `</div>`)
				} else {
					b.WriteString(`<div class="line">` + escaped + `</div>`)
				}
				shown++
			}
			return template.HTML(b.String())
		},
		"upper": strings.ToUpper,
		"add":   func(a, b int) int { return a + b },
		"pct": func(part, total int) float64 {
			if total == 0 {
				return 0
			}
			return float64(part) / float64(total) * 100
		},
	}
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en" data-theme="{{.DefaultTheme}}">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>verdiff: {{.RepoName}} {{.VersionA}}..{{.VersionB}}</title>
<style>
:root{--bg:#f8f9fa;--panel:#fff;--text:#1a1a2e;--muted:#6c757d;--border:#dee2e6;--accent:#4361ee;--green:#2d6a4f;--green-bg:#d8f3dc;--red:#9d0208;--red-bg:#fce4ec;--yellow:#e9c46a;--yellow-bg:#fff3cd;--blue-bg:#e3f2fd;--mono:ui-monospace,SFMono-Regular,Menlo,monospace;--radius:8px}
[data-theme=dark]{--bg:#0d1117;--panel:#161b22;--text:#e6edf3;--muted:#8b949e;--border:#30363d;--accent:#58a6ff;--green:#3fb950;--green-bg:rgba(63,185,80,.12);--red:#f85149;--red-bg:rgba(248,81,73,.12);--yellow:#d29922;--yellow-bg:rgba(210,153,34,.12);--blue-bg:rgba(88,166,255,.1)}
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;background:var(--bg);color:var(--text);line-height:1.6}
.container{max-width:1200px;margin:0 auto;padding:24px}
header{display:flex;justify-content:space-between;align-items:center;padding-bottom:16px;border-bottom:1px solid var(--border);margin-bottom:24px}
h1{font-size:22px;font-weight:700}
h1 .repo{color:var(--accent)}
h1 .arrow{color:var(--muted);margin:0 6px}
h2{font-size:16px;font-weight:600;margin:24px 0 12px;padding-bottom:6px;border-bottom:1px solid var(--border)}
.meta{font-size:12px;color:var(--muted);font-family:var(--mono)}
.stats{display:grid;grid-template-columns:repeat(auto-fit,minmax(160px,1fr));gap:12px;margin-bottom:24px}
.stat{background:var(--panel);border:1px solid var(--border);border-radius:var(--radius);padding:14px 16px;text-align:center}
.stat .value{font-size:28px;font-weight:700;font-family:var(--mono)}
.stat .label{font-size:12px;color:var(--muted);margin-top:2px}
.stat.green .value{color:var(--green)}
.stat.red .value{color:var(--red)}
.card{background:var(--panel);border:1px solid var(--border);border-radius:var(--radius);margin-bottom:16px;overflow:hidden}
.card-header{padding:10px 14px;font-weight:600;font-size:14px;cursor:pointer;display:flex;justify-content:space-between;align-items:center}
.card-header:hover{background:var(--blue-bg)}
.card-body{padding:0 14px 14px;display:none}
.card.open .card-body{display:block}
.card-header .chevron{transition:transform .2s}
.card.open .card-header .chevron{transform:rotate(90deg)}
table{width:100%;border-collapse:collapse;font-size:13px}
th,td{text-align:left;padding:6px 10px;border-bottom:1px solid var(--border)}
th{font-size:11px;text-transform:uppercase;letter-spacing:.04em;color:var(--muted);font-weight:600}
.pill{display:inline-block;font-size:11px;padding:1px 8px;border-radius:12px;font-weight:600;font-family:var(--mono)}
.pill.added{background:var(--green-bg);color:var(--green)}
.pill.deleted{background:var(--red-bg);color:var(--red)}
.pill.modified{background:var(--blue-bg);color:var(--accent)}
.pill.renamed{background:var(--yellow-bg);color:var(--yellow)}
.pill.major{background:var(--red-bg);color:var(--red)}
.pill.minor{background:var(--yellow-bg);color:var(--yellow)}
.pill.patch{background:var(--green-bg);color:var(--green)}
.pill.danger{background:var(--red-bg);color:var(--red)}
.pill.warning{background:var(--yellow-bg);color:var(--yellow)}
.pill.info{background:var(--blue-bg);color:var(--accent)}
code{font-family:var(--mono);font-size:.9em;background:var(--bg);padding:1px 4px;border-radius:3px}
.bar{height:6px;border-radius:3px;display:flex;overflow:hidden;background:var(--border)}
.bar .add{background:var(--green)}
.bar .del{background:var(--red)}
.dir-tree{font-size:13px;font-family:var(--mono)}
.dir-node{margin-left:16px}
.dir-row{display:flex;align-items:center;gap:6px;padding:3px 0;cursor:pointer;user-select:none}
.dir-row:hover{color:var(--accent)}
.dir-row .arrow{width:14px;text-align:center;font-size:10px;transition:transform .15s}
.dir-row.open>.arrow{transform:rotate(90deg)}
.dir-bar{height:4px;border-radius:2px;display:flex;overflow:hidden;background:var(--border);min-width:40px;max-width:120px;flex-shrink:0}
.dir-bar .a{background:var(--green)}.dir-bar .d{background:var(--red)}
.dir-files{margin-left:20px;display:none}
.dir-row.open+.dir-node>.dir-files,.dir-row.open+.dir-files{display:block}
.diff-block{background:var(--bg);border:1px solid var(--border);border-radius:4px;margin:6px 0;overflow-x:auto;font-family:var(--mono);font-size:11px;line-height:1.5;max-height:400px;overflow-y:auto}
.diff-block .line{padding:0 8px;white-space:pre}
.diff-block .line.add{background:var(--green-bg)}
.diff-block .line.del{background:var(--red-bg)}
.diff-toggle{font-size:11px;color:var(--accent);cursor:pointer;margin-top:4px;display:inline-block}
.search{padding:8px 12px;border:1px solid var(--border);border-radius:var(--radius);font-size:14px;width:100%;margin-bottom:16px;background:var(--panel);color:var(--text)}
.empty{padding:20px;text-align:center;color:var(--muted);font-style:italic}
.theme-toggle{cursor:pointer;background:var(--panel);border:1px solid var(--border);border-radius:var(--radius);padding:4px 10px;font-size:12px;color:var(--text)}
#file-list .file-row{padding:6px 10px;border-bottom:1px solid var(--border);font-size:13px;display:flex;align-items:center;gap:8px}
#file-list .file-row .path{flex:1;font-family:var(--mono);font-size:12px;word-break:break-all}
#file-list .file-row .nums{font-family:var(--mono);font-size:12px;white-space:nowrap}
.nums .a{color:var(--green)}.nums .d{color:var(--red)}
</style>
</head>
<body>
<div class="container">

<header>
  <h1><span class="repo">{{.RepoName}}</span> <span class="arrow">→</span> {{.VersionA}} <span class="arrow">..</span> {{.VersionB}}</h1>
  <div style="display:flex;gap:8px;align-items:center">
    <span class="meta">{{.GeneratedAt}}</span>
    <button class="theme-toggle" onclick="toggleTheme()">🌓</button>
  </div>
</header>

<div class="stats">
  <div class="stat"><div class="value">{{.TotalFiles}}</div><div class="label">Files Changed</div></div>
  <div class="stat green"><div class="value">+{{.TotalAdded}}</div><div class="label">Lines Added</div></div>
  <div class="stat red"><div class="value">-{{.TotalDeleted}}</div><div class="label">Lines Deleted</div></div>
  <div class="stat"><div class="value">{{.HotspotCount}}</div><div class="label">Hotspots</div></div>
</div>

{{if .Hotspots}}
<h2>🔥 Hotspots</h2>
<div class="card open">
<div class="card-body" style="display:block">
<table>
<tr><th>#</th><th>File</th><th>Added</th><th>Deleted</th><th>Total</th><th>Change</th></tr>
{{range $i,$h := .Hotspots}}
<tr>
  <td>{{add $i 1}}</td>
  <td><code>{{$h.Path}}</code></td>
  <td class="nums"><span class="a">+{{$h.LinesAdded}}</span></td>
  <td class="nums"><span class="d">-{{$h.LinesDeleted}}</span></td>
  <td>{{$h.TotalChange}}</td>
  <td><span class="pill {{change_class $h.ChangeType}}">{{$h.ChangeType}}</span></td>
</tr>
{{end}}
</table>
</div></div>
{{end}}

{{if .VersionChanges}}
<h2>📦 Dependency Changes ({{len .VersionChanges}})</h2>
<div class="card open">
<div class="card-body" style="display:block">
<table>
<tr><th>Name</th><th>Source</th><th>Old</th><th>New</th><th>Bump</th></tr>
{{range .VersionChanges}}
<tr>
  <td><code>{{.Name}}</code></td>
  <td><code>{{.Source}}</code></td>
  <td>{{if .OldVersion}}{{.OldVersion}}{{else}}—{{end}}</td>
  <td>{{if .NewVersion}}{{.NewVersion}}{{else}}—{{end}}</td>
  <td>{{if .BumpType}}<span class="pill {{bump_class .BumpType}}">{{.BumpType}}</span>{{end}} {{direction_symbol .Direction}}</td>
</tr>
{{end}}
</table>
</div></div>
{{end}}

{{if .BreakingChanges}}
<h2>⚠️ Breaking Change Candidates ({{len .BreakingChanges}})</h2>
{{range .BreakingChanges}}
<div class="card">
<div class="card-header" onclick="this.parentElement.classList.toggle('open')">
  <span><span class="pill {{severity_class .Severity}}">{{upper (print .Severity)}}</span> {{.Title}}</span>
  <span class="chevron">▸</span>
</div>
<div class="card-body">
  <p>{{.Description}}</p>
  {{if .FilePaths}}<p style="margin-top:4px">{{range .FilePaths}}<code>{{.}}</code> {{end}}</p>{{end}}
  {{if .OldValue}}<p style="margin-top:4px;font-size:12px;color:var(--muted)">Was: <code>{{.OldValue}}</code></p>{{end}}
  {{if .NewValue}}<p style="font-size:12px;color:var(--muted)">Now: <code>{{.NewValue}}</code></p>{{end}}
</div></div>
{{end}}
{{end}}

{{if .OtherFindings}}
<h2>ℹ️ Other Findings ({{len .OtherFindings}})</h2>
{{range .OtherFindings}}
<div class="card">
<div class="card-header" onclick="this.parentElement.classList.toggle('open')">
  <span><span class="pill {{severity_class .Severity}}">{{upper (print .Severity)}}</span> {{.Title}}</span>
  <span class="chevron">▸</span>
</div>
<div class="card-body">
  <p>{{.Description}}</p>
  {{if .FilePaths}}<p style="margin-top:4px">{{range .FilePaths}}<code>{{.}}</code> {{end}}</p>{{end}}
</div></div>
{{end}}
{{end}}

{{if .Tree}}
<h2>🌳 Directory Tree</h2>
<div class="dir-tree card open">
<div class="card-body" style="display:block">
{{template "dirnode" .Tree}}
</div></div>
{{end}}

<h2>📁 All Files ({{.TotalFiles}})</h2>
<input class="search" type="text" placeholder="Filter files..." oninput="filterFiles(this.value)" id="file-search">
<div id="file-list">
{{range .Files}}
<div class="file-row" data-path="{{.Path}}">
  <span class="pill {{change_class .ChangeType}}">{{.ChangeType}}</span>
  <span class="path">{{.Path}}{{if .OldPath}} <span style="color:var(--muted)">← {{.OldPath}}</span>{{end}}</span>
  <span class="nums"><span class="a">+{{.LinesAdded}}</span> <span class="d">-{{.LinesDeleted}}</span></span>
  {{if and .Patch (not .IsBinary)}}<span class="diff-toggle" onclick="toggleDiff(this)">show diff</span>
  <div class="diff-block" style="display:none">{{render_diff .Patch}}</div>{{end}}
</div>
{{end}}
</div>

</div>

{{define "dirnode"}}
<div class="dir-node">
{{range .Children}}
<div class="dir-row" onclick="this.classList.toggle('open')">
  <span class="arrow">▸</span>
  <strong>{{.Name}}/</strong>
  <span style="color:var(--muted);font-size:11px">{{.FileCount}} files</span>
  <span class="nums" style="font-size:11px"><span class="a">+{{.LinesAdded}}</span> <span class="d">-{{.LinesDeleted}}</span></span>
  <div class="dir-bar">
    {{if .LinesAdded}}<div class="a" style="flex:{{.LinesAdded}}"></div>{{end}}
    {{if .LinesDeleted}}<div class="d" style="flex:{{.LinesDeleted}}"></div>{{end}}
  </div>
</div>
{{template "dirnode" .}}
{{end}}
</div>
{{end}}

<script>
function toggleDiff(el){
  var block=el.nextElementSibling;
  if(block.style.display==='none'){block.style.display='block';el.textContent='hide diff'}
  else{block.style.display='none';el.textContent='show diff'}
}
function toggleTheme(){
  const h=document.documentElement;
  const t=h.getAttribute('data-theme');
  h.setAttribute('data-theme',t==='dark'?'light':'dark');
}
function filterFiles(q){
  q=q.toLowerCase();
  document.querySelectorAll('#file-list .file-row').forEach(r=>{
    r.style.display=r.dataset.path.toLowerCase().includes(q)?'':'none'
  });
}
(function(){
  const t=document.documentElement.getAttribute('data-theme');
  if(t==='auto'||!t){
    document.documentElement.setAttribute('data-theme',
      window.matchMedia('(prefers-color-scheme:dark)').matches?'dark':'light');
  }
})();
</script>
</body>
</html>`
