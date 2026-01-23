package handlers

import (
	"net/http"
	"html/template"
)

// ProfileHandler sert la page de profil utilisateur
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/layout.tmpl", "templates/head.tmpl", "templates/menu.tmpl", "templates/footer.tmpl", "templates/profile.tmpl")
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	data := map[string]interface{}{ "Title": "User Profil" }
	tmpl.ExecuteTemplate(w, "layout", data)
}
