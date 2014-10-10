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
)

type SiteResponse struct {
	Username string
}

func (s *SiteResponse) SetUsername(username string) {
	s.Username = username
}

func (s SiteResponse) GetUsername() string {
	return s.Username
}

type SR interface {
	SetUsername(string)
	GetUsername() string
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	// just ignore this crap
}

type Context struct {
	P    *Persistence
	User *User
}

func (c *Context) Populate(r *http.Request) {
	sess, _ := store.Get(r, "finch")
	username, found := sess.Values["user"]
	if found && username != "" {
		user, found := c.P.GetUser(username.(string))
		if found {
			c.User = user
		}
	}
}

func (c Context) PopulateResponse(sr SR) {
	if c.User != nil {
		sr.SetUsername(c.User.Username)
	}
}

type IndexResponse struct {
	Channels []*Channel
	Posts    []*Post
	SiteResponse
}

func indexHandler(w http.ResponseWriter, r *http.Request, ctx Context) {
	ctx.Populate(r)
	ir := IndexResponse{}
	ctx.PopulateResponse(&ir)
	if ctx.User != nil {
		ir.Channels = ctx.P.UserChannels(*ctx.User)
	}

	posts, err := ctx.P.GetAllPosts(50, 0)
	ir.Posts = posts
	if err != nil {
		log.Println(err)
		fmt.Fprintf(w, "error getting posts")
		return
	}
	tmpl := getTemplate("index.html")
	tmpl.Execute(w, ir)
}

