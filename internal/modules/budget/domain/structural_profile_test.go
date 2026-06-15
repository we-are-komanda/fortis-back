package domain

import (
	"errors"
	"testing"
)

func TestNewStructuralEchelonProfile_Success(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		layerID        string
		layerCode      string
		layerName      string
		objectCount    int
		unitCount      int
		categoryCount  int
		conflictCount  int
		coveredCount   int
	}{
		{
			name:          "basic echelon profile",
			layerID:       "echelon-1",
			layerCode:     "detection",
			layerName:     "Обнаружение",
			objectCount:   5,
			unitCount:     12,
			categoryCount: 3,
			conflictCount: 1,
			coveredCount:  4,
		},
		{
			name:          "zero values",
			layerID:       "echelon-2",
			layerCode:     "suppression",
			layerName:     "Подавление",
			objectCount:   0,
			unitCount:     0,
			categoryCount: 0,
			conflictCount: 0,
			coveredCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile, err := NewStructuralEchelonProfile(
				tt.layerID, tt.layerCode, tt.layerName,
				tt.objectCount, tt.unitCount, tt.categoryCount, tt.conflictCount, tt.coveredCount,
			)
			if err != nil {
				t.Fatalf("NewStructuralEchelonProfile() unexpected error: %v", err)
			}

			if profile.LayerID() != tt.layerID {
				t.Errorf("LayerID() = %s, want %s", profile.LayerID(), tt.layerID)
			}
			if profile.LayerCode() != tt.layerCode {
				t.Errorf("LayerCode() = %s, want %s", profile.LayerCode(), tt.layerCode)
			}
			if profile.LayerName() != tt.layerName {
				t.Errorf("LayerName() = %s, want %s", profile.LayerName(), tt.layerName)
			}
			if profile.ObjectCount() != tt.objectCount {
				t.Errorf("ObjectCount() = %d, want %d", profile.ObjectCount(), tt.objectCount)
			}
			if profile.UnitCount() != tt.unitCount {
				t.Errorf("UnitCount() = %d, want %d", profile.UnitCount(), tt.unitCount)
			}
			if profile.CategoryCount() != tt.categoryCount {
				t.Errorf("CategoryCount() = %d, want %d", profile.CategoryCount(), tt.categoryCount)
			}
			if profile.ConflictCount() != tt.conflictCount {
				t.Errorf("ConflictCount() = %d, want %d", profile.ConflictCount(), tt.conflictCount)
			}
			if profile.CoveredObjCount() != tt.coveredCount {
				t.Errorf("CoveredObjCount() = %d, want %d", profile.CoveredObjCount(), tt.coveredCount)
			}
		})
	}
}

func TestNewStructuralEchelonProfile_InvalidLayerID(t *testing.T) {
	t.Parallel()

	_, err := NewStructuralEchelonProfile("", "code", "name", 1, 1, 1, 1, 1)
	if !errors.Is(err, ErrInvalidEchelonProfile) {
		t.Errorf("error = %v, want ErrInvalidEchelonProfile", err)
	}
}

func TestNewStructuralEchelonProfile_NegativeValuesClamped(t *testing.T) {
	t.Parallel()

	profile, err := NewStructuralEchelonProfile("echelon-1", "code", "name", -5, -3, -2, -1, -4)
	if err != nil {
		t.Fatalf("NewStructuralEchelonProfile() unexpected error: %v", err)
	}

	if profile.ObjectCount() != 0 {
		t.Errorf("ObjectCount() = %d, want 0", profile.ObjectCount())
	}
	if profile.UnitCount() != 0 {
		t.Errorf("UnitCount() = %d, want 0", profile.UnitCount())
	}
	if profile.CategoryCount() != 0 {
		t.Errorf("CategoryCount() = %d, want 0", profile.CategoryCount())
	}
}

