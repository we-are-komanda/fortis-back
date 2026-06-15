package infrastructure

import (
	"time"

	"github.com/fortis/backend/internal/modules/user/domain"
)

// UserModel — GORM-модель для хранения User.
type UserModel struct {
	ID           string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Email        string    `gorm:"uniqueIndex;not null;size:255"`
	PasswordHash string    `gorm:"not null;size:255"`
	Name         string    `gorm:"not null;default:'';size:255"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

func (UserModel) TableName() string {
	return "users"
}

// ToDomain преобразует GORM-модель в доменный агрегат.
func (m *UserModel) ToDomain() *domain.User {
	u, err := domain.NewUser(
		m.ID,
		m.Email,
		m.PasswordHash,
		m.Name,
		m.CreatedAt,
		m.UpdatedAt,
	)
	if err != nil {
		// При чтении из БД данные уже валидны, ошибка невозможна
		panic("invalid user data in database: " + err.Error())
	}
	return u
}

// ToModel преобразует доменный агрегат в GORM-модель.
func ToModel(u *domain.User) *UserModel {
	return &UserModel{
		ID:           u.ID(),
		Email:        u.Email(),
		PasswordHash: u.PasswordHash(),
		Name:         u.Name(),
		CreatedAt:    u.CreatedAt(),
		UpdatedAt:    u.UpdatedAt(),
	}
}
