package handlers

import (
	"net/http"
	"html/template"
)

func WebNodePropertiesHandler(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.New("layout.html").
		ParseFiles(
			"web/layout.html",
			"web/head.tmpl",
			"web/menu.tmpl",
			"web/footer.tmpl",
			"web/node_properties.html",
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
		tmpl, err := template.New("layout.html").
		ParseFiles(
			"web/layout.html",
			"web/head.tmpl",
			"web/menu.tmpl",
			"web/footer.tmpl",
			"web/properties.html",
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
