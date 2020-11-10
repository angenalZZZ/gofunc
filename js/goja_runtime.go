package js

import (
	"github.com/angenalZZZ/gofunc/data"
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
	*goja.Runtime
	*log.Logger
	*sqlx.DB
}

// GoRuntimeParam all parameters of javascript runtime and register
type GoRuntimeParam struct {
	// parameter: *sqlx.DB
	DbType, DbConn string
	// parameter: *nats.Conn
	NatsConn    *nats.Conn
	NatsSubject string
	// parameter: *redis.Client
	RedisClient *redis.Client
	// parameter: *log.Logger
	Log *log.Logger
}

// Clear runtime interrupt and other global vars clear.
func (r *GoRuntime) Clear() {
	r.ClearInterrupt()
	if r.DB != nil {
		_ = r.DB.Close()
	}
}

// NewRuntime create a javascript runtime and register from other global vars.
func NewRuntime(parameter *GoRuntimeParam) *GoRuntime {
	var (
		db  *sqlx.DB
		vm  = goja.New()
		err error
	)

	// parameter: *log.Logger
	logger := log.Log
	if parameter != nil && parameter.Log != nil {
		logger = parameter.Log
	}
	if logger != nil {
		Logger(vm, logger)
	}

	// create all registers
	Console(vm)
	ID(vm)
	RD(vm)
	Ajax(vm)

	// parameter: *sqlx.DB
	dbType, dbConn := data.DbType, data.DbConn
	if parameter != nil && parameter.DbType != "" && parameter.DbConn != "" {
		dbType, dbConn = parameter.DbType, parameter.DbConn
	}
	if dbType != "" && dbConn != "" {
		db, err = sqlx.Connect(dbType, dbConn)
		if err != nil && logger != nil {
			logger.Error().Msgf("failed connect to db: %v\n", err)
		}
		Db(vm, db)
	}

	// parameter: *nats.Conn
	natConn, natSubject := nat.Conn, nat.Subject
	if parameter != nil && parameter.NatsConn != nil && parameter.NatsSubject != "" {
		natConn, natSubject = parameter.NatsConn, parameter.NatsSubject
	}
	if natConn != nil && natSubject != "" {
		Nats(vm, natConn, natSubject)
	}

	// parameter: *redis.Client
	redisClient := store.RedisClient
	if parameter != nil && parameter.RedisClient != nil {
		redisClient = parameter.RedisClient
	}
	if redisClient != nil {
		Redis(vm, redisClient)
	}

	return &GoRuntime{Runtime: vm, Logger: logger, DB: db}
}
