// Package quality exposes quality-check definitions and sample status output.
// This gives the first UI slice concrete status information before persistent
// quality result storage is introduced.
package quality

// CheckStatus summarizes the current state of a quality check.
type CheckStatus struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

// Service owns quality-check status retrieval.
type Service struct{}

// NewService creates a quality service.
func NewService() *Service {
	return &Service{}
}

// ListStatuses returns representative quality data for the first vertical slice.
func (s *Service) ListStatuses() []CheckStatus {
	return []CheckStatus{
		{
			ID:       "check_duplicate_transactions",
			Name:     "Duplicate Transactions",
			Status:   "passing",
			Severity: "high",
			Message:  "No duplicate transaction identifiers detected in the latest run.",
		},
		{
			ID:       "check_uncategorized_transactions",
			Name:     "Uncategorized Transactions",
			Status:   "warning",
			Severity: "medium",
			Message:  "Three recent transactions still need category assignment.",
		},
	}
}
