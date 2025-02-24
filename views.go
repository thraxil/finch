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
	Username          string
	AllowRegistration bool
}

func (s *siteResponse) SetUsername(username string) {
	s.Username = username
}

func (s siteResponse) GetUsername() string {
	return s.Username
}

func (s *siteResponse) SetAllowRegistration(allowReg bool) {
	s.AllowRegistration = allowReg
}

type sr interface {
	SetUsername(string)
	GetUsername() string
	SetAllowRegistration(bool)
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	// just ignore this crap
}

type siteContext struct {
	Site *site
	User *user
}

func (c *siteContext) Populate(r *http.Request) {
	sess, _ := c.Site.Store.Get(r, "finch")
	username, found := sess.Values["user"]
	if found && username != "" {
		user, err := c.Site.GetUser(username.(string))
		if err == nil {
			c.User = user
		}
	}
}

func (c siteContext) PopulateResponse(sr sr) {
	if c.User != nil {
		sr.SetUsername(c.User.Username)
	}
	sr.SetAllowRegistration(c.Site.AllowRegistration)
}

type paginationResponse struct {
	Page        int
	NextPage    int
	HasNextPage bool
	PrevPage    int
	HasPrevPage bool
}

func indexHandler(s *site) http.Handler {
	type indexResponse struct {
		Posts []*post
		siteResponse
		paginationResponse
	}

	tmpl := getTemplate("index.html")

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := siteContext{Site: s}
			ctx.Populate(r)
			ir := indexResponse{}
			ctx.PopulateResponse(&ir)
			spage := r.URL.Query().Get("page")
			page, err := strconv.Atoi(spage)
			if err != nil {
				page = 0
			}
			posts, err := s.GetAllPosts(s.ItemsPerPage, page*s.ItemsPerPage)
			ir.Posts = posts
			ir.Page = page + 1
			ir.PrevPage = page - 1
			ir.NextPage = page + 1
			ir.HasPrevPage = false
			ir.HasNextPage = false
			if ir.PrevPage > -1 {
				ir.HasPrevPage = true
			}
			// not the most accurate approach...
			// sometimes there will be an empty page at the end
			if len(posts) == s.ItemsPerPage {
				ir.HasNextPage = true
			}
			if err != nil {
				log.Println(err)
				fmt.Fprintf(w, "error getting posts")
				return
			}
			tmpl.Execute(w, ir)
		})
}

func searchHandler(s *site) http.Handler {
	type searchResponse struct {
		Posts []*post
		Q     string
		siteResponse
	}
	tmpl := getTemplate("search.html")

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := siteContext{Site: s}
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
			tmpl.Execute(w, sr)
		})
}

func bodyFromFields(url, title string) string {
	if url != "" {
		if title == "" {
			title = url
		}
		return "#### [" + title + "](" + url + ")\n\n"
	}
	return ""
}

func postFormHandler(s *site) http.Handler {
	type addResponse struct {
		Channels []*channel
		Body     string
		siteResponse
	}
	tmpl := getTemplate("add.html")

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := siteContext{Site: s}
			ctx.Populate(r)
			if ctx.User == nil {
				http.Redirect(w, r, "/login/", http.StatusFound)
				return
			}
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
			ar.Body = bodyFromFields(url, title)
			tmpl.Execute(w, ar)
			return
		})
}

func postHandler(s *site) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := siteContext{Site: s}
			ctx.Populate(r)
			if ctx.User == nil {
				http.Redirect(w, r, "/login/", http.StatusFound)
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
		})
}

func individualPostHandler(s *site) http.Handler {
	type postPageResponse struct {
		Post *post
		siteResponse
	}
	tmpl := getTemplate("post.html")
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			username := r.PathValue("username")
			puuid := r.PathValue("puuid")
			ctx := siteContext{Site: s}
			_, err := s.GetUser(username)
			if err != nil {
				http.Error(w, "user doesn't exist", 404)
				return
			}
			p, err := s.GetPostByUUID(puuid)
			if err != nil {
				http.Error(w, "post not found", 404)
				return
			}
			ctx.Populate(r)
			pr := postPageResponse{}
			ctx.PopulateResponse(&pr)
			pr.Post = p
			channels, err := ctx.Site.GetPostChannels(p)
			if err != nil {
				http.Error(w, "error retrieving channels", 500)
			}
			pr.Post.Channels = channels
			tmpl.Execute(w, pr)
		})
}

