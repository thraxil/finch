package main

import (
	"github.com/gorilla/sessions"
)

type Site struct {
	P       *Persistence
	BaseUrl string
	Store   sessions.Store

	// write operation channels
	createUserChan    chan *createUserOp
	deleteChannelChan chan *deleteChannelOp
	deletePostChan    chan *deletePostOp
	addChannelsChan   chan *addChannelsOp
	addPostChan       chan *addPostOp

	// read operation channels
	getUserChan       chan *getUserOp
	getPostByUUIDChan chan *getPostByUUIDOp
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

		getUserChan:       make(chan *getUserOp),
		getPostByUUIDChan: make(chan *getPostByUUIDOp),
	}
	go s.Run()
	return &s
}

/*
All database operations run through here to guarantee
that there is never more than one happening at a time.
*/
func (s *Site) Run() {
	for {
		select {
		// reads first
		case op := <-s.getUserChan:
			u, err := s.P.GetUser(op.Username)
			op.Resp <- userResponse{User: u, Err: err}
		case op := <-s.getPostByUUIDChan:
			p, err := s.P.GetPostByUUID(op.UUID)
			op.Resp <- postResponse{Post: p, Err: err}

		// then writes
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
			op.Resp <- postResponse{Post: post, Err: err}

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

type getUserOp struct {
	Username string
	Resp     chan userResponse
}

func (s *Site) GetUser(username string) (*User, error) {
	r := make(chan userResponse)
	op := &getUserOp{Username: username, Resp: r}
	s.getUserChan <- op
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

type postResponse struct {
	Post *Post
	Err  error
}

type addPostOp struct {
	User     User
	Body     string
	Channels []*Channel
	Resp     chan postResponse
}

func (s *Site) AddPost(u User, body string, channels []*Channel) (*Post, error) {
	r := make(chan postResponse)
	op := &addPostOp{User: u, Body: body, Channels: channels, Resp: r}
	s.addPostChan <- op
	ur := <-r
	return ur.Post, ur.Err
}

type getPostByUUIDOp struct {
	UUID string
	Resp chan postResponse
}

func (s *Site) GetPostByUUID(uu string) (*Post, error) {
	r := make(chan postResponse)
	op := &getPostByUUIDOp{UUID: uu, Resp: r}
	s.getPostByUUIDChan <- op
	ur := <-r
	return ur.Post, ur.Err
}
