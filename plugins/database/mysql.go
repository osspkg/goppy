package database

import (
	"fmt"
	"time"

	"github.com/dewep-online/goppy/plugins"
	"github.com/deweppro/go-logger"
	"github.com/deweppro/go-orm"
	"github.com/deweppro/go-orm/schema"
	"github.com/deweppro/go-orm/schema/mysql"
)

// ConfigMysql mysql config model
type ConfigMysql struct {
	Pool []mysql.Item `yaml:"mysql"`
}

// List getting all configs
func (v *ConfigMysql) List() (list []schema.ItemInterface) {
	for _, vv := range v.Pool {
		list = append(list, vv)
	}
	return
}

func (v *ConfigMysql) Default() {
	if len(v.Pool) == 0 {
		v.Pool = []mysql.Item{
			{
				Name:              "main",
				Host:              "127.0.0.1",
				Port:              3306,
				Schema:            "test_database",
				User:              "test",
				Password:          "test",
				MaxIdleConn:       5,
				MaxOpenConn:       5,
				MaxConnTTL:        time.Second * 50,
				InterpolateParams: false,
				Timezone:          "UTC",
				TxIsolationLevel:  "",
				Charset:           "utf8mb4,utf8",
				Timeout:           time.Second * 5,
				ReadTimeout:       time.Second * 5,
				WriteTimeout:      time.Second * 5,
			},
		}

	}
}

// WithMySQL launch MySQL connection pool
func WithMySQL() plugins.Plugin {
	return plugins.Plugin{
		Config: &ConfigMysql{},
		Inject: func(conf *ConfigMysql, log logger.Logger) (*mysqlProvider, MySQL) {
			conn := mysql.New(conf)
			o := orm.NewDB(conn, orm.Plugins{Logger: log})
			return &mysqlProvider{conn: conn, conf: *conf, log: log}, o
		},
	}
}

type (
	mysqlProvider struct {
		conn schema.Connector
		conf ConfigMysql
		log  logger.Logger
	}

	//MySQL connection MySQL interface
	MySQL interface {
		Pool(name string) *orm.Stmt
	}
)

func (v *mysqlProvider) Up() error {
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
		v.log.WithFields(
			logger.Fields{vv.Name: fmt.Sprintf("%s:%d", vv.Host, vv.Port)},
		).Infof("MySQL connect")
	}
	return nil
}

func (v *mysqlProvider) Down() error {
	return v.conn.Close()
}
