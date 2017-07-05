package metrics

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	Port      string
	Path      = "/metrics"
	Namespace = "alarm_assist"
	Subsystem = "metrics"

	requestCount    *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	errorCount      *prometheus.CounterVec
)

type (
	System string
	Cause  string
)

var (
	Alarm_api_phone      System = "alarm_api_phone"
	Alarm_api_im         System = "alarm_api_im"
	Alarm_api_mail       System = "alarm_api_mail"
	Alarm_api_pattern    System = "alarm_api_pattern"
	Alarm_api_user       System = "alarm_api_user"
	Alarm_api_subscrible System = "alarm_api_subscrible"

	NodataFit   Cause = "nodata_fit"
	InvalidJson Cause = "invalid_json"
	DbError     Cause = "db_error"
	ArgsError   Cause = "args_error"
)

func defineMetrics() {
	requestCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: Namespace,
		Subsystem: Subsystem,
		Name:      "request_count_total",
		Help:      "Counter of requests made.",
	}, []string{"system"})

	requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: Namespace,
		Subsystem: Subsystem,
		Name:      "request_duration_seconds",
		Help:      "Histogram of the time (in seconds) each request.",
		Buckets:   append([]float64{0.001, 0.003}, prometheus.DefBuckets...),
	}, []string{"system"})

	errorCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: Namespace,
		Subsystem: Subsystem,
		Name:      "error_count_total",
		Help:      "Counter of requests resulting in an error.",
	}, []string{"system", "cause"})

}

func Metrics() error {

	if Port == "" {
		return nil
	}

	_, err := strconv.Atoi(Port)
	if err != nil {
		fmt.Errorf("bad port for prometheus: %s", Port)
	}

	defineMetrics()

	prometheus.MustRegister(requestCount)
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(errorCount)

	http.Handle(Path, prometheus.Handler())
	go func() {
		fmt.Errorf("%s", http.ListenAndServe(":"+Port, nil))
	}()
	return nil
}

func ReportDuration(start time.Time, sys System) {
	if requestDuration == nil {
		return
	}

	requestDuration.WithLabelValues(string(sys)).Observe(float64(time.Since(start)) / float64(time.Second))
}

func ReportRequestCount(sys System) {
	if requestCount == nil {
		return
	}

	requestCount.WithLabelValues(string(sys)).Inc()
}

func ReportErrorCount(cause Cause, sys System) {
	if errorCount == nil {
		return
	}
	errorCount.WithLabelValues(string(sys), string(cause)).Inc()
}
