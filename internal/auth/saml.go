package auth

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"crypto"
	samlsp "github.com/crewjam/saml/samlsp"
	"go-dsc-pull/internal/schema"
)

// InitSamlMiddleware initialise le middleware SAML si activé, sinon retourne nil
func InitSamlMiddleware(appCfg *schema.AppConfig) (*samlsp.Middleware, error) {
	if appCfg == nil || !appCfg.SAML.Enabled {
		return nil, nil
	}
	// Charge la clé privée et le certificat
	cert, err := tls.LoadX509KeyPair(appCfg.SAML.SPCertFile, appCfg.SAML.SPKeyFile)
	if err != nil {
		return nil, fmt.Errorf("Erreur chargement clé/cert SP: %v", err)
	}
	// Utilise l'entity_id de la config pour l'URL du SP
	spURL := appCfg.SAML.EntityID
	idpMetadataURL := appCfg.SAML.IdpMetadataURL
	parsedIdpURL, err := url.Parse(idpMetadataURL)
	if err != nil {
		return nil, fmt.Errorf("Erreur parsing IdP Metadata URL: %v", err)
	}
	idpMetadata, err := samlsp.FetchMetadata(context.Background(), http.DefaultClient, *parsedIdpURL)
	if err != nil {
		return nil, fmt.Errorf("Erreur récupération metadata IdP: %v", err)
	}
	parsedSpURL, err := url.Parse(spURL)
	if err != nil {
		return nil, fmt.Errorf("Erreur parsing SP URL: %v", err)
	}
	samlOptions := samlsp.Options{
		URL: *parsedSpURL,
		Key: cert.PrivateKey.(crypto.Signer),
		Certificate: cert.Leaf,
		IDPMetadata: idpMetadata,
	}
	samlMiddleware, err := samlsp.New(samlOptions)
	if err != nil {
		return nil, fmt.Errorf("Erreur initialisation SAML middleware: %v", err)
	}
	if samlMiddleware != nil {
		// Force explicitement l'EntityID à la valeur de la config
		samlMiddleware.ServiceProvider.EntityID = appCfg.SAML.EntityID
	}
	return samlMiddleware, nil
}


