package handlers

import (
	"net/http"
	"html/template"
)

func WebModulesHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("layout.html").
		ParseFiles(
			"web/layout.html",
			"web/head.tmpl",
			"web/menu.tmpl",
			"web/footer.tmpl",
			"web/modules.tmpl",
		)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	data := map[string]interface{}{
		"Title": "Modules DSC",
	}
	err = tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}
