package internal

import (
   "encoding/json"
   "go-dsc-pull/internal/schema"
   "os"
   "path/filepath"
   "go-dsc-pull/utils"
)

// GetSAMLUserMapping lit le mapping user_mapping depuis la config
func GetSAMLUserMapping() (map[string]string, error) {
   var mapping map[string]string
   // On lit le champ brut du JSON
      exeDir, err := utils.ExePath()
      if err != nil {
         return nil, err
      }
      configPath := filepath.Join(filepath.Dir(exeDir), "config.json")
      f, err := os.Open(configPath)
      if err != nil {
         return nil, err
      }
      defer f.Close()
   dec := json.NewDecoder(f)
   var raw map[string]interface{}
   if err := dec.Decode(&raw); err != nil {
	   return nil, err
   }
   samlRaw, ok := raw["saml"].(map[string]interface{})
   if !ok {
	   return nil, nil
   }
   userMappingRaw, ok := samlRaw["user_mapping"].(map[string]interface{})
   if !ok {
	   return nil, nil
   }
   mapping = make(map[string]string)
   for k, v := range userMappingRaw {
	   if s, ok := v.(string); ok {
		   mapping[k] = s
	   }
   }
   return mapping, nil
}

func LoadAppConfig(path string) (*schema.AppConfig, error) {
      exeDir, err := utils.ExePath()
      if err != nil {
         return nil, err
      }
      absPath := filepath.Join(filepath.Dir(exeDir), path)
      f, err := os.Open(absPath)
      if err != nil {
         return nil, err
      }
      defer f.Close()
	var cfg schema.AppConfig
	dec := json.NewDecoder(f)
	if err := dec.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
