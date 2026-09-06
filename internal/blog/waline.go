package blog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type WalineClient struct {
	baseURL string
	hc      *http.Client
}

func NewWalineClient(base string) *WalineClient {
	return &WalineClient{
		baseURL: strings.TrimRight(base, "/"),
		hc:      &http.Client{Timeout: 10 * time.Second},
	}
}

type WalineComment struct {
	ObjectID  int64     `json:"objectId"`
	Nick      string    `json:"nick"`
	Comment   string    `json:"comment"`
	Link      string    `json:"link"`
	Avatar    string    `json:"avatar"`
	InsertedAt time.Time `json:"insertedAt"`
	Browser   string    `json:"browser"`
	OS        string    `json:"os"`
	Type      string    `json:"type"`
	Children  []WalineComment `json:"children"`
}

type walineListResp struct {
	Count    int             `json:"count"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
	Data     []WalineComment `json:"data"`
}

type walineErrorResp struct {
	ErrNo  int    `json:"errno"`
	ErrMsg string `json:"errmsg"`
}

func (c *WalineClient) List(path string) ([]WalineComment, error) {
	if path == "" {
		path = "/"
	}
	u := fmt.Sprintf("%s/comment?path=%s&pageSize=50", c.baseURL, url.QueryEscape(path))
	resp, err := c.hc.Get(u)
	if err != nil {
		return nil, fmt.Errorf("waline GET: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("waline GET status %d", resp.StatusCode)
	}
	var r walineListResp
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("waline decode: %w", err)
	}
	return r.Data, nil
}

type WalinePost struct {
	Nick    string `json:"nick"`
	Mail    string `json:"mail"`
	Link    string `json:"link,omitempty"`
	Comment string `json:"comment"`
	URL     string `json:"url"`
}

type walinePostResp struct {
	ErrNo  int            `json:"errno"`
	ErrMsg string         `json:"errmsg"`
	Data   WalineComment  `json:"data"`
}

func (c *WalineClient) Post(nick, mail, link, comment, urlPath string) error {
	body := WalinePost{
		Nick:    nick,
		Mail:    mail,
		Link:    link,
		Comment: comment,
		URL:     urlPath,
	}
	buf, _ := json.Marshal(body)
	resp, err := c.hc.Post(
		c.baseURL+"/comment",
		"application/json",
		strings.NewReader(string(buf)),
	)
	if err != nil {
		return fmt.Errorf("waline POST: %w", err)
	}
	defer resp.Body.Close()
	var r walinePostResp
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return fmt.Errorf("waline decode: %w", err)
	}
	if r.ErrNo != 0 {
		return fmt.Errorf("waline: %s", r.ErrMsg)
	}
	return nil
}

// StripHTML removes simple HTML tags for plain-text rendering in TUI.
func StripHTML(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}
