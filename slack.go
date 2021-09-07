package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

func doSlack(slackUrl, color, title, text string) error {
	payload := fmt.Sprintf(`
{
  "blocks": [],
  "attachments": [
    {
		"mrkdwn_in": ["text"],
		"color": "%s",
		"text": "*%s* \n %s"
	}
  ]
}
`, color, title, text)

	resp, err := http.PostForm(slackUrl, url.Values{"payload": {payload}})
	if err != nil {
		return errors.WithStack(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.WithStack(err)
	}

	if resp.StatusCode != 200 {
		return errors.Errorf("slack error: %s", string(body))
	}

	return nil
}
