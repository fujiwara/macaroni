package macaroni

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type slackConfigTest struct {
	env     map[string]string
	conf    *SlackConfig
	err     error
	payload *Payload
}

var slackConfigTests = []slackConfigTest{
	slackConfigTest{
		env:  map[string]string{},
		conf: nil,
		err:  nil,
	},
	slackConfigTest{
		env: map[string]string{
			"SLACK_ENDPOINT": "https://localhost/slack",
		},
		conf: nil,
		err:  nil,
	},
	slackConfigTest{
		env: map[string]string{
			"SLACK_CHANNEL": "#test",
		},
		conf: nil,
		err:  nil,
	},
	slackConfigTest{
		env: map[string]string{
			"SLACK_ENDPOINT":     "https://localhost/slack",
			"SLACK_CHANNEL":      "#general",
			"SLACK_USERNAME":     "macaroni",
			"SLACK_ICON_EMOJI":   ":x:",
			"SLACK_MENTION":      "@here",
			"SLACK_PASTEBIN_CMD": "tail -1",
		},
		conf: &SlackConfig{
			Endpoint:    "https://localhost/slack",
			Channel:     "#general",
			Username:    "macaroni",
			IconEmoji:   ":x:",
			Mention:     "@here",
			PasteBinCmd: "tail -1",
		},
		err: nil,
		payload: &Payload{
			Channel:   "#general",
			IconEmoji: ":x:",
			LinkNames: 1,
			Username:  "macaroni",
			Text:      "horenso reports success",
			Attachments: []Attachment{
				Attachment{
					Fallback: "95030\n perl -E 'say 1;warn \"$$\\n\";'",
					Color:    "#33cc33",
					Fields: []Field{
						Field{Title: "Hostname", Value: "webserver.example.com"},
						Field{Title: "Command", Value: `perl -E 'say 1;warn "$$\n";'`},
						Field{Title: "ExitCode", Value: "0"},
						Field{Title: "Output", Value: "```\n95030\n```"},
						Field{Title: "Started", Value: "2015-12-28T00:37:10.494282399+09:00"},
						Field{Title: "Ended", Value: "2015-12-28T00:37:10.546466379+09:00"},
					},
				},
			},
		},
	},
}

func TestSlack(t *testing.T) {
	defer func() { env = nil }()

	for i, suite := range slackConfigTests {
		env = suite.env

		sc, err := buildSlackConf()
		if diff := cmp.Diff(suite.conf, sc); diff != "" {
			t.Error(i, diff)
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

		if suite.payload != nil {
			payload := buildSlackPayload(&testReport, sc)
			t.Logf("%#v", payload)
			if diff := cmp.Diff(suite.payload, &payload); diff != "" {
				t.Error(diff)
			}
		}
	}
}
