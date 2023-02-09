package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/atom91/bookings/internal/config"
	"github.com/atom91/bookings/internal/handlers"
	"github.com/atom91/bookings/internal/helpers"
	"github.com/atom91/bookings/internal/models"
	"github.com/atom91/bookings/internal/renderer"
	"github.com/atom91/bookings/internal/driver"
)

const portNumber=":8080"
var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

func main() {
	db,err:=run()
	if err!=nil{
		log.Fatal(err)
	}
	defer db.SQL.Close();
	defer close(app.MailChan)
	log.Println("Starting mail listener")
	listenForMail()
	
	fmt.Printf("Starting application on port number %s...",portNumber)
	srv:= &http.Server{
		Addr: portNumber,
		Handler: routes(&app) ,
	}
	err=srv.ListenAndServe();
	log.Fatal(err)
}
func run ()(*driver.DB, error){
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(map[string]int{})
	mailChan:= make(chan models.MailData)
	app.MailChan=mailChan
	app.InProduction=false
	infoLog=log.New(os.Stdout,"INFO: ",log.Ldate|log.Ltime)
	app.InfoLog=infoLog
	errorLog=log.New(os.Stdout,"ERROR: ",log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog=errorLog
	session= scs.New()
	session.Lifetime=24*time.Hour
	session.Cookie.Persist=true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction
	app.Session=session

	//connect to the database
	log.Println("Connecting to the database...")
	db,err:=driver.ConnectSql("host=localhost port=5432 dbname=bookings user=postgres password=Sifra123")
	if err!=nil{
		log.Fatal("Cannot connect do DB. Dying...")
	}
	log.Println("Connected to DB")
	tc, err:=renderer.CreateTemplateCache()
	if err!=nil{
		log.Fatal("cannot create cache")
		return nil,err
	}
	app.TemplateCache=tc
	app.UseCache=false
	repo:=handlers.NewRepo(&app,db)
	handlers.NewHandlers(repo)

	renderer.NewTemplates(&app)
	helpers.NewHelpers(&app)
	
	return db,nil
}
