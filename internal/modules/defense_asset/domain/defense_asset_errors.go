package domain

import "errors"

var (
	// ErrDefenseAssetNotFound возвращается когда средство защиты не найдено.
	ErrDefenseAssetNotFound = errors.New("defense asset not found")

	// ErrDefenseAssetInvalidName возвращается при пустом названии средства защиты.
	ErrDefenseAssetInvalidName = errors.New("invalid defense asset name: must not be empty")

	// ErrDefenseAssetInvalidCategory возвращается при неверной категории.
	ErrDefenseAssetInvalidCategory = errors.New("invalid defense asset category")

	// ErrDefenseAssetInvalidCoverageType возвращается при неверном типе покрытия.
	ErrDefenseAssetInvalidCoverageType = errors.New("invalid defense asset coverage type")

	// ErrDefenseAssetInvalidDeploymentType возвращается при неверном типе развёртывания.
	ErrDefenseAssetInvalidDeploymentType = errors.New("invalid defense asset deployment type")

	// ErrDefenseAssetInvalidPlacementType возвращается при неверном типе размещения.
	ErrDefenseAssetInvalidPlacementType = errors.New("invalid defense asset placement type")

	// ErrDefenseAssetInvalidSpecification возвращается при несовместимости спецификации и категории.
	ErrDefenseAssetInvalidSpecification = errors.New("invalid defense asset specification for category")
)
