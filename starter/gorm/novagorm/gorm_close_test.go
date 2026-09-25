package novagorm

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestCloseAllClosesUnderlyingSQLPool(t *testing.T) {
	resetForTest()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	gormDB, err := gorm.Open(gormmysql.New(gormmysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("open gorm: %v", err)
	}

	const instanceName = "close-test"
	Register(instanceName, func(string) (*gorm.DB, error) { return gormDB, nil })
	if _, err := Named(instanceName).DB(); err != nil {
		t.Fatalf("open registered gorm instance: %v", err)
	}

	mock.ExpectClose()
	if err := CloseAll(); err != nil {
		t.Fatalf("close all: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("underlying SQL pool was not closed: %v", err)
	}
}

func TestPostgresBridgeCloseClosesUnderlyingSQLPool(t *testing.T) {
	resetForTest()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}

	gormDB, err := OpenPostgresFromSQLDB(sqlDB)
	if err != nil {
		t.Fatalf("open postgres bridge: %v", err)
	}

	const instanceName = "postgres-close-test"
	Register(instanceName, func(string) (*gorm.DB, error) { return gormDB, nil })
	if _, err := Named(instanceName).DB(); err != nil {
		t.Fatalf("open registered postgres instance: %v", err)
	}

	mock.ExpectClose()
	if err := CloseAll(); err != nil {
		t.Fatalf("close all: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("underlying postgres SQL pool was not closed: %v", err)
	}
}

func TestReloadClosesOldPostgresPoolAndKeepsReplacementUsable(t *testing.T) {
	resetForTest()

	firstSQLDB, firstMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create first sql mock: %v", err)
	}
	firstGormDB, err := OpenPostgresFromSQLDB(firstSQLDB)
	if err != nil {
		t.Fatalf("open first postgres bridge: %v", err)
	}

	secondSQLDB, secondMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create second sql mock: %v", err)
	}
	secondGormDB, err := OpenPostgresFromSQLDB(secondSQLDB)
	if err != nil {
		t.Fatalf("open second postgres bridge: %v", err)
	}

	builds := 0
	Register("postgres-reload-test", func(string) (*gorm.DB, error) {
		builds++
		if builds == 1 {
			return firstGormDB, nil
		}
		return secondGormDB, nil
	})
	handle := Named("postgres-reload-test")
	if _, err := handle.DB(); err != nil {
		t.Fatalf("open first registered postgres instance: %v", err)
	}

	firstMock.ExpectClose()
	if err := handle.Reload(); err != nil {
		t.Fatalf("reload postgres instance: %v", err)
	}
	if err := firstSQLDB.Ping(); err == nil {
		t.Fatal("old PostgreSQL pool remained usable after Reload")
	}
	if err := secondSQLDB.Ping(); err != nil {
		t.Fatalf("replacement PostgreSQL pool is not usable: %v", err)
	}

	secondMock.ExpectClose()
	if err := CloseAll(); err != nil {
		t.Fatalf("close replacement postgres instance: %v", err)
	}
	if err := firstMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("first pool expectations: %v", err)
	}
	if err := secondMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("second pool expectations: %v", err)
	}
}
