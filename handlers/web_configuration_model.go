package handlers

import (
	"net/http"
	"html/template"
)

func WebConfigurationModelHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("layout.html").
		ParseFiles(
			"templates/layout.html",
			"templates/head.tmpl",
			"templates/menu.tmpl",
			"templates/footer.tmpl",
			"templates/configuration_model.tmpl",
		)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	data := map[string]interface{}{
		"Title": "Configuration Models",
	}
	err = tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}
