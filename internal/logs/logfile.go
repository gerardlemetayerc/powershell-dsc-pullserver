package logs

import (
    "fmt"
    "os"
    "path/filepath"
	"time"
    "go-dsc-pull/utils"
)

// WriteLogFile écrit un message dans un fichier de log à côté du binaire
func WriteLogFile(message string) error {
    exePath, err := utils.ExePath()
    if err != nil {
        return fmt.Errorf("[LOG] Impossible de localiser l'exécutable: %v", err)
    }
    logPath := filepath.Join(filepath.Dir(exePath), "dsc-pull-server.log")
    f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return fmt.Errorf("[LOG] Impossible d'ouvrir le fichier log: %v", err)
    }
    defer f.Close()
    now := time.Now().Format("2006-01-02 15:04:05")
    logLine := fmt.Sprintf("[%s] %s\n", now, message)
    if _, err := f.WriteString(logLine); err != nil {
        return fmt.Errorf("[LOG] Impossible d'écrire dans le fichier log: %v", err)
    }
    return nil
}