package rdbms

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

type Executor interface {
	Exec(sql string, values ...interface{}) *gorm.DB
	Raw(sql string, values ...interface{}) *gorm.DB
	Model(value interface{}) *gorm.DB
	Create(value interface{}) *gorm.DB
	Save(value interface{}) *gorm.DB
	Delete(value interface{}, conds ...interface{}) *gorm.DB
	Find(dest interface{}, conds ...interface{}) *gorm.DB
	First(dest interface{}, conds ...interface{}) *gorm.DB
	Where(query interface{}, args ...interface{}) *gorm.DB
	WithContext(ctx context.Context) *gorm.DB

	Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error

	Select(query interface{}, args ...interface{}) *gorm.DB
	Distinct(args ...interface{}) *gorm.DB
	Omit(columns ...string) *gorm.DB

	Update(column string, value interface{}) *gorm.DB
	Updates(values interface{}) *gorm.DB
	UpdateColumn(column string, value interface{}) *gorm.DB
	UpdateColumns(values interface{}) *gorm.DB

	Scopes(funcs ...func(*gorm.DB) *gorm.DB) *gorm.DB
	Preload(query string, args ...interface{}) *gorm.DB
	Table(query string, args ...interface{}) *gorm.DB
	Joins(query string, args ...interface{}) *gorm.DB
	Group(name string) *gorm.DB
	Having(query interface{}, args ...interface{}) *gorm.DB
	Order(value interface{}) *gorm.DB
	Limit(limit int) *gorm.DB
	Offset(offset int) *gorm.DB
	Count(*int64) *gorm.DB
	Pluck(column string, dest interface{}) *gorm.DB
}
