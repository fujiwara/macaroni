package macaroni

import (
	"encoding/json"
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

	values := []*mackerel.MetricValue{
		// error occuered
		&mackerel.MetricValue{
			Name:  conf.MetricNamePrefix + ".error." + name,
			Time:  report.EndAt.Unix(),
			Value: boolToInt(report.ExitCode != 0),
		},
		// elapsed time
		&mackerel.MetricValue{
			Name:  conf.MetricNamePrefix + ".elapsed." + name,
			Time:  report.EndAt.Unix(),
			Value: report.EndAt.Sub(*report.StartAt).Seconds(),
		},
	}
	b, _ := json.Marshal(values)
	log.Printf("[debug] %s", b)

	if conf.Service != "" {
		log.Printf("[info] post service metrics to %s", conf.Service)
		return client.PostServiceMetricValues(conf.Service, values)
	}

	return nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
