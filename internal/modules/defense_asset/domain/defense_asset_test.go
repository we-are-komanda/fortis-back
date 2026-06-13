package domain

import (
	"errors"
	"testing"
	"time"
)

func validAssetBase() (string, string, string, string, DefenseAssetCategory, DefenseAssetCoverageType) {
	return "test-id", "Test Asset", "TA", "Test description",
		DefenseAssetCategoryRadar, DefenseAssetCoverageCircle
}

func TestNewDefenseAsset_WeaponSpec_Success(t *testing.T) {
	id, name, shortName, desc, _, ct := validAssetBase()
	cat := DefenseAssetCategoryArtillery

	caliber := "152mm"
	ammo := "HE"
	mode := "auto"
	modules := 2
	isManual := false

	ws := &WeaponSpecification{
		Caliber:        &caliber,
		AmmunitionType: &ammo,
		OperationMode:  &mode,
		ModuleCount:    &modules,
		IsManual:       &isManual,
	}

	now := time.Now().UTC()
	asset, err := NewDefenseAsset(
		id, name, shortName, desc, cat, nil, nil, "RUB", "", nil, nil, nil, nil, "",
		nil, nil, ct, nil, nil,
		DeploymentTypeStatic, PlacementTypeMapObject,
		"", "", nil, nil, nil,
		ws, nil, nil,
		nil, "", nil, nil,
		nil, false, now, now,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if asset.WeaponSpec() != ws {
		t.Error("weaponSpec not set correctly")
	}
	if asset.WeaponSpec().Caliber == nil || *asset.WeaponSpec().Caliber != "152mm" {
		t.Error("caliber not set correctly")
	}
}

func TestNewDefenseAsset_WeaponSpec_InvalidCategory(t *testing.T) {
	id, name, shortName, desc, cat, ct := validAssetBase()
	// Radar (from validAssetBase) is not a weapon category

	ws := &WeaponSpecification{}

	now := time.Now().UTC()
	_, err := NewDefenseAsset(
		id, name, shortName, desc, cat, nil, nil, "RUB", "", nil, nil, nil, nil, "",
		nil, nil, ct, nil, nil,
		DeploymentTypeStatic, PlacementTypeMapObject,
		"", "", nil, nil, nil,
		ws, nil, nil,
		nil, "", nil, nil,
		nil, false, now, now,
	)
	if !errors.Is(err, ErrDefenseAssetInvalidSpecification) {
		t.Errorf("expected ErrDefenseAssetInvalidSpecification, got %v", err)
	}
}

func TestNewDefenseAsset_DetectionSpec_Success(t *testing.T) {
	id, name, shortName, desc, _, ct := validAssetBase()
	cat := DefenseAssetCategoryRadiotechnical

	freq := "1-10 GHz"
	mode := "active"
	speed := 30.0
	fov := 120.0
	thermal := true

	ds := &DetectionSpecification{
		FrequencyRange:   &freq,
		DetectionMode:    &mode,
		RotationSpeed:    &speed,
		FieldOfView:      &fov,
		HasThermalImager: &thermal,
	}

	now := time.Now().UTC()
	asset, err := NewDefenseAsset(
		id, name, shortName, desc, cat, nil, nil, "RUB", "", nil, nil, nil, nil, "",
		nil, nil, ct, nil, nil,
		DeploymentTypeStatic, PlacementTypeMapObject,
		"", "", nil, nil, nil,
		nil, ds, nil,
		nil, "", nil, nil,
		nil, false, now, now,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if asset.DetectionSpec() != ds {
		t.Error("detectionSpec not set correctly")
	}
	if asset.DetectionSpec().HasThermalImager == nil || !*asset.DetectionSpec().HasThermalImager {
		t.Error("thermal imager flag not set correctly")
	}
}

func TestNewDefenseAsset_DetectionSpec_InvalidCategory(t *testing.T) {
	id, name, shortName, desc, _, ct := validAssetBase()
	cat := DefenseAssetCategoryEW

	ds := &DetectionSpecification{}

	now := time.Now().UTC()
	_, err := NewDefenseAsset(
		id, name, shortName, desc, cat, nil, nil, "RUB", "", nil, nil, nil, nil, "",
		nil, nil, ct, nil, nil,
		DeploymentTypeStatic, PlacementTypeMapObject,
		"", "", nil, nil, nil,
		nil, ds, nil,
		nil, "", nil, nil,
		nil, false, now, now,
	)
	if !errors.Is(err, ErrDefenseAssetInvalidSpecification) {
		t.Errorf("expected ErrDefenseAssetInvalidSpecification, got %v", err)
	}
}

func TestNewDefenseAsset_EWSpec_Success(t *testing.T) {
	id, name, shortName, desc, _, ct := validAssetBase()
	cat := DefenseAssetCategoryEW

	freq := "2-18 GHz"
	actionRange := 50.0
	azimuth := 120.0

	ews := &EWSpecification{
		FrequencyRange: &freq,
		ActionRange:    &actionRange,
		Azimuth:        &azimuth,
	}

	now := time.Now().UTC()
	asset, err := NewDefenseAsset(
		id, name, shortName, desc, cat, nil, nil, "RUB", "", nil, nil, nil, nil, "",
		nil, nil, ct, nil, nil,
		DeploymentTypeStatic, PlacementTypeMapObject,
		"", "", nil, nil, nil,
		nil, nil, ews,
		nil, "", nil, nil,
		nil, false, now, now,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if asset.EWSpec() != ews {
		t.Error("ewSpec not set correctly")
	}
	if asset.EWSpec().ActionRange == nil || *asset.EWSpec().ActionRange != 50.0 {
		t.Error("action range not set correctly")
	}
}

func TestNewDefenseAsset_EWSpec_InvalidCategory(t *testing.T) {
	id, name, shortName, desc, _, ct := validAssetBase()
	cat := DefenseAssetCategoryArtillery

	ews := &EWSpecification{}

	now := time.Now().UTC()
	_, err := NewDefenseAsset(
		id, name, shortName, desc, cat, nil, nil, "RUB", "", nil, nil, nil, nil, "",
		nil, nil, ct, nil, nil,
		DeploymentTypeStatic, PlacementTypeMapObject,
		"", "", nil, nil, nil,
		nil, nil, ews,
		nil, "", nil, nil,
		nil, false, now, now,
	)
	if !errors.Is(err, ErrDefenseAssetInvalidSpecification) {
		t.Errorf("expected ErrDefenseAssetInvalidSpecification, got %v", err)
	}
}

func TestDefenseAssetCompoundProfile_NewFields(t *testing.T) {
	now := time.Now().UTC()

	cp := &DefenseAssetCompoundProfile{
		Kind:           "compound-post",
		PostType:       "МОГ",
		PersonnelCount: "4",
		Accountability: "МО",
		Armament:       "Автомат/пулемёт/ПБС",
		WeaponUnits:    "2",
		SectorOrRange:  "до 4-8 км, сектор 90-360°",
		Azimuth:        45.0,
	}

	asset, err := NewDefenseAsset(
		"test-id", "Test", "", "", DefenseAssetCategoryFortification,
		nil, nil, "RUB", "", nil, nil, nil, nil, "",
		nil, nil, DefenseAssetCoverageNone, nil, nil,
		DeploymentTypeStatic, PlacementTypeMapObject,
		"", "", nil, nil, cp,
		nil, nil, nil,
		nil, "", nil, nil,
		nil, false, now, now,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if asset.CompoundProfile().Kind != "compound-post" {
		t.Errorf("expected kind 'compound-post', got %q", asset.CompoundProfile().Kind)
	}
	if asset.CompoundProfile().PostType != "МОГ" {
		t.Errorf("expected postType 'МОГ', got %q", asset.CompoundProfile().PostType)
	}
	if asset.CompoundProfile().PersonnelCount != "4" {
		t.Errorf("expected personnelCount '4', got %q", asset.CompoundProfile().PersonnelCount)
	}
	if asset.CompoundProfile().Accountability != "МО" {
		t.Errorf("expected accountability 'МО', got %q", asset.CompoundProfile().Accountability)
	}
	if asset.CompoundProfile().Armament != "Автомат/пулемёт/ПБС" {
		t.Errorf("expected armament 'Автомат/пулемёт/ПБС', got %q", asset.CompoundProfile().Armament)
	}
	if asset.CompoundProfile().WeaponUnits != "2" {
		t.Errorf("expected weaponUnits '2', got %q", asset.CompoundProfile().WeaponUnits)
	}
	if asset.CompoundProfile().SectorOrRange != "до 4-8 км, сектор 90-360°" {
		t.Errorf("expected sectorOrRange, got %q", asset.CompoundProfile().SectorOrRange)
	}
	if asset.CompoundProfile().Azimuth != 45.0 {
		t.Errorf("expected azimuth 45.0, got %f", asset.CompoundProfile().Azimuth)
	}
}

func TestSetWeaponSpec(t *testing.T) {
	now := time.Now().UTC()
	asset, err := NewDefenseAsset(
		"test-id", "Test", "", "", DefenseAssetCategoryRadar,
		nil, nil, "RUB", "", nil, nil, nil, nil, "",
		nil, nil, DefenseAssetCoverageCircle, nil, nil,
		DeploymentTypeStatic, PlacementTypeMapObject,
		"", "", nil, nil, nil,
		nil, nil, nil,
		nil, "", nil, nil,
		nil, false, now, now,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	caliber := "152mm"
	ws := &WeaponSpecification{Caliber: &caliber}
	asset.SetWeaponSpec(ws)

	if asset.WeaponSpec() != ws {
		t.Error("WeaponSpec not set via setter")
	}
	if *asset.WeaponSpec().Caliber != "152mm" {
		t.Error("caliber not set via setter")
	}
}

func TestSetDetectionSpec(t *testing.T) {
	now := time.Now().UTC()
	asset, err := NewDefenseAsset(
		"test-id", "Test", "", "", DefenseAssetCategoryRadar,
		nil, nil, "RUB", "", nil, nil, nil, nil, "",
		nil, nil, DefenseAssetCoverageCircle, nil, nil,
		DeploymentTypeStatic, PlacementTypeMapObject,
		"", "", nil, nil, nil,
		nil, nil, nil,
		nil, "", nil, nil,
		nil, false, now, now,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	freq := "1-10 GHz"
	ds := &DetectionSpecification{FrequencyRange: &freq}
	asset.SetDetectionSpec(ds)

	if asset.DetectionSpec() != ds {
		t.Error("DetectionSpec not set via setter")
	}
}

func TestSetEWSpec(t *testing.T) {
	now := time.Now().UTC()
	asset, err := NewDefenseAsset(
		"test-id", "Test", "", "", DefenseAssetCategoryRadar,
		nil, nil, "RUB", "", nil, nil, nil, nil, "",
		nil, nil, DefenseAssetCoverageCircle, nil, nil,
		DeploymentTypeStatic, PlacementTypeMapObject,
		"", "", nil, nil, nil,
		nil, nil, nil,
		nil, "", nil, nil,
		nil, false, now, now,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	freq := "2-18 GHz"
	ews := &EWSpecification{FrequencyRange: &freq}
	asset.SetEWSpec(ews)

	if asset.EWSpec() != ews {
		t.Error("EWSpec not set via setter")
	}
}
