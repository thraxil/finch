package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/gorilla/feeds"
	"github.com/russross/blackfriday"
)

type siteResponse struct {
	Username string
}

func (s *siteResponse) SetUsername(username string) {
	s.Username = username
}

func (s siteResponse) GetUsername() string {
	return s.Username
}

type sr interface {
	SetUsername(string)
	GetUsername() string
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	// just ignore this crap
}

type context struct {
	Site *site
	User *user
}

func (c *context) Populate(r *http.Request) {
	sess, _ := c.Site.Store.Get(r, "finch")
	username, found := sess.Values["user"]
	if found && username != "" {
		user, err := c.Site.GetUser(username.(string))
		if err == nil {
			c.User = user
		}
	}
}

func (c context) PopulateResponse(sr sr) {
	if c.User != nil {
		sr.SetUsername(c.User.Username)
	}
}

type indexResponse struct {
	Posts []*post
	siteResponse
}

func indexHandler(w http.ResponseWriter, r *http.Request, s *site) {
	ctx := context{Site: s}
	ctx.Populate(r)
	ir := indexResponse{}
	ctx.PopulateResponse(&ir)
	posts, err := s.GetAllPosts(s.ItemsPerPage, 0)
	ir.Posts = posts
	if err != nil {
		log.Println(err)
		fmt.Fprintf(w, "error getting posts")
		return
	}
	tmpl := getTemplate("index.html")
	tmpl.Execute(w, ir)
}

type searchResponse struct {
	Posts []*post
	Q     string
	siteResponse
}

func searchHandler(w http.ResponseWriter, r *http.Request, s *site) {
	ctx := context{Site: s}
	ctx.Populate(r)
	q := r.FormValue("q")
	sr := searchResponse{Q: q}
	ctx.PopulateResponse(&sr)
	posts, err := s.SearchPosts(q, s.ItemsPerPage, 0)
	if err != nil {
		http.Error(w, "search broke", 500)
		return
	}
	sr.Posts = posts
	tmpl := getTemplate("search.html")
	tmpl.Execute(w, sr)
}

type addResponse struct {
	Channels []*channel
	Body     string
	siteResponse
}

