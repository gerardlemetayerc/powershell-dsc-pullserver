package utils

import (
	"os"
	"path/filepath"
)

// ExePath retourne le chemin absolu de l'exécutable en cours d'exécution.
func ExePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Abs(exe)
}
