package macaroni

import (
	"context"
	"encoding/json"
	"io"
	"log"
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

func Run(conf *Config, src io.Reader) error {
	var report horenso.Report
	dec := json.NewDecoder(src)
	err := dec.Decode(&report)
	if err != nil {
		return errors.Wrap(err, "couldnot parse report")
	}

	eg := errgroup.Group{}
	if conf.Mackerel != nil {
		eg.Go(func() error {
			return reportToMackerel(&report, conf.Mackerel)
		})
	}
	if conf.Slack != nil {
		if report.ExitCode == 0 && conf.Slack.MuteOnNormal {
			// do not report
			log.Println("[debug] mute on normal exit")
		} else {
			eg.Go(func() error {
				return reportToSlack(&report, conf.Slack)
			})
		}
	}
	return eg.Wait()
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

func head(str string, n int) string {
	l := utf8.RuneCountInString(str)
	if l < n {
		return str
	}
	return string([]rune(str)[:n])
}