func userIndex(s *site) http.Handler {
	type userIndexResponse struct {
		User     *user
		Posts    []*post
		Channels []*channel
		siteResponse
		paginationResponse
	}
	tmpl := getTemplate("user.html")
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			username := r.PathValue("username")
			ctx := siteContext{Site: s}
			u, err := s.GetUser(username)
			if err != nil {
				http.Error(w, "user doesn't exist", 404)
				return
			}
			ctx.Populate(r)
			ir := userIndexResponse{User: u}
			ctx.PopulateResponse(&ir)

			spage := r.URL.Query().Get("page")
			page, err := strconv.Atoi(spage)
			if err != nil {
				page = 0
			}
			allPosts, err := ctx.Site.GetAllUserPosts(u, ctx.Site.ItemsPerPage, page*ctx.Site.ItemsPerPage)
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

			ir.Page = page + 1
			ir.PrevPage = page - 1
			ir.NextPage = page + 1
			ir.HasPrevPage = false
			ir.HasNextPage = false
			if ir.PrevPage > -1 {
				ir.HasPrevPage = true
			}
			// not the most accurate approach...
			// sometimes there will be an empty page at the end
			if len(ir.Posts) == s.ItemsPerPage {
				ir.HasNextPage = true
			}

			tmpl.Execute(w, ir)
		})
}

func userFeed(s *site) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			username := r.PathValue("username")
			ctx := siteContext{Site: s}
			u, err := s.GetUser(username)
			if err != nil {
				http.Error(w, "user doesn't exist", 404)
				return
			}
			ctx.Populate(r)
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
		})
}

func channelDelete(s *site) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			username := r.PathValue("username")
			slug := r.PathValue("slug")
			ctx := siteContext{Site: s}
			u, err := s.GetUser(username)
			if err != nil {
				http.Error(w, "user doesn't exist", 404)
				return
			}
			c, err := s.GetChannel(*u, slug)
			if err != nil {
				http.Error(w, "channel not found", 404)
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
		})
}

func postDelete(s *site) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			username := r.PathValue("username")
			puuid := r.PathValue("puuid")
			ctx := siteContext{Site: s}
			_, err := s.GetUser(username)
			if err != nil {
				http.Error(w, "user doesn't exist", 404)
				return
			}
			p, err := s.GetPostByUUID(puuid)
			if err != nil {
				http.Error(w, "post not found", 404)
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
		})
}

func channelFeed(s *site) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			username := r.PathValue("username")
			slug := r.PathValue("slug")
			ctx := siteContext{Site: s}
			u, err := s.GetUser(username)
			if err != nil {
				http.Error(w, "user doesn't exist", 404)
				return
			}
			c, err := s.GetChannel(*u, slug)
			if err != nil {
				http.Error(w, "channel not found", 404)
				return
			}
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
		})
}

func channelIndex(s *site) http.Handler {
	type channelIndexResponse struct {
		Channel *channel
		Posts   []*post
		siteResponse
		paginationResponse
	}
	tmpl := getTemplate("channel.html")
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			username := r.PathValue("username")
			slug := r.PathValue("slug")
			ctx := siteContext{Site: s}
			u, err := s.GetUser(username)
			if err != nil {
				http.Error(w, "user doesn't exist", 404)
				return
			}
			c, err := s.GetChannel(*u, slug)
			if err != nil {
				http.Error(w, "channel not found", 404)
				return
			}
			ctx.Populate(r)
			ir := channelIndexResponse{Channel: c}
			ctx.PopulateResponse(&ir)

			spage := r.URL.Query().Get("page")
			page, err := strconv.Atoi(spage)
			if err != nil {
				page = 0
			}
			allPosts, err := ctx.Site.GetAllPostsInChannel(*c, ctx.Site.ItemsPerPage, page*ctx.Site.ItemsPerPage)
			if err != nil {
				http.Error(w, "couldn't retrieve posts", 500)
				return
			}
			ir.Posts = allPosts
			ir.Page = page + 1
			ir.PrevPage = page - 1
			ir.NextPage = page + 1
			ir.HasPrevPage = false
			ir.HasNextPage = false
			if ir.PrevPage > -1 {
				ir.HasPrevPage = true
			}
			// not the most accurate approach...
			// sometimes there will be an empty page at the end
			if len(ir.Posts) == s.ItemsPerPage {
				ir.HasNextPage = true
			}
			tmpl.Execute(w, ir)
		})
}

func registerFormHandler(s *site) http.Handler {
	tmpl := getTemplate("register.html")
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := siteContext{Site: s}
			ctx.Populate(r)
			ir := siteResponse{}
			ctx.PopulateResponse(&ir)
			tmpl.Execute(w, ir)
		})
}

func registerHandler(s *site) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if !s.AllowRegistration {
				fmt.Fprintf(w, "registration not allowed")
				return
			}
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
			http.Redirect(w, r, "/", http.StatusFound)
		})
}

func loginFormHandler(s *site) http.Handler {
	tmpl := getTemplate("login.html")
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			tmpl.Execute(w, nil)
		})
}

func loginHandler(s *site) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
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
			http.Redirect(w, r, "/", http.StatusFound)
		})
}

func logoutHandler(s *site) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			sess, _ := s.Store.Get(r, "finch")
			delete(sess.Values, "user")
			sess.Save(r, w)
			http.Redirect(w, r, "/", http.StatusFound)
		})
}

func getTemplate(filename string) *template.Template {
	var t = template.New("base.html")
	return template.Must(t.ParseFiles(
		filepath.Join(templateDir, "base.html"),
		filepath.Join(templateDir, filename),
	))
}

func healthzHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
