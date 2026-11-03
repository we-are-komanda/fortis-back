package main

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestDefaultEnterpriseSeedIsIdempotent(t *testing.T) {
	db := frc02TestDatabase(t)
	fixture := newPostgresFixture(t, db)
	sql, err := os.ReadFile("../../migrations/1792270100_test_enterprise.up.sql")
	require.NoError(t, err)
	for i := 0; i < 2; i++ {
		require.NoError(t, db.Exec(string(sql)).Error)
	}
	var count int64
	require.NoError(t, db.Table("enterprises").Where("id = ?", "f0711500-0000-4000-8000-000000000001").Count(&count).Error)
	require.EqualValues(t, 1, count)
	for _, person := range fixture.people {
		require.NoError(t, db.Table("enterprises").Where("id = ?", person.e).Count(&count).Error)
		require.EqualValues(t, 1, count)
	}
}
