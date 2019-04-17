package macaroni

import (
	"log"
	"regexp"

	"github.com/Songmu/horenso"
	mackerel "github.com/mackerelio/mackerel-client-go"
)

var DefaultMetricNamePrefix = "horenso.report"
var MetricNameNoramlizeRegexp = regexp.MustCompile(`[^0-9a-zA-Z_-]`)

type MackerelConfig struct {
	ApiKey           string
	MetricNamePrefix string
	MetricName       string
	Service          string
	HostID           string
}

func normalize(name string) string {
	return head(MetricNameNoramlizeRegexp.ReplaceAllString(name, "_"), 64)
}

func reportToMackerel(report horenso.Report, conf *MackerelConfig) error {
	log.Println("[info] report to Mackerel")

	client := mackerel.NewClient(conf.ApiKey)
	var name string
	if conf.MetricName == "" {
		name = normalize(report.Command)
	} else {
		name = conf.MetricName
	}

	statusMV := &mackerel.MetricValue{
		Name: conf.MetricNamePrefix + ".error." + name,
		Time: report.EndAt.Unix(),
	}
	if report.ExitCode == 0 {
		statusMV.Value = 0
	} else {
		statusMV.Value = 1
	}
	log.Printf("[debug] %#v", *statusMV)

	elapsedMV := &mackerel.MetricValue{
		Name:  conf.MetricNamePrefix + ".elapsed." + name,
		Time:  report.EndAt.Unix(),
		Value: report.EndAt.Sub(*report.StartAt).Seconds(),
	}
	log.Printf("[debug] %#v", *elapsedMV)

	if conf.Service != "" {
		log.Printf("[info] post service metrics to %s", conf.Service)
		return client.PostServiceMetricValues(
			conf.Service,
			[]*mackerel.MetricValue{statusMV, elapsedMV},
		)
	}

	return nil
}
