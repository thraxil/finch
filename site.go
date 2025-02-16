package main

import (
	"strconv"

	"github.com/gorilla/sessions"
)

type site struct {
	p                 *persistence
	BaseURL           string
	Store             sessions.Store
	ItemsPerPage      int
	AllowRegistration bool

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
	getChannelByIDChan       chan *getChannelByIDOp
	getPostChannelsChan      chan *getPostChannelsOp
	searchPostsChan          chan *searchPostsOp
}

func newSite(p *persistence, base string, store sessions.Store, ipp string, allowRegistration string) *site {
	i, err := strconv.Atoi(ipp)
	if err != nil {
		i = 50
	}
	allowReg := false
	if allowRegistration == "true" {
		allowReg = true
	}
	s := site{
		p:                 p,
		BaseURL:           base,
		Store:             store,
		ItemsPerPage:      i,
		AllowRegistration: allowReg,
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
		getChannelByIDChan:       make(chan *getChannelByIDOp),
		getPostChannelsChan:      make(chan *getPostChannelsOp),
		searchPostsChan:          make(chan *searchPostsOp),
	}
	go s.Run()
	return &s
}

/*
All database operations run through here to guarantee
that there is never more than one happening at a time.
*/
func (s *site) Run() {
	for {
		select {
		// reads first
		case op := <-s.getUserChan:
			u, err := s.p.GetUser(op.Username)
			op.Resp <- userResponse{User: u, Err: err}
		case op := <-s.getPostByUUIDChan:
			p, err := s.p.GetPostByUUID(op.UUID)
			op.Resp <- postResponse{Post: p, Err: err}
		case op := <-s.getUserChannelsChan:
			channels, err := s.p.GetUserChannels(op.User)
			op.Resp <- channelsResponse{Channels: channels, Err: err}
		case op := <-s.getAllPostsChan:
			posts, err := s.p.GetAllPosts(op.Limit, op.Offset)
			op.Resp <- postsResponse{Posts: posts, Err: err}
		case op := <-s.getAllPostsInChannelChan:
			posts, err := s.p.GetAllPostsInChannel(op.Channel, op.Limit, op.Offset)
			op.Resp <- postsResponse{Posts: posts, Err: err}
		case op := <-s.getAllUserPostsChan:
			posts, err := s.p.GetAllUserPosts(op.User, op.Limit, op.Offset)
			op.Resp <- postsResponse{Posts: posts, Err: err}
		case op := <-s.getChannelChan:
			channel, err := s.p.GetChannel(op.User, op.Slug)
			op.Resp <- channelResponse{Channel: channel, Err: err}
		case op := <-s.getChannelByIDChan:
			channel, err := s.p.GetChannelByID(op.ID)
			op.Resp <- channelResponse{Channel: channel, Err: err}
		case op := <-s.getPostChannelsChan:
			channels, err := s.p.GetPostChannels(op.Post)
			op.Resp <- channelsResponse{Channels: channels, Err: err}
		case op := <-s.searchPostsChan:
			posts, err := s.p.SearchPosts(op.Q, op.Limit, op.Offset)
			op.Resp <- postsResponse{Posts: posts, Err: err}

		// then writes
		case op := <-s.createUserChan:
			u, err := s.p.CreateUser(op.Username, op.Password)
			op.Resp <- userResponse{User: u, Err: err}
		case op := <-s.deleteChannelChan:
			err := s.p.DeleteChannel(op.Channel)
			op.Resp <- deleteChannelResponse{Err: err}
		case op := <-s.deletePostChan:
			err := s.p.DeletePost(op.Post)
			op.Resp <- deletePostResponse{Err: err}
		case op := <-s.addChannelsChan:
			channels, err := s.p.AddChannels(op.User, op.Names)
			op.Resp <- channelsResponse{Channels: channels, Err: err}
		case op := <-s.addPostChan:
			post, err := s.p.AddPost(op.User, op.Body, op.Channels)
			op.Resp <- postResponse{Post: post, Err: err}

		}
	}
}

type userResponse struct {
	User *user
	Err  error
}

type createUserOp struct {
	Username string
	Password string
	Resp     chan userResponse
}

