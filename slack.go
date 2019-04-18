package macaroni

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Songmu/horenso"
	"github.com/pkg/errors"
)

type SlackConfig struct {
	Endpoint     string
	Username     string
	IconEmoji    string
	Channel      string
	Mention      string
	PasteBinCmd  string
	MuteOnNormal bool
}

type Payload struct {
	Text        string       `json:"text"`
	Channel     string       `json:"channel"`
	LinkNames   int          `json:"link_names"`
	Username    string       `json:"username,omitempty"`
	IconEmoji   string       `json:"icon_emoji,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type Attachment struct {
	Fallback string  `json:"fallback"`
	Color    string  `json:"color"`
	Fields   []Field `json:"fields"`
}

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

func buildSlackConf() (*SlackConfig, error) {
	endpoint := getenv("SLACK_ENDPOINT")
	channel := getenv("SLACK_CHANNEL")

	if endpoint == "" || channel == "" {
		if endpoint != "" || channel != "" {
			log.Println("[warn] enable to Slack reporter, required SLACK_ENDPOINT and SLACK_CHANNEL both. Slack reporter disabled.")
		}
		return nil, nil
	}
	// ignore error because default false
	muteOnNormal, _ := strconv.ParseBool(getenv("SLACK_MUTE_ON_NORMAL"))
	sc := &SlackConfig{
		Endpoint:     endpoint,
		Channel:      channel,
		Username:     getenv("SLACK_USERNAME"),
		IconEmoji:    getenv("SLACK_ICON_EMOJI"),
		Mention:      getenv("SLACK_MENTION"),
		PasteBinCmd:  getenv("SLACK_PASTEBIN_CMD"),
		MuteOnNormal: muteOnNormal,
	}
	return sc, nil
}

func buildSlackPayload(report horenso.Report, conf *SlackConfig) Payload {
	var message string
	if report.ExitCode == 0 {
		message = "horenso reports success"
	} else {
		message = "horenso reports error!"
		if conf.Mention != "" {
			message += " " + conf.Mention
		}
	}
	payload := Payload{
		Text:      message,
		Channel:   conf.Channel,
		LinkNames: 1,
	}
	if conf.Username != "" {
		payload.Username = conf.Username
	}
	if conf.IconEmoji != "" {
		payload.IconEmoji = conf.IconEmoji
	}

	var output string
	if conf.PasteBinCmd != "" {
		var err error
		output, err = writeToCommand(conf.PasteBinCmd, report.Output)
		if err != nil {
			log.Printf("[warn] failed to exec %v %s", conf.PasteBinCmd, err)
			output = report.Output
		}
	} else {
		output = report.Output
	}

	payload.Attachments = []Attachment{
		Attachment{
			Fallback: output + " " + report.Command,
			Color:    color(report.ExitCode),
			Fields: []Field{
				Field{
					Title: "command",
					Value: report.Command,
				},
				Field{
					Title: "hostname",
					Value: report.Hostname,
				},
				Field{
					Title: "exitCode",
					Value: strconv.Itoa(report.ExitCode),
				},
				Field{
					Title: "output",
					Value: "```\n" + tail(output, MaxOutputLength) + "```",
				},
				Field{
					Title: "started",
					Value: report.StartAt.Format(time.RFC3339),
				},
				Field{
					Title: "ended",
					Value: report.EndAt.Format(time.RFC3339),
				},
			},
		},
	}
	return payload
}

func reportToSlack(report horenso.Report, conf *SlackConfig) error {
	log.Println("[info] report to Slack")

	payload := buildSlackPayload(report, conf)

	b, _ := json.Marshal(payload)
	b = bytes.ReplaceAll(b, []byte{'&'}, []byte("&amp;"))
	b = bytes.ReplaceAll(b, []byte{'<'}, []byte("&lt;"))
	b = bytes.ReplaceAll(b, []byte{'>'}, []byte("&gt;"))
	log.Println("[debug] payload:", string(b))

	resp, err := http.Post(conf.Endpoint, "application/json", bytes.NewReader(b))
	if err != nil {
		return errors.Wrapf(err, "failed to post to Slack endpoint %s", conf.Endpoint)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to post to Slack with status %d", resp.StatusCode)
	}
	log.Println("[info] posted to Slack")

	return nil
}
