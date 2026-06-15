package domain

// UserEnterprise — value object для связи пользователя и предприятия.
type UserEnterprise struct {
	UserID       string
	EnterpriseID string
}

// NewUserEnterprise создаёт новую связь пользователя и предприятия.
func NewUserEnterprise(userID, enterpriseID string) *UserEnterprise {
	return &UserEnterprise{
		UserID:       userID,
		EnterpriseID: enterpriseID,
	}
}
