package handlers

import (
	"net/http"
	"html/template"
	"path/filepath"
	"go-dsc-pull/utils"
)

func WebNodePropertiesHandler(w http.ResponseWriter, r *http.Request) {
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
			   filepath.Join(baseDir, "templates/node_properties.tmpl"),
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
			   filepath.Join(baseDir, "templates/properties.tmpl"),
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
