package macaroni

import (
	"log"
	"os"
)

var env map[string]string

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

func getenv(name string) string {
	if env == nil {
		return os.Getenv(name)
	}
	if v, ok := env[name]; ok {
		return v
	}
	return ""
}
