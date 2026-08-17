package metrics

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	dto "github.com/prometheus/client_model/go"
)

// fakeDriver is a minimal database/sql driver for exercising the DB wrapper.
type fakeDriver struct{}

func (fakeDriver) Open(_ string) (driver.Conn, error) { return fakeConn{}, nil }

type fakeConn struct{}

func (fakeConn) Prepare(_ string) (driver.Stmt, error) { return fakeStmt{}, nil }
func (fakeConn) Close() error                          { return nil }
func (fakeConn) Begin() (driver.Tx, error)             { return nil, nil }
func (fakeConn) QueryContext(_ context.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
	return fakeRows{}, nil
}
func (fakeConn) ExecContext(_ context.Context, _ string, _ []driver.NamedValue) (driver.Result, error) {
	return fakeResult{}, nil
}

type fakeStmt struct{}

func (fakeStmt) Close() error  { return nil }
func (fakeStmt) NumInput() int { return -1 }
func (fakeStmt) Exec([]driver.Value) (driver.Result, error) {
	return fakeResult{}, nil
}
func (fakeStmt) Query([]driver.Value) (driver.Rows, error) { return fakeRows{}, nil }

type fakeRows struct{}

func (fakeRows) Columns() []string         { return []string{"x"} }
func (fakeRows) Close() error              { return nil }
func (fakeRows) Next([]driver.Value) error { return io.EOF }

type fakeResult struct{}

func (fakeResult) LastInsertId() (int64, error) { return 0, nil }
func (fakeResult) RowsAffected() (int64, error) { return 1, nil }

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	sql.Register("levelup-fake", fakeDriver{})
	db, err := sql.Open("levelup-fake", "x")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestDBWrapperObservesClickHouseMetrics(t *testing.T) {
	ctx := context.Background()
	wrapped := NewDB(newTestDB(t))

	if _, err := wrapped.QueryContext(ctx, "SELECT 1"); err != nil {
		t.Fatalf("QueryContext: %v", err)
	}
	if _, err := wrapped.ExecContext(ctx, "INSERT INTO x"); err != nil {
		t.Fatalf("ExecContext: %v", err)
	}
	_ = wrapped.QueryRowContext(ctx, "SELECT 1").Scan(new(int))

	queryCount := sampleCount(CHQueryDuration.WithLabelValues("query"))
	execCount := sampleCount(CHQueryDuration.WithLabelValues("exec"))
	rowCount := sampleCount(CHQueryDuration.WithLabelValues("queryrow"))
	if queryCount != 1 || execCount != 1 || rowCount != 1 {
		t.Fatalf("expected 1 sample each, got query=%d exec=%d row=%d", queryCount, execCount, rowCount)
	}
}

func TestBusinessCounters(t *testing.T) {
	UsersRegistered.Inc()
	QuestsCompleted.Inc()
	NodesCompleted.Inc()
	Purchases.Inc()
	RoadmapsInstalled.Inc()
	GoldSpent.Add(25)

	if got := testutil.ToFloat64(UsersRegistered); got != 1 {
		t.Fatalf("users registered = %v, want 1", got)
	}
	if got := testutil.ToFloat64(QuestsCompleted); got != 1 {
		t.Fatalf("quests completed = %v, want 1", got)
	}
	if got := testutil.ToFloat64(Purchases); got != 1 {
		t.Fatalf("purchases = %v, want 1", got)
	}
	if got := testutil.ToFloat64(GoldSpent); got != 25 {
		t.Fatalf("gold spent = %v, want 25", got)
	}
}

func sampleCount(obs prometheus.Observer) int {
	hist, ok := obs.(prometheus.Histogram)
	if !ok {
		return 0
	}
	var m dto.Metric
	_ = hist.Write(&m)
	if m.GetHistogram() == nil {
		return 0
	}
	return int(m.GetHistogram().GetSampleCount())
}
