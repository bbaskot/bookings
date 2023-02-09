package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestNoSurf(t *testing.T) {
	var myH myHandler
	h := NoSurf(&myH)
	switch v := h.(type) {
	case http.Handler:
		//do nothing
	default:
		t.Error(fmt.Sprintf("Type is not http.Handler but %T", v))
	}
}

func TestSessionLoad(t *testing.T) {
	var myH myHandler
	h := SessionLoad(&myH)
	switch v := h.(type) {
	case http.Handler:
		//do nothing
	default:
		t.Error(fmt.Sprintf("Type is not http.Handler but %T", v))
	}
}
func TestWriteToConsole(t *testing.T) {
	var myH myHandler
	h := WriteToConsole(&myH)
	switch v := h.(type) {
	case http.HandlerFunc:
		//do nothing
	default:
		t.Error(fmt.Sprintf("Type is not http.HandlerFunc, type is %T", v))
	}
}
