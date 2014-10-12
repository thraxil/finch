package main

import (
	"github.com/gorilla/sessions"
)

type Site struct {
	P       *Persistence
	BaseUrl string
	Store   sessions.Store
}
