package templates

import (
	"embed"
	"html/template"
)

//go:embed *.html
var files embed.FS

var Home = template.Must(template.ParseFS(files, "home.html", "post.html", "comment.html"))
