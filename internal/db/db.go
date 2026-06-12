package db

import (
	"fmt"
	"github.com/golang-migrate/migrate"
	pg "github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/fortis/backend/internal/config"
	gormPg "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLog "gorm.io/gorm/logger"
	"log/slog"
	"os"
)

type DataBase struct {
	GormORM *GormORM
}

type GormORM struct {
	*gorm.DB
}

//go:cover off
func NewDataBase(
	pgCfg config.Postgres,
) (db *DataBase) {
	db = new(DataBase)
	db.InitConnections(pgCfg)
	return
}

//go:cover off
func (db *DataBase) InitConnections(pgCfg config.Postgres) {
	db.initGorm(pgCfg)
	// uncomment row below to turn on migration
	// db.MigrationsUp()
}

//go:cover off
func (db *DataBase) initGorm(pgCfg config.Postgres) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		pgCfg.Host,
		pgCfg.Port,
		pgCfg.User,
		pgCfg.Password,
		pgCfg.DbName,
	)

	conn, err := gorm.Open(gormPg.Open(dsn), &gorm.Config{
		Logger: gormLog.Default.LogMode(gormLog.Info),
	})
	if err != nil {
		slog.Error(fmt.Sprintf("error creating connection: %s", err.Error()))
		os.Exit(1)
	}

	db.GormORM = &GormORM{conn}
}

//go:cover off
func (db *DataBase) MigrationsUp() {
	dbConnector, _ := db.GormORM.DB.DB()
	driver, err := pg.WithInstance(dbConnector, &pg.Config{})
	if err != nil {
		slog.Error(fmt.Sprintf("migrations: create driver error: %s", err))
		os.Exit(1)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		slog.Error(fmt.Sprintf("migrations: create connection error: %s", err))
		os.Exit(1)
	}
	err = m.Up()
	if err != nil && err.Error() != "no change" {
		slog.Error(fmt.Sprintf("migrations: apply error: %s", err))
		os.Exit(1)
	}
}
