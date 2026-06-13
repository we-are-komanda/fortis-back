package domain

// EstimateLine — строка расчёта стоимости одного размещённого объекта.
type EstimateLine struct {
	objectID     string
	assetID      string
	assetName    string
	echelonID    string
	echelonName  string
	typeID       string
	typeName     string
	quantity     int
	unitPriceMln float64
	lineTotalMln float64
}

// NewEstimateLine создаёт новую строку расчёта.
func NewEstimateLine(
	objectID, assetID, assetName string,
	echelonID, echelonName string,
	typeID, typeName string,
	quantity int,
	unitPriceMln float64,
	lineTotalMln float64,
) EstimateLine {
	return EstimateLine{
		objectID:     objectID,
		assetID:      assetID,
		assetName:    assetName,
		echelonID:    echelonID,
		echelonName:  echelonName,
		typeID:       typeID,
		typeName:     typeName,
		quantity:     quantity,
		unitPriceMln: unitPriceMln,
		lineTotalMln: lineTotalMln,
	}
}

// ObjectID возвращает ID размещённого объекта.
func (l EstimateLine) ObjectID() string { return l.objectID }

// AssetID возвращает ID средства защиты.
func (l EstimateLine) AssetID() string { return l.assetID }

// AssetName возвращает название средства защиты.
func (l EstimateLine) AssetName() string { return l.assetName }

// EchelonID возвращает ID эшелона (слоя).
func (l EstimateLine) EchelonID() string { return l.echelonID }

// EchelonName возвращает название эшелона.
func (l EstimateLine) EchelonName() string { return l.echelonName }

// TypeID возвращает ID типа защиты.
func (l EstimateLine) TypeID() string { return l.typeID }

// TypeName возвращает название типа защиты.
func (l EstimateLine) TypeName() string { return l.typeName }

// Quantity возвращает количество единиц.
func (l EstimateLine) Quantity() int { return l.quantity }

// UnitPriceMln возвращает цену за единицу в млн ₽.
func (l EstimateLine) UnitPriceMln() float64 { return l.unitPriceMln }

// LineTotalMln возвращает итоговую стоимость строки в млн ₽.
func (l EstimateLine) LineTotalMln() float64 { return l.lineTotalMln }

// EchelonEstimate — стоимость по эшелону (слою) защиты.
type EchelonEstimate struct {
	echelonID       string
	echelonName     string
	lines           []EstimateLine
	echelonTotalMln float64
}

// NewEchelonEstimate создаёт новую оценку стоимости по эшелону.
func NewEchelonEstimate(echelonID, echelonName string, lines []EstimateLine, totalMln float64) EchelonEstimate {
	return EchelonEstimate{
		echelonID:       echelonID,
		echelonName:     echelonName,
		lines:           lines,
		echelonTotalMln: totalMln,
	}
}

// EchelonID возвращает ID эшелона.
func (e EchelonEstimate) EchelonID() string { return e.echelonID }

// EchelonName возвращает название эшелона.
func (e EchelonEstimate) EchelonName() string { return e.echelonName }

// Lines возвращает строки расчёта в эшелоне.
func (e EchelonEstimate) Lines() []EstimateLine { return e.lines }

// EchelonTotalMln возвращает общую стоимость эшелона.
func (e EchelonEstimate) EchelonTotalMln() float64 { return e.echelonTotalMln }

// TypeEstimate — стоимость по типу защиты.
type TypeEstimate struct {
	typeID       string
	typeName     string
	lines        []EstimateLine
	typeTotalMln float64
}

// NewTypeEstimate создаёт новую оценку стоимости по типу.
func NewTypeEstimate(typeID, typeName string, lines []EstimateLine, totalMln float64) TypeEstimate {
	return TypeEstimate{
		typeID:       typeID,
		typeName:     typeName,
		lines:        lines,
		typeTotalMln: totalMln,
	}
}

// TypeID возвращает ID типа защиты.
func (t TypeEstimate) TypeID() string { return t.typeID }

// TypeName возвращает название типа защиты.
func (t TypeEstimate) TypeName() string { return t.typeName }

// Lines возвращает строки расчёта данного типа.
func (t TypeEstimate) Lines() []EstimateLine { return t.lines }

// TypeTotalMln возвращает общую стоимость по типу.
func (t TypeEstimate) TypeTotalMln() float64 { return t.typeTotalMln }

// CostCalculation — полный расчёт стоимости конфигурации проекта.
type CostCalculation struct {
	totalMln  float64
	byEchelon []EchelonEstimate
	byType    []TypeEstimate
	byObject  []EstimateLine
}

// NewCostCalculation создаёт новый полный расчёт стоимости.
func NewCostCalculation(
	totalMln float64,
	byEchelon []EchelonEstimate,
	byType []TypeEstimate,
	byObject []EstimateLine,
) CostCalculation {
	if byEchelon == nil {
		byEchelon = []EchelonEstimate{}
	}
	if byType == nil {
		byType = []TypeEstimate{}
	}
	if byObject == nil {
		byObject = []EstimateLine{}
	}
	return CostCalculation{
		totalMln:  totalMln,
		byEchelon: byEchelon,
		byType:    byType,
		byObject:  byObject,
	}
}

// TotalMln возвращает общую стоимость конфигурации.
func (c CostCalculation) TotalMln() float64 { return c.totalMln }

// ByEchelon возвращает стоимость по эшелонам.
func (c CostCalculation) ByEchelon() []EchelonEstimate { return c.byEchelon }

// ByType возвращает стоимость по типам защиты.
func (c CostCalculation) ByType() []TypeEstimate { return c.byType }

// ByObject возвращает стоимость по объектам.
func (c CostCalculation) ByObject() []EstimateLine { return c.byObject }
