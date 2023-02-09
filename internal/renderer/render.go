package renderer

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/atom91/bookings/internal/config"
	"github.com/atom91/bookings/internal/models"
	"github.com/justinas/nosurf"
)
var functions = template.FuncMap{
	"humanDate": HumanDate,
	"iterate": Iterate,
}
var app *config.AppConfig
var pathToTemplates = "./templates"

func NewTemplates(a *config.AppConfig) {
	app = a
}
//Returns Time in YYYY-MM-DD format
func HumanDate(t time.Time)string{
	return t.Format("2006-01-02")
}

func Iterate(count int)[]int{
	var i int
	var items []int
	for i=0;i<count;i++{
		items=append(items, i)
	}
	return items
}
func addDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.CSRFToken = nosurf.Token(r)
	if app.Session.Exists(r.Context(),"user_id"){
		td.IsAuthenticated=1
	}
	return td
}
func RenderTemplate(w http.ResponseWriter, tmpl string, td *models.TemplateData, r *http.Request) error {
	var tc map[string]*template.Template
	log.Printf(fmt.Sprintf("%t", app.UseCache))
	if app.UseCache {
		// get the template cache from the app config
		tc = app.TemplateCache
	} else {
		tc, _ = CreateTemplateCache()

	}
	//get template from cache
	t, ok := tc[tmpl]
	if !ok {
		log.Fatal("Could not get template from cache ")
		return errors.New("Could not get template from cache ")
	}

	buf := new(bytes.Buffer)
	td = addDefaultData(td, r)
	err := t.Execute(buf, td)
	if err != nil {
		log.Println(err)
		return err
	}
	//render the template
	_, err = buf.WriteTo(w)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil

}
func CreateTemplateCache() (map[string]*template.Template, error) {
	myCache := map[string]*template.Template{}
	pages, err := filepath.Glob(fmt.Sprintf("%s/*.html", pathToTemplates))
	if err != nil {
		return myCache, err
	}
	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
		if err != nil {

			return myCache, err
		}
		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
			if err != nil {

				return myCache, err
			}
		}
		myCache[name] = ts

	}
	return myCache, nil

}
