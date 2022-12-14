package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/atom91/bookings/internal/config"
	"github.com/atom91/bookings/internal/handlers"
	"github.com/atom91/bookings/internal/renderer"
)

const portNumber=":8080"
var app config.AppConfig
var session *scs.SessionManager

func main() {
	
	app.InProduction=false
	session= scs.New()
	session.Lifetime=24*time.Hour
	session.Cookie.Persist=true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction
	app.Session=session
	tc, err:=renderer.CreateTemplateCache()
	if err!=nil{
		log.Fatal("cannot create cache")
	}
	app.TemplateCache=tc
	app.UseCache=false
	repo:=handlers.NewRepo(&app)
	handlers.NewHandlers(repo)

	renderer.NewTemplates(&app)
	
	fmt.Printf("Starting application on port number %s...",portNumber)
	srv:= &http.Server{
		Addr: portNumber,
		Handler: routes(&app) ,
	}
	err=srv.ListenAndServe();
	log.Fatal(err)
}
