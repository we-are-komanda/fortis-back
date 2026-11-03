package application

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/fortis/backend/internal/modules/defense_project/domain"
	"github.com/stretchr/testify/require"
)

func projectJSONWithName(t *testing.T, name string) string {
	t.Helper()
	var value map[string]any
	require.NoError(t, json.Unmarshal([]byte(validProjectJSON()), &value))
	value["name"] = name
	encoded, err := json.Marshal(value)
	require.NoError(t, err)
	return string(encoded)
}

func TestConfigurationNameWriteBoundaries(t *testing.T) {
	ctx := context.Background()
	for _, name := range []string{"", " \n\t", strings.Repeat("я", 121), strings.Repeat("😀", 121)} {
		t.Run("reject_"+name, func(t *testing.T) {
			repo := newMockRepo()
			service := NewDefenseProjectService(repo, testAccess{})
			_, err := service.CreateFromJSON(ctx, "actor", name, "ent-1", validProjectJSON())
			require.ErrorIs(t, err, domain.ErrInvalidConfigName)
			_, err = service.Import(ctx, "actor", projectJSONWithName(t, name))
			require.ErrorIs(t, err, domain.ErrInvalidConfigName)
			require.Empty(t, repo.projects)
		})
	}
	repo := newMockRepo()
	service := NewDefenseProjectService(repo, testAccess{})
	name := strings.Repeat("😀", 120)
	project, err := service.CreateFromJSON(ctx, "actor", "  "+name+"  ", "ent-1", validProjectJSON())
	require.NoError(t, err)
	require.Equal(t, name, project.Name())
	version := project.Version()
	_, err = service.UpdateProject(ctx, "actor", project.ProjectID(), " ", nil, "", &version)
	require.ErrorIs(t, err, domain.ErrInvalidConfigName)
	_, err = service.UpdateProject(ctx, "actor", project.ProjectID(), "", nil, projectJSONWithName(t, ""), &version)
	require.ErrorIs(t, err, domain.ErrInvalidConfigName)
	require.Equal(t, version, project.Version())
	updated, err := service.UpdateProject(ctx, "actor", project.ProjectID(), "  Changed  ", nil, "", &version)
	require.NoError(t, err)
	require.Equal(t, "Changed", updated.Name())
	imported, err := service.Import(ctx, "actor", projectJSONWithName(t, "  Imported  "))
	require.NoError(t, err)
	require.Equal(t, "Imported", imported.Name())
}

func TestConfigurationNameLegacyReadAndUnchangedContent(t *testing.T) {
	repo := newMockRepo()
	service := NewDefenseProjectService(repo, testAccess{})
	project, err := service.Import(context.Background(), "actor", validProjectJSON())
	require.NoError(t, err)
	// Older imports omitted configuration name; new imports use their existing business label.
	require.Equal(t, "Тестовый проект", project.Name())
	legacyName := "  " + strings.Repeat("я", 150) + "  "
	project.SetName(legacyName)
	loaded, err := service.GetProject(context.Background(), "actor", project.ProjectID())
	require.NoError(t, err)
	require.Equal(t, legacyName, loaded.Name())
	version := loaded.Version()
	updated, err := service.UpdateProject(context.Background(), "actor", loaded.ProjectID(), "", nil, projectJSONWithName(t, legacyName), &version)
	require.NoError(t, err)
	require.Equal(t, legacyName, updated.Name())
}
