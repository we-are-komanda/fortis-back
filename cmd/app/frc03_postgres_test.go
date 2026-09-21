package main

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"testing"

	ea "github.com/fortis/backend/internal/modules/enterprise/application"
	ed "github.com/fortis/backend/internal/modules/enterprise/domain"
	ei "github.com/fortis/backend/internal/modules/enterprise/infrastructure"
	ua "github.com/fortis/backend/internal/modules/user/application"
	ui "github.com/fortis/backend/internal/modules/user/infrastructure"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func frc03Command(userID string) ea.ProvisionEnterpriseCommand {
	zero := 0.0
	return ea.ProvisionEnterpriseCommand{EnterpriseID: uuid.NewString(), UserID: userID, Name: "Synthetic FRC03 enterprise", Address: "Synthetic FRC03 address", Latitude: &zero, Longitude: &zero}
}

func TestFRC03PostgresProvision(t *testing.T) {
	db := frc04Database(t)
	ctx := context.Background()
	users := ua.NewUserService(ui.NewUserRepository(db), "synthetic-frc03-secret", 24)
	a, _, err := users.Register(ctx, "frc03-a@example.test", "fixture-password-only", "FRC03 A")
	require.NoError(t, err)
	b, _, err := users.Register(ctx, "frc03-b@example.test", "fixture-password-only", "FRC03 B")
	require.NoError(t, err)
	service := ea.NewProvisionService(ei.NewEnterpriseProvisionRepository(db))
	count := func(table string) int64 { var n int64; require.NoError(t, db.Table(table).Count(&n).Error); return n }
	command := frc03Command(a.ID())

	t.Run("missing user rolls back", func(t *testing.T) {
		missing := frc03Command(uuid.NewString())
		_, err := service.Provision(ctx, missing)
		require.ErrorIs(t, err, ed.ErrProvisionUserNotFound)
		require.Zero(t, count("enterprises"))
		require.Zero(t, count("user_enterprises"))
	})
	t.Run("membership failure rolls back enterprise", func(t *testing.T) {
		require.NoError(t, db.Callback().Create().Before("gorm:create").Register("frc03_membership_failure", func(tx *gorm.DB) {
			if tx.Statement.Table == "user_enterprises" {
				tx.AddError(errors.New("synthetic membership failure"))
			}
		}))
		_, err := service.Provision(ctx, command)
		require.ErrorIs(t, err, ea.ErrProvisionFailed)
		require.NoError(t, db.Callback().Create().Remove("frc03_membership_failure"))
		require.Zero(t, count("enterprises"))
		require.Zero(t, count("user_enterprises"))
	})
	t.Run("concurrent exact retries create one pair", func(t *testing.T) {
		start, failures := make(chan struct{}), make(chan error, 8)
		var wg sync.WaitGroup
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func() { defer wg.Done(); <-start; _, err := service.Provision(ctx, command); failures <- err }()
		}
		close(start)
		wg.Wait()
		close(failures)
		for err := range failures {
			require.NoError(t, err)
		}
		require.EqualValues(t, 1, count("enterprises"))
		require.EqualValues(t, 1, count("user_enterprises"))
		enterprise, err := ei.NewEnterpriseRepository(db).FindByID(ctx, command.EnterpriseID)
		require.NoError(t, err)
		require.Zero(t, enterprise.Latitude())
		require.Zero(t, enterprise.Longitude())
		list, total, err := ei.NewEnterpriseRepository(db).FindAllByUserID(ctx, a.ID(), 20, 0)
		require.NoError(t, err)
		require.EqualValues(t, 1, total)
		require.Equal(t, command.EnterpriseID, list[0].ID())
		_, total, err = ei.NewEnterpriseRepository(db).FindAllByUserID(ctx, b.ID(), 20, 0)
		require.NoError(t, err)
		require.Zero(t, total)
	})
	t.Run("successful retry does not change data or log input", func(t *testing.T) {
		var before, after ei.EnterpriseModel
		require.NoError(t, db.Where("id = ?", command.EnterpriseID).Take(&before).Error)
		var output bytes.Buffer
		verbose := db.Session(&gorm.Session{Logger: logger.New(log.New(&output, "", 0), logger.Config{LogLevel: logger.Info})})
		_, err := ea.NewProvisionService(ei.NewEnterpriseProvisionRepository(verbose)).Provision(ctx, command)
		require.NoError(t, err)
		require.Empty(t, output.String())
		require.NoError(t, db.Where("id = ?", command.EnterpriseID).Take(&after).Error)
		require.Equal(t, before, after)
	})
	t.Run("changed details or recipient conflict without mutations", func(t *testing.T) {
		for _, field := range []string{"name", "address", "latitude", "longitude", "user"} {
			changed := command
			switch field {
			case "name":
				changed.Name = "Other name"
			case "address":
				changed.Address = "Other address"
			case "latitude":
				value := 1.0
				changed.Latitude = &value
			case "longitude":
				value := 1.0
				changed.Longitude = &value
			case "user":
				changed.UserID = b.ID()
			}
			_, err := service.Provision(ctx, changed)
			require.ErrorIs(t, err, ed.ErrProvisionConflict, field)
		}
		require.EqualValues(t, 1, count("enterprises"))
		require.EqualValues(t, 1, count("user_enterprises"))
	})
	t.Run("retry cannot restore revoked membership", func(t *testing.T) {
		require.NoError(t, ei.NewEnterpriseRepository(db).RemoveUserFromEnterprise(ctx, a.ID(), command.EnterpriseID))
		_, err := service.Provision(ctx, command)
		require.ErrorIs(t, err, ed.ErrProvisionConflict)
		require.EqualValues(t, 1, count("enterprises"))
		require.Zero(t, count("user_enterprises"))
	})
}

func TestFRC03PostgresConfigurationName(t *testing.T) {
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	a := f.people[0]
	for _, name := range []string{"", " \t", strings.Repeat("я", 121), strings.Repeat("😀", 121)} {
		f.request(t, a, "POST", "/projects", map[string]any{"name": name, "enterpriseId": a.e, "projectJson": frc04Document(a.e)}, 400)
		f.request(t, a, "PUT", "/projects/update?id="+a.p, map[string]any{"name": name, "version": 1}, 400)
		f.request(t, a, "POST", "/projects/import", map[string]any{"projectJson": strings.Replace(frc04Document(a.e), `"projectName":`, `"name":`+js(name)+`,"projectName":`, 1)}, 400)
	}
	var count int64
	require.NoError(t, db.Table("defense_projects").Count(&count).Error)
	require.EqualValues(t, 2, count)
	require.NoError(t, db.Table("project_revisions").Count(&count).Error)
	require.EqualValues(t, 2, count)
	name := strings.Repeat("😀", 120)
	created := frc04Request(t, f, a, "POST", "/projects", "", map[string]any{"name": "  " + name + "  ", "enterpriseId": a.e, "projectJson": frc04Document(a.e)}, 200)
	require.Equal(t, name, created["name"])
	loaded := frc04Request(t, f, a, "GET", "/projects/get?id="+created["projectId"].(string), "", nil, 200)
	require.Equal(t, name, loaded["name"])
	// Inspect the wire bytes: decoding into map[string]any would round this integer to float64.
	require.Contains(t, f.request(t, a, "GET", "/projects/get?id="+created["projectId"].(string), nil, 200), "9007199254740993")
}
