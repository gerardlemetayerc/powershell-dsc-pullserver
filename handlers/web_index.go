package handlers

import (
	"net/http"
	"html/template"
	"path/filepath"
	"go-dsc-pull/utils"
)

func WebIndexHandler(w http.ResponseWriter, r *http.Request) {
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
			   filepath.Join(baseDir, "templates/index.tmpl"),
		   )
	   if err != nil {
		   http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		   return
	   }
	   data := map[string]interface{}{
		   "Title": "DSC Admin Panel",
	   }
	   err = tmpl.ExecuteTemplate(w, "layout", data)
	   if err != nil {
		   http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	   }
}

// NotFoundHandler affiche la page 404 avec le thème
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
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
		   filepath.Join(baseDir, "templates/404.tmpl"),
	   )
   if err != nil {
	   http.Error(w, "404 - Page non trouvée", http.StatusNotFound)
	   return
   }
   data := map[string]interface{}{ "Title": "404 - Page non trouvée" }
   w.WriteHeader(http.StatusNotFound)
   tmpl.ExecuteTemplate(w, "layout", data)
}