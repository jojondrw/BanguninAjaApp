package databasetest

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"sync"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	Begin    = "BEGIN"
	Commit   = "COMMIT"
	Rollback = "ROLLBACK"
)

var errPrepareUnsupported = errors.New("recorder does not prepare statements")

type Result struct {
	Columns  []string
	Rows     [][]driver.Value
	Affected int64
	Err      error
}

type Query struct {
	SQL  string
	Args []any
}

type Recorder struct {
	mutex   sync.Mutex
	results []Result
	queries []Query
}

func Open(t *testing.T, results ...Result) (*gorm.DB, *Recorder) {
	t.Helper()

	recorder := &Recorder{results: results}
	connection := sql.OpenDB(recorder)
	t.Cleanup(func() { _ = connection.Close() })

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: connection}), &gorm.Config{
		DisableAutomaticPing:   true,
		SkipDefaultTransaction: true,
		TranslateError:         true,
	})
	if err != nil {
		t.Fatalf("open recording database: %v", err)
	}
	return db, recorder
}

func (r *Recorder) Queries() []Query {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return append([]Query(nil), r.queries...)
}

func (r *Recorder) Connect(context.Context) (driver.Conn, error) {
	return &connection{recorder: r}, nil
}

func (r *Recorder) Driver() driver.Driver {
	return recordingDriver{recorder: r}
}

func (r *Recorder) answer(query string, args []driver.NamedValue) Result {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	values := make([]any, 0, len(args))
	for _, arg := range args {
		values = append(values, arg.Value)
	}
	r.queries = append(r.queries, Query{SQL: query, Args: values})

	if len(r.results) == 0 {
		return Result{}
	}
	next := r.results[0]
	r.results = r.results[1:]
	return next
}

func (r *Recorder) mark(statement string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.queries = append(r.queries, Query{SQL: statement})
}

type recordingDriver struct {
	recorder *Recorder
}

func (d recordingDriver) Open(string) (driver.Conn, error) {
	return &connection{recorder: d.recorder}, nil
}

type connection struct {
	recorder *Recorder
}

func (c *connection) Prepare(string) (driver.Stmt, error) {
	return nil, errPrepareUnsupported
}

func (c *connection) Close() error {
	return nil
}

func (c *connection) Begin() (driver.Tx, error) {
	c.recorder.mark(Begin)
	return transaction{recorder: c.recorder}, nil
}

func (c *connection) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	result := c.recorder.answer(query, args)
	if result.Err != nil {
		return nil, result.Err
	}
	return &rows{columns: result.Columns, values: result.Rows}, nil
}

func (c *connection) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	result := c.recorder.answer(query, args)
	if result.Err != nil {
		return nil, result.Err
	}
	return driver.RowsAffected(result.Affected), nil
}

type transaction struct {
	recorder *Recorder
}

func (t transaction) Commit() error {
	t.recorder.mark(Commit)
	return nil
}

func (t transaction) Rollback() error {
	t.recorder.mark(Rollback)
	return nil
}

type rows struct {
	columns  []string
	values   [][]driver.Value
	position int
}

func (r *rows) Columns() []string {
	return r.columns
}

func (r *rows) Close() error {
	return nil
}

func (r *rows) Next(dest []driver.Value) error {
	if r.position >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.position])
	r.position++
	return nil
}
