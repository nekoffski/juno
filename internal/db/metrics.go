package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

type poolCollector struct {
	pool        *pgxpool.Pool
	totalConns  *prometheus.Desc
	acquiredCns *prometheus.Desc
	idleConns   *prometheus.Desc
	maxConns    *prometheus.Desc
}

func NewPoolCollector(pool *pgxpool.Pool, reg prometheus.Registerer) {
	c := &poolCollector{
		pool: pool,
		totalConns: prometheus.NewDesc(
			"juno_db_pool_conns_total",
			"Total number of connections in the pool.",
			nil, nil,
		),
		acquiredCns: prometheus.NewDesc(
			"juno_db_pool_conns_acquired",
			"Number of currently acquired connections in the pool.",
			nil, nil,
		),
		idleConns: prometheus.NewDesc(
			"juno_db_pool_conns_idle",
			"Number of currently idle connections in the pool.",
			nil, nil,
		),
		maxConns: prometheus.NewDesc(
			"juno_db_pool_max_conns",
			"Maximum number of connections allowed in the pool.",
			nil, nil,
		),
	}
	reg.MustRegister(c)
}

func (c *poolCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.totalConns
	ch <- c.acquiredCns
	ch <- c.idleConns
	ch <- c.maxConns
}

func (c *poolCollector) Collect(ch chan<- prometheus.Metric) {
	stat := c.pool.Stat()
	ch <- prometheus.MustNewConstMetric(c.totalConns, prometheus.GaugeValue, float64(stat.TotalConns()))
	ch <- prometheus.MustNewConstMetric(c.acquiredCns, prometheus.GaugeValue, float64(stat.AcquiredConns()))
	ch <- prometheus.MustNewConstMetric(c.idleConns, prometheus.GaugeValue, float64(stat.IdleConns()))
	ch <- prometheus.MustNewConstMetric(c.maxConns, prometheus.GaugeValue, float64(stat.MaxConns()))
}

type queryTracerKey struct{}

type QueryTracer struct {
	hist *prometheus.HistogramVec
}

func NewQueryTracer(reg prometheus.Registerer) *QueryTracer {
	hist := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "juno_db_query_duration_seconds",
		Help:    "Duration of database queries in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"command"})
	reg.MustRegister(hist)
	return &QueryTracer{hist: hist}
}

func (t *QueryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryStartData) context.Context {
	return context.WithValue(ctx, queryTracerKey{}, time.Now())
}

func (t *QueryTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	start, ok := ctx.Value(queryTracerKey{}).(time.Time)
	if !ok {
		return
	}
	t.hist.WithLabelValues(data.CommandTag.String()).Observe(time.Since(start).Seconds())
}