func TestNewStructuralProfile_Success(t *testing.T) {
	t.Parallel()

	byEchelon := []StructuralEchelonProfile{
		mustEchelonProfile(t, "echelon-1", "detection", "Обнаружение", 5, 10, 3, 1, 4),
		mustEchelonProfile(t, "echelon-2", "suppression", "Подавление", 3, 6, 2, 0, 3),
	}

	profile := NewStructuralProfile(8, 16, 2, 3, 1, 7, 500.0, byEchelon)

	if profile.ObjectCount() != 8 {
		t.Errorf("ObjectCount() = %d, want 8", profile.ObjectCount())
	}
	if profile.UnitCount() != 16 {
		t.Errorf("UnitCount() = %d, want 16", profile.UnitCount())
	}
	if profile.EchelonCount() != 2 {
		t.Errorf("EchelonCount() = %d, want 2", profile.EchelonCount())
	}
	if profile.CategoryCount() != 3 {
		t.Errorf("CategoryCount() = %d, want 3", profile.CategoryCount())
	}
	if profile.ConflictCount() != 1 {
		t.Errorf("ConflictCount() = %d, want 1", profile.ConflictCount())
	}
	if profile.CoveredObjCount() != 7 {
		t.Errorf("CoveredObjCount() = %d, want 7", profile.CoveredObjCount())
	}
	if profile.TotalMln() != 500.0 {
		t.Errorf("TotalMln() = %f, want 500.0", profile.TotalMln())
	}
	if len(profile.ByEchelon()) != 2 {
		t.Errorf("ByEchelon() length = %d, want 2", len(profile.ByEchelon()))
	}
}

func TestNewStructuralProfile_Defaults(t *testing.T) {
	t.Parallel()

	// zero values
	profile := NewStructuralProfile(0, 0, 0, 0, 0, 0, 0, nil)

	if profile.ObjectCount() != 0 {
		t.Errorf("ObjectCount() = %d, want 0", profile.ObjectCount())
	}
	if len(profile.ByEchelon()) != 0 {
		t.Errorf("ByEchelon() length = %d, want 0", len(profile.ByEchelon()))
	}
}

func TestNewStructuralProfile_NegativeValuesClamped(t *testing.T) {
	t.Parallel()

	profile := NewStructuralProfile(-1, -2, -3, -4, -5, -6, -7, nil)
	if profile.ObjectCount() != 0 {
		t.Errorf("ObjectCount() = %d, want 0", profile.ObjectCount())
	}
	if profile.UnitCount() != 0 {
		t.Errorf("UnitCount() = %d, want 0", profile.UnitCount())
	}
	if profile.EchelonCount() != 0 {
		t.Errorf("EchelonCount() = %d, want 0", profile.EchelonCount())
	}
	if profile.CategoryCount() != 0 {
		t.Errorf("CategoryCount() = %d, want 0", profile.CategoryCount())
	}
	if profile.ConflictCount() != 0 {
		t.Errorf("ConflictCount() = %d, want 0", profile.ConflictCount())
	}
	if profile.CoveredObjCount() != 0 {
		t.Errorf("CoveredObjCount() = %d, want 0", profile.CoveredObjCount())
	}
	if profile.TotalMln() != 0 {
		t.Errorf("TotalMln() = %f, want 0", profile.TotalMln())
	}
}

func TestNewEchelonDiff(t *testing.T) {
	t.Parallel()

	diff := NewEchelonDiff("echelon-1", "detection", "Обнаружение", 2, 5, 1, 0, 2)

	if diff.LayerID() != "echelon-1" {
		t.Errorf("LayerID() = %s, want echelon-1", diff.LayerID())
	}
	if diff.LayerCode() != "detection" {
		t.Errorf("LayerCode() = %s, want detection", diff.LayerCode())
	}
	if diff.LayerName() != "Обнаружение" {
		t.Errorf("LayerName() = %s", diff.LayerName())
	}
	if diff.ObjectCountDelta() != 2 {
		t.Errorf("ObjectCountDelta() = %d, want 2", diff.ObjectCountDelta())
	}
	if diff.UnitCountDelta() != 5 {
		t.Errorf("UnitCountDelta() = %d, want 5", diff.UnitCountDelta())
	}
	if diff.CategoryCountDelta() != 1 {
		t.Errorf("CategoryCountDelta() = %d, want 1", diff.CategoryCountDelta())
	}
	if diff.ConflictCountDelta() != 0 {
		t.Errorf("ConflictCountDelta() = %d, want 0", diff.ConflictCountDelta())
	}
	if diff.CoveredObjDelta() != 2 {
		t.Errorf("CoveredObjDelta() = %d, want 2", diff.CoveredObjDelta())
	}
}

