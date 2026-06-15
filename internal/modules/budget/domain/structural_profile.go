package domain

// StructuralEchelonProfile — структурный профиль одного эшелона (value object).
type StructuralEchelonProfile struct {
	layerID         string
	layerCode       string
	layerName       string
	objectCount     int
	unitCount       int
	categoryCount   int
	conflictCount   int
	coveredObjCount int
}

// NewStructuralEchelonProfile создаёт новый профиль эшелона с валидацией.
func NewStructuralEchelonProfile(
	layerID, layerCode, layerName string,
	objectCount, unitCount, categoryCount, conflictCount, coveredObjCount int,
) (StructuralEchelonProfile, error) {
	if layerID == "" {
		return StructuralEchelonProfile{}, ErrInvalidEchelonProfile
	}
	if objectCount < 0 {
		objectCount = 0
	}
	if unitCount < 0 {
		unitCount = 0
	}
	if categoryCount < 0 {
		categoryCount = 0
	}
	if conflictCount < 0 {
		conflictCount = 0
	}
	if coveredObjCount < 0 {
		coveredObjCount = 0
	}

	return StructuralEchelonProfile{
		layerID:         layerID,
		layerCode:       layerCode,
		layerName:       layerName,
		objectCount:     objectCount,
		unitCount:       unitCount,
		categoryCount:   categoryCount,
		conflictCount:   conflictCount,
		coveredObjCount: coveredObjCount,
	}, nil
}

// LayerID возвращает ID эшелона.
func (p StructuralEchelonProfile) LayerID() string { return p.layerID }

// LayerCode возвращает код эшелона.
func (p StructuralEchelonProfile) LayerCode() string { return p.layerCode }

// LayerName возвращает название эшелона.
func (p StructuralEchelonProfile) LayerName() string { return p.layerName }

// ObjectCount возвращает количество позиций.
func (p StructuralEchelonProfile) ObjectCount() int { return p.objectCount }

// UnitCount возвращает количество единиц.
func (p StructuralEchelonProfile) UnitCount() int { return p.unitCount }

// CategoryCount возвращает количество distinct категорий.
func (p StructuralEchelonProfile) CategoryCount() int { return p.categoryCount }

// ConflictCount возвращает количество конфликтов.
func (p StructuralEchelonProfile) ConflictCount() int { return p.conflictCount }

// CoveredObjCount возвращает количество объектов с покрытием.
func (p StructuralEchelonProfile) CoveredObjCount() int { return p.coveredObjCount }

// StructuralProfile — полный структурный профиль конфигурации (value object).
type StructuralProfile struct {
	objectCount     int
	unitCount       int
	echelonCount    int
	categoryCount   int
	conflictCount   int
	coveredObjCount int
	totalMln        float64
	byEchelon       []StructuralEchelonProfile
}

// NewStructuralProfile создаёт новый структурный профиль.
func NewStructuralProfile(
	objectCount, unitCount, echelonCount, categoryCount, conflictCount, coveredObjCount int,
	totalMln float64,
	byEchelon []StructuralEchelonProfile,
) StructuralProfile {
	if byEchelon == nil {
		byEchelon = []StructuralEchelonProfile{}
	}
	if objectCount < 0 {
		objectCount = 0
	}
	if unitCount < 0 {
		unitCount = 0
	}
	if echelonCount < 0 {
		echelonCount = 0
	}
	if categoryCount < 0 {
		categoryCount = 0
	}
	if conflictCount < 0 {
		conflictCount = 0
	}
	if coveredObjCount < 0 {
		coveredObjCount = 0
	}
	if totalMln < 0 {
		totalMln = 0
	}

	return StructuralProfile{
		objectCount:     objectCount,
		unitCount:       unitCount,
		echelonCount:    echelonCount,
		categoryCount:   categoryCount,
		conflictCount:   conflictCount,
		coveredObjCount: coveredObjCount,
		totalMln:        totalMln,
		byEchelon:       byEchelon,
	}
}