func postHandler(w http.ResponseWriter, r *http.Request, ctx Context) {
	if r.Method != "POST" {
		fmt.Fprintf(w, "POST only")
		return
	}
	ctx.Populate(r)
	if ctx.User == nil {
		http.Redirect(w, r, "/login/", http.StatusFound)
		return
	}
	body := r.FormValue("body")
	nchan := make([]string, 3)
	nchan[0], nchan[1], nchan[2] = r.FormValue("new_channel0"), r.FormValue("new_channel1"), r.FormValue("new_channel2")
	channels, err := ctx.P.AddChannels(*ctx.User, nchan)
	if err != nil {
		log.Fatal(err)
		fmt.Fprintf(w, "error making channels")
		return
	}

	// and any existing selected channels
	for k, _ := range r.Form {
		if strings.HasPrefix(k, "channel_") {
			id, err := strconv.Atoi(strings.TrimPrefix(k, "channel_"))
			if err != nil {
				// couldn't parse it for some reason
				continue
			}
			c, err := ctx.P.GetChannelById(id)
			if err != nil {
				continue
			}
			channels = append(channels, c)
		}
	}

	_, err = ctx.P.AddPost(*ctx.User, body, channels)
	if err != nil {
		log.Fatal(err)
		fmt.Fprintf(w, "could not add post")
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func userDispatch(w http.ResponseWriter, r *http.Request, ctx Context) {
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) < 4 {
		http.Error(w, "bad request", 400)
		return
	}
	if parts[1] != "u" {
		http.Error(w, "bad request", 400)
		return
	}
	username := parts[2]
	u, found := ctx.P.GetUser(username)
	if !found {
		http.Error(w, "user doesn't exist", 404)
		return
	}
	if len(parts) == 4 {
		userWebIndex(w, r, ctx, u)
		return
	}

	if parts[3] == "feed" {
		userFeed(w, r, ctx, u)
		return
	}

	if parts[3] == "c" {
		slug := parts[4]
		channel, err := ctx.P.GetChannel(*u, slug)
		if err != nil {
			http.Error(w, "channel not found", 404)
		}
		if len(parts) == 5 {
			fmt.Fprintf(w, "channel index page for %s", channel.Label)
			return
		}
		if parts[5] == "feed" {
			channelFeed(w, r, ctx, u, channel)
			return
		}
	}

	if parts[3] == "p" {
		// individual post
	}

	fmt.Fprintf(w, "%v", u.Username)
}

func userWebIndex(w http.ResponseWriter, r *http.Request, ctx Context, u *User) {
	fmt.Fprintf(w, "web index for user %s", u.Username)
}

func userFeed(w http.ResponseWriter, r *http.Request, ctx Context, u *User) {
	base := "http://finch.thraxil.org"

	all_posts, err := ctx.P.GetAllUserPosts(u, 50, 0)
	if err != nil {
		http.Error(w, "couldn't retrieve posts", 500)
		return
	}
	if len(all_posts) == 0 {
		http.Error(w, "no posts", 404)
		return
	}
	latest := all_posts[0]

	feed := &feeds.Feed{
		Title:       "Finch Feed for " + u.Username,
		Link:        &feeds.Link{Href: base + "/u/" + u.Username + "/feed/"},
		Description: "Finch feed",
		Author:      &feeds.Author{u.Username, "anders@columbia.edu"},
		Created:     latest.Time(),
	}
	feed.Items = []*feeds.Item{}

	const layout = "Jan 2, 2006 at 3:04pm (MST)"
	for _, p := range all_posts {
		feed.Items = append(feed.Items,
			&feeds.Item{
				Title:       u.Username + ": " + p.Time().UTC().Format(layout),
				Link:        &feeds.Link{Href: base + p.URL()},
				Description: p.Body,
				Author:      &feeds.Author{u.Username, u.Username},
				Created:     p.Time(),
			})
	}
	atom, _ := feed.ToAtom()
	w.Header().Set("Content-Type", "application/atom+xml")
	fmt.Fprintf(w, atom)
}

func channelFeed(w http.ResponseWriter, r *http.Request, ctx Context, u *User, c *Channel) {
	base := "http://finch.thraxil.org"

	all_posts, err := ctx.P.GetAllPostsInChannel(*c, 50, 0)
	if err != nil {
		http.Error(w, "couldn't retrieve posts", 500)
		return
	}
	if len(all_posts) == 0 {
		http.Error(w, "no posts", 404)
		return
	}
	latest := all_posts[0]

	feed := &feeds.Feed{
		Title:       "Finch Feed for " + u.Username + " / " + c.Label,
		Link:        &feeds.Link{Href: base + "/u/" + u.Username + "/c/" + c.Slug + "/feed/"},
		Description: "Finch Channel feed",
		Author:      &feeds.Author{u.Username, "anders@columbia.edu"},
		Created:     latest.Time(),
	}
	feed.Items = []*feeds.Item{}

	const layout = "Jan 2, 2006 at 3:04pm (MST)"
	for _, p := range all_posts {
		feed.Items = append(feed.Items,
			&feeds.Item{
				Title:       u.Username + ": " + p.Time().UTC().Format(layout),
				Link:        &feeds.Link{Href: base + p.URL()},
				Description: p.Body,
				Author:      &feeds.Author{u.Username, u.Username},
				Created:     p.Time(),
			})
	}
	atom, _ := feed.ToAtom()
	w.Header().Set("Content-Type", "application/atom+xml")
	fmt.Fprintf(w, atom)
}

func registerForm(w http.ResponseWriter, r *http.Request, ctx Context) {
	ctx.Populate(r)
	ir := SiteResponse{}
	ctx.PopulateResponse(&ir)
	tmpl := getTemplate("register.html")
	tmpl.Execute(w, ir)
}

func registerHandler(w http.ResponseWriter, r *http.Request, ctx Context) {
	if r.Method == "GET" {
		registerForm(w, r, ctx)
		return
	}
	if r.Method == "POST" {
		username, password, pass2 := r.FormValue("username"), r.FormValue("password"), r.FormValue("pass2")
		if password != pass2 {
			fmt.Fprintf(w, "passwords don't match")
			return
		}
		user, err := ctx.P.CreateUser(username, password)

		if err != nil {
			fmt.Println(err)
			fmt.Fprintf(w, "could not create user")
			return
		}

		sess, _ := store.Get(r, "finch")
		sess.Values["user"] = user.Username
		sess.Save(r, w)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func loginForm(w http.ResponseWriter, req *http.Request) {
	tmpl := getTemplate("login.html")
	tmpl.Execute(w, nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request, ctx Context) {
	if r.Method == "GET" {
		loginForm(w, r)
		return
	}
	if r.Method == "POST" {
		username, password := r.FormValue("username"), r.FormValue("password")
		user, found := ctx.P.GetUser(username)

		if !found {
			fmt.Fprintf(w, "user not found")
			return
		}
		if !user.CheckPassword(password) {
			fmt.Fprintf(w, "login failed")
			return
		}

		// store userid in session
		sess, _ := store.Get(r, "finch")
		sess.Values["user"] = user.Username
		sess.Save(r, w)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func logoutHandler(w http.ResponseWriter, r *http.Request, ctx Context) {
	sess, _ := store.Get(r, "finch")
	delete(sess.Values, "user")
	sess.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func getTemplate(filename string) *template.Template {
	var t = template.New("base.html")
	return template.Must(t.ParseFiles(
		filepath.Join(template_dir, "base.html"),
		filepath.Join(template_dir, filename),
	))
}
