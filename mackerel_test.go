package macaroni

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	mackerel "github.com/mackerelio/mackerel-client-go"
	"github.com/pkg/errors"
)

var testMackerelApiKey = "dummy"

type mackerelConfigTest struct {
	env    map[string]string
	conf   *MackerelConfig
	err    error
	values []*mackerel.MetricValue
}

var mackerelConfigTests = []mackerelConfigTest{
	mackerelConfigTest{
		env:  map[string]string{},
		conf: nil,
		err:  nil,
	},
	mackerelConfigTest{
		env: map[string]string{
			"MACKEREL_TARGET":             "service:foo",
			"MACKEREL_METRIC_NAME_PREFIX": "macaroni",
			"MACKEREL_METRIC_NAME":        "my_foo",
		},
		conf: &MackerelConfig{
			ApiKey:           testMackerelApiKey,
			MetricNamePrefix: "macaroni",
			MetricName:       "my_foo",
			Service:          "foo",
		},
		err: nil,
		values: []*mackerel.MetricValue{
			&mackerel.MetricValue{
				Name:  "macaroni.error.my_foo",
				Value: 0,
				Time:  1451230630,
			},
			&mackerel.MetricValue{
				Name:  "macaroni.elapsed.my_foo",
				Value: 0.05218398,
				Time:  1451230630,
			},
		},
	},
	mackerelConfigTest{
		env: map[string]string{
			"MACKEREL_TARGET": "service:",
		},
		conf: nil,
		err:  errors.New("invalid MACKEREL_TARGET=service: service name required"),
	},
	mackerelConfigTest{
		env: map[string]string{
			"MACKEREL_TARGET": "host:abcdefg",
		},
		conf: &MackerelConfig{
			ApiKey:           testMackerelApiKey,
			MetricNamePrefix: "horenso.report",
			HostID:           "abcdefg",
		},
		err: nil,
		values: []*mackerel.MetricValue{
			&mackerel.MetricValue{
				Name:  "horenso.report.error.perl_-E_say_1_warn_n_",
				Value: 0,
				Time:  1451230630,
			},
			&mackerel.MetricValue{
				Name:  "horenso.report.elapsed.perl_-E_say_1_warn_n_",
				Value: 0.05218398,
				Time:  1451230630,
			},
		},
	},
}

func TestMackerel(t *testing.T) {
	defer func() { env = nil }()

	for _, suite := range mackerelConfigTests {
		env = suite.env
		env["MACKEREL_APIKEY"] = testMackerelApiKey

		mc, err := buildMackerelConf()
		if diff := cmp.Diff(suite.conf, mc); diff != "" {
			t.Error(diff)
		}
		if suite.err != nil && err == nil {
			t.Errorf("expected error: %s but got nil", suite.err)
		} else if suite.err == nil && err != nil {
			t.Errorf("unexpected error: got %s", err)
		} else if suite.err != nil {
			if !strings.HasPrefix(err.Error(), suite.err.Error()) {
				t.Errorf("unexpected error: expected: %s, got %s", suite.err, err)
			}
		}

		if suite.values != nil {
			values := buildMetricValues(&testReport, mc)
			t.Logf("%#v", values)
			if diff := cmp.Diff(suite.values, values); diff != "" {
				t.Error(diff)
			}
		}
	}
}
