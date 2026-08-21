package novagorm

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestCloseAllClosesUnderlyingSQLPool(t *testing.T) {
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
