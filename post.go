package main

import (
	"html/template"
	"time"

	"github.com/russross/blackfriday"
)

type Post struct {
	Id       int
	UUID     string
	User     *User
	Body     string
	Posted   int
	Channels []*Channel
}

func (p Post) RenderBody() template.HTML {
	return template.HTML(string(blackfriday.MarkdownCommon([]byte(p.Body))))
}

func (p Post) URL() string {
	return "/u/" + p.User.Username + "/p/" + p.UUID + "/"
}

func (p Post) Time() time.Time {
	return time.Unix(int64(p.Posted), 0)
}
