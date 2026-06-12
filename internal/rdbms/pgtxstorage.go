package rdbms

import (
	"github.com/fortis/backend/internal/db"
	"context"
	"fmt"

	"gorm.io/gorm"
)

type StorageFactory[Storage any] func(Executor) Storage

type GormTxStorage[Storage any] struct {
	db       *gorm.DB
	store    Storage
	sFactory StorageFactory[Storage]
}

func NewGormTxStorage[Storage any](database *db.GormORM, factory StorageFactory[Storage]) *GormTxStorage[Storage] {
	return &GormTxStorage[Storage]{
		db:       database.DB,
		store:    factory(database),
		sFactory: factory,
	}
}

func (txr *GormTxStorage[Storage]) Storage() Storage {
	return txr.store
}

func SQLIsolationLevel(isoLevel TXIsoLevel) (string, error) {
	switch isoLevel {
	case TXIsoLevelReadCommitted:
		return "READ COMMITTED", nil
	case TXIsoLevelRepeatableRead:
		return "REPEATABLE READ", nil
	case TXIsoLevelSerializable:
		return "SERIALIZABLE", nil
	case TXIsoLevelDefault:
		return "", nil
	default:
		return "", fmt.Errorf("unknown tx isolation level: %s", isoLevel)
	}
}

func (txr *GormTxStorage[Storage]) Begin(ctx context.Context, isoLevel TXIsoLevel) (Tx[Storage], error) {
	tx := txr.db.Session(&gorm.Session{
		SkipDefaultTransaction: true,
		Context:                ctx,
	})

	if isoLevel != TXIsoLevelDefault {
		isolationSQL, err := SQLIsolationLevel(isoLevel)
		if err != nil {
			return nil, err
		}

		if err := tx.Exec("SET TRANSACTION ISOLATION LEVEL " + isolationSQL).Error; err != nil {
			return nil, fmt.Errorf("failed to set transaction isolation level: %w", err)
		}
	}

	tx = tx.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	txObject := &GormTransaction[Storage]{
		tx:    tx,
		store: txr.sFactory(tx),
	}

	return txObject, nil
}

type GormTransaction[Storage any] struct {
	tx    *gorm.DB
	store Storage
}

func (gtx *GormTransaction[Storage]) Commit(ctx context.Context) error {
	result := gtx.tx.Commit()
	return result.Error
}

func (gtx *GormTransaction[Storage]) Rollback(ctx context.Context) error {
	result := gtx.tx.Rollback()
	return result.Error
}

func (gtx *GormTransaction[Storage]) Storage() Storage {
	return gtx.store
}
