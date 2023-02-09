package renderer

import (
	"net/http"
	"testing"

	"github.com/atom91/bookings/internal/models"
)

func TestAddDefaultData(t *testing.T) {
	var td models.TemplateData
	r,err:=getSession();
	if err!=nil{
		t.Error(err)
	}
	session.Put(r.Context(),"flash","123")
	result:= addDefaultData(&td,r)
	if result.Flash!="123"{
		t.Error("Value 123 not found in flash")
	}

}
func getSession()(*http.Request, error){
	r, err :=http.NewRequest("GET","/some-url",nil)
	if err!=nil{
		return nil, err
	}
	ctx:=r.Context()
	ctx,_=session.Load(ctx,r.Header.Get("X-Session"))
	r= r.WithContext(ctx)
	return r,err
}
/*
func TestRenderTemplate(t *testing.T){
	pathToTemplates="./../../templates"
	tc,err:=CreateTemplateCache()
	if err!=nil{
		t.Error(err)
	}
	app.TemplateCache=tc
	r, err:=getSession()
	if err!=nil{
		t.Error(err)
	}
	var ww myWriter
	err =RenderTemplate(&ww,"home.html",&models.TemplateData{},r)
	if err!=nil{
		t.Error("Error writing tmplt to browser")
	}
	err =RenderTemplate(&ww,"non-existant.html",&models.TemplateData{},r)
	if err==nil{
		t.Error("Rendered template that did not exist")
	}

}*/
