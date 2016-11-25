package main

import (
	"html/template"

	"net/http"
	"time"

	"github.com/go-playground/livereload"
	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
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
	currencies  = make([]currency.Type, 296)
	localeArray = []locales.Translator{
		en.New(),
		fr.New(),
	}
	nativeCurrencies = map[string][]currency.Type{
		"en": {
			currency.USD,
		},
		"fr": {
			currency.EUR,
		},
	}
	otherCurrencies = map[string][]currency.Type{
		"en": {
			currency.EUR,
		},
		"fr": {
			currency.USD,
		},
	}
	localeMapInfo = map[string]map[string]string{
		"en": {
			"MainTextTrans":                        "Please take a look at the l10n rules for locales you are familiar with and please indicated if each is correct or provide a correction:",
			"LocaleTrans":                          "Locale",
			"TimeTrans":                            "Time",
			"TimeSectionTrans":                     "Times",
			"NumberTrans":                          "Number",
			"NegativeNumberTrans":                  "Negative Number",
			"NumberSectionTrans":                   "Number Formatting",
			"PercentTrans":                         "Percent Number",
			"PercentSectionTrans":                  "Percentages",
			"DateSectionTrans":                     "Dates",
			"MonthsSectionTrans":                   "Months",
			"WeekdaysSectionTrans":                 "Weekdays",
			"PluralsSectionTrans":                  "Plural Rules",
			"NativeCurrencySectionTrans":           "Native Currencies",
			"OtherCurrencySectionTrans":            "Other Currencies",
			"NativeCurrencyAccountingSectionTrans": "Native Currency Accounting",
			"OtherCurrencyAccountingSectionTrans":  "Other Currency Accounting",
			"PositiveTrans":                        "Positive",
			"NegativeTrans":                        "Negative",
			"ShortTrans":                           "Short",
			"MediumTrans":                          "Medium",
			"LongTrans":                            "Long",
			"FullTrans":                            "Full",
			"NarrowTrans":                          "Narrow",
			"AbbreviatedTrans":                     "Abbreviated",
			"WideTrans":                            "Wide",
			"CardinalTrans":                        "Cardinal",
			"OrdinalTrans":                         "Ordinal",
			"RangeTrans":                           "Range",
		},
		"fr": {
			"MainTextTrans":                        "Veuillez prendre connaissance des règles l10n pour les lieux que vous connaissez et indiquer si chacune est correcte ou apporter une correction:",
			"LocaleTrans":                          "Lieu",
			"TimeTrans":                            "Temps",
			"TimeSectionTrans":                     "Fois",
			"NumberTrans":                          "Nombre",
			"NegativeNumberTrans":                  "Nombre négatif",
			"NumberSectionTrans":                   "Formatage des numéros",
			"PercentTrans":                         "Nombre de pourcentages",
			"PercentSectionTrans":                  "Pourcentages",
			"DateSectionTrans":                     "Rendez-vous",
			"MonthsSectionTrans":                   "Mois",
			"WeekdaysSectionTrans":                 "Jours de la semaine",
			"PluralsSectionTrans":                  "Règles plurielles",
			"NativeCurrencySectionTrans":           "Monnaies autochtones",
			"OtherCurrencySectionTrans":            "Autres monnaies",
			"NativeCurrencyAccountingSectionTrans": "Comptabilité en monnaie native",
			"OtherCurrencyAccountingSectionTrans":  "Autres Comptabilisation des devises",
			"PositiveTrans":                        "Positif",
			"NegativeTrans":                        "Négatif",
			"ShortTrans":                           "Court",
			"MediumTrans":                          "Moyen",
			"LongTrans":                            "Longue",
			"FullTrans":                            "Plein",
			"NarrowTrans":                          "Étroit",
			"AbbreviatedTrans":                     "Abrégé",
			"WideTrans":                            "Large",
			"CardinalTrans":                        "Cardinal",
			"OrdinalTrans":                         "Ordinal",
			"RangeTrans":                           "Gamme",
		},
	}
)

func init() {

	cLog := console.New()
	cLog.RedirectSTDLogOutput(true)

	log.RegisterHandler(cLog, log.AllLevels...)
}

func main() {

	var err error

	for i := 0; i < 295; i++ {
		currencies[i] = currency.Type(i)
	}

	var defaultLocale locales.Translator

	for _, l := range localeArray {

		if l.Locale() == "en" {
			defaultLocale = l
			break
		}
	}

	uni = ut.New(defaultLocale, localeArray...)
	setupTranslations()

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
		Locales          []locales.Translator
		Selected         ut.Translator
		Time             time.Time
		Number           float64
		NegativeNumber   float64
		Percent          float64
		NativeCurrencies []currency.Type
		OtherCurrencies  []currency.Type
	}{
		Locales:          localeArray,
		Selected:         loc,
		Time:             time.Now().UTC(),
		Number:           num,
		NegativeNumber:   num * -1,
		Percent:          45.67,
		NativeCurrencies: nativeCurrencies[loc.Locale()],
		OtherCurrencies:  otherCurrencies[loc.Locale()],
	}

	err := tpls.ExecuteTemplate(w, "root", s)
	if err != nil {
		log.Error(err)
	}
}

func setupTranslations() {

	var found bool
	var loc ut.Translator
	var err error

	for l, trans := range localeMapInfo {

		loc, found = uni.GetTranslator(l)
		if !found {
			log.WithFields(log.F("locale", l)).Error("Translator Not Found")
			continue
		}

		for k, v := range trans {
			err = loc.Add(k, v, false)
			if err != nil {
				log.WithFields(
					log.F("key", k),
					log.F("text", v),
					log.F("locale", l),
				).Error("Adding Translation")
			}
		}

		err = loc.VerifyTranslations()
		if err != nil {
			log.WithFields(log.F("error", err)).Error("Missing Translations")
		}
	}
}

func initTemplates() (*template.Template, error) {

	var jqFn, lrFn func() template.HTML

	if prod {
		jqFn = func() template.HTML {
			return template.HTML(productionJQuery)
		}
		lrFn = func() template.HTML {
			return template.HTML("")
		}
	} else {
		jqFn = func() template.HTML {
			return template.HTML(developmentJQuery)
		}
		lrFn = func() template.HTML {
			return template.HTML(livereloadScript)
		}
	}

	funcMap := template.FuncMap{
		"jquery":     jqFn,
		"livereload": lrFn,
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
