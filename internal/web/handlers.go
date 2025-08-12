package web

import (
	"net/http"

	"github.com/a-h/templ"
)

// Helper to render templ components consistently
func RenderHTML(w http.ResponseWriter, r *http.Request, c templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = c.Render(r.Context(), w)
}