func postHandler(w http.ResponseWriter, r *http.Request, s *site) {
	ctx := context{Site: s}
	ctx.Populate(r)
	if ctx.User == nil {
		http.Redirect(w, r, "/login/", http.StatusFound)
		return
	}
	if r.Method != "POST" {
		ar := addResponse{}
		ctx.PopulateResponse(&ar)
		c, err := s.GetUserChannels(*ctx.User)
		if err != nil {
			http.Error(w, "couldn't get channels", 500)
			return
		}
		ar.Channels = c
		url := r.FormValue("url")
		title := r.FormValue("title")
		if url != "" {
			if title == "" {
				title = url
			}
			ar.Body = "#### [" + title + "](" + url + ")\n\n"
		}
		tmpl := getTemplate("add.html")
		tmpl.Execute(w, ar)
		return
	}
	body := r.FormValue("body")
	nchan := make([]string, 3)
	nchan[0], nchan[1], nchan[2] = r.FormValue("new_channel0"), r.FormValue("new_channel1"), r.FormValue("new_channel2")
	channels, err := s.AddChannels(*ctx.User, nchan)
	if err != nil {
		log.Fatal(err)
		fmt.Fprintf(w, "error making channels")
		return
	}

	// and any existing selected channels
	for k := range r.Form {
		if strings.HasPrefix(k, "channel_") {
			id, err := strconv.Atoi(strings.TrimPrefix(k, "channel_"))
			if err != nil {
				// couldn't parse it for some reason
				continue
			}
			c, err := s.GetChannelByID(id)
			if err != nil {
				continue
			}
			channels = append(channels, c)
		}
	}

	_, err = s.AddPost(*ctx.User, body, channels)
	if err != nil {
		log.Fatal(err)
		fmt.Fprintf(w, "could not add post")
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func userDispatch(w http.ResponseWriter, r *http.Request, s *site) {
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) < 4 {
		http.Error(w, "bad request", 400)
		return
	}
	if parts[1] != "u" {
		http.Error(w, "bad request", 400)
		return
	}
	ctx := context{Site: s}
	username := parts[2]
	u, err := s.GetUser(username)
	if err != nil {
		http.Error(w, "user doesn't exist", 404)
		return
	}
	if len(parts) == 4 {
		userIndex(w, r, ctx, u)
		return
	}

	if parts[3] == "feed" {
		userFeed(w, r, ctx, u)
		return
	}

	if parts[3] == "c" {
		slug := parts[4]
		channel, err := s.GetChannel(*u, slug)
		if err != nil {
			http.Error(w, "channel not found", 404)
			return
		}
		if len(parts) == 6 {
			channelIndex(w, r, ctx, u, channel)
			return
		}
		if parts[5] == "delete" {
			channelDelete(w, r, ctx, u, channel)
			return
		}
		if parts[5] == "feed" {
			channelFeed(w, r, ctx, u, channel)
			return
		}
	}

	if parts[3] == "p" {
		individualPostHandler(w, r, s, parts, ctx, u)
		return
	}

	http.Error(w, "unknown page", 404)
}

func individualPostHandler(w http.ResponseWriter, r *http.Request, s *site, parts []string, ctx context, u *user) {
	if len(parts) < 4 {
		http.Error(w, "not found", 404)
		return
	}
	puuid := parts[4]
	p, err := s.GetPostByUUID(puuid)
	if err != nil {
		http.Error(w, "post not found", 404)
		return
	}

	if len(parts) == 6 {
		postPage(w, r, ctx, u, p)
		return
	}
	if parts[5] == "delete" {
		postDelete(w, r, ctx, u, p)
		return
	}
	http.Error(w, "unknown page", 404)
}

type postPageResponse struct {
	Post *post
	siteResponse
}

func postPage(w http.ResponseWriter, r *http.Request, ctx context, u *user, p *post) {
	ctx.Populate(r)
	pr := postPageResponse{}
	ctx.PopulateResponse(&pr)
	pr.Post = p
	channels, err := ctx.Site.GetPostChannels(p)
	if err != nil {
		http.Error(w, "error retrieving channels", 500)
	}
	pr.Post.Channels = channels
	tmpl := getTemplate("post.html")
	tmpl.Execute(w, pr)

}

type userIndexResponse struct {
	User     *user
	Posts    []*post
	Channels []*channel
	siteResponse
}

func userIndex(w http.ResponseWriter, r *http.Request, ctx context, u *user) {
	ctx.Populate(r)
	ir := userIndexResponse{User: u}
	ctx.PopulateResponse(&ir)

	allPosts, err := ctx.Site.GetAllUserPosts(u, ctx.Site.ItemsPerPage, 0)
	if err != nil {
		http.Error(w, "couldn't retrieve posts", 500)
		return
	}
	ir.Posts = allPosts
	c, err := ctx.Site.GetUserChannels(*u)
	if err != nil {
		http.Error(w, "couldn't get channels", 500)
		return
	}
	ir.Channels = c

	tmpl := getTemplate("user.html")
	tmpl.Execute(w, ir)
}

func userFeed(w http.ResponseWriter, r *http.Request, ctx context, u *user) {
	base := ctx.Site.BaseURL

	allPosts, err := ctx.Site.GetAllUserPosts(u, ctx.Site.ItemsPerPage, 0)
	if err != nil {
		http.Error(w, "couldn't retrieve posts", 500)
		return
	}
	if len(allPosts) == 0 {
		http.Error(w, "no posts", 404)
		return
	}
	latest := allPosts[0]

	feed := &feeds.Feed{
		Title:       "Finch Feed for " + u.Username,
		Link:        &feeds.Link{Href: base + "/u/" + u.Username + "/feed/"},
		Description: "Finch feed",
		Author:      &feeds.Author{Name: u.Username, Email: u.Username},
		Created:     latest.Time(),
	}
	feed.Items = []*feeds.Item{}

	const layout = "Jan 2, 2006 at 3:04pm (MST)"
	for _, p := range allPosts {
		feed.Items = append(feed.Items,
			&feeds.Item{
				Title:       u.Username + ": " + p.Time().UTC().Format(layout),
				Link:        &feeds.Link{Href: base + p.URL()},
				Description: string(blackfriday.MarkdownBasic([]byte(p.Body))),
				Author:      &feeds.Author{Name: u.Username, Email: u.Username},
				Created:     p.Time(),
			})
	}
	atom, _ := feed.ToAtom()
	w.Header().Set("Content-Type", "application/atom+xml")
	fmt.Fprintf(w, atom)
}

func channelDelete(w http.ResponseWriter, r *http.Request, ctx context, u *user, c *channel) {
	if r.Method != "POST" {
		fmt.Fprintf(w, "POST only")
		return
	}
	ctx.Populate(r)
	if ctx.User == nil {
		http.Redirect(w, r, "/login/", http.StatusFound)
		return
	}
	if ctx.User.ID != c.User.ID {
		http.Error(w, "you can only delete your own channels", 403)
		return
	}
	ctx.Site.DeleteChannel(c)
	http.Redirect(w, r, "/", http.StatusFound)
}

func postDelete(w http.ResponseWriter, r *http.Request, ctx context, u *user, p *post) {
	if r.Method != "POST" {
		fmt.Fprintf(w, "POST only")
		return
	}
	ctx.Populate(r)
	if ctx.User == nil {
		http.Redirect(w, r, "/login/", http.StatusFound)
		return
	}
	if ctx.User.ID != p.User.ID {
		http.Error(w, "you can only delete your own posts", 403)
		return
	}
	ctx.Site.DeletePost(p)
	http.Redirect(w, r, "/", http.StatusFound)
}

func channelFeed(w http.ResponseWriter, r *http.Request, ctx context, u *user, c *channel) {
	base := ctx.Site.BaseURL

	allPosts, err := ctx.Site.GetAllPostsInChannel(*c, ctx.Site.ItemsPerPage, 0)
	if err != nil {
		http.Error(w, "couldn't retrieve posts", 500)
		return
	}
	if len(allPosts) == 0 {
		http.Error(w, "no posts", 404)
		return
	}
	latest := allPosts[0]

	feed := &feeds.Feed{
		Title:       "Finch Feed for " + u.Username + " / " + c.Label,
		Link:        &feeds.Link{Href: base + "/u/" + u.Username + "/c/" + c.Slug + "/feed/"},
		Description: "Finch Channel feed",
		Author:      &feeds.Author{Name: u.Username, Email: u.Username},
		Created:     latest.Time(),
	}
	feed.Items = []*feeds.Item{}

	const layout = "Jan 2, 2006 at 3:04pm (MST)"
	for _, p := range allPosts {
		feed.Items = append(feed.Items,
			&feeds.Item{
				Title:       u.Username + ": " + p.Time().UTC().Format(layout),
				Link:        &feeds.Link{Href: base + p.URL()},
				Description: string(blackfriday.MarkdownBasic([]byte(p.Body))),
				Author:      &feeds.Author{Name: u.Username, Email: u.Username},
				Created:     p.Time(),
			})
	}
	atom, _ := feed.ToAtom()
	w.Header().Set("Content-Type", "application/atom+xml")
	fmt.Fprintf(w, atom)
}

type channelIndexResponse struct {
	Channel *channel
	Posts   []*post
	siteResponse
}

func channelIndex(w http.ResponseWriter, r *http.Request, ctx context, u *user, c *channel) {
	ctx.Populate(r)
	ir := channelIndexResponse{Channel: c}
	ctx.PopulateResponse(&ir)

	allPosts, err := ctx.Site.GetAllPostsInChannel(*c, ctx.Site.ItemsPerPage, 0)
	if err != nil {
		http.Error(w, "couldn't retrieve posts", 500)
		return
	}
	ir.Posts = allPosts
	tmpl := getTemplate("channel.html")
	tmpl.Execute(w, ir)
}

func registerForm(w http.ResponseWriter, r *http.Request, s *site) {
	ctx := context{Site: s}
	ctx.Populate(r)
	ir := siteResponse{}
	ctx.PopulateResponse(&ir)
	tmpl := getTemplate("register.html")
	tmpl.Execute(w, ir)
}

func registerHandler(w http.ResponseWriter, r *http.Request, s *site) {
	if r.Method == "GET" {
		registerForm(w, r, s)
		return
	}
	if r.Method == "POST" {
		username, password, pass2 := r.FormValue("username"), r.FormValue("password"), r.FormValue("pass2")
		if password != pass2 {
			fmt.Fprintf(w, "passwords don't match")
			return
		}
		user, err := s.CreateUser(username, password)

		if err != nil {
			fmt.Println(err)
			fmt.Fprintf(w, "could not create user")
			return
		}

		sess, _ := s.Store.Get(r, "finch")
		sess.Values["user"] = user.Username
		sess.Save(r, w)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func loginForm(w http.ResponseWriter, req *http.Request) {
	tmpl := getTemplate("login.html")
	tmpl.Execute(w, nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request, s *site) {
	if r.Method == "GET" {
		loginForm(w, r)
		return
	}
	if r.Method == "POST" {
		username, password := r.FormValue("username"), r.FormValue("password")
		user, err := s.GetUser(username)

		if err != nil {
			fmt.Fprintf(w, "user not found")
			return
		}
		if !user.CheckPassword(password) {
			fmt.Fprintf(w, "login failed")
			return
		}

		// store userid in session
		sess, _ := s.Store.Get(r, "finch")
		sess.Values["user"] = user.Username
		sess.Save(r, w)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func logoutHandler(w http.ResponseWriter, r *http.Request, s *site) {
	sess, _ := s.Store.Get(r, "finch")
	delete(sess.Values, "user")
	sess.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func getTemplate(filename string) *template.Template {
	var t = template.New("base.html")
	return template.Must(t.ParseFiles(
		filepath.Join(templateDir, "base.html"),
		filepath.Join(templateDir, filename),
	))
}
