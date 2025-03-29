package libpgx

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/aliocode/golib/libmode"
	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

var (
	_    Connector = (*pgxpool.Pool)(nil)
	PSQL           = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
)

type Connector interface {
	Transactable
	Executable
	Queryable
	Close()
}
type Transactable interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

type Executable interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
}

type Queryable interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

func NewPoolWrapper(ctx context.Context, dsn string, tracerName string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("invalid dsn: %w", err)
	}

	traceOpts := []otelpgx.Option{otelpgx.WithAttributes(semconv.ServiceName(tracerName))}
	if libmode.GetMode() != libmode.ServiceModeProd {
		// include query params in meta
		traceOpts = append(traceOpts, otelpgx.WithIncludeQueryParameters())
	}
	cfg.ConnConfig.Tracer = otelpgx.NewTracer(traceOpts...)

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("postgresql is not reachable with given config: %w", err)
	}

	return pool, nil
}
