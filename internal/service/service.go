package service

import (
	"fmt"
	"go-dsc-pull/internal/logs"
	"golang.org/x/sys/windows/svc"
)
// WriteServiceLog écrit un message dans un fichier de log à côté du binaire
func WriteServiceLog(message string) error {
   return logs.WriteLogFile(message)
}

// StartWindowsService démarre l'application en tant que service Windows
func StartWindowsService(serviceName string, runFunc func()) error {
	isService, err := svc.IsWindowsService()
	if err != nil {
		return fmt.Errorf("[SERVICE] Erreur détection mode service: %v", err)
	}
	if isService {
		return svc.Run(serviceName, &appService{runFunc: runFunc})
	}
	// Si pas en mode service, exécute normalement
	runFunc()
	return nil
}

// appService implémente svc.Handler
type appService struct {
	runFunc func()
}

func (s *appService) Execute(args []string, r <-chan svc.ChangeRequest, status chan<- svc.Status) (bool, uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	status <- svc.Status{State: svc.StartPending}
	go s.runFunc()
	status <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	for {
		c := <-r
		switch c.Cmd {
		case svc.Interrogate:
			status <- c.CurrentStatus
		case svc.Stop, svc.Shutdown:
			status <- svc.Status{State: svc.StopPending}
			return false, 0
		}
	}
}
