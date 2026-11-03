package application

// ProvisionEnterpriseCommand is a trusted local operator operation, never a member API operation.
// A stable EnterpriseID allows retry without creating another enterprise.
type ProvisionEnterpriseCommand struct {
	EnterpriseID string
	UserID       string
	Name         string
	Address      string
	Latitude     *float64
	Longitude    *float64
}
