package macaroni

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

func reportToSlack(report horenso.Report, conf *SlackConfig) error {
	log.Println("[info] report to Slack")
	var message string
	if report.ExitCode == 0 {
		message = fmt.Sprintf(
			":ok: [%s] horenso reports success",
			report.Hostname,
		)
	} else {
		message = fmt.Sprintf(
			":anger: [%s] horenso reports error! exit with %d",
			report.Hostname,
			report.ExitCode,
		)
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
