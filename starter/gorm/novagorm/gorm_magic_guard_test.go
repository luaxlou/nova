package novagorm

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type implicitTableNameModel struct {
	ID string `gorm:"column:id;primaryKey"`
}

type implicitColumnModel struct {
	ID string `gorm:"primaryKey"`
}

func (implicitColumnModel) TableName() string { return "implicit_columns" }

type automaticCreateTimeModel struct {
	ID        string    `gorm:"column:id;primaryKey"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (automaticCreateTimeModel) TableName() string { return "automatic_create_times" }

type automaticUpdateTimeModel struct {
	ID        string    `gorm:"column:id;primaryKey"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (automaticUpdateTimeModel) TableName() string { return "automatic_update_times" }

type embeddedGormModel struct {
	gorm.Model
}

func (embeddedGormModel) TableName() string { return "embedded_gorm_models" }

type softDeleteModel struct {
	ID        string         `gorm:"column:id;primaryKey"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (softDeleteModel) TableName() string { return "soft_delete_models" }

type lifecycleHookModel struct {
	ID string `gorm:"column:id;primaryKey"`
}

func (lifecycleHookModel) TableName() string { return "lifecycle_hook_models" }

func (*lifecycleHookModel) BeforeCreate(*gorm.DB) error { return nil }

type associationChildModel struct {
	ID       string `gorm:"column:id;primaryKey"`
	ParentID string `gorm:"column:parent_id;not null"`
}

func (associationChildModel) TableName() string { return "association_children" }

type associationModel struct {
	ID       string                  `gorm:"column:id;primaryKey"`
	Children []associationChildModel `gorm:"foreignKey:ParentID;references:ID"`
}

func (associationModel) TableName() string { return "associations" }

type explicitModel struct {
	ID        string    `gorm:"column:id;primaryKey"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime:false"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime:false"`
	Transient string    `gorm:"-"`
}

func (explicitModel) TableName() string { return "explicit_models" }

type recordingMigrator struct {
	gorm.Migrator
	calls int
}

func (m *recordingMigrator) AutoMigrate(...any) error {
	m.calls++
	return nil
}

func TestAutoMigrateRejectsImplicitTableNameBeforeDatabaseAccess(t *testing.T) {
	resetForTest()

	db, mock := openGuardTestDB(t)
	Register("guarded", func(string) (*gorm.DB, error) { return db, nil })

	guardedDB, err := Named("guarded").DB()
	if err != nil {
		t.Fatalf("get guarded db: %v", err)
	}

	err = guardedDB.AutoMigrate(&implicitTableNameModel{})
	if err == nil || !strings.Contains(err.Error(), "ORM magic blocked") || !strings.Contains(err.Error(), "TableName") {
		t.Fatalf("AutoMigrate() error = %v, want ORM magic TableName error", err)
	}
	if !errors.Is(err, ErrORMMagic) {
		t.Fatalf("errors.Is(AutoMigrate() error, ErrORMMagic) = false, error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("AutoMigrate reached database before guard rejection: %v", err)
	}
}

func TestOpenMySQLFromSQLDBInstallsAutoMigrateGuard(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	mock.ExpectQuery("SELECT VERSION\\(\\)").
		WillReturnRows(sqlmock.NewRows([]string{"VERSION()"}).AddRow("8.0.36"))

	db, err := OpenMySQLFromSQLDB(sqlDB)
	if err != nil {
		t.Fatalf("OpenMySQLFromSQLDB() error = %v", err)
	}

	err = db.AutoMigrate(&implicitTableNameModel{})
	if !errors.Is(err, ErrORMMagic) {
		t.Fatalf("AutoMigrate() error = %v, want ErrORMMagic", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("AutoMigrate reached database before guard rejection: %v", err)
	}
}

func TestAutoMigrateRejectsModelLevelORMMagic(t *testing.T) {
	tests := []struct {
		name  string
		model any
		want  string
	}{
		{name: "implicit column", model: &implicitColumnModel{}, want: "column"},
		{name: "automatic create time", model: &automaticCreateTimeModel{}, want: "autoCreateTime"},
		{name: "automatic update time", model: &automaticUpdateTimeModel{}, want: "autoUpdateTime"},
		{name: "embedded gorm model", model: &embeddedGormModel{}, want: "gorm.Model"},
		{name: "soft delete", model: &softDeleteModel{}, want: "gorm.DeletedAt"},
		{name: "lifecycle hook", model: &lifecycleHookModel{}, want: "BeforeCreate"},
		{name: "association", model: &associationModel{}, want: "association Children"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetForTest()

			db, mock := openGuardTestDB(t)
			Register("guarded", func(string) (*gorm.DB, error) { return db, nil })

			guardedDB, err := Named("guarded").DB()
			if err != nil {
				t.Fatalf("get guarded db: %v", err)
			}

			err = guardedDB.AutoMigrate(tt.model)
			if err == nil || !strings.Contains(err.Error(), "ORM magic blocked") || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("AutoMigrate() error = %v, want ORM magic error containing %q", err, tt.want)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("AutoMigrate reached database before guard rejection: %v", err)
			}
		})
	}
}

func TestAutoMigrateDelegatesExplicitModelToGORM(t *testing.T) {
	db, _ := openGuardTestDB(t)
	delegate := &recordingMigrator{}
	guard := &autoMigrateGuard{Migrator: delegate, db: db}

	if err := guard.AutoMigrate(&explicitModel{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v, want nil", err)
	}
	if delegate.calls != 1 {
		t.Fatalf("native AutoMigrate calls = %d, want 1", delegate.calls)
	}
}

func TestAutoMigrateValidatesAllModelsBeforeAnyDDL(t *testing.T) {
	resetForTest()

	db, mock := openGuardTestDB(t)
	Register("guarded", func(string) (*gorm.DB, error) { return db, nil })

	guardedDB, err := Named("guarded").DB()
	if err != nil {
		t.Fatalf("get guarded db: %v", err)
	}

	err = guardedDB.AutoMigrate(&explicitModel{}, &implicitColumnModel{})
	if !errors.Is(err, ErrORMMagic) || !strings.Contains(err.Error(), "implicitColumnModel") {
		t.Fatalf("AutoMigrate() error = %v, want second model rejected with ErrORMMagic", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("AutoMigrate executed DDL before validating every model: %v", err)
	}
}

func openGuardTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	db, err := gorm.Open(gormmysql.New(gormmysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("open gorm: %v", err)
	}

	return db, mock
}
