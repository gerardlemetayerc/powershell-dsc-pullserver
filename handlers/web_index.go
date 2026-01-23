package handlers

import (
	"net/http"
	"html/template"
)

func WebIndexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("layout.tmpl").
		ParseFiles(
			"templates/layout.tmpl",
			"templates/head.tmpl",
			"templates/menu.tmpl",
			"templates/footer.tmpl",
			"templates/index.tmpl",
		)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	data := map[string]interface{}{
		"Title": "DSC Admin Panel",
	}
	err = tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}
