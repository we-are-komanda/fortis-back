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

func TestDefaultEnterpriseRelocationIsIdempotent(t *testing.T) {
	db := frc02TestDatabase(t)
	seed, err := os.ReadFile("../../migrations/1792270100_test_enterprise.up.sql")
	require.NoError(t, err)
	require.NoError(t, db.Exec(string(seed)).Error)
	migration, err := os.ReadFile("../../migrations/1793700000_test_enterprise_new_zealand.up.sql")
	require.NoError(t, err)
	for i := 0; i < 2; i++ {
		require.NoError(t, db.Exec(string(migration)).Error)
	}
	var site struct {
		Latitude  float64
		Longitude float64
	}
	id := "f0711500-0000-4000-8000-000000000001"
	require.NoError(t, db.Table("enterprises").Where("id = ?", id).Take(&site).Error)
	require.Equal(t, -41.2865, site.Latitude)
	require.Equal(t, 174.7762, site.Longitude)
	// A user-customized location must not be overwritten when applying the migration again.
	require.NoError(t, db.Table("enterprises").Where("id = ?", id).Update("latitude", -42).Error)
	require.NoError(t, db.Exec(string(migration)).Error)
	require.NoError(t, db.Table("enterprises").Where("id = ?", id).Take(&site).Error)
	require.Equal(t, float64(-42), site.Latitude)
}
