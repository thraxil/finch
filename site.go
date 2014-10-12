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
	getUserChan              chan *getUserOp
	getPostByUUIDChan        chan *getPostByUUIDOp
	getUserChannelsChan      chan *getUserChannelsOp
	getAllPostsChan          chan *getAllPostsOp
	getAllPostsInChannelChan chan *getAllPostsInChannelOp
	getAllUserPostsChan      chan *getAllUserPostsOp
	getChannelChan           chan *getChannelOp
	getChannelByIdChan       chan *getChannelByIdOp
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

		getUserChan:              make(chan *getUserOp),
		getPostByUUIDChan:        make(chan *getPostByUUIDOp),
		getUserChannelsChan:      make(chan *getUserChannelsOp),
		getAllPostsChan:          make(chan *getAllPostsOp),
		getAllPostsInChannelChan: make(chan *getAllPostsInChannelOp),
		getAllUserPostsChan:      make(chan *getAllUserPostsOp),
		getChannelChan:           make(chan *getChannelOp),
		getChannelByIdChan:       make(chan *getChannelByIdOp),
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
		case op := <-s.getUserChannelsChan:
			channels, err := s.P.GetUserChannels(op.User)
			op.Resp <- channelsResponse{Channels: channels, Err: err}
		case op := <-s.getAllPostsChan:
			posts, err := s.P.GetAllPosts(op.Limit, op.Offset)
			op.Resp <- postsResponse{Posts: posts, Err: err}
		case op := <-s.getAllPostsInChannelChan:
			posts, err := s.P.GetAllPostsInChannel(op.Channel, op.Limit, op.Offset)
			op.Resp <- postsResponse{Posts: posts, Err: err}
		case op := <-s.getAllUserPostsChan:
			posts, err := s.P.GetAllUserPosts(op.User, op.Limit, op.Offset)
			op.Resp <- postsResponse{Posts: posts, Err: err}
		case op := <-s.getChannelChan:
			channel, err := s.P.GetChannel(op.User, op.Slug)
			op.Resp <- channelResponse{Channel: channel, Err: err}
		case op := <-s.getChannelByIdChan:
			channel, err := s.P.GetChannelById(op.Id)
			op.Resp <- channelResponse{Channel: channel, Err: err}

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
			op.Resp <- channelsResponse{Channels: channels, Err: err}
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

type channelsResponse struct {
	Channels []*Channel
	Err      error
}

type addChannelsOp struct {
	User  User
	Names []string
	Resp  chan channelsResponse
}

func (s *Site) AddChannels(u User, names []string) ([]*Channel, error) {
	r := make(chan channelsResponse)
	op := &addChannelsOp{User: u, Names: names, Resp: r}
	s.addChannelsChan <- op
	ur := <-r
	return ur.Channels, ur.Err
}

type channelResponse struct {
	Channel *Channel
	Err     error
}

type getChannelOp struct {
	User User
	Slug string
	Resp chan channelResponse
}

func (s *Site) GetChannel(u User, slug string) (*Channel, error) {
	r := make(chan channelResponse)
	op := &getChannelOp{User: u, Slug: slug, Resp: r}
	s.getChannelChan <- op
	ur := <-r
	return ur.Channel, ur.Err
}

type getChannelByIdOp struct {
	Id   int
	Resp chan channelResponse
}

func (s *Site) GetChannelById(id int) (*Channel, error) {
	r := make(chan channelResponse)
	op := &getChannelByIdOp{Id: id, Resp: r}
	s.getChannelByIdChan <- op
	ur := <-r
	return ur.Channel, ur.Err
}

type getUserChannelsOp struct {
	User User
	Resp chan channelsResponse
}

func (s *Site) GetUserChannels(u User) ([]*Channel, error) {
	r := make(chan channelsResponse)
	op := &getUserChannelsOp{User: u, Resp: r}
	s.getUserChannelsChan <- op
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

type postsResponse struct {
	Posts []*Post
	Err   error
}

type getAllPostsOp struct {
	Limit  int
	Offset int
	Resp   chan postsResponse
}

func (s *Site) GetAllPosts(limit, offset int) ([]*Post, error) {
	r := make(chan postsResponse)
	op := &getAllPostsOp{Limit: limit, Offset: offset, Resp: r}
	s.getAllPostsChan <- op
	ur := <-r
	return ur.Posts, ur.Err
}

type getAllPostsInChannelOp struct {
	Channel Channel
	Limit   int
	Offset  int
	Resp    chan postsResponse
}

func (s *Site) GetAllPostsInChannel(c Channel, limit, offset int) ([]*Post, error) {
	r := make(chan postsResponse)
	op := &getAllPostsInChannelOp{Channel: c, Limit: limit, Offset: offset, Resp: r}
	s.getAllPostsInChannelChan <- op
	ur := <-r
	return ur.Posts, ur.Err
}

type getAllUserPostsOp struct {
	User   *User
	Limit  int
	Offset int
	Resp   chan postsResponse
}

func (s *Site) GetAllUserPosts(u *User, limit, offset int) ([]*Post, error) {
	r := make(chan postsResponse)
	op := &getAllUserPostsOp{User: u, Limit: limit, Offset: offset, Resp: r}
	s.getAllUserPostsChan <- op
	ur := <-r
	return ur.Posts, ur.Err
}
