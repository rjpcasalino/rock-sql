// rock-sql: a tiny read-only PostgreSQL explorer + periodic table for rock.home.arpa
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
button,.btn{background:#2a2a37;color:var(--text);border:1px solid var(--cyan);padding:6px 14px;border-radius:5px;font-family:inherit;cursor:pointer}
.err{color:var(--rose);white-space:pre-wrap}.chips a{display:inline-block;margin:2px 6px 2px 0;padding:2px 8px;border:1px solid var(--border);border-radius:4px}
</style></head><body><div class="wrap">
<header><h1>🐡 rock-sql</h1><nav><a href="/">databases</a><a href="/elements">periodic table</a><a href="/sql?db=pagila">sql console</a><a href="/puzzles">puzzles</a></nav></header>{{end}}
{{define "bottom"}}</div></body></html>{{end}}
{{define "result"}}{{if .Err}}<div class="err">ERROR: {{.Err}}</div>{{else}}
<div class="muted">{{len .Rows}} row(s){{if .Truncate}} (capped at 1000){{end}} in {{.Elapsed}}</div>
<table><tr>{{range .Cols}}<th>{{.}}</th>{{end}}</tr>{{range .Rows}}<tr>{{range .}}<td title="{{.}}">{{.}}</td>{{end}}</tr>{{end}}</table>{{end}}{{end}}`

var tpl = template.Must(template.New("").Funcs(template.FuncMap{
	"q": url.QueryEscape,
}).Parse(base + `
{{define "home"}}{{template "top" .}}<div class="grid">{{range .Dbs}}<div class="card">
<h2><a href="/db/{{.Name}}">{{.Name}}</a></h2><div class="muted">{{.Blurb}}</div>
<div style="margin-top:8px">{{.Size}} · {{.Tables}} tables · <a href="/sql?db={{.Name}}">query →</a></div></div>{{end}}</div>{{template "bottom" .}}{{end}}

{{define "db"}}{{template "top" .}}<div class="card"><h2>{{.Name}}</h2><div class="muted">{{.Blurb}}</div></div>
<div class="card"><h2>Tables</h2><table><tr><th>table</th><th>~rows</th><th>size</th><th>columns</th></tr>
{{$d:=.Name}}{{range .Tables}}<tr><td><a href="/db/{{$d}}/t/{{.Schema}}/{{.Name}}">{{if ne .Schema "public"}}{{.Schema}}.{{end}}{{.Name}}</a></td><td>{{.Rows}}</td><td>{{.Size}}</td><td class="muted">{{.Cols}}</td></tr>{{end}}</table></div>{{template "bottom" .}}{{end}}

{{define "table"}}{{template "top" .}}<div class="card"><h2><a href="/db/{{.Db}}">{{.Db}}</a> / {{.Table}}</h2>
<div>page {{.Page}} · {{if gt .Page 1}}<a href="?page={{.Prev}}">← prev</a> · {{end}}<a href="?page={{.Next}}">next →</a> · <a href="/sql?db={{.Db}}&q={{q .SQL}}">open in console</a></div></div>
<div class="card">{{template "result" .Res}}</div>{{template "bottom" .}}{{end}}

{{define "sql"}}{{template "top" .}}<div class="card"><form method="GET" action="/sql">
<div style="margin-bottom:8px">database: <select name="db">{{$c:=.Db}}{{range .Dbs}}<option {{if eq .Name $c}}selected{{end}}>{{.Name}}</option>{{end}}</select>
<span class="muted"> read-only · 15s timeout · 1000 row cap · Ctrl+Enter runs</span></div>
<textarea name="q" id="q" autofocus>{{.Q}}</textarea><div style="margin-top:8px"><button type="submit">▶ Run</button></div></form></div>
{{if .Ran}}<div class="card">{{template "result" .Res}}</div>{{end}}
<script>document.getElementById('q').addEventListener('keydown',e=>{if(e.key==='Enter'&&(e.ctrlKey||e.metaKey))e.target.form.submit()})</script>{{template "bottom" .}}{{end}}

{{define "puzzles"}}{{template "top" .}}{{range .}}<div class="card"><h2>{{.Title}} <span class="muted">({{.Db}})</span></h2>
<div>{{.Ask}}</div><div class="chips" style="margin-top:6px"><a href="/sql?db={{.Db}}&q={{q .Starter}}">try it →</a>
<a href="/sql?db={{.Db}}&q={{q .Answer}}">peek at an answer</a></div></div>{{end}}{{template "bottom" .}}{{end}}

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
document.getElementById('info').innerHTML=h+'</table><div style="margin-top:8px"><a href="/sql?db=elements&q='+encodeURIComponent('SELECT * FROM periodic_table WHERE "AtomicNumber" = '+n)+'">open in console →</a></div>'}</script>
{{template "bottom" .}}{{end}}
`))

type puzzle struct{ Title, Db, Ask, Starter, Answer string }

var puzzles = []puzzle{
	{"Top 3 films per category", "pagila", "For each category, list the 3 films rented the most. (window functions: RANK)",
		"SELECT c.name, f.title\nFROM film f\nJOIN film_category fc USING (film_id)\nJOIN category c USING (category_id)\nLIMIT 20;",
		"SELECT name, title, rentals FROM (\n  SELECT c.name, f.title, count(r.rental_id) AS rentals,\n         rank() OVER (PARTITION BY c.name ORDER BY count(r.rental_id) DESC) AS rk\n  FROM film f\n  JOIN film_category fc USING (film_id)\n  JOIN category c USING (category_id)\n  JOIN inventory i USING (film_id)\n  JOIN rental r USING (inventory_id)\n  GROUP BY c.name, f.title\n) t WHERE rk <= 3 ORDER BY name, rk;"},
	{"Running revenue", "pagila", "Show daily payment totals and a running cumulative total. (SUM() OVER ORDER BY)",
		"SELECT payment_date::date AS day, sum(amount)\nFROM payment\nGROUP BY 1 ORDER BY 1;",
		"SELECT day, total, sum(total) OVER (ORDER BY day) AS running\nFROM (SELECT payment_date::date AS day, sum(amount) AS total FROM payment GROUP BY 1) t\nORDER BY day;"},
	{"Best-selling genres", "chinook", "Which genres earn the most money? (multi-join + GROUP BY)",
		"SELECT * FROM genre;",
		"SELECT g.name, sum(il.unit_price * il.quantity) AS revenue\nFROM invoice_line il\nJOIN track t USING (track_id)\nJOIN genre g USING (genre_id)\nGROUP BY g.name ORDER BY revenue DESC;"},
	{"Who survived?", "titanic", "Survival rate by passenger class and sex. (conditional aggregates)",
		"SELECT * FROM titanic LIMIT 10;",
		"SELECT * FROM titanic LIMIT 1; -- inspect columns, then:\n-- SELECT pclass, sex, round(avg(survived::int)*100,1) AS pct FROM titanic GROUP BY 1,2 ORDER BY 1,2;"},
	{"Salary raises", "employees", "Find the 10 biggest single salary jumps ever given. (LAG)",
		"SELECT * FROM employees.salary LIMIT 10;",
		"SELECT employee_id, from_date, amount,\n       amount - lag(amount) OVER (PARTITION BY employee_id ORDER BY from_date) AS raise\nFROM employees.salary\nORDER BY raise DESC NULLS LAST LIMIT 10;"},
	{"Discovery timeline", "elements", "How many elements were discovered each century, and by whom most often?",
		"SELECT \"Element\", \"Year\", \"Discoverer\" FROM periodic_table ORDER BY \"Year\";",
		"SELECT (\"Year\"/100+1) || 'th c.' AS century, count(*), mode() WITHIN GROUP (ORDER BY \"Discoverer\") AS top_discoverer\nFROM periodic_table WHERE \"Year\" IS NOT NULL\nGROUP BY 1 ORDER BY 1;"},
	{"Biggest LEGO sets", "lego", "Largest sets by number of parts, per decade.",
		"SELECT * FROM sets LIMIT 10;",
		"SELECT decade, name, num_parts FROM (\n  SELECT (year/10)*10 AS decade, name, num_parts,\n         row_number() OVER (PARTITION BY (year/10)*10 ORDER BY num_parts DESC) rn\n  FROM sets) t WHERE rn = 1 ORDER BY decade;"},
	{"Mandelbrot in SQL", "elements", "Just for fun: recursive CTE that draws the Mandelbrot set in ASCII.",
		"WITH RECURSIVE x(i) AS (SELECT 0 UNION ALL SELECT i+1 FROM x WHERE i < 9) SELECT * FROM x;",
		"WITH RECURSIVE\n x(i) AS (VALUES(0) UNION ALL SELECT i+1 FROM x WHERE i<67),\n y(i) AS (VALUES(0) UNION ALL SELECT i+1 FROM y WHERE i<23),\n m(iter,cx,cy,x,y) AS (\n  SELECT 0, x.i*3.0/67-2.2, y.i*2.4/23-1.2, 0.0::float, 0.0::float FROM x, y\n  UNION ALL\n  SELECT iter+1, cx, cy, x*x-y*y+cx, 2*x*y+cy FROM m WHERE x*x+y*y<4 AND iter<28)\nSELECT string_agg(substr(' .:-=+*#%@', least(10, 1+max_iter/3), 1), '' ORDER BY cx) AS mandelbrot\nFROM (SELECT cx, cy, max(iter) max_iter FROM m GROUP BY cx, cy) t\nGROUP BY cy ORDER BY cy;"},
}

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

	http.HandleFunc("/sql", func(w http.ResponseWriter, r *http.Request) {
		db := r.URL.Query().Get("db")
		if db == "" {
			db = "pagila"
		}
		q := r.URL.Query().Get("q")
		data := map[string]any{"Title": "console", "Db": db, "Q": q, "Dbs": databases, "Ran": false}
		if strings.TrimSpace(q) != "" {
			if p, err := pool(db); err == nil {
				data["Res"], data["Ran"] = runQuery(p, q), true
			}
		}
		_ = tpl.ExecuteTemplate(w, "sql", data)
	})

	http.HandleFunc("/puzzles", func(w http.ResponseWriter, r *http.Request) {
		_ = tpl.ExecuteTemplate(w, "puzzles", puzzles)
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

	log.Println("rock-sql listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
