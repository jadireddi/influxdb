package http

import (
	"net/http"
	"net/http/httptest"
	_ "net/http/pprof"
	"testing"

	"github.com/influxdata/influxdb/kit/prom"
	"github.com/influxdata/influxdb/kit/prom/promtest"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

func TestHandler_ServeHTTP(t *testing.T) {
	type fields struct {
		name           string
		MetricsHandler http.Handler
		ReadyHandler   http.Handler
		HealthHandler  http.Handler
		DebugHandler   http.Handler
		Handler        http.Handler
		requests       *prometheus.CounterVec
		requestDur     *prometheus.HistogramVec
		Logger         *zap.Logger
	}
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "howdy",
			fields: fields{
				name:    "doody",
				Handler: http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
				Logger:  zap.NewNop(),
			},
			args: args{
				r: httptest.NewRequest(http.MethodGet, "/", nil),
				w: httptest.NewRecorder(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				name:    tt.fields.name,
				Handler: tt.fields.Handler,
				Logger:  tt.fields.Logger,
			}
			h.initMetrics()
			reg := prom.NewRegistry()
			reg.MustRegister(h)

			h.ServeHTTP(tt.args.w, tt.args.r)

			mfs, err := reg.Gather()
			if err != nil {
				t.Fatal(err)
			}

			c := promtest.MustFindMetric(t, mfs, "my_random_counter", nil)
			if got := c.GetCounter().GetValue(); got != 1 {
				t.Fatalf("expected counter to be 1, got %v", got)
			}

			g := promtest.MustFindMetric(t, mfs, "my_random_gaugevec", map[string]string{"label1": "one", "label2": "two"})
			if got := g.GetGauge().GetValue(); got != 3 {
				t.Fatalf("expected gauge to be 3, got %v", got)
			}

		})

	}
}
