package easy_db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/jimmyseraph/sparkle/dialect"

	"go.uber.org/zap"
)

// https://github.com/golang/go/wiki/SQLDrivers
// _ "github.com/go-sql-driver/mysql"
// _ "github.com/lib/pq"
// _ "github.com/sijms/go-ora"
// _ "modernc.org/sqlite"

// Driver is a dialect.Driver implementation for SQL based databases.
type Driver struct {
	Conn
	dialect string
	log     *zap.Logger
}

// Open wraps the database/sql.Open method and returns a dialect.Driver.
func Open(driver, source string) (*Driver, error) {
	log, _ := zap.NewDevelopment()
	db, err := sql.Open(driver, source)
	if err != nil {
		log.Error("cannot open db connection", zap.String("driver", driver), zap.String("source", source))
		return nil, err
	}
	return &Driver{
		Conn{db},
		driver,
		log,
	}, nil
}

// OpenDB wraps the given database/sql.DB method with a Driver.
func OpenDB(driver string, db *sql.DB) *Driver {
	log, _ := zap.NewDevelopment()
	return &Driver{
		Conn{db},
		driver,
		log,
	}
}

// DB returns the underlying *sql.DB instance.
func (d Driver) DB() *sql.DB {
	return d.ExecQuerier.(*sql.DB)
}

// Tx starts and returns a transaction.
func (d *Driver) Tx(ctx context.Context) (dialect.Tx, error) {
	return d.BeginTx(ctx, nil)
}

// BeginTx starts a transaction with options.
func (d *Driver) BeginTx(ctx context.Context, opts *TxOptions) (dialect.Tx, error) {
	tx, err := d.DB().BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{
		ExecQuerier: Conn{tx},
		Tx:          tx,
	}, nil
}

// Dialect implements the dialect.Dialect method.
func (d Driver) Dialect() string {
	// If the underlying driver is wrapped with opencensus driver.
	for _, name := range []string{dialect.MySQL, dialect.SQLite, dialect.Postgres, dialect.Oracle} {
		if strings.HasPrefix(d.dialect, name) {
			return name
		}
	}
	return d.dialect
}

// Close closes the underlying connection.
func (d *Driver) Close() error { return d.DB().Close() }

// Tx implements dialect.Tx interface.
type Tx struct {
	dialect.ExecQuerier
	driver.Tx
}

// ExecQuerier wraps the standard Exec and Query methods.
type ExecQuerier interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// Conn implements dialect.ExecQuerier given ExecQuerier.
type Conn struct {
	ExecQuerier
}

// Exec implements the dialect.Exec method.
func (c Conn) Exec(ctx context.Context, query string, args, v interface{}) error {
	argv, ok := args.([]interface{})
	if !ok {
		return fmt.Errorf("dialect/sql: invalid type %T. expect []interface{} for args", v)
	}
	switch v := v.(type) {
	case nil:
		if _, err := c.ExecContext(ctx, query, argv...); err != nil {
			return err
		}
	case *sql.Result:
		res, err := c.ExecContext(ctx, query, argv...)
		if err != nil {
			return err
		}
		*v = res
	default:
		return fmt.Errorf("dialect/sql: invalid type %T. expect *sql.Result", v)
	}
	return nil
}

// Query implements the dialect.Query method.
func (c Conn) Query(ctx context.Context, query string, args, v interface{}) error {
	vr, ok := v.(*Rows)
	if !ok {
		return fmt.Errorf("dialect/sql: invalid type %T. expect *sql.Rows", v)
	}
	argv, ok := args.([]interface{})
	if !ok {
		return fmt.Errorf("dialect/sql: invalid type %T. expect []interface{} for args", args)
	}
	rows, err := c.QueryContext(ctx, query, argv...)
	if err != nil {
		return err
	}
	*vr = Rows{rows}
	return nil
}

var _ dialect.Driver = (*Driver)(nil)

type (
	// Rows wraps the sql.Rows to avoid locks copy.
	Rows struct{ *sql.Rows }
	// Result is an alias to sql.Result.
	Result = sql.Result
	// NullBool is an alias to sql.NullBool.
	NullBool = sql.NullBool
	// NullInt64 is an alias to sql.NullInt64.
	NullInt64 = sql.NullInt64
	// NullString is an alias to sql.NullString.
	NullString = sql.NullString
	// NullFloat64 is an alias to sql.NullFloat64.
	NullFloat64 = sql.NullFloat64
	// NullTime represents a time.Time that may be null.
	NullTime = sql.NullTime
	// TxOptions holds the transaction options to be used in DB.BeginTx.
	TxOptions = sql.TxOptions
)
