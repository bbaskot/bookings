package renderer

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/atom91/bookings/internal/config"
	"github.com/atom91/bookings/internal/models"
	"github.com/justinas/nosurf"
)
var app * config.AppConfig
func NewTemplates(a *config.AppConfig){
   app=a
}
func addDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData{
	td.CSRFToken=nosurf.Token(r)
	return td
}
func RenderTemplate(w http.ResponseWriter, tmpl string, td *models.TemplateData, r *http.Request) {
	var tc map[string]*template.Template
	if app.UseCache{
		// get the template cache from the app config
		tc=app.TemplateCache
	}else{
		tc,_=CreateTemplateCache()
		
	}
	//get template from cache
	t, ok := tc[tmpl]
	if !ok{
		log.Fatal("Could not get template from cache ")
	}

	buf := new(bytes.Buffer)
	td=addDefaultData(td, r)
	err:= t.Execute(buf, td)
	if err!=nil{
		log.Println(err)
	}
	//render the template
	_,err=buf.WriteTo(w)
	if err!=nil{
		log.Println(err)
	}

}
func CreateTemplateCache()(map[string]*template.Template,error){
	myCache:=map[string]*template.Template{}
	pages, err := filepath.Glob("./templates/*.html")
	if err!=nil{
		return myCache,err
	}
	for _, page := range pages{
		name:=filepath.Base(page)
		ts, err :=template.New(name).ParseFiles(page)
		if err!=nil{
			return myCache,err
		}
		matches, err := filepath.Glob("./templates/*.layout.tmpl")
		if err!=nil{
			return myCache,err
		}
		if len(matches)>0{
			ts,err= ts.ParseGlob("./templates/*.layout.tmpl")
			if err!=nil{
				return myCache,err
			}
		}
		myCache[name]=ts

	}
	return myCache,nil

}