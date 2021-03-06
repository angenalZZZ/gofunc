package js

import (
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/angenalZZZ/gofunc/f"

	"github.com/angenalZZZ/gofunc/data"
	"github.com/angenalZZZ/gofunc/data/cache/fastcache"
	"github.com/angenalZZZ/gofunc/data/cache/store"
	"github.com/angenalZZZ/gofunc/log"
	nat "github.com/angenalZZZ/gofunc/rpc/nats"
	"github.com/dop251/goja"
	"github.com/go-redis/redis/v7"
	"github.com/jmoiron/sqlx"
	"github.com/nats-io/nats.go"
)

// Runtime vm for javascript runtime and register
var Runtime *GoRuntime

// GoRuntime javascript runtime and register
type GoRuntime struct {
	// is already registered
	registered bool
	// new javascript runtime
	*goja.Runtime
	// load javascript modules
	Modules map[string]goja.Value
	// field: *log.Logger
	*log.Logger
	// field: *sqlx.DB
	*sqlx.DB
	DbType, DbConn string
	// field: *nats.Conn
	NatConn    *nats.Conn
	NatSubject string
	// field: *redis.Client
	RedisClient *redis.Client
	// field: *fastcache.Cache new fast thread-safe inmemory cache optimized for big number of entries
	*fastcache.Cache
	// field: CacheDir sets cache persist to disk directory
	CacheDir string
}

// GoRuntimeParam all parameters of javascript runtime and register
type GoRuntimeParam struct {
	// parameter: *log.Logger
	Log *log.Logger
	// parameter: *sqlx.DB
	*sqlx.DB
	DbType, DbConn string
	// parameter: *nats.Conn
	NatConn    *nats.Conn
	NatSubject string
	// parameter: *redis.Client
	RedisClient *redis.Client
	// parameter: *fastcache.Cache new fast thread-safe inmemory cache optimized for big number of entries
	*fastcache.Cache
	// parameter: CacheDir sets cache persist to disk directory
	CacheDir string
}

// NewRuntime create a javascript runtime and register from parameter or global vars.
func NewRuntime(parameter *GoRuntimeParam) *GoRuntime {
	var (
		db  *sqlx.DB
		vm  = goja.New()
		r   = new(GoRuntime)
		err error
	)

	// new javascript runtime
	r.Runtime, r.Modules = vm, make(map[string]goja.Value)

	// parameter: *log.Logger
	logger := log.Log
	if parameter != nil && parameter.Log != nil {
		logger = parameter.Log
	}
	if logger != nil {
		r.Logger = logger
	}

	// parameter: *sqlx.DB
	dbType, dbConn := data.DbType, data.DbConn
	if parameter != nil && parameter.DbType != "" && parameter.DbConn != "" {
		dbType, dbConn = parameter.DbType, parameter.DbConn
	}
	if parameter != nil && parameter.DB != nil {
		r.DB = parameter.DB
	} else if dbType != "" && dbConn != "" {
		r.DbType, r.DbConn = dbType, dbConn
		db, err = sqlx.Open(dbType, dbConn)
		if err != nil && logger != nil {
			logger.Error().Msgf("failed connect to db: %v\n", err)
		}
		if db != nil {
			db.SetConnMaxLifetime(time.Minute)
			db.SetMaxIdleConns(4)
			db.SetMaxOpenConns(40)
			r.DB = db
		}
		if parameter != nil {
			parameter.DB = db
		}
	}

	// parameter: *nats.Conn
	natConn, natSubject := nat.Conn, nat.Subject
	if parameter != nil && parameter.NatConn != nil && parameter.NatSubject != "" {
		natConn, natSubject = parameter.NatConn, parameter.NatSubject
	}
	if natConn != nil && natSubject != "" {
		r.NatConn, r.NatSubject = natConn, natSubject
	}

	// parameter: *redis.Client
	redisClient := store.RedisClient
	if parameter != nil && parameter.RedisClient != nil {
		redisClient = parameter.RedisClient
	}
	if redisClient != nil {
		r.RedisClient = redisClient
	}

	// parameter: *fastcache.Cache
	var (
		cache    *fastcache.Cache
		cacheDir string
		maxBytes = 1073741824 // 1GB cache capacity
	)
	if parameter != nil && parameter.Cache != nil {
		cache = parameter.Cache
	} else {
		cache = fastcache.New(maxBytes)
	}
	if parameter != nil && parameter.CacheDir != "" {
		cacheDir = parameter.CacheDir
	}
	r.Cache, r.CacheDir = cache, cacheDir

	r.Register()
	return r
}

// Register init runtime register if not registered.
func (r *GoRuntime) Register() {
	if r.Runtime == nil || r.registered {
		return
	}

	r.loadModules()

	if r.Logger != nil {
		Logger(r.Runtime, r.Logger)
	}

	if r.DB != nil {
		Db(r.Runtime, r.DB)
	} else if r.DbType != "" && r.DbConn != "" {
		Db(r.Runtime, nil, r.DbType, r.DbConn)
	}

	if r.NatConn != nil && r.NatSubject != "" {
		Nats(r.Runtime, r.NatConn, r.NatSubject)
	}

	if r.RedisClient != nil {
		Redis(r.Runtime, r.RedisClient)
	}

	// create all registers
	Console(r.Runtime)
	ID(r.Runtime)
	RD(r.Runtime)
	Ajax(r.Runtime)
	Cache(r.Runtime, r.Cache, r.CacheDir)

	// sets registered
	r.registered = true
}

// loadModules load javascript modules.
func (r *GoRuntime) loadModules() {
	r.Runtime.Set("require", func(c goja.FunctionCall) goja.Value {
		v, p := goja.Undefined(), c.Argument(0).String()
		if p == "" {
			return v
		}

		p = filepath.Clean(p)
		if pkg, ok := r.Modules[p]; ok {
			return pkg
		}

		code, err1 := ioutil.ReadFile(p)
		if err1 != nil {
			return v
		}

		text := "(function(module,exports){\n" + f.String(code) + "\nif(exports)module.exports=exports;})"
		prg, err2 := goja.Compile(p, text, false)
		if err2 != nil {
			return v
		}

		res, err3 := r.Runtime.RunProgram(prg)
		if err3 != nil {
			return v
		}

		fun, ok := goja.AssertFunction(res)
		if !ok {
			return v
		}

		m, e := r.Runtime.NewObject(), r.Runtime.NewObject()
		_ = m.Set("exports", e)
		_, err4 := fun(e, m, v)
		if err4 != nil {
			return v
		}

		return m.Get("exports")
	})
}

// Clear runtime interrupt and fields.
func (r *GoRuntime) Clear() {
	r.ClearInterrupt()
	// field: *log.Logger
	if r.Logger != nil {
	}
	// field: *sqlx.DB
	if r.DB != nil {
		//_ = r.DB.Close()
	}
	// field: *nats.Conn
	if r.NatConn != nil {
		//_ = r.NatConn.FlushTimeout(time.Second)
	}
	// field: *redis.Client
	if r.RedisClient != nil {
	}
}
