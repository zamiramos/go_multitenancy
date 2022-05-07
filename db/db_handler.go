package db

import (
	"errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"sync"
	"time"
	gocache "zgo.at/zcache"
)

var cacheHandler *gocache.Cache

const (
	PostgreSQL string = "postgresql"
	SQLite     string = "sqlite"
)

var (
	ErrUnknownDBType = errors.New("unknown db type")
)

type Handler interface {
	CreateDBHandler(ctx *Context) *gorm.DB
	GetDBHandler(ctx *Context) (*gorm.DB, bool)
}

type GormDBLifecycle interface {
	InitDB(ctx *Context) (*gorm.DB, error)
}

func HandlerFactory(dbType string) (Handler, error) {
	cacheHandler = gocache.New(10*time.Minute, 10*time.Minute)
	cacheHandler.OnEvicted(func(s string, i interface{}) {
		// https://github.com/go-gorm/gorm/issues/3145
		sql, err := i.(*gorm.DB).DB()
		if err != nil {
			panic(err)
		}
		sql.Close()
	})

	switch dbType {
	case SQLite:
		handler := &handler{
			cache: cacheHandler,
			gormDBlc: &SQLiteHandler{
				dbPath: ".\\",
			},
		}

		return handler, nil
	}

	return nil, ErrUnknownDBType
}

type SQLiteHandler struct {
	dbPath string
}

type handler struct {
	CreationMutex sync.Mutex
	cache         *gocache.Cache
	gormDBlc      GormDBLifecycle
}

type Context struct {
	RealmId string
}

func (h *SQLiteHandler) InitDB(ctx *Context) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(h.dbPath+ctx.RealmId+".db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (h *handler) CreateDBHandler(ctx *Context) *gorm.DB {
	// Maybe we could use a better mechanism to do this.
	h.CreationMutex.Lock()
	defer h.CreationMutex.Unlock()

	// Double check before moving to Creation procedures
	if db, found := h.cache.Touch(ctx.RealmId, gocache.DefaultExpiration); found {
		return db.(*gorm.DB)
	}

	db, err := h.gormDBlc.InitDB(ctx)

	if err != nil {
		panic(err)
	}

	cacheHandler.Set(ctx.RealmId, db, gocache.DefaultExpiration)

	return db
}

func (h *handler) GetDBHandler(ctx *Context) (*gorm.DB, bool) {
	db, found := h.cache.Touch(ctx.RealmId, gocache.DefaultExpiration)
	if found {
		return db.(*gorm.DB), true
	}

	return h.CreateDBHandler(ctx), false
}
