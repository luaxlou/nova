//go:build postgres_integration

package novagorm

import (
	"errors"
	"os"
	"strings"
	"testing"

	"gorm.io/gorm"
)

type postgresIntegrationRecord struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement"`
	ExternalKey string `gorm:"column:external_key;uniqueIndex:uq_nova_pg18_records_external_key;not null"`
	Payload     string `gorm:"column:payload;type:jsonb;not null"`
}

func (postgresIntegrationRecord) TableName() string { return "nova_pg18_records" }

func TestPostgres18Integration(t *testing.T) {
	dsn := os.Getenv("POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Fatal("POSTGRES_TEST_DSN is required for the postgres_integration test")
	}

	resetForTest()
	t.Cleanup(resetForTest)

	definitions, selectedName := buildDefinitions(map[string]any{
		"driver": "postgres",
		"postgres": map[string]any{
			"dsn":                    dsn,
			"max_open":               4,
			"max_idle":               2,
			"conn_max_lifetime":      60,
			"conn_max_idle_time":     30,
			"prefer_simple_protocol": true,
		},
	})
	if selectedName != singletonName || definitions[singletonName] == nil {
		t.Fatalf("configured definitions = %#v, selected = %q", definitions, selectedName)
	}

	Register("postgres-integration", definitions[singletonName])
	db, err := Named("postgres-integration").DB()
	if err != nil {
		t.Fatalf("open configured PostgreSQL connection: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get underlying PostgreSQL pool: %v", err)
	}
	if got := sqlDB.Stats().MaxOpenConnections; got != 4 {
		t.Fatalf("max open connections = %d, want 4", got)
	}

	t.Cleanup(func() {
		cleanupDB, cleanupErr := newPostgresConnection(gormConfig{Postgres: postgresConfig{DSN: dsn}})
		if cleanupErr != nil {
			t.Logf("cleanup connection failed: %v", cleanupErr)
			return
		}
		_ = cleanupDB.Exec("DROP TABLE IF EXISTS nova_pg18_records").Error
		cleanupSQLDB, cleanupErr := cleanupDB.DB()
		if cleanupErr == nil {
			_ = cleanupSQLDB.Close()
		}
	})

	if err := db.Exec("DROP TABLE IF EXISTS nova_pg18_records").Error; err != nil {
		t.Fatalf("clean integration table: %v", err)
	}

	var serverVersion string
	if err := db.Raw("SHOW server_version_num").Scan(&serverVersion).Error; err != nil {
		t.Fatalf("read PostgreSQL server version: %v", err)
	}
	if !strings.HasPrefix(serverVersion, "180") {
		t.Fatalf("server_version_num = %q, want PostgreSQL 18", serverVersion)
	}
	t.Logf("verified PostgreSQL server_version_num=%s", serverVersion)

	if err := db.AutoMigrate(&postgresIntegrationRecord{}); err != nil {
		t.Fatalf("AutoMigrate explicit PostgreSQL model: %v", err)
	}
	if !db.Migrator().HasTable(&postgresIntegrationRecord{}) {
		t.Fatal("AutoMigrate did not create nova_pg18_records")
	}
	if !db.Migrator().HasColumn(&postgresIntegrationRecord{}, "payload") {
		t.Fatal("AutoMigrate did not create payload column")
	}
	if !db.Migrator().HasIndex(&postgresIntegrationRecord{}, "uq_nova_pg18_records_external_key") {
		t.Fatal("AutoMigrate did not create unique external_key index")
	}

	wantPayload := `{"component":"nova","version":18}`
	first := postgresIntegrationRecord{ExternalKey: "component:nova", Payload: wantPayload}
	if err := db.Create(&first).Error; err != nil {
		t.Fatalf("insert PostgreSQL record: %v", err)
	}

	var got postgresIntegrationRecord
	if err := db.Where("external_key = ?", first.ExternalKey).First(&got).Error; err != nil {
		t.Fatalf("read PostgreSQL record: %v", err)
	}
	var component string
	if err := db.Raw("SELECT payload->>'component' FROM nova_pg18_records WHERE external_key = ?", first.ExternalKey).
		Scan(&component).Error; err != nil {
		t.Fatalf("query JSONB component: %v", err)
	}
	if component != "nova" {
		t.Fatalf("JSONB component = %q, want nova", component)
	}

	duplicate := postgresIntegrationRecord{ExternalKey: first.ExternalKey, Payload: `{"duplicate":true}`}
	if err := db.Create(&duplicate).Error; err == nil {
		t.Fatal("duplicate unique key insert succeeded, want error")
	}

	rollbackErr := db.Transaction(func(txDB *gorm.DB) error {
		if err := txDB.Create(&postgresIntegrationRecord{
			ExternalKey: "component:rolled-back",
			Payload:     `{"rolled_back":true}`,
		}).Error; err != nil {
			return err
		}
		return errors.New("force rollback")
	})
	if rollbackErr == nil {
		t.Fatal("transaction returned nil, want forced rollback error")
	}

	var rollbackCount int64
	if err := db.Model(&postgresIntegrationRecord{}).
		Where("external_key = ?", "component:rolled-back").
		Count(&rollbackCount).Error; err != nil {
		t.Fatalf("count rolled-back records: %v", err)
	}
	if rollbackCount != 0 {
		t.Fatalf("rolled-back record count = %d, want 0", rollbackCount)
	}

	if err := db.Exec("DROP TABLE nova_pg18_records").Error; err != nil {
		t.Fatalf("drop integration table: %v", err)
	}
	if err := CloseAll(); err != nil {
		t.Fatalf("close PostgreSQL instance: %v", err)
	}
	if err := sqlDB.Ping(); err == nil {
		t.Fatal("ping after CloseAll succeeded, want closed pool error")
	}
}
