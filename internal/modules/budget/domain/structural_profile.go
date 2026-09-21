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
