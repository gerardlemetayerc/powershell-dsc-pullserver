package handlers

import (
	"net/http"
	"html/template"
)

func WebModulesHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("layout.tmpl").
		ParseFiles(
			"templates/layout.tmpl",
			"templates/head.tmpl",
			"templates/menu.tmpl",
			"templates/footer.tmpl",
			"templates/modules.tmpl",
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
