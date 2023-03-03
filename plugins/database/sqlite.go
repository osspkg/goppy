package database

import (
	"fmt"

	"github.com/deweppro/go-sdk/log"
	"github.com/deweppro/go-sdk/orm"
	"github.com/deweppro/go-sdk/orm/schema"
	"github.com/deweppro/go-sdk/orm/schema/sqlite"
	"github.com/deweppro/goppy/plugins"
)

// ConfigSqlite sqlite config model
type ConfigSqlite struct {
	Pool    []sqlite.Item       `yaml:"sqlite"`
	Migrate []ConfigMigrateItem `yaml:"sqlite_migrate"`
}

func (v *ConfigSqlite) Default() {
	if len(v.Pool) == 0 {
		v.Pool = []sqlite.Item{
			{
				Name:        "main",
				File:        "./sqlite.db",
				Cache:       "private",
				Mode:        "rwc",
				Journal:     "TRUNCATE",
				LockingMode: "EXCLUSIVE",
				OtherParams: "",
			},
		}
	}
	if len(v.Migrate) == 0 {
		v.Migrate = []ConfigMigrateItem{
			{
				Pool: "main",
				Dir:  "./migrations",
			},
		}
	}
}

// List getting all configs
func (v *ConfigSqlite) List() (list []schema.ItemInterface) {
	for _, vv := range v.Pool {
		list = append(list, vv)
	}
	return
}

// WithSQLite launch SQLite connection pool
func WithSQLite() plugins.Plugin {
	return plugins.Plugin{
		Config: &ConfigSqlite{},
		Inject: func(c *ConfigSqlite, l log.Logger) (*sqliteProvider, *migrate, SQLite) {
			conn := sqlite.New(c)
			o := orm.New(conn, orm.UsePluginLogger(l))
			m := newMigrate(o, c.Migrate, l)
			return &sqliteProvider{conn: conn, conf: *c, log: l}, m, o
		},
	}
}

type (
	sqliteProvider struct {
		conn schema.Connector
		conf ConfigSqlite
		log  log.Logger
	}

	//SQLite connection SQLite interface
	SQLite interface {
		Pool(name string) orm.Stmt
		Dialect() string
	}
)

func (v *sqliteProvider) Up() error {
	if err := v.conn.Reconnect(); err != nil {
		return err
	}
	for _, vv := range v.conf.Pool {
		p, err := v.conn.Pool(vv.Name)
		if err != nil {
			return fmt.Errorf("pool `%s`: %w", vv.Name, err)
		}
		if err = p.Ping(); err != nil {
			return fmt.Errorf("pool `%s`: %w", vv.Name, err)
		}
		v.log.WithFields(log.Fields{vv.Name: vv.File}).Infof("SQLite connect")
	}
	return nil
}

func (v *sqliteProvider) Down() error {
	return v.conn.Close()
}
