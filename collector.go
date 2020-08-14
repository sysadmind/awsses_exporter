package main

import (
	"net/http"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	namespace = "awsses_exporter"
)

// Exporter collects metrics from aws ses.
type Exporter struct {
	accessKey       string
	secretAccessKey string

	max24hoursend   *prometheus.Desc
	maxsendrate     *prometheus.Desc
	sentlast24hours *prometheus.Desc
}

// NewExporter returns an initialized exporter.
func NewExporter() *Exporter {
	return &Exporter{
		max24hoursend: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "max24hoursend"),
			"The maximum number of emails allowed to be sent in a rolling 24 hours.",
			[]string{"aws_region"},
			nil,
		),
		maxsendrate: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "maxsendrate"),
			"The maximum rate of emails allowed to be sent per second.",
			[]string{"aws_region"},
			nil,
		),
		sentlast24hours: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "sentlast24hours"),
			"The number of emails sent in the last 24 hours.",
			[]string{"aws_region"},
			nil,
		),
	}
}

// Describe describes all the metrics exported by the memcached exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.max24hoursend
	ch <- e.maxsendrate
	ch <- e.sentlast24hours
}

// Collect fetches the statistics from the configured memcached server, and
// delivers them as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	awsSesRegions := []string{"us-east-1", "us-west-2", "eu-west-1"}
	for _, regionName := range awsSesRegions {
		svc := ses.New(session.New(), aws.NewConfig().WithRegion(regionName))
		input := &ses.GetSendQuotaInput{}

		result, err := svc.GetSendQuota(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				default:
					log.Errorf("Failed to get sending quota from region %s: %s", regionName, aerr)
				}
			} else {
				log.Errorf("Failed to get sending quota from region %s: %s", regionName, err)
			}
			return
		}

		ch <- prometheus.MustNewConstMetric(e.max24hoursend, prometheus.GaugeValue, *result.Max24HourSend, regionName)
		ch <- prometheus.MustNewConstMetric(e.maxsendrate, prometheus.GaugeValue, *result.MaxSendRate, regionName)
		ch <- prometheus.MustNewConstMetric(e.sentlast24hours, prometheus.GaugeValue, *result.SentLast24Hours, regionName)
	}
}

func main() {
	var (
		listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9199").String()
		metricsPath   = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
	)
	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("awsses_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.Infoln("Starting awsses_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	prometheus.MustRegister(NewExporter())

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>AWS SES Exporter</title></head>
             <body>
             <h1>AWS SES Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	log.Infoln("Starting HTTP server on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
