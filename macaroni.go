package macaroni

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Songmu/horenso"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

var MaxOutputLength = 1000
var CommandTimeout = 30 * time.Second

type Config struct {
	Slack    *SlackConfig
	Mackerel *MackerelConfig
}

type SlackConfig struct {
	Endpoint     string
	Username     string
	IconEmoji    string
	Channel      string
	Mention      string
	PasteBinCmd  string
	MuteOnNormal bool
}

type MackerelConfig struct {
	ApiKey           string
	MetricType       string
	MetricNamePrefix string
	Service          string
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

func Run(conf *Config, src io.Reader) error {
	var report horenso.Report
	dec := json.NewDecoder(src)
	err := dec.Decode(&report)
	if err != nil {
		return errors.Wrap(err, "couldnot parse report")
	}

	eg := errgroup.Group{}
	if conf.Mackerel.ApiKey != "" {
		eg.Go(func() error {
			return reportToMackerel(report, conf.Mackerel)
		})
	}
	if conf.Slack.Endpoint != "" {
		if report.ExitCode == 0 && conf.Slack.MuteOnNormal {
			// do not report
			log.Println("[debug] mute on normal exit")
		} else {
			eg.Go(func() error {
				return reportToSlack(report, conf.Slack)
			})
		}
	}
	return eg.Wait()
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

func reportToMackerel(report horenso.Report, conf *MackerelConfig) error {
	log.Println("[info] report to Mackerel")
	return nil
}

func color(code int) (color string) {
	switch code {
	case 0:
		color = "#33cc33"
	default:
		color = "#d22a3c"
	}
	return
}

func writeToCommand(command string, input string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), CommandTimeout)
	defer cancel()

	var commands []string
	if strings.Index(command, " ") != -1 {
		commands = []string{"sh", "-c", command}
	} else {
		commands = []string{command}
	}
	log.Println("[debug] execute:", strings.Join(commands, " "))
	cmd := exec.CommandContext(ctx, commands[0], commands[1:]...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, input)
	}()

	b, err := cmd.CombinedOutput()
	return string(b), err
}

func tail(str string, n int) string {
	l := utf8.RuneCountInString(str)
	if l < n {
		return str
	}
	return string([]rune(str)[l-n:])
}
