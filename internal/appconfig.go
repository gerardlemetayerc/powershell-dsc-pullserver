

package internal

import (
   "encoding/json"
   "os"
)

// GetSAMLUserMapping lit le mapping user_mapping depuis la config
func GetSAMLUserMapping() (map[string]string, error) {
   var mapping map[string]string
   // On lit le champ brut du JSON
   f, err := os.Open("config.json")
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

type AppConfig struct {
	Driver   string      `json:"driver"`
	Server   string      `json:"server"`
	Port     int         `json:"port"`
	User     string      `json:"user"`
	Password string      `json:"password"`
	Database string      `json:"database"`
	DSCPort  int         `json:"dsc_port"`
	WebPort  int         `json:"web_port"`
	SAML     SAMLConfig  `json:"saml"`
}

func LoadAppConfig(path string) (*AppConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var cfg AppConfig
	dec := json.NewDecoder(f)
	if err := dec.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
