package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/atom91/bookings/internal/config"
	"github.com/atom91/bookings/internal/driver"
	"github.com/atom91/bookings/internal/forms"
	"github.com/atom91/bookings/internal/helpers"
	"github.com/atom91/bookings/internal/models"
	"github.com/atom91/bookings/internal/renderer"
	"github.com/atom91/bookings/internal/repository"
	"github.com/atom91/bookings/internal/repository/dbrepo"
	"github.com/go-chi/chi/v5"
)

type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

var Repo *Repository

func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
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

func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})
	res := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	data["reservation"] = res
	stringMap := make(map[string]string)
	stringMap["start_date"] = res.StartDate.Format("2006-01-02")
	stringMap["end_date"] = res.EndDate.Format("2006-01-02")
	m.App.Session.Put(r.Context(), "reservation", res)
	renderer.RenderTemplate(w, "make-reservation.html", &models.TemplateData{
		Form:      *forms.New(nil),
		Data:      data,
		StringMap: stringMap,
	}, r)
}
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	sd := r.Form.Get("start_date")
	ed := r.Form.Get("end_date")
	log.Printf("datumi su %s %s\n", sd, ed)
	layout := "2006-01-02"
	startDate, err := time.Parse(layout, sd)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	endDate, err := time.Parse(layout, ed)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Email:     r.Form.Get("email"),
		Phone:     r.Form.Get("phone"),
		StartDate: startDate,
		EndDate:   endDate,
		RoomId:    roomID,
	}
	form := forms.New(r.PostForm)
	form.Required("first_name", "last_name", "email", "phone")
	form.MinLength("first_name", 2, *r)
	form.IsEmail("email")
	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation
		renderer.RenderTemplate(w, "make-reservation.html", &models.TemplateData{
			Form: *form,
			Data: data,
		}, r)
		return
	}
	newReservationId, err := m.DB.InsertReservation(reservation)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	restriction := models.RoomRestriction{
		StartDate:     startDate,
		EndDate:       endDate,
		ReservationId: newReservationId,
		RoomId:        roomID,
		RestrictionId: 1,
	}
	err = m.DB.InsertRoomRestriction(&restriction)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	//send notifications
	htmlMessage := fmt.Sprintf(`
	<strong> Potvrda rezervacija</strong><br>
	Postovani %s,<br>
	Rezervisali ste odmor u nasoj sobi %s<br>
	od %s<br>
	do %s<br>
	Radujemo se vasem dolasku!
	`, reservation.FirstName, reservation.Room.RoomName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))
	msg := models.MailData{
		To:       reservation.Email,
		From:     "gogi@gogic.com",
		Subject:  "Potvrda rezervacije",
		Content:  htmlMessage,
		Template: "basic.html",
	}

	m.App.MailChan <- msg

	m.App.Session.Put(r.Context(), "reservation", reservation)
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)

}

func (m *Repository) Room(w http.ResponseWriter, r *http.Request) {
	renderer.RenderTemplate(w, "room.html", &models.TemplateData{}, r)
}

func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.ErrorLog.Println("Couldn't get reservation data from session")
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	room := m.DB.SearchRoomById(reservation.RoomId)
	reservation.Room = room
	log.Printf("Rezervisana soba je %s\n", reservation.Room.RoomName)
	stringMap := make(map[string]string)
	stringMap["start_date"] = reservation.StartDate.Format("2006-01-02")
	stringMap["end_date"] = reservation.EndDate.Format("2006-01-02")
	m.App.Session.Remove(r.Context(), "reservation")
	data := make(map[string]interface{})
	data["reservation"] = reservation
	renderer.RenderTemplate(w, "reservation-summary.html", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	}, r)

}