// ObjectCount возвращает количество позиций.
func (p StructuralProfile) ObjectCount() int { return p.objectCount }

// UnitCount возвращает количество единиц.
func (p StructuralProfile) UnitCount() int { return p.unitCount }

// EchelonCount возвращает количество эшелонов с объектами.
func (p StructuralProfile) EchelonCount() int { return p.echelonCount }

// CategoryCount возвращает количество distinct категорий.
func (p StructuralProfile) CategoryCount() int { return p.categoryCount }

// ConflictCount возвращает количество конфликтов.
func (p StructuralProfile) ConflictCount() int { return p.conflictCount }

// CoveredObjCount возвращает количество объектов с покрытием.
func (p StructuralProfile) CoveredObjCount() int { return p.coveredObjCount }

// TotalMln возвращает общую стоимость конфигурации в млн ₽.
func (p StructuralProfile) TotalMln() float64 { return p.totalMln }

// ByEchelon возвращает профили по эшелонам.
func (p StructuralProfile) ByEchelon() []StructuralEchelonProfile { return p.byEchelon }

// EchelonDiff — разница структурных метрик по одному эшелону (value object).
type EchelonDiff struct {
	layerID          string
	layerCode        string
	layerName        string
	objectCountDelta int
	unitCountDelta   int
	categoryCountDelta int
	conflictCountDelta int
	coveredObjDelta  int
}

// NewEchelonDiff создаёт новый EchelonDiff.
func NewEchelonDiff(
	layerID, layerCode, layerName string,
	objectCountDelta, unitCountDelta, categoryCountDelta, conflictCountDelta, coveredObjDelta int,
) EchelonDiff {
	return EchelonDiff{
		layerID:            layerID,
		layerCode:          layerCode,
		layerName:          layerName,
		objectCountDelta:   objectCountDelta,
		unitCountDelta:     unitCountDelta,
		categoryCountDelta: categoryCountDelta,
		conflictCountDelta: conflictCountDelta,
		coveredObjDelta:    coveredObjDelta,
	}
}

// LayerID возвращает ID эшелона.
func (d EchelonDiff) LayerID() string { return d.layerID }

// LayerCode возвращает код эшелона.
func (d EchelonDiff) LayerCode() string { return d.layerCode }

// LayerName возвращает название эшелона.
func (d EchelonDiff) LayerName() string { return d.layerName }

// ObjectCountDelta возвращает разницу в количестве позиций.
func (d EchelonDiff) ObjectCountDelta() int { return d.objectCountDelta }

// UnitCountDelta возвращает разницу в количестве единиц.
func (d EchelonDiff) UnitCountDelta() int { return d.unitCountDelta }

// CategoryCountDelta возвращает разницу в количестве категорий.
func (d EchelonDiff) CategoryCountDelta() int { return d.categoryCountDelta }

// ConflictCountDelta возвращает разницу в количестве конфликтов.
func (d EchelonDiff) ConflictCountDelta() int { return d.conflictCountDelta }

// CoveredObjDelta возвращает разницу в количестве покрытых объектов.
func (d EchelonDiff) CoveredObjDelta() int { return d.coveredObjDelta }

// ConfigDiff — разница структурных метрик между двумя конфигурациями (value object).
type ConfigDiff struct {
	objectCountDelta     int
	unitCountDelta       int
	echelonCountDelta    int
	categoryCountDelta    int
	conflictCountDelta    int
	coveredObjCountDelta  int
	costDeltaMln         float64
	byEchelon            []EchelonDiff
}

