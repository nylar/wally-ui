package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	rdb "github.com/dancannon/gorethink"
	"github.com/fatih/color"
	"github.com/nylar/wally"
)

var (
	session *rdb.Session
)

func init() {
	//wally.ItemsPerPage = 10
	var err error

	confData, err := ioutil.ReadFile("config.yml")
	if err != nil {
		color.Set(color.FgRed)
		log.Fatalln(err.Error())
		color.Unset()
	}
	wally.Conf, err = wally.LoadConfig(confData)
	if err != nil {
		color.Set(color.FgRed)
		log.Fatalln(err.Error())
		color.Unset()
	}

	session, err = rdb.Connect(rdb.ConnectOpts{
		Address:  wally.Conf.Database.Host,
		Database: wally.Conf.Database.Name,
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

func indexHandler(w http.ResponseWriter, r *http.Request) {
	t := template.New("index.html")
	t = t.Funcs(funcs)

	tmpl, err := t.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "index.html", nil); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	ctx := struct {
		Res   *wally.Results
		Query string
	}{
		nil,
		"",
	}

	query := r.URL.Query().Get("query")
	pageParam := r.URL.Query().Get("page")
	page, err := strconv.Atoi(pageParam)
	if err != nil {
		page = 1
	}

	res := new(wally.Results)
	res, err = wally.Search(query, session, page)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	ctx.Res = res
	ctx.Query = query

	t := template.New("search.html")
	t = t.Funcs(funcs)

	tmpl, err := t.ParseFiles("templates/search.html")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "search.html", ctx); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func main() {
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	http.HandleFunc("/search/", searchHandler)
	http.HandleFunc("/", indexHandler)
	http.ListenAndServe(":8008", nil)
}
