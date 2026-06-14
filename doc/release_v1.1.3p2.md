# Release v1.1.3p2

Date: 2026-06-13

Patch de stabilisation axe sur la securite de l interface scheduler et la logique de statut apres reception des rapports agent.

## Changements

### Securite

- Correction d une sanitization HTML incomplete sur les attributs des boutons d action du scheduler.
- Ajout d un echappement adapte au contexte attribut pour les valeurs injectees dans data-task.

Fichier impacte:
- web/scheduler_config.js

### Comportement du statut agent apres SendReport

- last_communication est mise a jour pour chaque rapport recu (heartbeat).
- Pour les rapports avec metadata MOF:
  - statut success => state Success, has_error_last_report = 0
  - tout autre statut => state Failure, has_error_last_report = 1
- Pour les rapports sans metadata MOF:
  - si report.Errors contient des erreurs non vides => state Failure, has_error_last_report = 1
  - sinon, l etat agent n est pas modifie.

Fichier impacte:
- handlers/sendreport.go

## Base de donnees

- Ajout du script de migration pour la version DB:
  - db/migration_v1.1.3p2.sql

## Procedure de mise a jour

Executer ces etapes dans cet ordre:

1. Arret du service.
2. Execution du script SQL de migration:
   - db/migration_v1.1.3p2.sql
3. Ecrasement des fichiers existants avec ceux du ZIP de release.
4. Demarrage du service.

## Verifications post mise a jour (recommande)

- Verifier que dsc_infra_info.db_version vaut 1.1.3p2.
- Verifier que les actions de la page scheduler s affichent correctement.
- Verifier la logique de statut noeud:
  - MOF success => OK
  - MOF non success => KO
  - sans MOF + erreurs => KO
  - sans MOF + sans erreurs => inchange
