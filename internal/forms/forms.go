package forms

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/asaskevich/govalidator"
)

type Form struct {
	url.Values
	Errors errors
}

func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}
func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "This field is required")
		}
	}
}
func (f *Form) Has(field string, r http.Request) bool {
	x := r.Form.Get(field)
	if x == "" {
		return false
	}
	return true
}
func (f *Form) MinLength(field string, length int, r http.Request) bool {
	value := r.Form.Get(field)
	if len(value) < length {
		f.Errors.Add(field, fmt.Sprintf("This field must be at least %d characters long.", length))
		return false
	}
	return true
}
func (f *Form) IsEmail(field string) bool {
	if !govalidator.IsExistingEmail(f.Get(field)) {
		f.Errors.Add(field, "Wrong email format")
		return false
	}
	return true
}
