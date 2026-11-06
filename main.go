// rock-sql: a tiny read-only PostgreSQL explorer + periodic table + astronomy for rock.home.arpa
package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

var databases = []struct{ Name, Blurb string }{
	{"astronomy", "Solar system, planetary moons, exoplanets & Messier deep sky"},
	{"math", "Constants, polyhedra, number sequences, landmark theorems & curves"},
	{"anatomy", "Human osteology, skeletal joints, prime mover muscles & organ systems"},
	{"elements", "The periodic table: 118 elements, 28 properties each"},
	{"library", "Your nookbook library: books, authors, book_authors"},
	{"pagila", "DVD rental store: films, actors, customers, rentals, payments"},
	{"chinook", "Digital music store: artists, albums, tracks, invoices"},
	{"employees", "~300k employees with salaries & titles over time"},
	{"lego", "Every LEGO set, part, color and theme"},
	{"netflix", "Netflix catalog: shows & movies"},
	{"titanic", "Titanic passenger manifest & survival"},
	{"happiness_index", "World Happiness Report by country"},
	{"misc", "Odds & ends: holiday_songs"},
}

var (
	pools   = map[string]*sql.DB{}
	poolsMu sync.Mutex
	identRe = regexp.MustCompile(`^[A-Za-z0-9_]+$`)
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func sanitizeSpeechText(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	s = strings.Map(func(r rune) rune {
		if r == 0 || r < 0x20 {
			return -1
		}
		return r
	}, s)
	if len(s) > 600 {
		s = s[:600]
	}
	return s
}

func pool(name string) (*sql.DB, error) {
	ok := false
	for _, d := range databases {
		if d.Name == name {
			ok = true
		}
	}
	if !ok {
		return nil, fmt.Errorf("unknown database %q", name)
	}
	poolsMu.Lock()
	defer poolsMu.Unlock()
	if p, ok := pools[name]; ok {
		return p, nil
	}
	host := getEnv("PGHOST", "127.0.0.1")
	port := getEnv("PGPORT", "5432")
	user := getEnv("PGUSER", "explorer")
	pass := getEnv("PGPASSWORD", "explorer")
	ssl := getEnv("PGSSLMODE", "disable")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, pass, name, ssl)
	p, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	p.SetMaxOpenConns(4)
	pools[name] = p
	return p, nil
}

type Result struct {
	Cols     []string
	Rows     [][]string
	Err      string
	Elapsed  string
	Truncate bool
}

// runQuery executes SQL inside a READ ONLY transaction with a timeout and row cap.
func runQuery(db *sql.DB, q string, args ...any) Result {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return Result{Err: err.Error()}
	}
	defer tx.Rollback()
	_, _ = tx.ExecContext(ctx, "SET LOCAL statement_timeout = '15s'")
	rows, err := tx.QueryContext(ctx, q, args...)
	if err != nil {
		return Result{Err: err.Error()}
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	res := Result{Cols: cols}
	for rows.Next() {
		if len(res.Rows) >= 1000 {
			res.Truncate = true
			break
		}
		vals := make([]sql.NullString, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			res.Err = err.Error()
			break
		}
		row := make([]string, len(cols))
		for i, v := range vals {
			if v.Valid {
				row[i] = v.String
			} else {
				row[i] = "∅"
			}
		}
		res.Rows = append(res.Rows, row)
	}
	if err := rows.Err(); err != nil && res.Err == "" {
		res.Err = err.Error()
	}
	res.Elapsed = time.Since(start).Round(time.Millisecond).String()
	return res
}

const base = `{{define "top"}}<!DOCTYPE html><html><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1, viewport-fit=cover"><title>rock-sql{{if .Title}} // {{.Title}}{{end}}</title>
<style>
:root{--bg:#16161c;--panel:#1f1f28;--border:#2e2e3e;--text:#dcd7ba;--muted:#727169;--mint:#8abf9c;--rose:#d97c8a;--gold:#d4b47b;--cyan:#78b5ba;--blue:#7e9cd8}
*{box-sizing:border-box}body{margin:0;padding:18px;background:var(--bg);color:var(--text);font:13px/1.5 ui-monospace,Menlo,Consolas,monospace}
a{color:var(--cyan);text-decoration:none}a:hover{text-decoration:underline}
.wrap{max-width:1400px;margin:0 auto}header{display:flex;gap:18px;align-items:center;padding:12px 16px;background:var(--panel);border:1px solid var(--border);border-radius:8px;margin-bottom:16px;flex-wrap:wrap}
header h1{margin:0;font-size:17px;color:var(--cyan)}nav a{margin-right:14px}
.card{background:var(--panel);border:1px solid var(--border);border-radius:8px;padding:14px 16px;margin-bottom:16px;overflow-x:auto;-webkit-overflow-scrolling:touch}
h2{margin:0 0 10px;font-size:14px;color:var(--blue)}.muted{color:var(--muted)}
table{border-collapse:collapse;font-size:12px;width:100%}th,td{text-align:left;padding:4px 8px;border-bottom:1px solid #252532;vertical-align:top;max-width:420px;overflow:hidden;text-overflow:ellipsis}
th{color:var(--gold);font-weight:normal;position:sticky;top:0;background:var(--panel)}
.grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(300px,1fr));gap:14px}
textarea{width:100%;min-height:140px;background:#101015;color:var(--text);border:1px solid var(--border);border-radius:6px;padding:10px;font:13px ui-monospace,monospace}
button,.btn{background:#2a2a37;color:var(--text);border:1px solid var(--cyan);padding:6px 14px;border-radius:5px;font-family:inherit;cursor:pointer;display:inline-block}
button:hover,.btn:hover{background:#353545}
.btn.active{background:var(--cyan);color:#16161c;font-weight:bold}
.err{color:var(--rose);white-space:pre-wrap}.chips a{display:inline-block;margin:2px 6px 2px 0;padding:2px 8px;border:1px solid var(--border);border-radius:4px}
@media (max-width: 640px) {
  body { padding: 8px; font-size: 12px; }
  .wrap { width: 100%; }
  header { padding: 10px 12px; gap: 8px; flex-direction: column; align-items: flex-start; }
  header h1 { font-size: 16px; margin: 0; }
  header nav { display: flex; flex-wrap: wrap; gap: 6px 12px; }
  header nav a { margin-right: 0; font-size: 12px; }
  .card { padding: 10px 12px; margin-bottom: 12px; }
  .grid { grid-template-columns: 1fr; gap: 10px; }
  table { font-size: 11px; }
  th, td { padding: 3px 6px; max-width: 180px; }
}
</style></head><body><div class="wrap">
<header><h1>🐡 rock-sql</h1><nav><a href="/">databases</a><a href="/elements">periodic table</a><a href="/astronomy">astronomy</a><a href="/math">mathematics</a><a href="/anatomy">anatomy</a></nav></header>{{end}}
{{define "bottom"}}</div></body></html>{{end}}
{{define "result"}}{{if .Err}}<div class="err">ERROR: {{.Err}}</div>{{else}}
<div class="muted">{{len .Rows}} row(s){{if .Truncate}} (capped at 1000){{end}} in {{.Elapsed}}</div>
<table><tr>{{range .Cols}}<th>{{.}}</th>{{end}}</tr>{{range .Rows}}<tr>{{range .}}<td title="{{.}}">{{.}}</td>{{end}}</tr>{{end}}</table>{{end}}{{end}}`

