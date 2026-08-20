package novagorm

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	"github.com/luaxlou/nova/starter/internal/registry"
	"github.com/luaxlou/nova/starter/novamysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Instance interface {
	DB() (*gorm.DB, error)
	Reload() error
	Close() error
}

type gormInstance struct {
	handle *registry.Instance[*gorm.DB]
}

var (
	initialized bool
	initMu      sync.Mutex
	reg         = registry.New[*gorm.DB]()
)

func Init() error {
	initMu.Lock()
	defer initMu.Unlock()
	if initialized {
		return nil
	}

	if err := novamysql.Init(); err != nil {
		return err
	}

	definitions := map[string]registry.Builder[*gorm.DB]{
		"": newGormConnection,
	}
	reg.Init("", definitions)
	initialized = true
	log.Printf("GORM Starter initialized, source=novamysql")
	return nil
}

func Get() *gormInstance {
	_ = ensureInit()
	return &gormInstance{handle: reg.Get()}
}

func Named(name string) *gormInstance {
	_ = ensureInit()
	if name != "" {
		ensureDefinition(name)
	}
	return &gormInstance{handle: reg.Named(name)}
}

func DB() (*gorm.DB, error) {
	return Named("").DB()
}

func (h *gormInstance) DB() (*gorm.DB, error) {
	return h.handle.Get()
}

func (h *gormInstance) Reload() error {
	return h.handle.Reload()
}

func (h *gormInstance) Close() error {
	return h.handle.Close()
}

func Reload() {
	_ = ensureInit()
	_ = Get().Reload()
}

func Close() error {
	_ = ensureInit()
	return Get().Close()
}

func CloseAll() error {
	_ = ensureInit()
	return reg.CloseAll()
}

func newGormConnection(name string) (*gorm.DB, error) {
	sqlDB, err := mysqlDB(name)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(gormmysql.New(gormmysql.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open gorm from novamysql connection: %w", err)
	}
	return db, nil
}

func mysqlDB(name string) (*sql.DB, error) {
	if name == "" {
		return novamysql.DB()
	}
	return novamysql.Named(name).DB()
}

func ensureDefinition(name string) {
	for _, defined := range reg.Definitions() {
		if defined == name {
			return
		}
	}
	reg.Register(name, newGormConnection)
}

func ensureInit() error {
	if initialized {
		return nil
	}
	return Init()
}
