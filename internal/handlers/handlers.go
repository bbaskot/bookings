package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/atom91/bookings/internal/config"
	"github.com/atom91/bookings/internal/models"
	"github.com/atom91/bookings/internal/renderer"
)

type Repository struct {
	App *config.AppConfig
}

var Repo *Repository

func NewRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
	}
}
func NewHandlers(r *Repository) {
	Repo = r
}
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	remoteIp := r.RemoteAddr
	m.App.Session.Put(r.Context(), "remote_ip", remoteIp)
	renderer.RenderTemplate(w, "home.html", &models.TemplateData{}, r)
}
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	renderer.RenderTemplate(w, "about.html", &models.TemplateData{}, r)
}
func (m *Repository) Room(w http.ResponseWriter, r *http.Request) {
	renderer.RenderTemplate(w, "room.html", &models.TemplateData{}, r)
}

type JSONData struct {
	OK      bool   `json:"ok"`
	Message string `json:"Available"`
}

func (m *Repository) AvailabilityJson(w http.ResponseWriter, r *http.Request) {
	resp := JSONData{
		OK:      true,
		Message: "Dostupno",
	}
	out, err := json.MarshalIndent(resp, "", "     ")
	if err != nil {
		log.Println(err)
	}
	w.Header().Set("Content-Type", "aplication/json")
	w.Write(out)

}
func (m *Repository) Apartment(w http.ResponseWriter, r *http.Request) {
	renderer.RenderTemplate(w, "apartment.html", &models.TemplateData{}, r)
}
func (m *Repository) SearchAvailability(w http.ResponseWriter, r *http.Request) {
	renderer.RenderTemplate(w, "search-availability.html", &models.TemplateData{}, r)
}
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start")
	end := r.Form.Get("end")
	w.Write([]byte(fmt.Sprintf("Pocetni datum je %s a zavrsni datum je %s ", start, end)))
}
