package metrics

// https://robert-scherbarth.medium.com/measure-request-duration-with-prometheus-and-golang-adc6f4ca05fe

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus"
)

//statusRecorder to record the status code from the ResponseWriter
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *statusRecorder) WriteHeader(statusCode int) {
	rec.statusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}

type Measurer struct {
	histogram *prometheus.HistogramVec
}

func NewMeasurer() (*Measurer, error) {
	histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_server_request_duration_seconds",
		Help:    "Histogram of response time for handler in seconds",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	}, []string{"route", "method", "status_code"})

	err := prometheus.Register(histogram)
	if err != nil {
		return nil, err
	}

	return &Measurer{
		histogram: histogram,
	}, nil
}

func (m *Measurer) MeasureDuration(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := statusRecorder{w, 200}

		next.ServeHTTP(&rec, r)

		statusCode := strconv.Itoa(rec.statusCode)
		route := getRoutePattern(r)
		m.histogram.WithLabelValues(route, r.Method, statusCode).Observe(time.Since(start).Seconds())
	})
}

// getRoutePattern returns the route pattern from the chi context there are 3 conditions
// a) static routes "/example" => "/example"
// b) dynamic routes "/example/:id" => "/example/{id}"
// c) if nothing matches the output is undefined
func getRoutePattern(r *http.Request) string {
	reqContext := chi.RouteContext(r.Context())
	if pattern := reqContext.RoutePattern(); pattern != "" {
		return pattern
	}

	return "undefined"
}