type JSONData struct {
	OK        bool   `json:"ok"`
	Message   string `json:"Available"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	RoomId    string `json:"room_id"`
}

func (m *Repository) AvailabilityJson(w http.ResponseWriter, r *http.Request) {

	sd := r.Form.Get("start")
	ed := r.Form.Get("end")
	layout := "2006-02-02"

	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)
	roomId, _ := strconv.Atoi(r.Form.Get("room_id"))
	available, _ := m.DB.SearchAvailabilityByDatesByRoomId(startDate, endDate, roomId)
	log.Printf("First date is %s, second date is %s, availability is %b\n", sd, ed, available)
	resp := JSONData{
		OK:        available,
		Message:   "",
		StartDate: sd,
		EndDate:   ed,
		RoomId:    r.Form.Get("room_id"),
	}
	out, err := json.MarshalIndent(resp, "", "     ")
	if err != nil {
		helpers.ServerError(w, err)
		return
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
	layout := "2006-01-02"
	startDate, err := time.Parse(layout, start)
	if err != nil {
		helpers.ServerError(w, err)
	}
	endDate, err := time.Parse(layout, end)
	if err != nil {
		helpers.ServerError(w, err)
	}
	var rooms []models.Room
	rooms, err = m.DB.SearchAvailabilityOfAllRooms(startDate, endDate)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	if len(rooms) == 0 {
		m.App.Session.Put(r.Context(), "error", "No availability")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["rooms"] = rooms

	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}
	m.App.Session.Put(r.Context(), "reservation", res)
	renderer.RenderTemplate(w, "choose-room.html", &models.TemplateData{
		Data: data,
	}, r)

}
func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {

	roomId, err := strconv.Atoi(chi.URLParamFromCtx(r.Context(), "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	room := m.DB.SearchRoomById(roomId)

	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	res.Room = room
	res.RoomId = roomId
	if !ok {
		helpers.ServerError(w, err)
		return
	}

	m.App.Session.Put(r.Context(), "reservation", res)
	http.Redirect(w, r, "/reservation", http.StatusSeeOther)

}
func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {

	roomId, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	layout := "2006-01-02"
	startDate, err := time.Parse(layout, r.URL.Query().Get("s"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	endDate, err := time.Parse(layout, r.URL.Query().Get("e"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	var reservation models.Reservation
	reservation.RoomId = roomId
	room := m.DB.SearchRoomById(roomId)
	reservation.Room = room
	reservation.StartDate = startDate
	reservation.EndDate = endDate
	m.App.Session.Put(r.Context(), "reservation", reservation)
	http.Redirect(w, r, "/reservation", http.StatusSeeOther)
}
func (m *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	renderer.RenderTemplate(w, "login.html", &models.TemplateData{
		Form: *forms.New(nil),
	}, r)

}
func (m *Repository) PostShowLogin(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
	}
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")
	if !form.Valid() {
		renderer.RenderTemplate(w, "login.html", &models.TemplateData{
			Form: *form,
		}, r)
		return
	}
	id, _, err := m.DB.Authenticate(email, password)
	if err != nil {

		m.App.Session.Put(r.Context(), "error", "Invalid login credentials")
		renderer.RenderTemplate(w, "login.html", &models.TemplateData{
			Form: *form,
		}, r)
		return
	}
	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "flash", "Logged in succesfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (m *Repository) UserLogout(w http.ResponseWriter, r *http.Request) {
	m.App.Session.Destroy(r.Context())
	m.App.Session.RenewToken(r.Context())
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
func (m *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	renderer.RenderTemplate(w, "admin-dashboard.html", &models.TemplateData{}, r)
}
func (m *Repository) AdminNewReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.NewReservations()
	if err != nil {
		log.Println(err)
		return
	}
	data := make(map[string]interface{})
	data["reservations"] = reservations

	renderer.RenderTemplate(w, "admin-new-reservations.html", &models.TemplateData{
		Data: data,
	}, r)
}
func (m *Repository) AdminAllReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.AllReservations()
	if err != nil {
		log.Println(err)
		return
	}
	data := make(map[string]interface{})
	data["reservations"] = reservations

	renderer.RenderTemplate(w, "admin-all-reservations.html", &models.TemplateData{
		Data: data,
	}, r)
}
func (m *Repository) AdminShowReservation(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	src := exploded[3]
	stringMap := make(map[string]string)
	stringMap["src"] = src
	res, err := m.DB.GetReservationById(id)
	if err != nil {
		helpers.ServerError(w, err)
	}
	stringMap["start_date"] = res.StartDate.Format("2006-01-02")
	stringMap["end_date"] = res.EndDate.Format("2006-01-02")
	stringMap["room"] = res.Room.RoomName
	data := make(map[string]interface{})
	data["reservation"] = res
	renderer.RenderTemplate(w, "admin-reservations-show.html", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		Form:      *forms.New(nil),
	}, r)
}
func (m *Repository) AdminPostShowReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	src := exploded[3]
	stringMap := make(map[string]string)
	stringMap["src"] = src
	res, err := m.DB.GetReservationById(id)
	if err != nil {
		helpers.ServerError(w, err)
	}
	res.FirstName = r.Form.Get("first_name")
	res.LastName = r.Form.Get("last_name")
	res.Email = r.Form.Get("email")
	res.Phone = r.Form.Get("phone")
	err = m.DB.UpdateReservation(res)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	m.App.Session.Put(r.Context(), "flash", "Changes saved")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)

}
func (m *Repository) AdminProcessReservation(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")
	_ = m.DB.UpdateProcessed(id, 1)
	m.App.Session.Put(r.Context(), "flash", "Reservation marked as processed")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)

}
func (m *Repository) AdminDeleteReservation(w http.ResponseWriter, r *http.Request) {
	log.Println("ovdje")
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")
	_ = m.DB.DeleteReservation(id)
	m.App.Session.Put(r.Context(), "flash", "Reservation deleted")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)

}
func (m *Repository) AdminReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	now :=time.Now()
	if r.URL.Query().Get("y") != "" {
		year,_:=strconv.Atoi(r.URL.Query().Get("y"))
		month,_:=strconv.Atoi(r.URL.Query().Get("m"))
		now = time.Date(year, time.Month(month),1,0,0,0,0,time.UTC)
	}
	next := now.AddDate(0,1,0)
	last:= now.AddDate(0,-1,0)
	nextMonth:= next.Format("01")
	nextMonthYear:=next.Format("2006")
	lastMonth:=last.Format("01")
	lastMonthYear:=last.Format("2006")
	stringMap:=make(map[string]string)
	stringMap["next_month"]=nextMonth
	stringMap["next_month_year"]=nextMonthYear
	stringMap["last_month"]=lastMonth
	stringMap["last_month_year"]=lastMonthYear
	stringMap["this_month"]=now.Format("01")
	stringMap["this_month_year"]=now.Format("2006")

	curentYear, currentMonth,_:= now.Date()
	currentLoaction := now.Location()
	firstOfMonth:= time.Date(curentYear,currentMonth,1,0,0,0,0,currentLoaction)
	lastOfMonth:=firstOfMonth.AddDate(0,1,-1)
	intMap:= make(map[string]int)
	data:= make(map[string]interface{})
	intMap["days_in_month"]= lastOfMonth.Day()
	var allDays []int
	var i int
	for i=1;i<=intMap["days_in_month"];i++{
		allDays=append(allDays, i)
	}
	data["all_days"]=allDays
	rooms, err := m.DB.AllRooms()
	if err!=nil{
		helpers.ServerError(w,err)
		return
	}
	data["rooms"]=rooms
	for _,x:=range rooms{
		reservationMap:=make(map[string]int)
		blockMap:= make(map[string]int)
		
		
		for d:=firstOfMonth;d.After(lastOfMonth)==false;d=d.AddDate(0,0,1){
			
			reservationMap[d.Format("2006-01-2")]=0
			blockMap[d.Format("2006-01-2")]=0
			
		} 
		restrictions,err:=m.DB.GetRestrictionsForRoomByDate(x.ID,firstOfMonth,lastOfMonth)
		if err!=nil{
			helpers.ServerError(w,err)
			return
		}
		for _, y:=range restrictions{
			if y.ReservationId>0{
				for d:=y.StartDate;d.After(y.EndDate)==false;d=d.AddDate(0,0,1){
					reservationMap[d.Format("2006-01-2")]=y.ReservationId
				}
			}else{
				
					blockMap[y.StartDate.Format("2006-01-2")]=y.ID
				
			}
			
		}
		data[fmt.Sprintf("reservation_map_%d",x.ID)]=reservationMap
		data[fmt.Sprintf("block_map_%d",x.ID)]=blockMap
		m.App.Session.Put(r.Context(),fmt.Sprintf("block_map_%d",x.ID), blockMap)
		
	}
	
	renderer.RenderTemplate(w, "admin-reservations-calendar.html", &models.TemplateData{
		StringMap: stringMap,
		IntMap: intMap,
		Data: data,
	}, r)
}
func (m *Repository)AdminPostReservationsCalendar(w http.ResponseWriter, r *http.Request){
	err:=r.ParseForm()
	if err!=nil{
		helpers.ServerError(w,err)
		return
	}
	year,_:=strconv.Atoi(r.Form.Get("y"))
	month,_:=strconv.Atoi(r.Form.Get("m"))

	rooms, err:= m.DB.AllRooms()
	if err!=nil{
		helpers.ServerError(w,err)
		return
	}
	form := forms.New(r.PostForm)
	for _,x:= range rooms{
		curMap:=m.App.Session.Get(r.Context(), fmt.Sprintf("block_map_%d",x.ID)).(map[string]int)
		for name, value :=range curMap{
			if val, ok :=curMap[name]; ok{
				if val > 0{
					if !form.Has(fmt.Sprintf("remove_block_%d_%s",x.ID,name),*r){
						m.DB.DeleteBlocks(value)
					}
				}
			}
		}
	}
	for name,_:=range r.PostForm{
		if strings.HasPrefix(name,"add_block"){
			exploded:= strings.Split(name,"_")
			roomId,_:=strconv.Atoi(exploded[2])
			 layout:="2006-01-2"
			 date,err:=time.Parse(layout,exploded[3])
			 if err!=nil{
				helpers.ServerError(w,err)
			 }
			 m.DB.InsertBlock(date,roomId)
			//log.Println("would insert block for room id ",roomId," for date ",exploded[3])
		}
	}

	m.App.Session.Put(r.Context(),"flash","Changes Saved")
	http.Redirect(w,r,fmt.Sprintf("/admin/reservation-calendar?y=%d&m=%d",year,month),http.StatusSeeOther)


}