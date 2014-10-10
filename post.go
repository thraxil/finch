package main

import (
	"html/template"

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