var tpl = template.Must(template.New("").Funcs(template.FuncMap{
	"q": url.QueryEscape,
}).Parse(base + `
{{define "home"}}{{template "top" .}}<div class="grid">{{range .Dbs}}<div class="card">
<h2><a href="/db/{{.Name}}">{{.Name}}</a></h2><div class="muted">{{.Blurb}}</div>
<div style="margin-top:8px">{{.Size}} · {{.Tables}} tables · <a href="/db/{{.Name}}">explore →</a></div></div>{{end}}</div>{{template "bottom" .}}{{end}}

{{define "db"}}{{template "top" .}}<div class="card"><h2>{{.Name}}</h2><div class="muted">{{.Blurb}}</div></div>
<div class="card"><h2>Tables</h2><table><tr><th>table</th><th>~rows</th><th>size</th><th>columns</th></tr>
{{$d:=.Name}}{{range .Tables}}<tr><td><a href="/db/{{$d}}/t/{{.Schema}}/{{.Name}}">{{if ne .Schema "public"}}{{.Schema}}.{{end}}{{.Name}}</a></td><td>{{.Rows}}</td><td>{{.Size}}</td><td class="muted">{{.Cols}}</td></tr>{{end}}</table></div>{{template "bottom" .}}{{end}}

{{define "table"}}{{template "top" .}}<div class="card"><h2><a href="/db/{{.Db}}">{{.Db}}</a> / {{.Table}}</h2>
<div>page {{.Page}} · {{if gt .Page 1}}<a href="?page={{.Prev}}">← prev</a> · {{end}}<a href="?page={{.Next}}">next →</a></div></div>
<div class="card">{{template "result" .Res}}</div>{{template "bottom" .}}{{end}}

{{define "elements"}}{{template "top" .}}
<style>.pt{display:grid;grid-template-columns:repeat(18,1fr);gap:3px}.el{border-radius:4px;padding:4px 3px;min-height:58px;font-size:10px;color:#16161c;cursor:pointer;position:relative}
.el b{display:block;font-size:16px}.el .n{position:absolute;top:2px;right:4px;font-size:9px}.el:hover{outline:2px solid #fff}
#info{min-height:120px}.legend span{display:inline-block;padding:1px 8px;margin:2px;border-radius:3px;color:#16161c;font-size:11px}</style>
<div class="card"><h2>Periodic Table <span class="muted">(click an element · colored by type)</span></h2>
<div class="legend">{{range $t,$c := .Colors}}<span style="background:{{$c}}">{{$t}}</span>{{end}}</div>
<div class="pt" style="margin-top:10px">{{range .Cells}}{{if .Num}}<div class="el" style="grid-column:{{.Col}};grid-row:{{.Row}};background:{{.Color}}" data-num="{{.Num}}" onclick="show(this.dataset.num)">
<span class="n">{{.Num}}</span><b>{{.Sym}}</b>{{.Name}}</div>{{end}}{{end}}</div></div>
<div class="card" id="info"><span class="muted">Select an element…</span></div>
<script>const E={{.JSON}};function show(n){const e=E[n];if(!e)return;const wrap=document.getElementById('info');wrap.innerHTML='';const h2=document.createElement('h2');h2.textContent=e.Element+' ('+e.Symbol+') — #'+n;wrap.appendChild(h2);const table=document.createElement('table');for(const k in e){if(e[k]!==null&&e[k]!==''){const tr=document.createElement('tr');const th=document.createElement('th');th.style.width='220px';th.textContent=k;const td=document.createElement('td');td.textContent=String(e[k]);tr.appendChild(th);tr.appendChild(td);table.appendChild(tr);}}wrap.appendChild(table);}</script>
{{template "bottom" .}}{{end}}

{{define "astronomy"}}{{template "top" .}}
<style>
.tabbar{display:flex;gap:8px;flex-wrap:wrap;margin-bottom:14px;align-items:center}
.astro-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(210px,1fr));gap:10px}
.astro-card{background:#181822;border:1px solid var(--border);border-radius:6px;padding:10px;cursor:pointer;transition:transform .12s,border-color .15s}
.astro-card:hover{border-color:var(--cyan);outline:1px solid var(--cyan);transform:translateY(-2px)}
.astro-head{display:flex;align-items:center;gap:10px;margin-bottom:6px}
.astro-orb{width:22px;height:22px;border-radius:50%;display:inline-block;flex-shrink:0;box-shadow:0 0 6px rgba(0,0,0,0.6)}
.astro-title{font-size:13px;font-weight:bold;color:var(--text)}
.astro-sub{font-size:11px;color:var(--muted)}
.astro-badge{display:inline-block;padding:1px 6px;border-radius:3px;font-size:10px;margin-top:4px}
.badge-hz{background:#2b4436;color:var(--mint)}
.badge-ocean{background:#23374d;color:var(--cyan)}
.badge-dwarf{background:#3a3223;color:var(--gold)}
.badge-giant{background:#3d2b45;color:#c4a7e7}
.badge-terrestrial{background:#452828;color:var(--rose)}
.badge-deepsky{background:#223344;color:var(--blue)}

/* LCARS Starship Bridge Mode Styles */
.lcars-toggle{margin-left:auto;background:#3a2818;border:2px solid #ff9900;color:#ffcc66;font-weight:bold;border-radius:20px;padding:6px 16px;cursor:pointer;font-size:13px;box-shadow:0 0 12px rgba(255,153,0,0.3);transition:all .2s}
.lcars-toggle.on{background:#ff9900;color:#111;box-shadow:0 0 16px rgba(255,153,0,0.7)}
.lcars-frame{border-left:18px solid #ff9900;border-top:10px solid #ff9900;border-bottom:10px solid #cc6699;border-top-left-radius:28px;border-bottom-left-radius:28px;background:#0b0c14;padding:18px;margin-bottom:16px;position:relative;box-shadow:inset 0 0 40px rgba(0,0,0,0.9)}
.lcars-header{display:flex;justify-content:space-between;align-items:center;border-bottom:2px solid #33334d;padding-bottom:10px;margin-bottom:16px;flex-wrap:wrap;gap:10px}
.lcars-title{color:#ff9900;font-size:18px;font-weight:bold;letter-spacing:1.5px;text-transform:uppercase;margin:0}
.lcars-code{background:#cc6699;color:#111;font-weight:bold;padding:3px 10px;border-radius:12px;font-size:11px;letter-spacing:1px}
.lcars-layout{display:grid;grid-template-columns:minmax(280px, 420px) 1fr;gap:18px;align-items:start}
@media(max-width:860px){.lcars-layout{grid-template-columns:1fr}}
.viewscreen-wrap{position:relative;background:radial-gradient(circle at center, #101729 0%, #05060a 100%);border:2px solid #445577;border-radius:16px;overflow:hidden;box-shadow:0 0 24px rgba(120,181,186,0.2);text-align:center;user-select:none;touch-action:none;width:100%;max-width:420px;margin:0 auto}
#planet-canvas{width:100%;max-width:420px;height:auto;aspect-ratio:420/360;display:block;margin:0 auto;cursor:grab}
#planet-canvas:active{cursor:grabbing}
.viewscreen-hint{position:absolute;bottom:8px;left:0;right:0;text-align:center;font-size:11px;color:#88aacc;pointer-events:none;background:rgba(5,6,10,0.65);padding:3px 0}
.voice-box{background:#141624;border:1px solid #ff9900;border-left:8px solid #ff9900;border-radius:10px;padding:12px 16px;margin-bottom:14px}
.voice-status{display:flex;align-items:center;gap:10px;margin-bottom:6px}
.voice-label{color:#ff9900;font-weight:bold;font-size:12px;letter-spacing:1px;text-transform:uppercase}
.voice-bars{display:flex;gap:3px;align-items:flex-end;height:18px}
.voice-bar{width:4px;background:#66ccff;border-radius:2px;height:4px;transition:height .08s}
.voice-bars.speaking .voice-bar{background:#ff9900}
.voice-text{font-size:15px;line-height:1.5;color:#f0ebd8;min-height:44px}
.cadet-pad{display:grid;grid-template-columns:repeat(auto-fill,minmax(160px,1fr));gap:10px;margin-bottom:14px}
.lcars-btn{border:none;border-radius:22px;padding:12px 14px;font-family:inherit;font-size:13px;font-weight:bold;color:#111;cursor:pointer;text-align:left;display:flex;align-items:center;gap:8px;transition:transform .1s,filter .15s;box-shadow:0 3px 10px rgba(0,0,0,0.4)}
.lcars-btn:hover{filter:brightness(1.15);transform:scale(1.02)}
.lcars-btn:active{transform:scale(0.96)}
.lcars-btn .ico{font-size:20px;line-height:1}
.lcars-amber{background:#ff9900}
.lcars-cyan{background:#66ccff}
.lcars-mint{background:#8abf9c}
.lcars-rose{background:#dd6688}
.lcars-gold{background:#ffcc66}
.lcars-purple{background:#cc99ff}
.lcars-red{background:#e05555;color:#fff}
.telemetry-row{cursor:pointer;transition:background .15s}
.telemetry-row:hover{background:#232738}
.voice-controls{margin-top:10px;padding-top:10px;border-top:1px dashed #2e3450}
.lcars-select{background:#0e101b;color:#ffcc66;border:1px solid #ff9900;border-radius:6px;padding:4px 8px;font-family:inherit;font-size:12px;max-width:260px;cursor:pointer}
.lcars-select:focus{outline:none;border-color:#66ccff;box-shadow:0 0 8px rgba(102,204,255,0.4)}

@media (max-width: 640px) {
  .astro-grid{grid-template-columns:repeat(auto-fill,minmax(135px,1fr));gap:8px}
  .astro-card{padding:8px}
  .astro-title{font-size:12px}
  .astro-sub{font-size:10px}
  .tabbar{gap:6px}
  .tabbar button,.tabbar .btn{padding:5px 8px;font-size:11px}
  .lcars-toggle{width:100%;text-align:center;margin:6px 0;font-size:12px;padding:7px 12px}
  .lcars-frame{
    border-left:8px solid #ff9900;
    border-top:6px solid #ff9900;
    border-bottom:6px solid #cc6699;
    border-top-left-radius:16px;
    border-bottom-left-radius:16px;
    padding:10px 8px;
    margin-bottom:12px;
  }
  .lcars-header{flex-direction:column;align-items:flex-start;gap:6px}
  .lcars-title{font-size:13px;letter-spacing:1px}
  .lcars-code{font-size:10px}
  .viewscreen-wrap{max-width:100%;border-radius:12px}
  #planet-canvas{width:100%;height:auto;aspect-ratio:420/340}
  .voice-box{padding:10px;margin-bottom:10px}
  .voice-text{font-size:13px;min-height:36px}
  .cadet-pad{grid-template-columns:1fr 1fr;gap:8px}
  .lcars-btn{padding:9px 8px;font-size:11px;border-radius:14px;gap:6px}
  .lcars-btn .ico{font-size:16px}
  .lcars-select{max-width:100%;width:100%;font-size:12px}
}
</style>
<div class="card">
<div style="display:flex;align-items:center;justify-content:space-between;flex-wrap:wrap;gap:10px;margin-bottom:10px">
  <h2 style="margin:0">🔭 Astronomy & Solar System Explorer <span class="muted">(click Earth 🌍 or any body!)</span></h2>
  <button type="button" id="starship-toggle" class="lcars-toggle on" onclick="toggleStarshipMode()">🖖 Starship Computer Mode: ON</button>
</div>
<div class="tabbar">
  <button type="button" class="btn active" onclick="filterAstro('all', this)">All Solar System</button>
  <button type="button" class="btn" onclick="filterAstro('terrestrial', this)">Terrestrial</button>
  <button type="button" class="btn" onclick="filterAstro('giant', this)">Gas & Ice Giants</button>
  <button type="button" class="btn" onclick="filterAstro('dwarf', this)">Dwarf Planets</button>
  <button type="button" class="btn" onclick="filterAstro('ocean', this)">🌊 Ocean Moons</button>
  <button type="button" class="btn" onclick="filterAstro('exoplanet', this)">🔭 Exoplanets (HZ)</button>
  <button type="button" class="btn" onclick="filterAstro('messier', this)">🌌 Messier Deep Sky</button>
  <button type="button" class="btn" onclick="filterAstro('constellation', this)">✨ Constellations</button>
</div>
<div class="astro-grid" id="astro-list">
{{range .Planets}}
<div class="astro-card" data-cat="{{.Cat}}" data-kind="planet" data-key="{{.Name}}" onclick="inspectAstro(this.dataset.kind, this.dataset.key)">
  <div class="astro-head">
    <span class="astro-orb" style="background:{{.Color}}"></span>
    <div>
      <div class="astro-title">{{.Icon}} {{.Name}}</div>
      <div class="astro-sub">{{.Dist}} AU · {{.Moons}} moon(s)</div>
    </div>
  </div>
  <span class="astro-badge {{.BadgeClass}}">{{.Type}}</span>
</div>
{{end}}
{{range .Moons}}
<div class="astro-card" data-cat="moon{{if .Ocean}} ocean{{end}}" data-kind="moon" data-key="{{.Name}}" onclick="inspectAstro(this.dataset.kind, this.dataset.key)">
  <div class="astro-head">
    <span class="astro-orb" style="background:#a5a5b5"></span>
    <div>
      <div class="astro-title">{{.Name}}{{if .Ocean}} 🌊{{end}}</div>
      <div class="astro-sub">{{.Planet}} · {{.Diam}} km</div>
    </div>
  </div>
  <span class="astro-badge {{if .Ocean}}badge-ocean{{else}}badge-dwarf{{end}}">{{if .Ocean}}Subsurface Ocean{{else}}Major Moon{{end}}</span>
</div>
{{end}}
{{range .Exoplanets}}
<div class="astro-card" data-cat="exoplanet" data-kind="exoplanet" data-key="{{.Name}}" onclick="inspectAstro(this.dataset.kind, this.dataset.key)">
  <div class="astro-head">
    <span class="astro-orb" style="background:#5577aa"></span>
    <div>
      <div class="astro-title">{{.Name}}{{if .HZ}} 🌱{{end}}</div>
      <div class="astro-sub">{{.Host}} · {{.Dist}} ly</div>
    </div>
  </div>
  <span class="astro-badge {{if .HZ}}badge-hz{{else}}badge-giant{{end}}">{{.Type}}{{if .HZ}} · Habitable{{end}}</span>
</div>
{{end}}
{{range .Messier}}
<div class="astro-card" data-cat="messier" data-kind="messier" data-key="{{.Id}}" onclick="inspectAstro(this.dataset.kind, this.dataset.key)">
  <div class="astro-head">
    <span class="astro-orb" style="background:#b893ce"></span>
    <div>
      <div class="astro-title">{{.Id}}: {{.Name}}</div>
      <div class="astro-sub">{{.Constellation}} · mag {{.Mag}}</div>
    </div>
  </div>
  <span class="astro-badge badge-deepsky">{{.Type}}</span>
</div>
{{end}}
{{range .Constellations}}
<div class="astro-card" data-cat="constellation" data-kind="constellation" data-key="{{.Abbr}}" onclick="inspectAstro(this.dataset.kind, this.dataset.key)">
  <div class="astro-head">
    <span class="astro-orb" style="background:#d4b47b"></span>
    <div>
      <div class="astro-title">{{.Name}} ({{.Abbr}})</div>
      <div class="astro-sub">★ {{.BrightStar}} · {{.Area}} sq°</div>
    </div>
  </div>
  <span class="astro-badge badge-dwarf">{{.Quadrant}}</span>
</div>
{{end}}
</div>
</div>

<div class="card" id="astro-detail">
<span class="muted">Select a celestial body above (try clicking <b>🌍 Earth</b>!) to activate the Starship Enterprise viewscreen & computer voice…</span>
</div>

<script>
const AstroData = {{.JSON}};
let starshipMode = true;
let audioCtx = null;
let animFrameId = null;
let voiceInterval = null;
let cachedVoices = [];
let selectedVoiceURI = '';
try { selectedVoiceURI = localStorage.getItem('rocksql_voice_uri') || ''; } catch(e){}

function populateVoiceDropdown() {
  const sel = document.getElementById('voice-select');
  if (!sel) return;
  const voices = cachedVoices.length ? cachedVoices : (('speechSynthesis' in window) ? window.speechSynthesis.getVoices() : []);
  if (!voices || voices.length === 0) {
    sel.innerHTML = '<option value="">Standard Browser Voice</option>';
    return;
  }

  const isHighQuality = (v) => /natural|neural|google|premium|enhanced|samantha|serena|karen|daniel|victoria|alex|moira|oliver|fiona/i.test(v.name);

  const sorted = [...voices].sort((a, b) => {
    const aEn = /en/i.test(a.lang);
    const bEn = /en/i.test(b.lang);
    if (aEn && !bEn) return -1;
    if (!aEn && bEn) return 1;
    const aHQ = isHighQuality(a);
    const bHQ = isHighQuality(b);
    if (aHQ && !bHQ) return -1;
    if (!aHQ && bHQ) return 1;
    return a.name.localeCompare(b.name);
  });

  sel.innerHTML = '';
  const defOpt = document.createElement('option');
  defOpt.value = '';
  defOpt.textContent = '🌟 Best Natural English Voice';
  sel.appendChild(defOpt);

  const streamOpt = document.createElement('option');
  streamOpt.value = 'server_stream';
  streamOpt.textContent = '📻 Starship Audio Stream (Mobile Audio)';
  if (selectedVoiceURI === 'server_stream') streamOpt.selected = true;
  sel.appendChild(streamOpt);

  sorted.forEach(v => {
    const opt = document.createElement('option');
    opt.value = v.voiceURI || v.name;
    const isEn = /en/i.test(v.lang);
    const star = isHighQuality(v) ? ' ⭐' : '';
    opt.textContent = (isEn ? '🗣️ ' : '🌐 ') + v.name + star + ' (' + v.lang + ')';
    if (opt.value === selectedVoiceURI) {
      opt.selected = true;
    }
    sel.appendChild(opt);
  });
}

function onVoiceChanged(val) {
  selectedVoiceURI = val;
  try { localStorage.setItem('rocksql_voice_uri', val); } catch(e){}
  computerSpeak('Voice ready.', 'chirp');
}

function testSelectedVoice() {
  computerSpeak('Starship computer voice test on Earth.', 'chirp');
}

if ('speechSynthesis' in window) {
  cachedVoices = window.speechSynthesis.getVoices();
  window.speechSynthesis.onvoiceschanged = () => {
    cachedVoices = window.speechSynthesis.getVoices();
    populateVoiceDropdown();
  };
}

// Interactive Globe State
const Globe = {
  kind: 'planet',
  name: 'Earth',
  obj: null,
  angle: 0.4,
  tilt: 0.2,
  spinSpeed: 0.007,
  dragging: false,
  lastX: 0,
  lastY: 0,
  clouds: true,
  shields: false,
  warp: false,
  moonAngle: 0.8,
  moonBeam: 0,
  scanWave: 0,
  lightning: 0,
  popups: [],
  stars: Array.from({length: 70}, () => ({
    x: Math.random(),
    y: Math.random(),
    s: Math.random() * 2 + 0.6,
    p: Math.random() * 6.28
  }))
};

// Simplified Earth continent polygons (lon in [-PI, PI], lat in [-PI/2, PI/2])
const EarthContinents = [
  // North America
  { color: '#4ca66b', pts: [[-2.2,0.85],[-1.6,1.05],[-1.1,0.85],[-1.3,0.45],[-1.55,0.28],[-1.85,0.45],[-2.1,0.65]] },
  // South America
  { color: '#58b368', pts: [[-1.4,0.2],[-1.15,0.05],[-0.65,-0.15],[-0.8,-0.65],[-1.15,-0.95],[-1.3,-0.45]] },
  // Europe
  { color: '#5ca36c', pts: [[-0.15,0.65],[0.1,0.95],[0.65,1.05],[0.7,0.75],[0.3,0.65],[0.0,0.62]] },
  // Africa
  { color: '#8ab864', pts: [[-0.3,0.58],[0.25,0.55],[0.65,0.35],[0.85,0.18],[0.65,-0.45],[0.35,-0.62],[0.15,-0.1],[-0.25,0.2]] },
  // Asia
  { color: '#52a367', pts: [[0.65,0.98],[1.4,1.1],[2.35,1.05],[2.5,0.65],[1.9,0.25],[1.35,0.15],[1.0,0.42],[0.68,0.68]] },
  // Australia
  { color: '#c99e5a', pts: [[1.95,-0.25],[2.45,-0.22],[2.6,-0.55],[2.25,-0.65],[1.92,-0.5]] },
  // Greenland
  { color: '#d8f0f8', pts: [[-0.95,1.05],[-0.45,1.25],[-0.35,1.05],[-0.75,0.98]] }
];

const EarthDiscoveries = [
  { icon: '🐋', title: 'Blue Whale!', speech: 'Sensor lock! A giant blue whale singing in the deep Pacific Ocean!' },
  { icon: '🦖', title: 'Dinosaur Fossil!', speech: 'Subsurface scan! We found a Tyrannosaurus Rex fossil buried in the rock!' },
  { icon: '🐘', title: 'African Elephant!', speech: 'Lifeform detected! A family of elephants splashing water with their trunks!' },
  { icon: '🐧', title: 'Emperor Penguins!', speech: 'Polar scan! Playful penguins waddling across the sparkling snow and ice!' },
  { icon: '🌋', title: 'Rumbling Volcano!', speech: 'Thermal reading! A glowing volcano bubbling with warm orange lava!' },
  { icon: '🌳', title: 'Amazon Rainforest!', speech: 'Botanical scan! Millions of tall green trees, colorful parrots, and tree frogs!' },
  { icon: '🚀', title: 'Space Station!', speech: 'Orbital contact! Astronauts waving hello from the International Space Station!' },
  { icon: '🐬', title: 'Jumping Dolphins!', speech: 'Marine sensors detect happy dolphins leaping over the ocean waves!' }
];
let discoveryIdx = 0;

function ensureAudio() {
  if (!audioCtx) {
    const AC = window.AudioContext || window.webkitAudioContext;
    if (AC) audioCtx = new AC();
  }
  if (audioCtx && audioCtx.state === 'suspended') {
    audioCtx.resume();
  }
}

// Synthesize authentic Starfleet LCARS computer bleeps & sound effects
function playLcarsSound(type) {
  try {
    ensureAudio();
    if (!audioCtx) return;
    const now = audioCtx.currentTime;
    const playTone = (freq, start, dur, wave, gainVal) => {
      const osc = audioCtx.createOscillator();
      const gain = audioCtx.createGain();
      osc.type = wave || 'sine';
      osc.frequency.setValueAtTime(freq, now + start);
      gain.gain.setValueAtTime(gainVal || 0.12, now + start);
      gain.gain.exponentialRampToValueAtTime(0.001, now + start + dur);
      osc.connect(gain);
      gain.connect(audioCtx.destination);
      osc.start(now + start);
      osc.stop(now + start + dur);
    };

    if (type === 'chirp') {
      // Classic Enterprise two-tone computer activation chirp
      playTone(1320, 0.00, 0.06, 'sine', 0.12);
      playTone(1760, 0.065, 0.08, 'sine', 0.14);
      playTone(2217, 0.15, 0.11, 'sine', 0.10);
    } else if (type === 'scan') {
      // Tricorder warble sweep
      for (let i = 0; i < 6; i++) {
        playTone(900 + (i % 2) * 350 + i * 60, i * 0.06, 0.055, 'triangle', 0.11);
      }
    } else if (type === 'shield') {
      playTone(220, 0.0, 0.45, 'sawtooth', 0.06);
      playTone(440, 0.05, 0.40, 'sine', 0.10);
      playTone(660, 0.10, 0.35, 'sine', 0.08);
    } else if (type === 'moon') {
      playTone(880, 0.0, 0.12, 'sine', 0.12);
      playTone(1320, 0.12, 0.25, 'sine', 0.12);
    } else if (type === 'warp') {
      const osc = audioCtx.createOscillator();
      const gain = audioCtx.createGain();
      osc.type = 'sawtooth';
      osc.frequency.setValueAtTime(120, now);
      osc.frequency.exponentialRampToValueAtTime(980, now + 0.55);
      gain.gain.setValueAtTime(0.10, now);
      gain.gain.exponentialRampToValueAtTime(0.001, now + 0.6);
      osc.connect(gain);
      gain.connect(audioCtx.destination);
      osc.start(now);
      osc.stop(now + 0.6);
    } else if (type === 'weather') {
      playTone(520, 0.0, 0.1, 'sine', 0.1);
      playTone(390, 0.1, 0.2, 'triangle', 0.1);
    }
  } catch (e) {}
}

function setVoiceVisualizer(active) {
  const box = document.getElementById('voice-bars');
  const statusEl = document.getElementById('voice-state');
  if (!box) return;
  if (voiceInterval) {
    clearInterval(voiceInterval);
    voiceInterval = null;
  }
  if (active) {
    box.classList.add('speaking');
    if (statusEl) statusEl.textContent = 'USS ENTERPRISE COMPUTER // AUDIO ACTIVE';
    voiceInterval = setInterval(() => {
      box.querySelectorAll('.voice-bar').forEach(b => {
        b.style.height = (Math.floor(Math.random() * 15) + 4) + 'px';
      });
    }, 80);
  } else {
    box.classList.remove('speaking');
    if (statusEl) statusEl.textContent = 'USS ENTERPRISE COMPUTER // STANDING BY';
    box.querySelectorAll('.voice-bar').forEach(b => { b.style.height = '4px'; });
  }
}

window.activeSpeechUtterance = null;
let streamAudioEl = null;
let speechSessionId = 0;
let speechStartTimer = null;
let speechFallbackTimer = null;

function clearSpeechTimers() {
  if (speechStartTimer) {
    clearTimeout(speechStartTimer);
    speechStartTimer = null;
  }
  if (speechFallbackTimer) {
    clearTimeout(speechFallbackTimer);
    speechFallbackTimer = null;
  }
}

function stopComputerVoiceInternal(playStopTone) {
  speechSessionId++;
  clearSpeechTimers();
  window.activeSpeechUtterance = null;
  if ('speechSynthesis' in window) {
    try { window.speechSynthesis.cancel(); } catch(e){}
  }
  if (streamAudioEl) {
    try {
      streamAudioEl.pause();
      streamAudioEl.currentTime = 0;
      streamAudioEl.onended = null;
      streamAudioEl.onerror = null;
      streamAudioEl.removeAttribute('src');
      streamAudioEl.load();
    } catch(e){}
  }
  setVoiceVisualizer(false);
  if (playStopTone) {
    playLcarsSound('chirp');
  }
}

function playAudioStream(text, onDone, sessionId) {
  const sid = sessionId || speechSessionId;
  try {
    ensureAudio();
    if (!streamAudioEl) {
      streamAudioEl = new Audio();
    }
    streamAudioEl.pause();
    streamAudioEl.currentTime = 0;
    streamAudioEl.src = '/api/tts?text=' + encodeURIComponent(text);
    setVoiceVisualizer(true);
    streamAudioEl.onended = () => {
      if (sid !== speechSessionId) return;
      setVoiceVisualizer(false);
      if (onDone) onDone();
    };
    streamAudioEl.onerror = () => {
      if (sid !== speechSessionId) return;
      setVoiceVisualizer(false);
      if (onDone) onDone();
    };
    const p = streamAudioEl.play();
    if (p && p.catch) {
      p.catch(err => {
        if (sid !== speechSessionId) return;
        setVoiceVisualizer(false);
        console.warn('Audio stream error:', err);
      });
    }
  } catch(e) {
    setVoiceVisualizer(false);
  }
}

function computerSpeak(text, soundType) {
  playLcarsSound(soundType || 'chirp');
  const txtEl = document.getElementById('voice-transcript');
  if (txtEl) txtEl.textContent = text;
  stopComputerVoiceInternal(false);
  const sid = speechSessionId;

  if (selectedVoiceURI === 'server_stream') {
    playAudioStream(text, null, sid);
    return;
  }

  const hasSynth = ('speechSynthesis' in window) && (window.speechSynthesis !== null);
  if (!hasSynth) {
    playAudioStream(text, null, sid);
    return;
  }

  try {
    if (window.speechSynthesis.paused) window.speechSynthesis.resume();
  } catch(e) {}

  try {
    window.speechSynthesis.cancel();
  } catch(e) {}

  const utter = new SpeechSynthesisUtterance(text);
  window.activeSpeechUtterance = utter; // Prevent iOS WebKit GC bug
  utter.lang = 'en-US';

  // Pick voice: user selection or best natural English voice
  const voices = cachedVoices.length ? cachedVoices : (('speechSynthesis' in window) ? window.speechSynthesis.getVoices() : []);
  if (voices && voices.length > 0) {
    let chosen = null;
    if (selectedVoiceURI) {
      chosen = voices.find(v => (v.voiceURI && v.voiceURI === selectedVoiceURI) || v.name === selectedVoiceURI);
    }
    if (!chosen) {
      chosen = voices.find(v => /google.*english|natural|neural|premium|enhanced/i.test(v.name) && /en/i.test(v.lang))
        || voices.find(v => /samantha|serena|karen|daniel|victoria|alex|moira|zira/i.test(v.name) && /en/i.test(v.lang))
        || voices.find(v => /en/i.test(v.lang))
        || voices[0];
    }
    if (chosen) utter.voice = chosen;
  }
  // Keep pitch and rate natural (1.0) for clear speech without robotic distortion
  utter.rate = 1.0;
  utter.pitch = 1.0;

  let started = false;
  utter.onstart = () => {
    if (sid !== speechSessionId) return;
    started = true;
    setVoiceVisualizer(true);
  };
  utter.onend = () => {
    if (sid !== speechSessionId) return;
    setVoiceVisualizer(false);
    window.activeSpeechUtterance = null;
  };
  utter.onerror = (e) => {
    if (sid !== speechSessionId) return;
    setVoiceVisualizer(false);
    window.activeSpeechUtterance = null;
    console.warn('Speech synthesis failed, falling back to audio stream:', e);
    playAudioStream(text, null, sid);
  };

  setVoiceVisualizer(true);

  speechStartTimer = setTimeout(() => {
    speechStartTimer = null;
    if (sid !== speechSessionId) return;
    try {
      window.speechSynthesis.resume();
      window.speechSynthesis.speak(utter);
      // Fallback timer: if browser speech didn't start in 550ms (typical of iOS WKWebView in DuckDuckGo/Firefox), fallback to audio stream!
      speechFallbackTimer = setTimeout(() => {
        speechFallbackTimer = null;
        if (sid !== speechSessionId) return;
        if (!started && !window.speechSynthesis.speaking) {
          playAudioStream(text, null, sid);
        }
      }, 550);
    } catch(err) {
      if (sid !== speechSessionId) return;
      playAudioStream(text, null, sid);
    }
  }, 40);
}

function stopComputerVoice() {
  stopComputerVoiceInternal(true);
}

function toggleStarshipMode() {
  starshipMode = !starshipMode;
  const btn = document.getElementById('starship-toggle');
  if (btn) {
    btn.classList.toggle('on', starshipMode);
    btn.textContent = starshipMode ? '🖖 Starship Computer Mode: ON' : '🖖 Starship Computer Mode: OFF';
  }
  playLcarsSound('chirp');
  if (Globe.obj) {
    inspectAstro(Globe.kind, Globe.name);
  }
}

function getOrdinal(n) {
  const num = parseInt(n);
  if (isNaN(num)) return '';
  const s = ["th", "st", "nd", "rd"];
  const v = num % 100;
  return num + (s[(v - 20) % 10] || s[v] || s[0]);
}

function buildComputerBriefing(kind, obj) {
  if (kind === 'planet' && obj.name === 'Earth') {
    return "Earth is our home planet! It is covered in blue oceans, green continents, white clouds, and icy poles. It has one moon in space, fresh air to breathe, and millions of wonderful animals, plants, and people.";
  }
  if (kind === 'planet') {
    let s = obj.name + " is a " + obj.type + ", and the " + getOrdinal(obj.order_from_sun) + " planet from the Sun. ";
    if (obj.moons_count !== undefined) {
      const m = parseInt(obj.moons_count);
      s += (m === 0) ? "It has no moons. " : (m === 1 ? "It has one moon. " : "It has " + m + " moons. ");
    }
    if (obj.description) s += obj.description;
    return s;
  }
  if (kind === 'moon') {
    return obj.name + " is a moon orbiting " + obj.planet_name + ". " + (obj.description || '');
  }
  if (kind === 'exoplanet') {
    return obj.name + " is an exoplanet orbiting the star " + obj.host_star + " at a distance of " + obj.distance_light_years + " light years. " + (obj.description || '');
  }
  if (kind === 'messier') {
    return obj.messier_id + ", the " + obj.common_name + ", a " + obj.object_type + " in the constellation " + obj.constellation + ". " + (obj.description || '');
  }
  if (kind === 'constellation') {
    return "The constellation " + obj.name + ", meaning " + (obj.meaning || obj.name) + ". Its brightest star is " + obj.brightest_star + ". " + (obj.mythology || '');
  }
  return obj.name || 'Celestial object';
}

function triggerCadetAction(action) {
  if (action === 'briefing') {
    computerSpeak(buildComputerBriefing(Globe.kind, Globe.obj), 'chirp');
  } else if (action === 'life') {
    Globe.scanWave = 1.0;
    const d = EarthDiscoveries[discoveryIdx % EarthDiscoveries.length];
    discoveryIdx++;
    Globe.popups.push({ text: d.icon + ' ' + d.title, x: 210 + (Math.random() - 0.5) * 110, y: 185 + (Math.random() - 0.5) * 90, alpha: 1.5 });
    if (Globe.kind === 'planet' && Globe.name === 'Earth') {
      computerSpeak(d.speech, 'scan');
    } else {
      computerSpeak("Biosensor sweep complete on " + Globe.name + ". No Earth animals detected here, Cadet! Only space rocks and cosmic dust!", 'scan');
    }
  } else if (action === 'weather') {
    Globe.clouds = !Globe.clouds;
    Globe.lightning = 1.0;
    if (Globe.clouds) {
      computerSpeak("Atmospheric weather systems active! Look at the fluffy white clouds swirling around the planet, bringing rain for the trees and flowers.", 'weather');
    } else {
      computerSpeak("Weather clouds cleared! We now have a crystal clear view of the surface from orbit.", 'weather');
    }
  } else if (action === 'shields') {
    Globe.shields = !Globe.shields;
    if (Globe.shields) {
      computerSpeak("Planetary magnetic shield raised! Earth has an invisible forcefield that protects us from solar wind and makes the glowing green Northern Lights!", 'shield');
    } else {
      computerSpeak("Magnetic aurora visualizer deactivated. Standard orbital view restored.", 'shield');
    }
  } else if (action === 'moon') {
    Globe.moonBeam = 1.0;
    if (Globe.kind === 'planet' && Globe.name === 'Earth') {
      computerSpeak("Pinging Earth's Moon! Beep boop! The Moon travels all the way around Earth and pulls on the oceans to make waves at the beach!", 'moon');
    } else {
      const m = Globe.obj.moons_count || 0;
      computerSpeak(Globe.name + " has " + m + (m == 1 ? " moon" : " moons") + " orbiting around it in space!", 'moon');
    }
  } else if (action === 'warp') {
    Globe.warp = !Globe.warp;
    Globe.spinSpeed = Globe.warp ? 0.055 : 0.007;
    if (Globe.warp) {
      computerSpeak("Engage Warp Speed! Spinning the planet super fast! Hang on, Cadets!", 'warp');
    } else {
      computerSpeak("Dropping out of warp. Returning to normal orbital cruise speed.", 'chirp');
    }
  }
}

function initPlanetCanvas() {
  const canvas = document.getElementById('planet-canvas');
  if (!canvas) return;
  if (animFrameId) cancelAnimationFrame(animFrameId);

  const getPos = (e) => {
    const rect = canvas.getBoundingClientRect();
    const clientX = e.touches ? e.touches[0].clientX : e.clientX;
    const clientY = e.touches ? e.touches[0].clientY : e.clientY;
    return {
      x: (clientX - rect.left) * (canvas.width / rect.width),
      y: (clientY - rect.top) * (canvas.height / rect.height)
    };
  };

  let downPos = null;
  const onDown = (e) => {
    Globe.dragging = true;
    const p = getPos(e);
    Globe.lastX = p.x;
    Globe.lastY = p.y;
    downPos = p;
  };
  const onMove = (e) => {
    if (!Globe.dragging) return;
    const p = getPos(e);
    const dx = p.x - Globe.lastX;
    const dy = p.y - Globe.lastY;
    Globe.angle += dx * 0.015;
    Globe.tilt = Math.max(-0.8, Math.min(0.8, Globe.tilt + dy * 0.01));
    Globe.lastX = p.x;
    Globe.lastY = p.y;
  };
  const onUp = (e) => {
    if (Globe.dragging && downPos) {
      const dx = Math.abs(Globe.lastX - downPos.x);
      const dy = Math.abs(Globe.lastY - downPos.y);
      if (dx < 8 && dy < 8) {
        if (Globe.name === 'Earth') {
          // Kid tapped Earth! Scan for animals
          Globe.scanWave = 1.0;
          const d = EarthDiscoveries[discoveryIdx % EarthDiscoveries.length];
          discoveryIdx++;
          Globe.popups.push({
            text: d.icon + ' ' + d.title,
            x: downPos.x,
            y: downPos.y,
            alpha: 1.5
          });
          computerSpeak(d.speech, 'scan');
        } else {
          // For other planets: tap reads info
          computerSpeak(buildComputerBriefing(Globe.kind, Globe.obj), 'chirp');
        }
      }
    }
    Globe.dragging = false;
  };

  canvas.onmousedown = onDown;
  window.onmousemove = onMove;
  window.onmouseup = onUp;
  canvas.ontouchstart = (e) => { e.preventDefault(); onDown(e); };
  canvas.ontouchmove = (e) => { e.preventDefault(); onMove(e); };
  canvas.ontouchend = onUp;

  function renderFrame() {
    if (!document.getElementById('planet-canvas')) return;
    const ctx = canvas.getContext('2d');
    const W = canvas.width;
    const H = canvas.height;
    const cx = W / 2;
    const cy = H / 2;
    const R = 112;

    ctx.clearRect(0, 0, W, H);

    // 1. Starfield & Warp Streaks
    Globe.stars.forEach(st => {
      st.p += 0.04;
      if (Globe.warp) {
        st.x -= 0.025;
        if (st.x < 0) st.x = 1;
        ctx.strokeStyle = '#99ddff';
        ctx.lineWidth = st.s;
        ctx.beginPath();
        ctx.moveTo(st.x * W, st.y * H);
        ctx.lineTo(st.x * W + 28, st.y * H);
        ctx.stroke();
      } else {
        const alpha = 0.4 + 0.6 * Math.abs(Math.sin(st.p));
        ctx.fillStyle = 'rgba(220,235,255,' + alpha.toFixed(2) + ')';
        ctx.beginPath();
        ctx.arc(st.x * W, st.y * H, st.s, 0, Math.PI * 2);
        ctx.fill();
      }
    });

    if (!Globe.dragging) {
      Globe.angle += Globe.spinSpeed;
    }
    Globe.moonAngle += Globe.warp ? 0.06 : 0.015;

    const moonX = cx + Math.cos(Globe.moonAngle) * 168;
    const moonY = cy + Math.sin(Globe.moonAngle) * 42;
    const moonBehind = Math.sin(Globe.moonAngle) < 0;

    const drawMoon = () => {
      if (Globe.kind !== 'planet' || (Globe.obj && parseInt(Globe.obj.moons_count || '0') === 0)) return;
      if (Globe.moonBeam > 0.01) {
        ctx.strokeStyle = 'rgba(255, 204, 102, ' + Globe.moonBeam.toFixed(2) + ')';
        ctx.lineWidth = 4;
        ctx.setLineDash([6, 4]);
        ctx.beginPath();
        ctx.moveTo(cx, cy);
        ctx.lineTo(moonX, moonY);
        ctx.stroke();
        ctx.setLineDash([]);
      }
      ctx.save();
      ctx.beginPath();
      ctx.arc(moonX, moonY, 16, 0, Math.PI * 2);
      const mg = ctx.createRadialGradient(moonX - 5, moonY - 5, 2, moonX, moonY, 16);
      mg.addColorStop(0, '#f0f0f5');
      mg.addColorStop(1, '#707080');
      ctx.fillStyle = mg;
      ctx.fill();
      if (Globe.moonBeam > 0.05) {
        ctx.strokeStyle = '#ffcc66';
        ctx.lineWidth = 3;
        ctx.stroke();
      }
      // Moon craters
      ctx.fillStyle = 'rgba(80,80,95,0.45)';
      ctx.beginPath(); ctx.arc(moonX - 4, moonY - 2, 3.5, 0, Math.PI * 2); ctx.fill();
      ctx.beginPath(); ctx.arc(moonX + 5, moonY + 4, 2.5, 0, Math.PI * 2); ctx.fill();
      ctx.restore();
    };

    if (moonBehind) drawMoon();

    // 2. Planetary Shields / Aurora Halo
    if (Globe.shields) {
      const sg = ctx.createRadialGradient(cx, cy, R - 4, cx, cy, R + 32);
      sg.addColorStop(0, 'rgba(102, 255, 178, 0.0)');
      sg.addColorStop(0.5, 'rgba(102, 255, 178, 0.38)');
      sg.addColorStop(0.8, 'rgba(180, 120, 255, 0.28)');
      sg.addColorStop(1, 'rgba(102, 204, 255, 0.0)');
      ctx.fillStyle = sg;
      ctx.beginPath();
      ctx.arc(cx, cy, R + 32, 0, Math.PI * 2);
      ctx.fill();
    } else {
      // Subtle atmospheric glow
      const ag = ctx.createRadialGradient(cx, cy, R - 2, cx, cy, R + 14);
      const glowCol = (Globe.name === 'Earth') ? 'rgba(100, 180, 255, 0.35)' : 'rgba(200, 200, 220, 0.18)';
      ag.addColorStop(0, glowCol);
      ag.addColorStop(1, 'rgba(0,0,0,0)');
      ctx.fillStyle = ag;
      ctx.beginPath();
      ctx.arc(cx, cy, R + 14, 0, Math.PI * 2);
      ctx.fill();
    }

    // 3. Planet Sphere Base (Clipped)
    ctx.save();
    ctx.beginPath();
    ctx.arc(cx, cy, R, 0, Math.PI * 2);
    ctx.clip();

    if (Globe.name === 'Earth') {
      // Deep ocean gradient
      const og = ctx.createRadialGradient(cx - 35, cy - 35, 15, cx, cy, R);
      og.addColorStop(0, '#2b7de9');
      og.addColorStop(0.65, '#13499e');
      og.addColorStop(1, '#071d49');
      ctx.fillStyle = og;
      ctx.fillRect(cx - R, cy - R, R * 2, R * 2);

      // Project 3D spherical continents
      const projectPt = (lon, lat) => {
        const l = lon + Globe.angle;
        const cosLat = Math.cos(lat);
        const sinLat = Math.sin(lat);
        const cosLon = Math.cos(l);
        const sinLon = Math.sin(l);
        const x = cosLat * sinLon;
        const y = sinLat * Math.cos(Globe.tilt) - cosLat * cosLon * Math.sin(Globe.tilt);
        const z = sinLat * Math.sin(Globe.tilt) + cosLat * cosLon * Math.cos(Globe.tilt);
        return { x: cx + x * R, y: cy - y * R, z: z };
      };

      EarthContinents.forEach(cont => {
        const projected = cont.pts.map(pt => projectPt(pt[0], pt[1]));
        const visible = projected.some(p => p.z > -0.15);
        if (!visible) return;
        ctx.fillStyle = cont.color;
        ctx.beginPath();
        projected.forEach((p, idx) => {
          if (idx === 0) ctx.moveTo(p.x, p.y);
          else ctx.lineTo(p.x, p.y);
        });
        ctx.closePath();
        ctx.fill();
      });

      // Polar Ice Caps
      ctx.fillStyle = 'rgba(240, 248, 255, 0.85)';
      ctx.beginPath();
      ctx.ellipse(cx, cy - R * 0.88 + Globe.tilt * 25, R * 0.55, 18, 0, 0, Math.PI * 2);
      ctx.fill();
      ctx.beginPath();
      ctx.ellipse(cx, cy + R * 0.90 + Globe.tilt * 25, R * 0.50, 16, 0, 0, Math.PI * 2);
      ctx.fill();

      // Swirling Weather Clouds
      if (Globe.clouds) {
        ctx.fillStyle = 'rgba(255, 255, 255, 0.38)';
        for (let i = 0; i < 7; i++) {
          const cw = ((Globe.angle * 1.35 + i * 1.1) % (Math.PI * 2));
          const sx = cx + Math.sin(cw) * (R * 0.72);
          const sy = cy + (i - 3) * 28 + Math.cos(cw * 2) * 6;
          if (Math.cos(cw) > -0.2) {
            ctx.beginPath();
            ctx.ellipse(sx, sy, 38, 10, 0.15, 0, Math.PI * 2);
            ctx.fill();
          }
        }
      }

      // Lightning flash if weather triggered
      if (Globe.lightning > 0.05) {
        ctx.fillStyle = 'rgba(255, 255, 210, ' + (Globe.lightning * 0.55).toFixed(2) + ')';
        ctx.fillRect(cx - R, cy - R, R * 2, R * 2);
        Globe.lightning *= 0.86;
      }
    } else {
      // Other Planets / Bodies
      const baseColor = (Globe.obj && Globe.obj.color_hex) ? Globe.obj.color_hex : '#78b5ba';
      const pg = ctx.createRadialGradient(cx - 35, cy - 35, 12, cx, cy, R);
      pg.addColorStop(0, baseColor);
      pg.addColorStop(1, '#12131c');
      ctx.fillStyle = pg;
      ctx.fillRect(cx - R, cy - R, R * 2, R * 2);

      // Banding for gas giants or craters for rocky worlds
      ctx.fillStyle = 'rgba(255,255,255,0.09)';
      for (let b = -3; b <= 3; b++) {
        ctx.fillRect(cx - R, cy + b * 26 + Math.sin(Globe.angle + b) * 4, R * 2, 10);
      }
    }

    // 3D Day/Night Terminator Shadow
    const sh = ctx.createLinearGradient(cx - R, cy - R, cx + R, cy + R);
    sh.addColorStop(0, 'rgba(255,255,255,0.12)');
    sh.addColorStop(0.55, 'rgba(0,0,0,0.0)');
    sh.addColorStop(1, 'rgba(0,0,12,0.68)');
    ctx.fillStyle = sh;
    ctx.fillRect(cx - R, cy - R, R * 2, R * 2);

    // Sensor Scan Wave
    if (Globe.scanWave > 0.02) {
      const scanY = cy - R + (1.0 - Globe.scanWave) * (R * 2);
      ctx.strokeStyle = 'rgba(102, 255, 178, ' + Globe.scanWave.toFixed(2) + ')';
      ctx.lineWidth = 4;
      ctx.beginPath();
      ctx.moveTo(cx - R, scanY);
      ctx.lineTo(cx + R, scanY);
      ctx.stroke();
      Globe.scanWave *= 0.95;
    }

    ctx.restore();

    // Saturn / Giant Rings if applicable
    if (Globe.obj && (Globe.obj.has_rings === true || Globe.obj.has_rings === 'true' || Globe.obj.has_rings === 't')) {
      ctx.save();
      ctx.strokeStyle = 'rgba(234, 214, 184, 0.65)';
      ctx.lineWidth = 12;
      ctx.beginPath();
      ctx.ellipse(cx, cy, R * 1.55, R * 0.35, 0.2, 0, Math.PI * 2);
      ctx.stroke();
      ctx.restore();
    }

    if (!moonBehind) drawMoon();
    if (Globe.moonBeam > 0.01) Globe.moonBeam *= 0.94;

    // 4. Floating Kid Discovery Popups
    for (let i = Globe.popups.length - 1; i >= 0; i--) {
      const p = Globe.popups[i];
      p.y -= 0.7;
      p.alpha -= 0.015;
      if (p.alpha <= 0) {
        Globe.popups.splice(i, 1);
        continue;
      }
      ctx.save();
      ctx.font = 'bold 15px ui-monospace, monospace';
      const a = Math.min(1, p.alpha);
      ctx.fillStyle = 'rgba(20, 24, 38, ' + (a * 0.9).toFixed(2) + ')';
      const tw = ctx.measureText(p.text).width + 18;
      ctx.fillRect(p.x - tw / 2, p.y - 16, tw, 28);
      ctx.strokeStyle = 'rgba(255, 153, 0, ' + a.toFixed(2) + ')';
      ctx.lineWidth = 2;
      ctx.strokeRect(p.x - tw / 2, p.y - 16, tw, 28);
      ctx.fillStyle = 'rgba(255, 235, 180, ' + a.toFixed(2) + ')';
      ctx.textAlign = 'center';
      ctx.fillText(p.text, p.x, p.y + 3);
      ctx.restore();
    }

    animFrameId = requestAnimationFrame(renderFrame);
  }

  renderFrame();
}

function filterAstro(cat, btn) {
  document.querySelectorAll('.tabbar button').forEach(b => b.classList.remove('active'));
  if (btn) btn.classList.add('active');
  const cards = document.querySelectorAll('.astro-card');
  cards.forEach(c => {
    const cats = (c.getAttribute('data-cat') || '').split(' ');
    if (cat === 'all') {
      c.style.display = (cats.includes('terrestrial') || cats.includes('giant') || cats.includes('dwarf')) ? 'block' : 'none';
    } else if (cats.includes(cat)) {
      c.style.display = 'block';
    } else {
      c.style.display = 'none';
    }
  });
}
filterAstro('all', document.querySelector('.tabbar button'));

function inspectAstro(kind, key) {
  const map = AstroData[kind] || {};
  const obj = map[key];
  if (!obj) return;

  Globe.kind = kind;
  Globe.name = key;
  Globe.obj = obj;
  Globe.warp = false;
  Globe.spinSpeed = 0.007;
  Globe.popups = [];

  let title = key;
  if (kind === 'planet') {
    const icons = {'Mercury':'☿','Venus':'♀','Earth':'🌍','Mars':'♂','Ceres':'⚪','Jupiter':'♃','Saturn':'🪐','Uranus':'♅','Neptune':'♆','Pluto':'♇','Haumea':'🥚','Makemake':'❄️','Eris':'🪨'};
    const icon = icons[obj.name] || '●';
    title = icon + ' ' + obj.name + ' (' + obj.type + ')';
  } else if (kind === 'moon') {
    title = '🌕 ' + obj.name + ' (Moon of ' + obj.planet_name + ')';
  } else if (kind === 'exoplanet') {
    title = '🔭 ' + obj.name + ' (' + obj.planet_type + ' in ' + obj.constellation + ')';
  } else if (kind === 'messier') {
    title = '🌌 ' + obj.messier_id + ' — ' + obj.common_name + ' (' + obj.object_type + ')';
  } else if (kind === 'constellation') {
    title = '✨ ' + obj.name + ' (' + obj.abbreviation + ' · ' + obj.genitive + ')';
  }

  let h = '';
  if (starshipMode) {
    const briefing = buildComputerBriefing(kind, obj);
    let barsHtml = '';
    for (let i = 0; i < 18; i++) barsHtml += '<div class="voice-bar"></div>';

    const canvasHint = (key === 'Earth')
      ? '👆 Drag to spin Earth · Tap to scan for animals!'
      : '👆 Drag to spin ' + key + ' · Tap to read info';

    let controlsHtml = '';
    if (key === 'Earth') {
      controlsHtml = '<div class="cadet-pad">' +
        '<button type="button" class="lcars-btn lcars-amber" onclick="triggerCadetAction(\'briefing\')"><span class="ico">🔊</span><span>Read Info</span></button>' +
        '<button type="button" class="lcars-btn lcars-mint" onclick="triggerCadetAction(\'life\')"><span class="ico">🐾</span><span>Scan for Animals!</span></button>' +
        '<button type="button" class="lcars-btn lcars-cyan" onclick="triggerCadetAction(\'weather\')"><span class="ico">🌦️</span><span>Weather & Clouds</span></button>' +
        '<button type="button" class="lcars-btn lcars-purple" onclick="triggerCadetAction(\'shields\')"><span class="ico">🛡️</span><span>Aurora Forcefield</span></button>' +
        '<button type="button" class="lcars-btn lcars-gold" onclick="triggerCadetAction(\'moon\')"><span class="ico">🌙</span><span>Ping The Moon</span></button>' +
        '<button type="button" class="lcars-btn lcars-rose" onclick="triggerCadetAction(\'warp\')"><span class="ico">🚀</span><span>Warp Speed Spin!</span></button>' +
        '<button type="button" class="lcars-btn lcars-red" onclick="stopComputerVoice()"><span class="ico">⏹️</span><span>Stop Voice</span></button>' +
      '</div>' +
      '<div class="muted" style="font-size:11px">💡 Tip: Tap Earth to find hidden animals, or click any SQL table row below to hear it read aloud!</div>';
    } else {
      controlsHtml = '<div style="display:flex;gap:10px;align-items:center;margin-bottom:14px;flex-wrap:wrap">' +
        '<button type="button" class="lcars-btn lcars-amber" style="padding:10px 20px;font-size:14px" onclick="triggerCadetAction(\'briefing\')"><span class="ico">🔊</span><span>Read Info</span></button>' +
        '<button type="button" class="lcars-btn lcars-red" style="padding:10px 16px" onclick="stopComputerVoice()"><span class="ico">⏹️</span><span>Stop Voice</span></button>' +
      '</div>' +
      '<div class="muted" style="font-size:11px">💡 Click <b>Read Info</b> to hear the facts aloud, or click any row in the SQL table below!</div>';
    }

    h += '<div class="lcars-frame">' +
      '<div class="lcars-header">' +
        '<h3 class="lcars-title">🖖 USS ENTERPRISE // MAIN BRIDGE VIEWSCREEN — ' + key.toUpperCase() + '</h3>' +
        '<span class="lcars-code">LCARS 47110 · SOL SECTOR</span>' +
      '</div>' +
      '<div class="lcars-layout">' +
        '<div class="viewscreen-wrap">' +
          '<canvas id="planet-canvas" width="420" height="380"></canvas>' +
          '<div class="viewscreen-hint">' + canvasHint + '</div>' +
        '</div>' +
        '<div>' +
          '<div class="voice-box">' +
            '<div class="voice-status">' +
              '<span class="voice-label" id="voice-state">USS ENTERPRISE COMPUTER // STANDING BY</span>' +
              '<div class="voice-bars" id="voice-bars">' + barsHtml + '</div>' +
            '</div>' +
            '<div class="voice-text" id="voice-transcript">' + briefing + '</div>' +
            '<div class="voice-controls">' +
              '<div style="display:flex;gap:8px;align-items:center;flex-wrap:wrap">' +
                '<span style="font-size:11px;color:#ff9900;font-weight:bold;letter-spacing:1px">🎙️ VOICE:</span>' +
                '<select id="voice-select" class="lcars-select" onchange="onVoiceChanged(this.value)"><option value="">Loading voices…</option></select>' +
                '<button type="button" class="btn" style="padding:3px 10px;font-size:11px;border-color:#ff9900;color:#ff9900" onclick="testSelectedVoice()">Test 🔊</button>' +
              '</div>' +
              '<div style="font-size:11px;color:#d4b47b;margin-top:6px;line-height:1.4">📱 <b>Phone sound note:</b> If quiet, make sure your iPhone silent/mute switch on the side of the phone is OFF!</div>' +
            '</div>' +
          '</div>' +
          controlsHtml +
        '</div>' +
      '</div>' +
    '</div>';
  }

  h += '<h2>' + title + ' — Database Telemetry</h2><table>';
  for (const k in obj) {
    if (obj[k] !== null && obj[k] !== '') {
      const safeKey = k.replace(/_/g, ' ');
      const valStr = String(obj[k]).replace(/'/g, "\\'");
      h += '<tr class="telemetry-row" onclick="computerSpeak(\'' + safeKey + ' is ' + valStr + '\', \'chirp\')" title="Click to hear the Computer read this row">' +
           '<th style="width:240px">🔊 ' + k + '</th><td>' + obj[k] + '</td></tr>';
    }
  }
  h += '</table>';
  document.getElementById('astro-detail').innerHTML = h;
  document.getElementById('astro-detail').scrollIntoView({ behavior: 'smooth', block: 'nearest' });

  if (starshipMode) {
    initPlanetCanvas();
    populateVoiceDropdown();
  }
}
</script>
{{template "bottom" .}}{{end}}

{{define "math"}}{{template "top" .}}
<style>
.tabbar{display:flex;gap:8px;flex-wrap:wrap;margin-bottom:14px}
.math-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(220px,1fr));gap:10px}
.math-card{background:#181822;border:1px solid var(--border);border-radius:6px;padding:10px;cursor:pointer;transition:border-color .15s}
.math-card:hover{border-color:var(--cyan);outline:1px solid var(--cyan)}
.math-head{display:flex;align-items:center;gap:10px;margin-bottom:6px}
.math-sym{min-width:32px;height:32px;border-radius:6px;background:#252535;color:var(--gold);display:flex;align-items:center;justify-content:center;font-size:16px;font-weight:bold;flex-shrink:0;border:1px solid var(--border)}
.math-title{font-size:13px;font-weight:bold;color:var(--text)}
.math-sub{font-size:11px;color:var(--muted)}
.math-badge{display:inline-block;padding:1px 6px;border-radius:3px;font-size:10px;margin-top:4px}
.badge-const{background:#263c45;color:var(--cyan)}
.badge-poly{background:#3a3223;color:var(--gold)}
.badge-seq{background:#2b4436;color:var(--mint)}
.badge-thm{background:#3d2b45;color:#c4a7e7}
.badge-crv{background:#452828;color:var(--rose)}
</style>
<div class="card">
<h2>📐 Mathematics & Geometry Explorer <span class="muted">(click any item to inspect formulas & properties)</span></h2>
<div class="tabbar">
  <button type="button" class="btn active" onclick="filterMath('all', this)">All Topics</button>
  <button type="button" class="btn" onclick="filterMath('const', this)">π Constants</button>
  <button type="button" class="btn" onclick="filterMath('poly', this)">📦 Polyhedra (3D)</button>
  <button type="button" class="btn" onclick="filterMath('seq', this)">🔢 Sequences & Classes</button>
  <button type="button" class="btn" onclick="filterMath('thm', this)">📜 Landmark Theorems</button>
  <button type="button" class="btn" onclick="filterMath('crv', this)">🌀 Geometric Curves</button>
</div>
<div class="math-grid" id="math-list">
{{range .Constants}}
<div class="math-card" data-cat="const" data-kind="const" data-key="{{.Symbol}}" onclick="inspectMath(this.dataset.kind, this.dataset.key)">
  <div class="math-head">
    <div class="math-sym">{{.Symbol}}</div>
    <div>
      <div class="math-title">{{.Name}}</div>
      <div class="math-sub">{{.ValApprox}}</div>
    </div>
  </div>
  <span class="math-badge badge-const">{{.Domain}}</span>
</div>
{{end}}
{{range .Polyhedra}}
<div class="math-card" data-cat="poly" data-kind="poly" data-key="{{.Name}}" onclick="inspectMath(this.dataset.kind, this.dataset.key)">
  <div class="math-head">
    <div class="math-sym" style="color:var(--gold)">{{.Schlafli}}</div>
    <div>
      <div class="math-title">{{.Name}}</div>
      <div class="math-sub">V:{{.V}} · E:{{.E}} · F:{{.F}}</div>
    </div>
  </div>
  <span class="math-badge badge-poly">{{.Family}}</span>
</div>
{{end}}
{{range .NumberClasses}}
<div class="math-card" data-cat="seq" data-kind="seq" data-key="{{.Name}}" onclick="inspectMath(this.dataset.kind, this.dataset.key)">
  <div class="math-head">
    <div class="math-sym" style="color:var(--mint)">{{.Sym}}</div>
    <div>
      <div class="math-title">{{.Name}}</div>
      <div class="math-sub">{{.Oeis}}</div>
    </div>
  </div>
  <span class="math-badge badge-seq">{{.FirstPreview}}</span>
</div>
{{end}}
{{range .Theorems}}
<div class="math-card" data-cat="thm" data-kind="thm" data-key="{{.Name}}" onclick="inspectMath(this.dataset.kind, this.dataset.key)">
  <div class="math-head">
    <div class="math-sym" style="color:#c4a7e7">QED</div>
    <div>
      <div class="math-title">{{.Name}}</div>
      <div class="math-sub">{{.Prover}} ({{.Year}})</div>
    </div>
  </div>
  <span class="math-badge badge-thm">{{.Field}}</span>
</div>
{{end}}
{{range .Curves}}
<div class="math-card" data-cat="crv" data-kind="crv" data-key="{{.Name}}" onclick="inspectMath(this.dataset.kind, this.dataset.key)">
  <div class="math-head">
    <div class="math-sym" style="color:var(--rose)">ƒ(x)</div>
    <div>
      <div class="math-title">{{.Name}}</div>
      <div class="math-sub">{{.Discoverer}}</div>
    </div>
  </div>
  <span class="math-badge badge-crv">{{.Type}}</span>
</div>
{{end}}
</div>
</div>

<div class="card" id="math-detail">
<span class="muted">Select a constant, polyhedron, sequence, theorem, or curve above to inspect its properties…</span>
</div>

<script>
const MathData = {{.JSON}};
function filterMath(cat, btn) {
  document.querySelectorAll('.tabbar button').forEach(b => b.classList.remove('active'));
  if (btn) btn.classList.add('active');
  const cards = document.querySelectorAll('.math-card');
  cards.forEach(c => {
    const ccat = c.getAttribute('data-cat') || '';
    if (cat === 'all' || ccat === cat) {
      c.style.display = 'block';
    } else {
      c.style.display = 'none';
    }
  });
}

function inspectMath(kind, key) {
  const map = MathData[kind] || {};
  const obj = map[key];
  if (!obj) return;
  let title = key;
  if (kind === 'const') {
    title = obj.symbol + ' — ' + obj.name + ' (' + obj.domain + ')';
  } else if (kind === 'poly') {
    title = '📦 ' + obj.name + ' (' + obj.family + ')';
  } else if (kind === 'seq') {
    title = '🔢 ' + obj.name + ' (' + (obj.symbol || '') + ')';
  } else if (kind === 'thm') {
    title = '📜 ' + obj.name + ' (' + obj.field + ')';
  } else if (kind === 'crv') {
    title = '🌀 ' + obj.name + ' (' + obj.curve_type + ')';
  }

  const detail = document.getElementById('math-detail');
  if (!detail) return;
  detail.innerHTML = '';
  const h2 = document.createElement('h2');
  h2.textContent = title;
  detail.appendChild(h2);
  const table = document.createElement('table');
  for (const k in obj) {
    if (obj[k] !== null && obj[k] !== '') {
      const tr = document.createElement('tr');
      const th = document.createElement('th');
      th.style.width = '240px';
      th.textContent = k;
      const td = document.createElement('td');
      td.textContent = String(obj[k]);
      tr.appendChild(th);
      tr.appendChild(td);
      table.appendChild(tr);
    }
  }
  detail.appendChild(table);
  detail.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
}
</script>
{{template "bottom" .}}{{end}}

{{define "anatomy"}}{{template "top" .}}
<style>
.tabbar{display:flex;gap:8px;flex-wrap:wrap;margin-bottom:14px}
.anat-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(230px,1fr));gap:10px}
.anat-card{background:#181822;border:1px solid var(--border);border-radius:6px;padding:10px;cursor:pointer;transition:border-color .15s}
.anat-card:hover{border-color:var(--cyan);outline:1px solid var(--cyan)}
.anat-head{display:flex;align-items:center;gap:10px;margin-bottom:6px}
.anat-icon{min-width:32px;height:32px;border-radius:6px;background:#252535;color:var(--text);display:flex;align-items:center;justify-content:center;font-size:16px;flex-shrink:0;border:1px solid var(--border)}
.anat-title{font-size:13px;font-weight:bold;color:var(--text)}
.anat-sub{font-size:11px;color:var(--muted)}
.anat-badge{display:inline-block;padding:1px 6px;border-radius:3px;font-size:10px;margin-top:4px}
.badge-axial{background:#3a2835;color:#e8a2af}
.badge-append{background:#263c45;color:var(--cyan)}
.badge-joint{background:#3a3223;color:var(--gold)}
.badge-muscle{background:#452828;color:var(--rose)}
.badge-organ{background:#2b4436;color:var(--mint)}
</style>
<div class="card">
<h2>🦴 Human Anatomy & Osteology Explorer <span class="muted">(click any structure to inspect anatomical relationships & functions)</span></h2>
<div class="tabbar">
  <button type="button" class="btn active" onclick="filterAnat('all', this)">All Structures</button>
  <button type="button" class="btn" onclick="filterAnat('axial', this)">💀 Axial Skeleton</button>
  <button type="button" class="btn" onclick="filterAnat('append', this)">🦴 Appendicular Limbs</button>
  <button type="button" class="btn" onclick="filterAnat('joint', this)">🔗 Joints & Articulations</button>
  <button type="button" class="btn" onclick="filterAnat('muscle', this)">💪 Prime Movers (Muscles)</button>
  <button type="button" class="btn" onclick="filterAnat('organ', this)">🫀 Organ Systems</button>
</div>
<div class="anat-grid" id="anat-list">
{{range .Bones}}
<div class="anat-card" data-cat="{{.Cat}}" data-kind="bone" data-key="{{.Name}}" onclick="inspectAnat(this.dataset.kind, this.dataset.key)">
  <div class="anat-head">
    <div class="anat-icon">{{if eq .Cat "axial"}}💀{{else}}🦴{{end}}</div>
    <div>
      <div class="anat-title">{{.Name}}</div>
      <div class="anat-sub">{{.LatinName}} · ({{.Count}}×)</div>
    </div>
  </div>
  <span class="anat-badge {{.BadgeClass}}">{{.Region}} · {{.Shape}}</span>
</div>
{{end}}
{{range .Joints}}
<div class="anat-card" data-cat="joint" data-kind="joint" data-key="{{.Name}}" onclick="inspectAnat(this.dataset.kind, this.dataset.key)">
  <div class="anat-head">
    <div class="anat-icon" style="color:var(--gold)">🔗</div>
    <div>
      <div class="anat-title">{{.Name}}</div>
      <div class="anat-sub">{{.TypeShort}}</div>
    </div>
  </div>
  <span class="anat-badge badge-joint">{{.ClassShort}}</span>
</div>
{{end}}
{{range .Muscles}}
<div class="anat-card" data-cat="muscle" data-kind="muscle" data-key="{{.Name}}" onclick="inspectAnat(this.dataset.kind, this.dataset.key)">
  <div class="anat-head">
    <div class="anat-icon" style="color:var(--rose)">💪</div>
    <div>
      <div class="anat-title">{{.Name}}</div>
      <div class="anat-sub">{{.LatinName}}</div>
    </div>
  </div>
  <span class="anat-badge badge-muscle">{{.Region}}</span>
</div>
{{end}}
{{range .Organs}}
<div class="anat-card" data-cat="organ" data-kind="organ" data-key="{{.Name}}" onclick="inspectAnat(this.dataset.kind, this.dataset.key)">
  <div class="anat-head">
    <div class="anat-icon" style="color:var(--mint)">🫀</div>
    <div>
      <div class="anat-title">{{.Name}}</div>
      <div class="anat-sub">{{.OrgansShort}}</div>
    </div>
  </div>
  <span class="anat-badge badge-organ">Organ System</span>
</div>
{{end}}
</div>
</div>

<div class="card" id="anat-detail">
<span class="muted">Select any bone, joint, muscle, or organ system above to inspect details…</span>
</div>

<script>
const AnatData = {{.JSON}};
function filterAnat(cat, btn) {
  document.querySelectorAll('.tabbar button').forEach(b => b.classList.remove('active'));
  if (btn) btn.classList.add('active');
  const cards = document.querySelectorAll('.anat-card');
  cards.forEach(c => {
    const ccat = c.getAttribute('data-cat') || '';
    if (cat === 'all' || ccat === cat) {
      c.style.display = 'block';
    } else {
      c.style.display = 'none';
    }
  });
}

function inspectAnat(kind, key) {
  const map = AnatData[kind] || {};
  const obj = map[key];
  if (!obj) return;
  let title = key;
  if (kind === 'bone') {
    title = '🦴 ' + obj.name + ' (' + obj.latin_name + ' · ' + obj.division + ')';
  } else if (kind === 'joint') {
    title = '🔗 ' + obj.name + ' (' + obj.joint_type + ')';
  } else if (kind === 'muscle') {
    title = '💪 ' + obj.name + ' (' + obj.latin_name + ')';
  } else if (kind === 'organ') {
    title = '🫀 ' + obj.name;
  }

  const detail = document.getElementById('anat-detail');
  if (!detail) return;
  detail.innerHTML = '';
  const h2 = document.createElement('h2');
  h2.textContent = title;
  detail.appendChild(h2);
  const table = document.createElement('table');
  for (const k in obj) {
    if (obj[k] !== null && obj[k] !== '') {
      const tr = document.createElement('tr');
      const th = document.createElement('th');
      th.style.width = '240px';
      th.textContent = k;
      const td = document.createElement('td');
      td.textContent = String(obj[k]);
      tr.appendChild(th);
      tr.appendChild(td);
      table.appendChild(tr);
    }
  }
  detail.appendChild(table);
  detail.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
}
</script>
{{template "bottom" .}}{{end}}
`))

