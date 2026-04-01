package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed html/*
var htmlFS embed.FS

//go:embed translation/*
var i18nFS embed.FS

func GetHTMLFilesystem() http.FileSystem {
	sub, _ := fs.Sub(htmlFS, "html")
	return http.FS(sub)
}

func GetI18nFilesystem() fs.FS {
	return i18nFS
}
