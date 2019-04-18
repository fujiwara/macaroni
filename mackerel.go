package macaroni

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/Songmu/horenso"
	"github.com/mackerelio/mkr/mackerelclient"

	agentConfig "github.com/mackerelio/mackerel-agent/config"
	mackerel "github.com/mackerelio/mackerel-client-go"
)

var (
	DefaultMetricNamePrefix   = "horenso.report"
	MetricNameNoramlizeRegexp = regexp.MustCompile(`[^0-9a-zA-Z_-]`)
	MetricNameTruncateRegexp  = regexp.MustCompile(`_{2,}`)
)

type MackerelConfig struct {
	ApiKey           string
	MetricNamePrefix string
	MetricName       string
	Service          string
	HostID           string
}

func buildMackerelConf() (*MackerelConfig, error) {
	target := getenv("MACKEREL_TARGET")
	if target == "" {
		// disabled
		return nil, nil
	}

	var prefix string
	if prefix = getenv("MACKEREL_METRIC_NAME_PREFIX"); prefix == "" {
		prefix = DefaultMetricNamePrefix
	}
	mc := &MackerelConfig{
		MetricNamePrefix: prefix,
		MetricName:       getenv("MACKEREL_METRIC_NAME"),
	}
	if strings.HasPrefix(target, "host:") {
		n := strings.SplitN(target, ":", 2)
		if mc.HostID = n[1]; mc.HostID == "" {
			mc.HostID = mackerelclient.LoadHostIDFromConfig(agentConfig.DefaultConfig.Conffile)
		}
	} else if strings.HasPrefix(target, "service:") {
		n := strings.SplitN(target, ":", 2)
		if mc.Service = n[1]; mc.Service == "" {
			return nil, fmt.Errorf("invalid MACKEREL_TARGET=%s service name required", target)
		}
	} else {
		return nil, fmt.Errorf("invalid MACKEREL_TARGET=%s service: or host: is required", target)
	}

	mc.ApiKey = mackerelclient.LoadApikeyFromEnvOrConfig(
		agentConfig.DefaultConfig.Conffile,
	)
	if mc.ApiKey == "" {
		mc.ApiKey = getenv("MACKEREL_APIKEY") // works for test only
	}
	if mc.ApiKey == "" {
		return nil, errors.New("unable to get Mackerel API key")
	}
	return mc, nil
}

func buildMetricValues(report horenso.Report, conf *MackerelConfig) []*mackerel.MetricValue {
	var name string
	if conf.MetricName == "" {
		name = normalize(report.Command)
	} else {
		name = conf.MetricName
	}

	return []*mackerel.MetricValue{
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
}

func reportToMackerel(report horenso.Report, conf *MackerelConfig) error {
	log.Println("[info] report to Mackerel")

	values := buildMetricValues(report, conf)
	b, _ := json.Marshal(values)
	log.Printf("[debug] %s", b)

	client := mackerel.NewClient(conf.ApiKey)

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

func normalize(name string) string {
	name = MetricNameNoramlizeRegexp.ReplaceAllString(name, "_")
	name = MetricNameTruncateRegexp.ReplaceAllString(name, "_")
	return head(name, 64)
}
