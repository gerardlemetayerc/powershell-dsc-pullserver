package handlers

import (
	"net/http"
	"html/template"
)

func WebNodePropertiesHandler(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.New("layout.tmpl").
		ParseFiles(
			"templates/layout.tmpl",
			"templates/head.tmpl",
			"templates/menu.tmpl",
			"templates/footer.tmpl",
			"templates/node_properties.tmpl",
		)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	data := map[string]interface{}{
		"Title": "Node properties",
	}
	err = tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

func WebPropertiesHandler(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.New("layout.tmpl").
		ParseFiles(
			"templates/layout.tmpl",
			"templates/head.tmpl",
			"templates/menu.tmpl",
			"templates/footer.tmpl",
			"templates/properties.tmpl",
		)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	data := map[string]interface{}{
		"Title": "Properties",
	}
	err = tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}
