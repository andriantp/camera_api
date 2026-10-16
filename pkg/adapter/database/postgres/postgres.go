package postgres

import (
	"context"
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
			Logger: newLogger,
			NowFunc: func() time.Time {
				return time.Now().UTC()
			},
		},
	)
	if err != nil {
		return nil, err
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