// NewConfigDiff создаёт новый ConfigDiff.
func NewConfigDiff(
	objectCountDelta, unitCountDelta, echelonCountDelta, categoryCountDelta, conflictCountDelta, coveredObjCountDelta int,
	costDeltaMln float64,
	byEchelon []EchelonDiff,
) ConfigDiff {
	if byEchelon == nil {
		byEchelon = []EchelonDiff{}
	}
	return ConfigDiff{
		objectCountDelta:     objectCountDelta,
		unitCountDelta:       unitCountDelta,
		echelonCountDelta:    echelonCountDelta,
		categoryCountDelta:    categoryCountDelta,
		conflictCountDelta:    conflictCountDelta,
		coveredObjCountDelta:  coveredObjCountDelta,
		costDeltaMln:         costDeltaMln,
		byEchelon:            byEchelon,
	}
}

// ObjectCountDelta возвращает разницу в количестве позиций.
func (d ConfigDiff) ObjectCountDelta() int { return d.objectCountDelta }

// UnitCountDelta возвращает разницу в количестве единиц.
func (d ConfigDiff) UnitCountDelta() int { return d.unitCountDelta }

// EchelonCountDelta возвращает разницу в количестве эшелонов.
func (d ConfigDiff) EchelonCountDelta() int { return d.echelonCountDelta }

// CategoryCountDelta возвращает разницу в количестве категорий.
func (d ConfigDiff) CategoryCountDelta() int { return d.categoryCountDelta }

// ConflictCountDelta возвращает разницу в количестве конфликтов.
func (d ConfigDiff) ConflictCountDelta() int { return d.conflictCountDelta }

// CoveredObjCountDelta возвращает разницу в количестве покрытых объектов.
func (d ConfigDiff) CoveredObjCountDelta() int { return d.coveredObjCountDelta }

// CostDeltaMln возвращает разницу в стоимости в млн ₽.
func (d ConfigDiff) CostDeltaMln() float64 { return d.costDeltaMln }

// ByEchelon возвращает разницу по эшелонам.
func (d ConfigDiff) ByEchelon() []EchelonDiff { return d.byEchelon }

// ConfigSnapshot — слепок одной конфигурации (value object).
type ConfigSnapshot struct {
	projectID         string
	projectName       string
	structuralProfile StructuralProfile
	costCalculation   CostCalculation
}

// NewConfigSnapshot создаёт новый слепок конфигурации.
func NewConfigSnapshot(
	projectID, projectName string,
	structuralProfile StructuralProfile,
	costCalculation CostCalculation,
) ConfigSnapshot {
	return ConfigSnapshot{
		projectID:         projectID,
		projectName:       projectName,
		structuralProfile: structuralProfile,
		costCalculation:   costCalculation,
	}
}

// ProjectID возвращает ID проекта.
func (s ConfigSnapshot) ProjectID() string { return s.projectID }

// ProjectName возвращает название конфигурации.
func (s ConfigSnapshot) ProjectName() string { return s.projectName }

// StructuralProfile возвращает структурный профиль.
func (s ConfigSnapshot) StructuralProfile() StructuralProfile { return s.structuralProfile }

// CostCalculation возвращает расчёт стоимости.
func (s ConfigSnapshot) CostCalculation() CostCalculation { return s.costCalculation }

// ConfigComparison — результат сравнения двух конфигураций (value object).
type ConfigComparison struct {
	projectA ConfigSnapshot
	projectB ConfigSnapshot
	diff     ConfigDiff
}

// NewConfigComparison создаёт новый результат сравнения.
func NewConfigComparison(projectA, projectB ConfigSnapshot, diff ConfigDiff) ConfigComparison {
	return ConfigComparison{
		projectA: projectA,
		projectB: projectB,
		diff:     diff,
	}
}

// ProjectA возвращает слепок первой конфигурации.
func (c ConfigComparison) ProjectA() ConfigSnapshot { return c.projectA }

// ProjectB возвращает слепок второй конфигурации.
func (c ConfigComparison) ProjectB() ConfigSnapshot { return c.projectB }

// Diff возвращает разницу между конфигурациями.
func (c ConfigComparison) Diff() ConfigDiff { return c.diff }
