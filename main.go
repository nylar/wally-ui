package main

import (
	"html/template"
	"log"
	"net/http"
	"os"

	rdb "github.com/dancannon/gorethink"
	"github.com/fatih/color"
	"github.com/nylar/wally"
)

var session *rdb.Session

func init() {
	var err error
	session, err = rdb.Connect(rdb.ConnectOpts{
		Address:  os.Getenv("RETHINKDB_URL"),
		Database: wally.Database,
	})
	if err != nil {
		color.Set(color.FgRed)
		log.Fatalln(err.Error())
		color.Unset()
	}
}

func TruncateContent(value string) template.HTML {
	return template.HTML(wally.TruncateText(value, " ...", 200))
}

var funcs = template.FuncMap{
	"truncateContent": TruncateContent,
}

func handler(w http.ResponseWriter, r *http.Request) {
	res := new(wally.Results)
	ctx := struct {
		Res   *wally.Results
		Query string
	}{
		nil,
		"",
	}

	if r.Method == "POST" {
		var err error
		query := r.PostFormValue("query")
		res, err = wally.Search(query, session)
		ctx.Res = res
		ctx.Query = query
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}

	t := template.New("template.html")
	t = t.Funcs(funcs)

	tmpl, err := t.ParseFiles("template.html")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "template.html", ctx); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8008", nil)
}
