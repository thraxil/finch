package main

import (
	"github.com/gorilla/sessions"
)

type Site struct {
	P       *Persistence
	BaseUrl string
	Store   sessions.Store

	createUserChan    chan *createUserOp
	deleteChannelChan chan *deleteChannelOp
	deletePostChan    chan *deletePostOp
	addChannelsChan   chan *addChannelsOp
	addPostChan       chan *addPostOp
}

func NewSite(p *Persistence, base string, store sessions.Store) *Site {
	s := Site{
		P:                 p,
		BaseUrl:           base,
		Store:             store,
		createUserChan:    make(chan *createUserOp),
		deleteChannelChan: make(chan *deleteChannelOp),
		deletePostChan:    make(chan *deletePostOp),
		addChannelsChan:   make(chan *addChannelsOp),
		addPostChan:       make(chan *addPostOp),
	}
	go s.Run()
	return &s
}

func (s *Site) Run() {
	for {
		select {
		case op := <-s.createUserChan:
			u, err := s.P.CreateUser(op.Username, op.Password)
			op.Resp <- userResponse{User: u, Err: err}
		case op := <-s.deleteChannelChan:
			err := s.P.DeleteChannel(op.Channel)
			op.Resp <- deleteChannelResponse{Err: err}
		case op := <-s.deletePostChan:
			err := s.P.DeletePost(op.Post)
			op.Resp <- deletePostResponse{Err: err}
		case op := <-s.addChannelsChan:
			channels, err := s.P.AddChannels(op.User, op.Names)
			op.Resp <- addChannelsResponse{Channels: channels, Err: err}
		case op := <-s.addPostChan:
			post, err := s.P.AddPost(op.User, op.Body, op.Channels)
			op.Resp <- addPostResponse{Post: post, Err: err}

		}
	}
}

type userResponse struct {
	User *User
	Err  error
}

type createUserOp struct {
	Username string
	Password string
	Resp     chan userResponse
}

func (s *Site) CreateUser(username, password string) (*User, error) {
	r := make(chan userResponse)
	cuo := &createUserOp{Username: username, Password: password, Resp: r}
	s.createUserChan <- cuo
	ur := <-r
	return ur.User, ur.Err
}

type deleteChannelResponse struct {
	Err error
}

type deleteChannelOp struct {
	Channel *Channel
	Resp    chan deleteChannelResponse
}

func (s *Site) DeleteChannel(c *Channel) error {
	r := make(chan deleteChannelResponse)
	op := &deleteChannelOp{Channel: c, Resp: r}
	s.deleteChannelChan <- op
	ur := <-r
	return ur.Err
}

type deletePostResponse struct {
	Err error
}

type deletePostOp struct {
	Post *Post
	Resp chan deletePostResponse
}

func (s *Site) DeletePost(c *Post) error {
	r := make(chan deletePostResponse)
	op := &deletePostOp{Post: c, Resp: r}
	s.deletePostChan <- op
	ur := <-r
	return ur.Err
}

type addChannelsResponse struct {
	Channels []*Channel
	Err      error
}

type addChannelsOp struct {
	User  User
	Names []string
	Resp  chan addChannelsResponse
}

func (s *Site) AddChannels(u User, names []string) ([]*Channel, error) {
	r := make(chan addChannelsResponse)
	op := &addChannelsOp{User: u, Names: names, Resp: r}
	s.addChannelsChan <- op
	ur := <-r
	return ur.Channels, ur.Err
}

type addPostResponse struct {
	Post *Post
	Err  error
}

type addPostOp struct {
	User     User
	Body     string
	Channels []*Channel
	Resp     chan addPostResponse
}

func (s *Site) AddPost(u User, body string, channels []*Channel) (*Post, error) {
	r := make(chan addPostResponse)
	op := &addPostOp{User: u, Body: body, Channels: channels, Resp: r}
	s.addPostChan <- op
	ur := <-r
	return ur.Post, ur.Err
}
