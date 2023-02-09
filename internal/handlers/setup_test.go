package handlers

import (
	"encoding/gob"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/atom91/bookings/internal/config"
	"github.com/atom91/bookings/internal/models"
	"github.com/atom91/bookings/internal/renderer"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/justinas/nosurf"
)

var app config.AppConfig
var pathToTemplates="./../../templates"
var session *scs.SessionManager
func getRoutes() http.Handler{
	gob.Register(models.Reservation{})
	app.InProduction=false
	session= scs.New()
	session.Lifetime=24*time.Hour
	session.Cookie.Persist=true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = false
	app.Session=session
	tc, err:=CreateTestTemplateCache()
	if err!=nil{
		log.Fatal("cannot create cache")
	}
	app.TemplateCache=tc
	app.UseCache=true
	repo:=NewRepo(&app)
	NewHandlers(repo)

	renderer.NewTemplates(&app)

	mux := chi.NewRouter()
	mux.Use(middleware.Recoverer)
	//mux.Use(NoSurf)
	mux.Use(SessionLoad)
	mux.Get("/", Repo.Home)
	mux.Get("/about", Repo.About)
	mux.Get("/room", Repo.Room)
	
	mux.Get("/apartment", Repo.Apartment)

	mux.Get("/reservation", Repo.Reservation)
	mux.Post("/reservation", Repo.PostReservation)
	mux.Get("/reservation-summary", Repo.ReservationSummary)
	

	mux.Get("/search-availability", Repo.SearchAvailability)
	mux.Post("/search-availability", Repo.PostAvailability)
	mux.Post("/search-availability-json", Repo.AvailabilityJson)
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))
	return mux
}
func WriteToConsole(next http.Handler)http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		fmt.Println("hit the page")
		next.ServeHTTP(w,r)
	})
}

func NoSurf(next http.Handler) http.Handler{
	csrfHandler:= nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path: "/",
		Secure: app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})
	return csrfHandler
}
func SessionLoad(next http.Handler) http.Handler{
	return session.LoadAndSave(next)
}
func CreateTestTemplateCache()(map[string]*template.Template,error){
	myCache:=map[string]*template.Template{}
	pages, err := filepath.Glob(fmt.Sprintf("%s/*.html",pathToTemplates))
	if err!=nil{
		return myCache,err
	}
	for _, page := range pages{
		name:=filepath.Base(page)
		ts, err :=template.New(name).ParseFiles(page)
		if err!=nil{
			return myCache,err
		}
		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl",pathToTemplates))
		if err!=nil{
			return myCache,err
		}
		if len(matches)>0{
			ts,err= ts.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl",pathToTemplates))
			if err!=nil{
				return myCache,err
			}
		}
		myCache[name]=ts

	}
	return myCache,nil

}