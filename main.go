package main

import (
	"html/template"

	"net/http"
	"time"

	"github.com/go-playground/livereload"
	"github.com/go-playground/locales"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/fr"
	"github.com/go-playground/log"
	"github.com/go-playground/log/handlers/console"
	"github.com/go-playground/pure"
	mw "github.com/go-playground/pure/examples/middleware/logging-recovery"
	"github.com/go-playground/pure/middleware"
	"github.com/go-playground/universal-translator"
)

const (
	productionJQuery  = "<script src=\"https://ajax.googleapis.com/ajax/libs/jquery/3.1.0/jquery.min.js\"></script>"
	developmentJQuery = "<script src=\"/assets/js/jquery-3.1.0.min.js\"></script>"
	livereloadScript  = "<script src=\"/assets/js/livereload.js?host=localhost\"></script>"
	defaultlocale     = "en"
)

var (
	prod        bool
	tpls        *template.Template
	uni         *ut.UniversalTranslator
	localeArray = []locales.Translator{
		en.New(),
		fr.New(),
	}
)

func init() {

	cLog := console.New()
	cLog.RedirectSTDLogOutput(true)

	log.RegisterHandler(cLog, log.AllLevels...)
}

func main() {

	var err error

	uni = ut.New(localeArray[0], localeArray...)

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

	var found bool
	var loc ut.Translator

	l := r.URL.Query().Get("locale")

	loc, found = uni.GetTranslator(l)
	if !found && len(l) > 0 {
		loc, _ = uni.FindTranslator(pure.AcceptedLanguages(r)...)
	}

	num := 1987654321.51

	s := struct {
		Locales           []locales.Translator
		Selected          ut.Translator
		TimeSectionHeader string
		Time              time.Time
		Number            float64
		NegativeNumber    float64
		Percent           float64
	}{
		Locales:           localeArray,
		Selected:          loc,
		TimeSectionHeader: "",
		Time:              time.Now().UTC(),
		Number:            num,
		NegativeNumber:    num * -1,
		Percent:           45.67,
	}

	err := tpls.ExecuteTemplate(w, "root", s)
	if err != nil {
		log.Error(err)
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
