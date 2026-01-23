package routes

import (
	"net/http"
	"html/template"
)

// WebLoginHandler serves the login page
func WebLoginHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/login.tmpl")
}

// WebUsersHandler serves the users page
func WebUsersHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("layout.tmpl").
		ParseFiles(
			"templates/layout.tmpl",
			"templates/head.tmpl",
			"templates/menu.tmpl",
			"templates/footer.tmpl",
			"templates/users.tmpl",
		)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	data := map[string]interface{}{ "Title": "Utilisateurs" }
	tmpl.ExecuteTemplate(w, "layout", data)
}

// WebUserEditHandler serves the user edit page
func WebUserEditHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("layout.tmpl").
		ParseFiles(
			"templates/layout.tmpl",
			"templates/head.tmpl",
			"templates/menu.tmpl",
			"templates/footer.tmpl",
			"templates/user_edit.tmpl",
		)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	data := map[string]interface{}{ "Title": "Édition utilisateur" }
	tmpl.ExecuteTemplate(w, "layout", data)
}

// WebUserPasswordHandler serves the user password page
func WebUserPasswordHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("layout.tmpl").
		ParseFiles(
			"templates/layout.tmpl",
			"templates/head.tmpl",
			"templates/menu.tmpl",
			"templates/footer.tmpl",
			"templates/user_password.tmpl",
		)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	data := map[string]interface{}{ "Title": "Mot de passe utilisateur" }
	tmpl.ExecuteTemplate(w, "layout", data)
}
