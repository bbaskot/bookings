package handlers

import (
	"net/http"

	"github.com/atom91/bookings/pkg/config"
	"github.com/atom91/bookings/pkg/models"
	"github.com/atom91/bookings/pkg/renderer"
)


type Repository struct{
	App *config.AppConfig
}
var Repo *Repository

func NewRepo(a *config.AppConfig) *Repository{
	return &Repository{
		App: a,
	}
}
func NewHandlers(r *Repository){
	Repo = r;
}
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	remoteIp:= r.RemoteAddr
	m.App.Session.Put(r.Context(),"remote_ip",remoteIp)
	renderer.RenderTemplate(w, "home.html", &models.TemplateData{})
}
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	//performe some logic
	stringMap:=make(map[string]string)
	stringMap["test"]="Hello again"
	remoteIp:=m.App.Session.GetString(r.Context(),"remote_ip")
	stringMap["remote_ip"]=remoteIp
	//send data to the template
	renderer.RenderTemplate(w, "about.html",&models.TemplateData{
		StringMap: stringMap,
	})
}
