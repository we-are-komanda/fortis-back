package application

type SubmitCommand struct {
	Name, Organization, Email, Comment, ConsentVersion, IdempotencyKey string
	Consent                                                            bool
}
type Accepted struct {
	RequestID string
	Replayed  bool
}
type Config struct {
	Enabled         bool
	ConsentVersions []string
}