func main() {
	http.HandleFunc("/api/tts", func(w http.ResponseWriter, r *http.Request) {
		text := sanitizeSpeechText(r.URL.Query().Get("text"))
		if text == "" {
			http.Error(w, "missing text", 400)
			return
		}
		espeakPath, err := exec.LookPath("espeak")
		if err != nil {
			espeakPath, err = exec.LookPath("espeak-ng")
		}
		if err != nil {
			http.Error(w, "tts not available", 501)
			return
		}
		cmd := exec.CommandContext(r.Context(), espeakPath, "--stdout", "-s", "145", text)
		w.Header().Set("Content-Type", "audio/wav")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		cmd.Stdout = w
		_ = cmd.Run()
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		type row struct{ Name, Blurb, Size, Tables string }
		var out []row
		for _, d := range databases {
			rw := row{Name: d.Name, Blurb: d.Blurb, Size: "?", Tables: "?"}
			if p, err := pool(d.Name); err == nil {
				res := runQuery(p, `SELECT pg_size_pretty(pg_database_size(current_database())),
				  (SELECT count(*) FROM information_schema.tables WHERE table_schema NOT IN ('pg_catalog','information_schema') AND table_type='BASE TABLE')`)
				if res.Err == "" && len(res.Rows) == 1 {
					rw.Size, rw.Tables = res.Rows[0][0], res.Rows[0][1]
				} else {
					rw.Size = "not loaded"
				}
			}
			out = append(out, rw)
		}
		_ = tpl.ExecuteTemplate(w, "home", map[string]any{"Dbs": out})
	})

	http.HandleFunc("/db/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/db/"), "/")
		name := parts[0]
		p, err := pool(name)
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		if len(parts) == 4 && parts[1] == "t" { // /db/{db}/t/{schema}/{table}
			s, t := parts[2], parts[3]
			if !identRe.MatchString(s) || !identRe.MatchString(t) {
				http.Error(w, "bad identifier", 400)
				return
			}
			page, _ := strconv.Atoi(r.URL.Query().Get("page"))
			if page < 1 {
				page = 1
			}
			q := fmt.Sprintf(`SELECT * FROM "%s"."%s" LIMIT 100 OFFSET %d`, s, t, (page-1)*100)
			_ = tpl.ExecuteTemplate(w, "table", map[string]any{"Title": t, "Db": name, "Table": s + "." + t,
				"Page": page, "Prev": page - 1, "Next": page + 1, "SQL": fmt.Sprintf(`SELECT * FROM "%s"."%s" LIMIT 100`, s, t), "Res": runQuery(p, q)})
			return
		}
		res := runQuery(p, `SELECT n.nspname, c.relname, greatest(c.reltuples,0)::bigint, pg_size_pretty(pg_total_relation_size(c.oid)),
		  (SELECT string_agg(attname, ', ' ORDER BY attnum) FROM pg_attribute WHERE attrelid=c.oid AND attnum>0 AND NOT attisdropped)
		  FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace
		  WHERE c.relkind IN ('r','p','v','m') AND n.nspname NOT IN ('pg_catalog','information_schema') AND n.nspname NOT LIKE 'pg_toast%'
		  ORDER BY n.nspname, c.relname`)
		type tr struct{ Schema, Name, Rows, Size, Cols string }
		var tables []tr
		for _, rw := range res.Rows {
			tables = append(tables, tr{rw[0], rw[1], rw[2], rw[3], rw[4]})
		}
		blurb := ""
		for _, d := range databases {
			if d.Name == name {
				blurb = d.Blurb
			}
		}
		_ = tpl.ExecuteTemplate(w, "db", map[string]any{"Title": name, "Name": name, "Blurb": blurb, "Tables": tables})
	})

	http.HandleFunc("/elements", func(w http.ResponseWriter, r *http.Request) {
		p, err := pool("elements")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		res := runQuery(p, `SELECT "AtomicNumber","Symbol","Element","Period","Group","Type", row_to_json(t)::text FROM periodic_table t ORDER BY 1`)
		if res.Err != "" {
			http.Error(w, res.Err, 500)
			return
		}
		palette := []string{"#8abf9c", "#d97c8a", "#d4b47b", "#78b5ba", "#7e9cd8", "#b893ce", "#c4a7e7", "#e6c384", "#98bb6c", "#7fb4ca", "#ffa066", "#a3d4d5"}
		colors := map[string]string{}
		type cell struct {
			Num, Row, Col          int
			Sym, Name, Color, Type string
		}
		var cells []cell
		var js strings.Builder
		js.WriteString("{")
		var types []string
		for _, rw := range res.Rows {
			if _, ok := colors[rw[5]]; !ok {
				types = append(types, rw[5])
				colors[rw[5]] = ""
			}
		}
		sort.Strings(types)
		for i, t := range types {
			colors[t] = palette[i%len(palette)]
		}
		for i, rw := range res.Rows {
			n, _ := strconv.Atoi(rw[0])
			period, _ := strconv.Atoi(rw[3])
			group, _ := strconv.Atoi(rw[4])
			row, col := period, group
			// lanthanides / actinides go to separate rows below the main table
			if (n >= 57 && n <= 71) || (n >= 89 && n <= 103) {
				if n <= 71 {
					row, col = 9, n-57+3
				} else {
					row, col = 10, n-89+3
				}
			}
			if col == 0 {
				col = 3
			}
			cells = append(cells, cell{n, row, col, rw[1], rw[2], colors[rw[5]], rw[5]})
			if i > 0 {
				js.WriteString(",")
			}
			fmt.Fprintf(&js, "%q:%s", rw[0], rw[6])
		}
		js.WriteString("}")
		_ = tpl.ExecuteTemplate(w, "elements", map[string]any{"Title": "periodic table", "Cells": cells, "Colors": colors, "JSON": template.JS(js.String())})
	})

	http.HandleFunc("/astronomy", func(w http.ResponseWriter, r *http.Request) {
		p, err := pool("astronomy")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		type planetCard struct {
			Name, Type, Color, Dist, Moons, Cat, BadgeClass, Icon string
			HasRings                                              bool
		}
		type moonCard struct {
			Name, Planet, Diam string
			Ocean, Atm         bool
		}
		type exoCard struct {
			Name, Host, Dist, Type string
			HZ                     bool
		}
		type messierCard struct {
			Id, Name, Type, Constellation, Mag string
		}
		type constCard struct {
			Abbr, Name, Genitive, BrightStar, Area, Quadrant string
		}

		planetIcons := map[string]string{
			"Mercury":  "☿",
			"Venus":    "♀",
			"Earth":    "🌍",
			"Mars":     "♂",
			"Ceres":    "⚪",
			"Jupiter":  "♃",
			"Saturn":   "🪐",
			"Uranus":   "♅",
			"Neptune":  "♆",
			"Pluto":    "♇",
			"Haumea":   "🥚",
			"Makemake": "❄️",
			"Eris":     "🪨",
		}

		var planets []planetCard
		var moons []moonCard
		var exoplanets []exoCard
		var messier []messierCard
		var constellations []constCard

		var js strings.Builder
		js.WriteString(`{"planet":{`)

		// 1. Planets
		pres := runQuery(p, `SELECT name, type, distance_from_sun_au, moons_count, has_rings, color_hex, row_to_json(t)::text FROM planets t ORDER BY order_from_sun`)
		for i, rw := range pres.Rows {
			cat := "terrestrial"
			bclass := "badge-terrestrial"
			if strings.Contains(rw[1], "Gas") || strings.Contains(rw[1], "Ice") {
				cat = "giant"
				bclass = "badge-giant"
			} else if strings.Contains(rw[1], "Dwarf") {
				cat = "dwarf"
				bclass = "badge-dwarf"
			}
			pName := rw[0]
			icon := planetIcons[pName]
			if icon == "" {
				icon = "●"
			}
			planets = append(planets, planetCard{
				Name:       pName,
				Type:       rw[1],
				Dist:       rw[2],
				Moons:      rw[3],
				HasRings:   rw[4] == "t" || rw[4] == "true",
				Color:      rw[5],
				Cat:        cat,
				BadgeClass: bclass,
				Icon:       icon,
			})
			if i > 0 {
				js.WriteString(",")
			}
			fmt.Fprintf(&js, "%q:%s", rw[0], rw[6])
		}

		// 2. Moons
		js.WriteString(`},"moon":{`)
		mres := runQuery(p, `SELECT name, planet_name, diameter_km, has_subsurface_ocean, has_atmosphere, row_to_json(t)::text FROM moons t ORDER BY diameter_km DESC`)
		for i, rw := range mres.Rows {
			moons = append(moons, moonCard{
				Name:   rw[0],
				Planet: rw[1],
				Diam:   rw[2],
				Ocean:  rw[3] == "true",
				Atm:    rw[4] == "true",
			})
			if i > 0 {
				js.WriteString(",")
			}
			fmt.Fprintf(&js, "%q:%s", rw[0], rw[5])
		}

		// 3. Exoplanets
		js.WriteString(`},"exoplanet":{`)
		eres := runQuery(p, `SELECT name, host_star, distance_light_years, planet_type, habitable_zone, row_to_json(t)::text FROM exoplanets t ORDER BY distance_light_years ASC`)
		for i, rw := range eres.Rows {
			exoplanets = append(exoplanets, exoCard{
				Name: rw[0],
				Host: rw[1],
				Dist: rw[2],
				Type: rw[3],
				HZ:   rw[4] == "true",
			})
			if i > 0 {
				js.WriteString(",")
			}
			fmt.Fprintf(&js, "%q:%s", rw[0], rw[5])
		}

		// 4. Messier
		js.WriteString(`},"messier":{`)
		cres := runQuery(p, `SELECT messier_id, common_name, object_type, constellation, apparent_magnitude, row_to_json(t)::text FROM messier_catalog t ORDER BY apparent_magnitude ASC`)
		for i, rw := range cres.Rows {
			messier = append(messier, messierCard{
				Id:            rw[0],
				Name:          rw[1],
				Type:          rw[2],
				Constellation: rw[3],
				Mag:           rw[4],
			})
			if i > 0 {
				js.WriteString(",")
			}
			fmt.Fprintf(&js, "%q:%s", rw[0], rw[5])
		}
		// 5. Constellations
		js.WriteString(`},"constellation":{`)
		csres := runQuery(p, `SELECT abbreviation, name, genitive, brightest_star, area_sq_deg, quadrant, row_to_json(t)::text FROM constellations t ORDER BY name ASC`)
		for i, rw := range csres.Rows {
			constellations = append(constellations, constCard{
				Abbr:       rw[0],
				Name:       rw[1],
				Genitive:   rw[2],
				BrightStar: rw[3],
				Area:       rw[4],
				Quadrant:   rw[5],
			})
			if i > 0 {
				js.WriteString(",")
			}
			fmt.Fprintf(&js, "%q:%s", rw[0], rw[6])
		}
		js.WriteString(`}}`)

		data := map[string]any{
			"Title":          "astronomy",
			"Planets":        planets,
			"Moons":          moons,
			"Exoplanets":     exoplanets,
			"Messier":        messier,
			"Constellations": constellations,
			"JSON":           template.JS(js.String()),
		}
		_ = tpl.ExecuteTemplate(w, "astronomy", data)
	})

	http.HandleFunc("/math", func(w http.ResponseWriter, r *http.Request) {
		p, err := pool("math")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		type constCard struct {
			Symbol, Name, ValApprox, Domain string
		}
		type polyCard struct {
			Name, Family, Schlafli string
			V, E, F                int
		}
		type numCard struct {
			Name, Sym, Oeis, FirstPreview string
		}
		type thmCard struct {
			Name, Field, Prover, Year string
		}
		type crvCard struct {
			Name, Type, Discoverer string
		}

		var constants []constCard
		var polyhedra []polyCard
		var numClasses []numCard
		var theorems []thmCard
		var curves []crvCard

		var js strings.Builder
		js.WriteString(`{"const":{`)

		// 1. Constants
		cres := runQuery(p, `SELECT symbol, name, value_approx, domain, row_to_json(t)::text FROM constants t ORDER BY symbol ASC`)
		for i, rw := range cres.Rows {
			val := rw[2]
			if len(val) > 20 {
				val = val[:20] + "…"
			}
			constants = append(constants, constCard{
				Symbol:    rw[0],
				Name:      rw[1],
				ValApprox: val,
				Domain:    rw[3],
			})
			if i > 0 {
				js.WriteString(",")
			}
			fmt.Fprintf(&js, "%q:%s", rw[0], rw[4])
		}

		// 2. Polyhedra
		js.WriteString(`},"poly":{`)
		pres := runQuery(p, `SELECT name, family, schlafli_symbol, vertices, edges, faces, row_to_json(t)::text FROM polyhedra t ORDER BY faces ASC`)
		for i, rw := range pres.Rows {
			v, _ := strconv.Atoi(rw[3])
			e, _ := strconv.Atoi(rw[4])
			f, _ := strconv.Atoi(rw[5])
			polyhedra = append(polyhedra, polyCard{
				Name:     rw[0],
				Family:   rw[1],
				Schlafli: rw[2],
				V:        v,
				E:        e,
				F:        f,
			})
			if i > 0 {
				js.WriteString(",")
			}
			fmt.Fprintf(&js, "%q:%s", rw[0], rw[6])
		}

		// 3. Number Classes
		js.WriteString(`},"seq":{`)
		nres := runQuery(p, `SELECT name, symbol, oeis_id, first_terms, row_to_json(t)::text FROM number_classes t ORDER BY name ASC`)
		for i, rw := range nres.Rows {
			prev := rw[3]
			if len(prev) > 22 {
				prev = prev[:22] + "…"
			}
			numClasses = append(numClasses, numCard{
				Name:         rw[0],
				Sym:          rw[1],
				Oeis:         rw[2],
				FirstPreview: prev,
			})
			if i > 0 {
				js.WriteString(",")
			}
			fmt.Fprintf(&js, "%q:%s", rw[0], rw[4])
		}

		// 4. Theorems
		js.WriteString(`},"thm":{`)
		tres := runQuery(p, `SELECT name, field, prover, year_proven, row_to_json(t)::text FROM theorems t ORDER BY year_proven ASC`)
		for i, rw := range tres.Rows {
			theorems = append(theorems, thmCard{
				Name:   rw[0],
				Field:  rw[1],
				Prover: rw[2],
				Year:   rw[3],
			})
			if i > 0 {
				js.WriteString(",")
			}
			fmt.Fprintf(&js, "%q:%s", rw[0], rw[4])
		}

		// 5. Curves
		js.WriteString(`},"crv":{`)
		crres := runQuery(p, `SELECT name, curve_type, coalesce(discoverer, 'Classical'), row_to_json(t)::text FROM curves t ORDER BY name ASC`)
		for i, rw := range crres.Rows {
			curves = append(curves, crvCard{
				Name:       rw[0],
				Type:       rw[1],
				Discoverer: rw[2],
			})
			if i > 0 {
				js.WriteString(",")
			}
			fmt.Fprintf(&js, "%q:%s", rw[0], rw[3])
		}
		js.WriteString(`}}`)

		data := map[string]any{
			"Title":         "mathematics",
			"Constants":     constants,
			"Polyhedra":     polyhedra,
			"NumberClasses": numClasses,
			"Theorems":      theorems,
			"Curves":        curves,
			"JSON":          template.JS(js.String()),
		}
		_ = tpl.ExecuteTemplate(w, "math", data)
	})

	http.HandleFunc("/anatomy", func(w http.ResponseWriter, r *http.Request) {
		p, err := pool("anatomy")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		type boneCard struct {
			Name, LatinName, Division, Region, Shape, Cat, BadgeClass string
			Count                                                     int
		}
		type jointCard struct {
			Name, JointType, TypeShort, ClassShort string
		}
		type muscleCard struct {
			Name, LatinName, Region string
		}
		type organCard struct {
			Name, OrgansShort string
		}

		var bones []boneCard
		var joints []jointCard
		var muscles []muscleCard
		var organs []organCard

		var js strings.Builder
		js.WriteString(`{"bone":{`)

		// 1. Bones
		bres := runQuery(p, `SELECT name, latin_name, division, region, shape, count_in_body, row_to_json(t)::text FROM bones t ORDER BY division ASC, region ASC, name ASC`)
		for i, rw := range bres.Rows {
			cnt, _ := strconv.Atoi(rw[5])
			cat := "axial"
			bclass := "badge-axial"
			if strings.Contains(rw[2], "Appendicular") {
				cat = "append"
				bclass = "badge-append"
			}
			bones = append(bones, boneCard{
				Name:       rw[0],
				LatinName:  rw[1],
				Division:   rw[2],
				Region:     rw[3],
				Shape:      rw[4],
				Count:      cnt,
				Cat:        cat,
				BadgeClass: bclass,
			})
			if i > 0 {
				js.WriteString(",")
			}
			fmt.Fprintf(&js, "%q:%s", rw[0], rw[6])
		}

		// 2. Joints
		js.WriteString(`},"joint":{`)
		jres := runQuery(p, `SELECT name, joint_type, functional_class, row_to_json(t)::text FROM joints t ORDER BY name ASC`)
		for i, rw := range jres.Rows {
			tshort := rw[1]
			if idx := strings.Index(tshort, " ("); idx != -1 {
				tshort = tshort[:idx]
			}
			cshort := rw[2]
			if idx := strings.Index(cshort, " ("); idx != -1 {
				cshort = cshort[:idx]
			}
			joints = append(joints, jointCard{
				Name:       rw[0],
				JointType:  rw[1],
				TypeShort:  tshort,
				ClassShort: cshort,
			})
			if i > 0 {
				js.WriteString(",")
			}
			fmt.Fprintf(&js, "%q:%s", rw[0], rw[3])
		}

		// 3. Muscles
		js.WriteString(`},"muscle":{`)
		mres := runQuery(p, `SELECT name, latin_name, region, row_to_json(t)::text FROM muscles t ORDER BY region ASC, name ASC`)
		for i, rw := range mres.Rows {
			muscles = append(muscles, muscleCard{
				Name:      rw[0],
				LatinName: rw[1],
				Region:    rw[2],
			})
			if i > 0 {
				js.WriteString(",")
			}
			fmt.Fprintf(&js, "%q:%s", rw[0], rw[3])
		}

		// 4. Organ Systems
		js.WriteString(`},"organ":{`)
		ores := runQuery(p, `SELECT name, primary_organs, row_to_json(t)::text FROM organ_systems t ORDER BY name ASC`)
		for i, rw := range ores.Rows {
			oshort := rw[1]
			if len(oshort) > 36 {
				oshort = oshort[:36] + "…"
			}
			organs = append(organs, organCard{
				Name:        rw[0],
				OrgansShort: oshort,
			})
			if i > 0 {
				js.WriteString(",")
			}
			fmt.Fprintf(&js, "%q:%s", rw[0], rw[2])
		}
		js.WriteString(`}}`)

		data := map[string]any{
			"Title":   "anatomy",
			"Bones":   bones,
			"Joints":  joints,
			"Muscles": muscles,
			"Organs":  organs,
			"JSON":    template.JS(js.String()),
		}
		_ = tpl.ExecuteTemplate(w, "anatomy", data)
	})

	port := getEnv("PORT", "8080")
	addr := getEnv("LISTEN_ADDR", ":"+port)
	log.Printf("rock-sql listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
