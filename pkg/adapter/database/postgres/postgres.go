package postgres

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"camera_api/pkg/adapter/database/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	ormlog "gorm.io/gorm/logger"
)

type Setting struct {
	Migrate bool
	DSN     string

	// Connection pool. Zero values fall back to the defaults below.
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// Pool defaults: bounded so a traffic spike can't exhaust PostgreSQL's
// max_connections, and recycled so connections don't live forever (e.g.
// across a database failover or behind a proxy that drops idle links).
const (
	defaultMaxOpenConns    = 25
	defaultMaxIdleConns    = 10
	defaultConnMaxLifetime = 30 * time.Minute
	defaultConnMaxIdleTime = 5 * time.Minute

	pingTimeout = 5 * time.Second
)

// withPoolDefaults fills unset pool settings with the defaults.
func (s Setting) withPoolDefaults() Setting {
	if s.MaxOpenConns <= 0 {
		s.MaxOpenConns = defaultMaxOpenConns
	}

	if s.MaxIdleConns <= 0 {
		s.MaxIdleConns = defaultMaxIdleConns
	}

	// Idle connections above the open limit would never be used.
	if s.MaxIdleConns > s.MaxOpenConns {
		s.MaxIdleConns = s.MaxOpenConns
	}

	if s.ConnMaxLifetime <= 0 {
		s.ConnMaxLifetime = defaultConnMaxLifetime
	}

	if s.ConnMaxIdleTime <= 0 {
		s.ConnMaxIdleTime = defaultConnMaxIdleTime
	}

	return s
}

type postgresDB struct {
	db *gorm.DB
}

type contextKey string

const txContext contextKey = "txContext"

type Database interface {
	Db(ctx context.Context) *gorm.DB
	WithTransaction(
		ctx context.Context,
		fn func(ctxWithTx context.Context, db *gorm.DB) error,
	) error
}

func NewPostgres(setting Setting) (Database, error) {
	newLogger := ormlog.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		ormlog.Config{
			SlowThreshold:             5 * time.Second,
			Colorful:                  true,
			IgnoreRecordNotFoundError: true,
			LogLevel:                  ormlog.Error,
		},
	)

	db, err := gorm.Open(
		postgres.Open(setting.DSN),
		&gorm.Config{
			// Map driver errors such as unique violations to gorm.ErrDuplicatedKey.
			TranslateError: true,
			Logger:         newLogger,
			NowFunc: func() time.Time {
				return time.Now().UTC()
			},
		},
	)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}

	pool := setting.withPoolDefaults()
	sqlDB.SetMaxOpenConns(pool.MaxOpenConns)
	sqlDB.SetMaxIdleConns(pool.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(pool.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(pool.ConnMaxIdleTime)

	// Fail fast at startup if the database is unreachable.
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if setting.Migrate {
		if err := migrate(db); err != nil {
			return nil, err
		}
	}

	return &postgresDB{
		db: db,
	}, nil
}

func migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.Role{},
		&model.User{},
		&model.Session{},
		&model.UserActivity{},

		&model.Protocol{},
		&model.Camera{},
		&model.CameraAccess{},

		&model.Direction{},
		&model.PanTilt{},
	)
}

func (d *postgresDB) Db(ctx context.Context) *gorm.DB {
	tx := ctx.Value(txContext)

	if tx == nil {
		return d.db.WithContext(ctx)
	}

	db, ok := tx.(*gorm.DB)
	if !ok {
		return d.db.WithContext(ctx)
	}

	return db.WithContext(ctx)
}

func (d *postgresDB) WithTransaction(ctx context.Context, fn func(ctxWithTx context.Context, db *gorm.DB) error) error {
	tx := d.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	ctxWithTx := context.WithValue(ctx, txContext, tx)

	if err := fn(ctxWithTx, tx); err != nil {
		if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
			log.Printf(
				"failed on rollback transaction: %v",
				rollbackErr,
			)
		}

		return err
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf(
			"failed on commit transaction: %v",
			err,
		)

		return err
	}

	return nil
}
