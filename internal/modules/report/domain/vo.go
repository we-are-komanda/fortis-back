package domain

type ReportKey struct {
	projectID      string
	projectVersion int
	hideCost       bool
	mapView        string
}

func NewReportKey(id string, version int, hidden bool, view string) (ReportKey, error) {
	if id == "" || version <= 0 || (view != "all" && view != "saved") {
		return ReportKey{}, ErrInvalidReport
	}
	return ReportKey{id, version, hidden, view}, nil
}
func (k ReportKey) ProjectID() string   { return k.projectID }
func (k ReportKey) ProjectVersion() int { return k.projectVersion }
func (k ReportKey) HideCost() bool      { return k.hideCost }
func (k ReportKey) MapView() string     { return k.mapView }
