package schema

// Pour insertion/lecture SQL (avec JSON pour les tableaux)
type DscReportDB struct {
	ID                  int64  // id en base
	AgentId             string
	JobId               string
	ReportFormatVersion string
	OperationType       string
	RefreshMode         string
	Status              string
	StartTime           string
	EndTime             string
	RebootRequested     string
	Errors              string // JSON array
	StatusData          string // JSON array
	AdditionalData      string // JSON array
	MofApplied          int    // 0 ou 1 : indique si un MOF a été appliqué
	CreatedAt           string
	RawJson             string
}
