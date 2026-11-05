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
	p, err := sql.Open("postgres", fmt.Sprintf("host=127.0.0.1 user=explorer password=explorer dbname=%s sslmode=disable", name))
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
<meta name="viewport" content="width=device-width, initial-scale=1"><title>rock-sql{{if .Title}} // {{.Title}}{{end}}</title>
<style>
:root{--bg:#16161c;--panel:#1f1f28;--border:#2e2e3e;--text:#dcd7ba;--muted:#727169;--mint:#8abf9c;--rose:#d97c8a;--gold:#d4b47b;--cyan:#78b5ba;--blue:#7e9cd8}
*{box-sizing:border-box}body{margin:0;padding:18px;background:var(--bg);color:var(--text);font:13px/1.5 ui-monospace,Menlo,Consolas,monospace}
a{color:var(--cyan);text-decoration:none}a:hover{text-decoration:underline}
.wrap{max-width:1400px;margin:0 auto}header{display:flex;gap:18px;align-items:center;padding:12px 16px;background:var(--panel);border:1px solid var(--border);border-radius:8px;margin-bottom:16px;flex-wrap:wrap}
header h1{margin:0;font-size:17px;color:var(--cyan)}nav a{margin-right:14px}
.card{background:var(--panel);border:1px solid var(--border);border-radius:8px;padding:14px 16px;margin-bottom:16px;overflow-x:auto}
h2{margin:0 0 10px;font-size:14px;color:var(--blue)}.muted{color:var(--muted)}
table{border-collapse:collapse;font-size:12px;width:100%}th,td{text-align:left;padding:4px 8px;border-bottom:1px solid #252532;vertical-align:top;max-width:420px;overflow:hidden;text-overflow:ellipsis}
th{color:var(--gold);font-weight:normal;position:sticky;top:0;background:var(--panel)}
.grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(300px,1fr));gap:14px}
textarea{width:100%;min-height:140px;background:#101015;color:var(--text);border:1px solid var(--border);border-radius:6px;padding:10px;font:13px ui-monospace,monospace}
button,.btn{background:#2a2a37;color:var(--text);border:1px solid var(--cyan);padding:6px 14px;border-radius:5px;font-family:inherit;cursor:pointer;display:inline-block}
button:hover,.btn:hover{background:#353545}
.btn.active{background:var(--cyan);color:#16161c;font-weight:bold}
.err{color:var(--rose);white-space:pre-wrap}.chips a{display:inline-block;margin:2px 6px 2px 0;padding:2px 8px;border:1px solid var(--border);border-radius:4px}
</style></head><body><div class="wrap">
<header><h1>🐡 rock-sql</h1><nav><a href="/">databases</a><a href="/elements">periodic table</a><a href="/astronomy">astronomy</a></nav></header>{{end}}
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
<div class="pt" style="margin-top:10px">{{range .Cells}}{{if .Num}}<div class="el" style="grid-column:{{.Col}};grid-row:{{.Row}};background:{{.Color}}" onclick="show({{.Num}})">
<span class="n">{{.Num}}</span><b>{{.Sym}}</b>{{.Name}}</div>{{end}}{{end}}</div></div>
<div class="card" id="info"><span class="muted">Select an element…</span></div>
<script>const E={{.JSON}};function show(n){const e=E[n];let h='<h2>'+e.Element+' ('+e.Symbol+') — #'+n+'</h2><table>';
for(const k in e){if(e[k]!==null&&e[k]!=='')h+='<tr><th style="width:220px">'+k+'</th><td>'+e[k]+'</td></tr>'}
document.getElementById('info').innerHTML=h+'</table>'}</script>
{{template "bottom" .}}{{end}}

{{define "astronomy"}}{{template "top" .}}
<style>
.tabbar{display:flex;gap:8px;flex-wrap:wrap;margin-bottom:14px}
.astro-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(210px,1fr));gap:10px}
.astro-card{background:#181822;border:1px solid var(--border);border-radius:6px;padding:10px;cursor:pointer;transition:border-color .15s}
.astro-card:hover{border-color:var(--cyan);outline:1px solid var(--cyan)}
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
</style>
<div class="card">
<h2>🪐 Astronomy & Solar System Explorer <span class="muted">(click any celestial body to inspect & query)</span></h2>
<div class="tabbar">
  <button type="button" class="btn active" onclick="filterAstro('all', this)">All Solar System</button>
  <button type="button" class="btn" onclick="filterAstro('terrestrial', this)">Terrestrial</button>
  <button type="button" class="btn" onclick="filterAstro('giant', this)">Gas & Ice Giants</button>
  <button type="button" class="btn" onclick="filterAstro('dwarf', this)">Dwarf Planets</button>
  <button type="button" class="btn" onclick="filterAstro('ocean', this)">🌊 Ocean Moons</button>
  <button type="button" class="btn" onclick="filterAstro('exoplanet', this)">🔭 Exoplanets (HZ)</button>
  <button type="button" class="btn" onclick="filterAstro('messier', this)">🌌 Messier Deep Sky</button>
</div>
<div class="astro-grid" id="astro-list">
{{range .Planets}}
<div class="astro-card" data-cat="{{.Cat}}" onclick="inspectAstro('planet', '{{.Name}}')">
  <div class="astro-head">
    <span class="astro-orb" style="background:{{.Color}}"></span>
    <div>
      <div class="astro-title">{{.Name}}{{if .HasRings}} 🪐{{end}}</div>
      <div class="astro-sub">{{.Dist}} AU · {{.Moons}} moon(s)</div>
    </div>
  </div>
  <span class="astro-badge {{.BadgeClass}}">{{.Type}}</span>
</div>
{{end}}
{{range .Moons}}
<div class="astro-card" data-cat="moon{{if .Ocean}} ocean{{end}}" onclick="inspectAstro('moon', '{{.Name}}')">
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
<div class="astro-card" data-cat="exoplanet" onclick="inspectAstro('exoplanet', '{{.Name}}')">
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
<div class="astro-card" data-cat="messier" onclick="inspectAstro('messier', '{{.Id}}')">
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
</div>
</div>

<div class="card" id="astro-detail">
<span class="muted">Select a celestial body above to inspect its properties and launch queries…</span>
</div>

<script>
const AstroData = {{.JSON}};
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
  let title = key;
  if (kind === 'planet') {
    title = '🪐 ' + obj.name + ' (' + obj.type + ')';
  } else if (kind === 'moon') {
    title = '🌕 ' + obj.name + ' (Moon of ' + obj.planet_name + ')';
  } else if (kind === 'exoplanet') {
    title = '🔭 ' + obj.name + ' (' + obj.planet_type + ' in ' + obj.constellation + ')';
  } else if (kind === 'messier') {
    title = '🌌 ' + obj.messier_id + ' — ' + obj.common_name + ' (' + obj.object_type + ')';
  }
  let h = '<h2>' + title + '</h2><table>';
  for (const k in obj) {
    if (obj[k] !== null && obj[k] !== '') {
      h += '<tr><th style="width:240px">' + k + '</th><td>' + obj[k] + '</td></tr>';
    }
  }
  h += '</table>';
  document.getElementById('astro-detail').innerHTML = h;
  document.getElementById('astro-detail').scrollIntoView({ behavior: 'smooth', block: 'nearest' });
}
</script>
{{template "bottom" .}}{{end}}
`))

func main() {
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
			Name, Type, Color, Dist, Moons, Cat, BadgeClass string
			HasRings                                        bool
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

		var planets []planetCard
		var moons []moonCard
		var exoplanets []exoCard
		var messier []messierCard

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
			planets = append(planets, planetCard{
				Name:       rw[0],
				Type:       rw[1],
				Dist:       rw[2],
				Moons:      rw[3],
				HasRings:   rw[4] == "true",
				Color:      rw[5],
				Cat:        cat,
				BadgeClass: bclass,
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
		js.WriteString(`}}`)

		data := map[string]any{
			"Title":      "astronomy",
			"Planets":    planets,
			"Moons":      moons,
			"Exoplanets": exoplanets,
			"Messier":    messier,
			"JSON":       template.JS(js.String()),
		}
		_ = tpl.ExecuteTemplate(w, "astronomy", data)
	})

	log.Println("rock-sql listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
