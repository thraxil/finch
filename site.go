package main

import (
	"github.com/gorilla/sessions"
)

type userResponse struct {
	User *User
	Err  error
}

type createUserOp struct {
	Username string
	Password string
	Resp     chan userResponse
}

type Site struct {
	P       *Persistence
	BaseUrl string
	Store   sessions.Store

	createUserChan chan *createUserOp
}

func NewSite(p *Persistence, base string, store sessions.Store) *Site {
	s := Site{
		P:              p,
		BaseUrl:        base,
		Store:          store,
		createUserChan: make(chan *createUserOp),
	}
	go s.Run()
	return &s
}

func (s *Site) Run() {
	for {
		select {
		case cuo := <-s.createUserChan:
			u, err := s.P.CreateUser(cuo.Username, cuo.Password)
			cuo.Resp <- userResponse{User: u, Err: err}
		}
	}
}

func (s *Site) CreateUser(username, password string) (*User, error) {
	r := make(chan userResponse)
	cuo := &createUserOp{Username: username, Password: password, Resp: r}
	s.createUserChan <- cuo
	ur := <-r
	return ur.User, ur.Err
}
