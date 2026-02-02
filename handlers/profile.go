package handlers

import (
	"net/http"
	"html/template"
	"path/filepath"
	"go-dsc-pull/utils"
)

// ProfileHandler sert la page de profil utilisateur
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	   exeDir, err := utils.ExePath()
	   if err != nil {
		   http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		   return
	   }
	   baseDir := filepath.Dir(exeDir)
	   tmpl, err := template.ParseFiles(
		   filepath.Join(baseDir, "templates/layout.tmpl"),
		   filepath.Join(baseDir, "templates/head.tmpl"),
		   filepath.Join(baseDir, "templates/menu.tmpl"),
		   filepath.Join(baseDir, "templates/footer.tmpl"),
		   filepath.Join(baseDir, "templates/profile.tmpl"),
	   )
	   if err != nil {
		   http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		   return
	   }
	   data := map[string]interface{}{ "Title": "User Profil" }
	   tmpl.ExecuteTemplate(w, "layout", data)
}