func (s *site) CreateUser(username, password string) (*user, error) {
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

func (s *site) GetUser(username string) (*user, error) {
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
	Channel *channel
	Resp    chan deleteChannelResponse
}

func (s *site) DeleteChannel(c *channel) error {
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
	Post *post
	Resp chan deletePostResponse
}

func (s *site) DeletePost(c *post) error {
	r := make(chan deletePostResponse)
	op := &deletePostOp{Post: c, Resp: r}
	s.deletePostChan <- op
	ur := <-r
	return ur.Err
}

type channelsResponse struct {
	Channels []*channel
	Err      error
}

type addChannelsOp struct {
	User  user
	Names []string
	Resp  chan channelsResponse
}

func (s *site) AddChannels(u user, names []string) ([]*channel, error) {
	r := make(chan channelsResponse)
	op := &addChannelsOp{User: u, Names: names, Resp: r}
	s.addChannelsChan <- op
	ur := <-r
	return ur.Channels, ur.Err
}

type getPostChannelsOp struct {
	Post *post
	Resp chan channelsResponse
}

func (s *site) GetPostChannels(p *post) ([]*channel, error) {
	r := make(chan channelsResponse)
	op := &getPostChannelsOp{Post: p, Resp: r}
	s.getPostChannelsChan <- op
	ur := <-r
	return ur.Channels, ur.Err
}

type channelResponse struct {
	Channel *channel
	Err     error
}

type getChannelOp struct {
	User user
	Slug string
	Resp chan channelResponse
}

func (s *site) GetChannel(u user, slug string) (*channel, error) {
	r := make(chan channelResponse)
	op := &getChannelOp{User: u, Slug: slug, Resp: r}
	s.getChannelChan <- op
	ur := <-r
	return ur.Channel, ur.Err
}

type getChannelByIDOp struct {
	ID   int
	Resp chan channelResponse
}

func (s *site) GetChannelByID(id int) (*channel, error) {
	r := make(chan channelResponse)
	op := &getChannelByIDOp{ID: id, Resp: r}
	s.getChannelByIDChan <- op
	ur := <-r
	return ur.Channel, ur.Err
}

type getUserChannelsOp struct {
	User user
	Resp chan channelsResponse
}

func (s *site) GetUserChannels(u user) ([]*channel, error) {
	r := make(chan channelsResponse)
	op := &getUserChannelsOp{User: u, Resp: r}
	s.getUserChannelsChan <- op
	ur := <-r
	return ur.Channels, ur.Err
}

type postResponse struct {
	Post *post
	Err  error
}

type addPostOp struct {
	User     user
	Body     string
	Channels []*channel
	Resp     chan postResponse
}

func (s *site) AddPost(u user, body string, channels []*channel) (*post, error) {
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

func (s *site) GetPostByUUID(uu string) (*post, error) {
	r := make(chan postResponse)
	op := &getPostByUUIDOp{UUID: uu, Resp: r}
	s.getPostByUUIDChan <- op
	ur := <-r
	return ur.Post, ur.Err
}

type postsResponse struct {
	Posts []*post
	Err   error
}

type getAllPostsOp struct {
	Limit  int
	Offset int
	Resp   chan postsResponse
}

func (s *site) GetAllPosts(limit, offset int) ([]*post, error) {
	r := make(chan postsResponse)
	op := &getAllPostsOp{Limit: limit, Offset: offset, Resp: r}
	s.getAllPostsChan <- op
	ur := <-r
	return ur.Posts, ur.Err
}

type getAllPostsInChannelOp struct {
	Channel channel
	Limit   int
	Offset  int
	Resp    chan postsResponse
}

func (s *site) GetAllPostsInChannel(c channel, limit, offset int) ([]*post, error) {
	r := make(chan postsResponse)
	op := &getAllPostsInChannelOp{Channel: c, Limit: limit, Offset: offset, Resp: r}
	s.getAllPostsInChannelChan <- op
	ur := <-r
	return ur.Posts, ur.Err
}

type searchPostsOp struct {
	Q      string
	Limit  int
	Offset int
	Resp   chan postsResponse
}

func (s *site) SearchPosts(q string, limit, offset int) ([]*post, error) {
	r := make(chan postsResponse)
	op := &searchPostsOp{Q: q, Limit: limit, Offset: offset, Resp: r}
	s.searchPostsChan <- op
	ur := <-r
	return ur.Posts, ur.Err
}

type getAllUserPostsOp struct {
	User   *user
	Limit  int
	Offset int
	Resp   chan postsResponse
}

func (s *site) GetAllUserPosts(u *user, limit, offset int) ([]*post, error) {
	r := make(chan postsResponse)
	op := &getAllUserPostsOp{User: u, Limit: limit, Offset: offset, Resp: r}
	s.getAllUserPostsChan <- op
	ur := <-r
	return ur.Posts, ur.Err
}
