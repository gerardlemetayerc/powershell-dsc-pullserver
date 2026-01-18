package internal

// SAMLUserMapping définit le mapping obligatoire des claims SAML vers les champs utilisateur locaux
// Les clés sont les noms des claims SAML attendus, les valeurs les champs locaux
// Exemple: {"email": "email", "givenName": "first_name", "sn": "last_name"}

type SAMLUserMapping map[string]string

// AppConfig inclut la config SAML
// ...existing code...
