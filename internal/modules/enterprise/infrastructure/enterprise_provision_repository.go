package infrastructure

import (
	"context"
	"errors"

	"github.com/fortis/backend/internal/modules/enterprise/domain"
	userInfra "github.com/fortis/backend/internal/modules/user/infrastructure"
	"github.com/fortis/backend/internal/rdbms"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

func NewEnterpriseProvisionRepository(executor rdbms.Executor) domain.EnterpriseProvisionRepository {
	return &EnterpriseRepository{executor: executor}
}

func (r *EnterpriseRepository) Provision(ctx context.Context, enterprise *domain.Enterprise, userID string) error {
	// Runtime SQL logging must not expose the operator's name/address or user details.
	db := r.executor.WithContext(ctx).Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})
	return db.Transaction(func(tx *gorm.DB) error {
		var user userInfra.UserModel
		err := tx.Select("id").Clauses(clause.Locking{Strength: "KEY SHARE"}).Where("id = ?", userID).Take(&user).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrProvisionUserNotFound
		}
		if err != nil {
			return err
		}

		model := ToModel(enterprise)
		insert := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoNothing: true}).Create(model)
		if insert.Error != nil {
			return insert.Error
		}
		if insert.RowsAffected == 1 {
			return tx.Create(&userEnterpriseModel{UserID: userID, EnterpriseID: enterprise.ID()}).Error
		}

		var existing EnterpriseModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", enterprise.ID()).Take(&existing).Error; err != nil {
			return err
		}
		stored, err := domain.NewEnterprise(existing.ID, existing.Name, existing.Address, domain.EnterpriseStatus(existing.Status), existing.Latitude, existing.Longitude, existing.CreatedAt, existing.UpdatedAt)
		if err != nil || !enterprise.MatchesProvision(stored) {
			return domain.ErrProvisionConflict
		}
		var member userEnterpriseModel
		err = tx.Clauses(clause.Locking{Strength: "KEY SHARE"}).Where("user_id = ? AND enterprise_id = ?", userID, enterprise.ID()).Take(&member).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrProvisionConflict
		}
		return err
	})
}
