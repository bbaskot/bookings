package helpers

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/atom91/bookings/internal/config"
)

var app *config.AppConfig

func NewHelpers(a *config.AppConfig){
	app=a

}
func ClientError(w http.ResponseWriter, status int){
	app.InfoLog.Printf("Client error with code %d",status)
	http.Error(w,http.StatusText(status),status)
}
func ServerError(w http.ResponseWriter,err error){
	trace:=fmt.Sprintf("%s\n%s",err.Error(),debug.Stack())
	app.ErrorLog.Printf("%s",trace)
	http.Error(w,http.StatusText(http.StatusInternalServerError),http.StatusInternalServerError)
}
func IsAuthenticated(r *http.Request)bool{
	return app.Session.Exists(r.Context(),"user_id")
}