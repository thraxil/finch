package main

import (
	"html/template"
	"time"

	"github.com/russross/blackfriday"
)

type post struct {
	ID       int
	UUID     string
	User     *user
	Body     string
	Posted   int
	Channels []*channel
}

func (p post) RenderBody() template.HTML {
	return template.HTML(string(blackfriday.MarkdownCommon([]byte(p.Body))))
}

func (p post) URL() string {
	return "/u/" + p.User.Username + "/p/" + p.UUID + "/"
}

func (p post) Time() time.Time {
	return time.Unix(int64(p.Posted), 0)
}
