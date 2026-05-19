package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

// ValidateDSCRegistrationKey vérifie la signature DSC Authorization selon la spec
// Retourne true si la signature est valide, false sinon, et un motif d'erreur non sensible
func ValidateDSCRegistrationKey(body []byte, xmsDate, authHeader, registrationKeyPlain string) (bool, string) {
	registrationKey := []byte(registrationKeyPlain)
	// Step 1: SHA256 hash of body
	sha256Hash := sha256.Sum256(body)
	bodyHashB64 := base64.StdEncoding.EncodeToString(sha256Hash[:])
	// Step 2: stringToSign = base64(body hash) + "\n" + xmsDate
	stringToSign := bodyHashB64 + "\n" + xmsDate
	// Step 3: HMAC-SHA256 of stringToSign with registrationKey
	mac := hmac.New(sha256.New, registrationKey)
	mac.Write([]byte(stringToSign))
	expectedSig := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	receivedSig := strings.TrimPrefix(authHeader, "Shared ")
	if !hmac.Equal([]byte(receivedSig), []byte(expectedSig)) {
		return false, "[REGISTER][AUTH] invalid DSC registration signature"
	}
	return true, ""
}
