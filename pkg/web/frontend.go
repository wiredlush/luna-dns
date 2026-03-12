//go:build web

package web

import "embed"

//go:embed static/*
var staticFS embed.FS
