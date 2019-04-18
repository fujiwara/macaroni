package macaroni

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	agentConfig "github.com/mackerelio/mackerel-agent/config"
	"github.com/mackerelio/mkr/mackerelclient"
)

var Env *sync.Map

func buildMackerelConf() (*MackerelConfig, error) {
	target := Getenv("MACKEREL_TARGET")
	if target == "" {
		// disabled
		return nil, nil
	}

	var prefix string
	if prefix = Getenv("MACKEREL_METRIC_NAME_PREFIX"); prefix == "" {
		prefix = DefaultMetricNamePrefix
	}
	mc := &MackerelConfig{
		MetricNamePrefix: prefix,
		MetricName:       Getenv("MACKEREL_METRIC_NAME"),
	}
	if strings.HasPrefix(target, "host:") {
		n := strings.SplitN(target, ":", 2)
		if mc.HostID = n[1]; mc.HostID == "" {
			mc.HostID = mackerelclient.LoadHostIDFromConfig(agentConfig.DefaultConfig.Conffile)
		}
	} else if strings.HasPrefix(target, "service:") {
		n := strings.SplitN(target, ":", 2)
		if mc.Service = n[1]; mc.Service == "" {
			return nil, fmt.Errorf("invalid MACKEREL_TARGET=%s. service name required", target)
		}
	} else {
		return nil, fmt.Errorf("invalid MACKEREL_TARGET=%s. service: or host: is required", target)
	}

	mc.ApiKey = mackerelclient.LoadApikeyFromEnvOrConfig(
		agentConfig.DefaultConfig.Conffile,
	)
	if mc.ApiKey == "" {
		return nil, errors.New("unable to get Mackerel API key")
	}
	return mc, nil
}

func buildSlackConf() (*SlackConfig, error) {
	endpoint := Getenv("SLACK_ENDPOINT")
	channel := Getenv("SLACK_CHANNEL")
	if endpoint == "" || channel == "" {
		// disabled
		return nil, nil
	}
	// ignore error because default false
	muteOnNormal, _ := strconv.ParseBool(Getenv("SLACK_MUTE_ON_NORMAL"))
	sc := &SlackConfig{
		Endpoint:     endpoint,
		Channel:      channel,
		Username:     Getenv("SLACK_USERNAME"),
		IconEmoji:    Getenv("SLACK_ICON_EMOJI"),
		Mention:      Getenv("SLACK_MENTION"),
		PasteBinCmd:  Getenv("SLACK_PASTEBIN_CMD"),
		MuteOnNormal: muteOnNormal,
	}
	return sc, nil
}

func BuildConfig() *Config {
	conf := &Config{}

	if mc, err := buildMackerelConf(); err != nil {
		log.Printf("[warn] Mackerel reporter disabled. %s", err)
	} else {
		conf.Mackerel = mc
	}
	if sc, err := buildSlackConf(); err != nil {
		log.Printf("[warn] Slack reporter disabled. %s", err)
	} else {
		conf.Slack = sc
	}

	return conf
}

func Getenv(name string) string {
	if Env == nil {
		return os.Getenv(name)
	}
	if v, ok := Env.Load(name); ok {
		return v.(string)
	}
	return ""
}
