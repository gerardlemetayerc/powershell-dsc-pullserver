package routes

import (
	"net/http"
	"html/template"
	"path/filepath"
	"go-dsc-pull/utils"
)

// WebLoginHandler serves the login page
func WebLoginHandler(w http.ResponseWriter, r *http.Request) {
	   exeDir, err := utils.ExePath()
	   if err != nil {
		   http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		   return
	   }
	   baseDir := filepath.Dir(exeDir)
	   loginPath := filepath.Join(baseDir, "templates/login.tmpl")
	   http.ServeFile(w, r, loginPath)
}

// WebUsersHandler serves the users page
func WebUsersHandler(w http.ResponseWriter, r *http.Request) {
	   exeDir, err := utils.ExePath()
	   if err != nil {
		   http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		   return
	   }
	   baseDir := filepath.Dir(exeDir)
	   tmpl, err := template.New("layout.tmpl").
		   ParseFiles(
			   filepath.Join(baseDir, "templates/layout.tmpl"),
			   filepath.Join(baseDir, "templates/head.tmpl"),
			   filepath.Join(baseDir, "templates/menu.tmpl"),
			   filepath.Join(baseDir, "templates/footer.tmpl"),
			   filepath.Join(baseDir, "templates/users.tmpl"),
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
	   exeDir, err := utils.ExePath()
	   if err != nil {
		   http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		   return
	   }
	   baseDir := filepath.Dir(exeDir)
	   tmpl, err := template.New("layout.tmpl").
		   ParseFiles(
			   filepath.Join(baseDir, "templates/layout.tmpl"),
			   filepath.Join(baseDir, "templates/head.tmpl"),
			   filepath.Join(baseDir, "templates/menu.tmpl"),
			   filepath.Join(baseDir, "templates/footer.tmpl"),
			   filepath.Join(baseDir, "templates/user_edit.tmpl"),
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
	   exeDir, err := utils.ExePath()
	   if err != nil {
		   http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		   return
	   }
	   baseDir := filepath.Dir(exeDir)
	   tmpl, err := template.New("layout.tmpl").
		   ParseFiles(
			   filepath.Join(baseDir, "templates/layout.tmpl"),
			   filepath.Join(baseDir, "templates/head.tmpl"),
			   filepath.Join(baseDir, "templates/menu.tmpl"),
			   filepath.Join(baseDir, "templates/footer.tmpl"),
			   filepath.Join(baseDir, "templates/user_password.tmpl"),
		   )
	   if err != nil {
		   http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		   return
	   }
	   data := map[string]interface{}{ "Title": "Mot de passe utilisateur" }
	   tmpl.ExecuteTemplate(w, "layout", data)
}
