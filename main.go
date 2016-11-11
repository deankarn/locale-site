package main

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/go-playground/livereload"
	"github.com/go-playground/locales"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/fr"
	"github.com/go-playground/pure"
	mw "github.com/go-playground/pure/examples/middleware/logging-recovery"
	"github.com/go-playground/pure/middleware"
)

const (
	productionJQuery  = "<script src=\"https://ajax.googleapis.com/ajax/libs/jquery/3.1.0/jquery.min.js\"></script>"
	developmentJQuery = "<script src=\"/assets/js/jquery-3.1.0.min.js\"></script>"
	livereloadScript  = "<script src=\"/assets/js/livereload.js?host=localhost\"></script>"
	defaultlocale     = "en"
)

var (
	prod      bool
	tpls      *template.Template
	localeMap = map[string]locales.Translator{
		"en": en.New(),
		"fr": fr.New(),
	}
)

func main() {

	var err error

	err = lr()
	if err != nil {
		log.Fatal(err)
	}

	tpls, err = initTemplates()
	if err != nil {
		log.Fatal(err)
	}

	p := pure.New()
	p.Use(mw.LoggingAndRecovery(true), middleware.Gzip)
	p.Get("/", root)

	assets := p.Group("/assets/", nil)
	assets.Use(middleware.Gzip)
	assets.Get("*", http.StripPrefix("/assets", http.FileServer(http.Dir("assets"))).ServeHTTP)

	log.Println(http.ListenAndServe(":8080", p.Serve()))
}

func root(w http.ResponseWriter, r *http.Request) {

	var ok bool
	var loc locales.Translator

	l := r.URL.Query().Get("locale")

	loc, ok = localeMap[l]
	if !ok {
		languages := pure.AcceptedLanguages(r)

		for i := 0; i < len(languages); i++ {
			if loc, ok = localeMap[languages[i]]; ok {
				break
			}
		}

		if len(l) == 0 {
			loc = localeMap[defaultlocale]
		}
	}

	s := struct {
		Locales  map[string]locales.Translator
		Selected locales.Translator
		Time     time.Time
	}{
		Locales:  localeMap,
		Selected: loc,
		Time:     time.Now().UTC(),
	}

	err := tpls.ExecuteTemplate(w, "root", s)
	if err != nil {
		log.Fatal(err)
	}
}

func initTemplates() (*template.Template, error) {

	funcMap := template.FuncMap{
		"jquery": func() template.HTML {
			if prod {
				return template.HTML(productionJQuery)
			}

			return template.HTML(developmentJQuery)
		},
		"livereload": func() template.HTML {
			if !prod {
				return template.HTML(livereloadScript)
			}

			return template.HTML("")
		},
	}

	templates, err := template.New("").Funcs(funcMap).ParseGlob("./*tmpl")
	if err != nil {
		return nil, err
	}

	return templates.Funcs(funcMap), err
}

func lr() error {
	if prod {
		return nil
	}

	paths := []string{
		"./",
	}

	tmplFn := func(name string) (bool, error) {

		templates, err := initTemplates()
		if err != nil {
			return false, err
		}

		*tpls = *templates

		return true, nil

	}

	mappings := livereload.ReloadMapping{
		".css":  nil,
		".js":   nil,
		".tmpl": tmplFn,
	}

	_, err := livereload.ListenAndServe(livereload.DefaultPort, paths, mappings)

	return err
}
