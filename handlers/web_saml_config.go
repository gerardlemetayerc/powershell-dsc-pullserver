package handlers

import (
	"net/http"
	"html/template"
)

// WebSAMLConfigHandler serves the SAML config admin page
func WebSAMLConfigHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("layout.tmpl").
		ParseFiles(
			"templates/layout.tmpl",
			"templates/head.tmpl",
			"templates/menu.tmpl",
			"templates/footer.tmpl",
			"templates/saml_config.tmpl",
		)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	data := map[string]interface{}{ "Title": "SAML Configuration" }
	tmpl.ExecuteTemplate(w, "layout", data)
}