func TestNewConfigDiff(t *testing.T) {
	t.Parallel()

	byEchelon := []EchelonDiff{
		NewEchelonDiff("echelon-1", "detection", "Обнаружение", 2, 4, 1, 0, 1),
	}

	diff := NewConfigDiff(2, 4, 1, 1, 0, 1, 150.0, byEchelon)

	if diff.ObjectCountDelta() != 2 {
		t.Errorf("ObjectCountDelta() = %d, want 2", diff.ObjectCountDelta())
	}
	if diff.UnitCountDelta() != 4 {
		t.Errorf("UnitCountDelta() = %d, want 4", diff.UnitCountDelta())
	}
	if diff.EchelonCountDelta() != 1 {
		t.Errorf("EchelonCountDelta() = %d, want 1", diff.EchelonCountDelta())
	}
	if diff.CategoryCountDelta() != 1 {
		t.Errorf("CategoryCountDelta() = %d, want 1", diff.CategoryCountDelta())
	}
	if diff.ConflictCountDelta() != 0 {
		t.Errorf("ConflictCountDelta() = %d, want 0", diff.ConflictCountDelta())
	}
	if diff.CoveredObjCountDelta() != 1 {
		t.Errorf("CoveredObjCountDelta() = %d, want 1", diff.CoveredObjCountDelta())
	}
	if diff.CostDeltaMln() != 150.0 {
		t.Errorf("CostDeltaMln() = %f, want 150.0", diff.CostDeltaMln())
	}
	if len(diff.ByEchelon()) != 1 {
		t.Errorf("ByEchelon() length = %d, want 1", len(diff.ByEchelon()))
	}
}

func TestNewConfigDiff_NilByEchelon(t *testing.T) {
	t.Parallel()

	diff := NewConfigDiff(0, 0, 0, 0, 0, 0, 0, nil)
	if diff.ByEchelon() == nil {
		t.Error("ByEchelon() should not be nil")
	}
	if len(diff.ByEchelon()) != 0 {
		t.Errorf("ByEchelon() length = %d, want 0", len(diff.ByEchelon()))
	}
}

func TestNewConfigSnapshot(t *testing.T) {
	t.Parallel()

	calc := NewCostCalculation(100, nil, nil, nil)
	profile := NewStructuralProfile(5, 10, 1, 3, 1, 4, 100, nil)
	snapshot := NewConfigSnapshot("proj-1", "Config A", profile, calc)

	if snapshot.ProjectID() != "proj-1" {
		t.Errorf("ProjectID() = %s, want proj-1", snapshot.ProjectID())
	}
	if snapshot.ProjectName() != "Config A" {
		t.Errorf("ProjectName() = %s, want Config A", snapshot.ProjectName())
	}
	if snapshot.StructuralProfile().ObjectCount() != 5 {
		t.Errorf("StructuralProfile().ObjectCount() = %d, want 5", snapshot.StructuralProfile().ObjectCount())
	}
	if snapshot.CostCalculation().TotalMln() != 100 {
		t.Errorf("CostCalculation().TotalMln() = %f, want 100", snapshot.CostCalculation().TotalMln())
	}
}

func TestNewConfigComparison(t *testing.T) {
	t.Parallel()

	calcA := NewCostCalculation(100, nil, nil, nil)
	calcB := NewCostCalculation(200, nil, nil, nil)
	profA := NewStructuralProfile(5, 10, 1, 3, 1, 4, 100, nil)
	profB := NewStructuralProfile(8, 15, 2, 4, 2, 6, 200, nil)

	snapA := NewConfigSnapshot("proj-1", "Config A", profA, calcA)
	snapB := NewConfigSnapshot("proj-2", "Config B", profB, calcB)
	diff := NewConfigDiff(3, 5, 1, 1, 1, 2, 100, nil)

	comp := NewConfigComparison(snapA, snapB, diff)

	if comp.ProjectA().ProjectID() != "proj-1" {
		t.Errorf("ProjectA().ProjectID() = %s, want proj-1", comp.ProjectA().ProjectID())
	}
	if comp.ProjectB().ProjectID() != "proj-2" {
		t.Errorf("ProjectB().ProjectID() = %s, want proj-2", comp.ProjectB().ProjectID())
	}
	if comp.Diff().ObjectCountDelta() != 3 {
		t.Errorf("Diff().ObjectCountDelta() = %d, want 3", comp.Diff().ObjectCountDelta())
	}
}

func mustEchelonProfile(t *testing.T, layerID, layerCode, layerName string, objCount, unitCount, catCount, confCount, coveredCount int) StructuralEchelonProfile {
	t.Helper()
	p, err := NewStructuralEchelonProfile(layerID, layerCode, layerName, objCount, unitCount, catCount, confCount, coveredCount)
	if err != nil {
		t.Fatalf("NewStructuralEchelonProfile() error: %v", err)
	}
	return p
}
