package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

var (
	ErrInvalidRequest = errors.New("invalid invalid request")
	ErrUnauthorized   = errors.New("unauthorized")
)

type FreshRSS struct {
	host      string
	authToken string
	client    *http.Client
}

func NewFreshRSS(host string) *FreshRSS {
	return &FreshRSS{
		host: host,
		client: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (g FreshRSS) Login(ctx context.Context, email, password string) (string, error) {
	body := url.Values{}
	body.Set("Email", email)
	body.Set("Passwd", password)

	var resp string
	if err := g.postRequest(ctx, "/accounts/ClientLogin", body, &resp); err != nil {
		return "", err
	}

	for line := range strings.SplitSeq(resp, "\n") {
		if after, ok := strings.CutPrefix(line, "Auth="); ok {
			return after, nil
		}
	}

	return "", ErrUnauthorized
}

func (g *FreshRSS) SetAuthToken(token string) {
	// todo: validate token
	g.authToken = token
}

func (g FreshRSS) GetWriteToken(ctx context.Context) (string, error) {
	var resp string
	err := g.request(ctx, "/reader/api/0/token", nil, &resp)
	return resp, err
}

type subscriptionList struct {
	Subscriptions []Subscriptions `json:"subscriptions"`
}

type Subscriptions struct {
	Categories []SubscriptionCategory `json:"categories"`
	ID         string                 `json:"id"`
	HTMLURL    string                 `json:"htmlUrl"`
	Title      string                 `json:"title"`
	URL        string                 `json:"url"`

	// IconURL    string                 `json:"iconUrl"`
}

type SubscriptionCategory struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

func (g FreshRSS) SubscriptionList(ctx context.Context) ([]Subscriptions, error) {
	params := url.Values{}
	params.Set("output", "json")

	var jsonResp subscriptionList
	err := g.request(ctx, "/reader/api/0/subscription/list", params, &jsonResp)
	return jsonResp.Subscriptions, err
}

type tagList struct {
	Tags []Tag `json:"tags"`
}

type Tag struct {
	ID   string `json:"id"`
	Type string `json:"type,omitempty"`
}

func (g FreshRSS) TagList(ctx context.Context) ([]Tag, error) {
	params := url.Values{}
	params.Set("output", "json")

	var resp tagList
	err := g.request(ctx, "/reader/api/0/tag/list", params, &resp)
	return resp.Tags, err
}

type ContentItem struct {
	ID            string
	Published     int64
	Title         string
	Author        string
	Canonical     []string
	Content       string
	Categories    []string
	TimestampUsec string
	Origin        struct {
		HTMLURL  string
		StreamID string
		Title    string
	}

	// CrawlTimeMsec string `json:"crawlTimeMsec"`
}

func (g FreshRSS) StreamContents(ctx context.Context, steamID, excludeTarget string, lastModified int64, n int) ([]ContentItem, error) {
	params := url.Values{}
	setOption(&params, "xt", excludeTarget)
	setOptionInt64(&params, "ot", lastModified)
	setOptionInt(&params, "n", n)
	params.Set("r", "n")

	var jsonResp string
	if err := g.request(ctx, "/reader/api/0/stream/contents/"+steamID, params, &jsonResp); err != nil {
		return nil, err
	}

	items := gjson.GetBytes([]byte(jsonResp), "items").Array()
	if len(items) == 0 {
		return []ContentItem{}, nil
	}

	res := make([]ContentItem, len(items))
	for i, item := range items {
		var ci ContentItem
		ci.ID = item.Get("id").String()
		ci.Title = item.Get("title").String()
		ci.Published = item.Get("published").Int()
		ci.Author = item.Get("author").String()
		ci.Content = item.Get("summary.content").String()
		ci.Origin.StreamID = item.Get("origin.streamId").String()
		ci.Origin.HTMLURL = item.Get("origin.htmlUrl").String()
		ci.Origin.Title = item.Get("origin.title").String()
		ci.TimestampUsec = item.Get("timestampUsec").String()

		for _, href := range item.Get("canonical.#.href").Array() {
			if h := href.String(); h != "" {
				ci.Canonical = append(ci.Canonical, h)
			}
		}
		for _, cat := range item.Get("categories").Array() {
			ci.Categories = append(ci.Categories, cat.String())
		}

		res[i] = ci
	}

	return res, nil
}

func (g FreshRSS) StreamIDs(ctx context.Context, includeTarget, excludeTarget string, n int) ([]string, error) {
	params := url.Values{}
	setOption(&params, "xt", excludeTarget)
	setOption(&params, "s", includeTarget)
	setOptionInt(&params, "n", n)
	params.Set("r", "n")

	var jsonResp string
	if err := g.request(ctx, "/reader/api/0/stream/items/ids", params, &jsonResp); err != nil {
		return nil, err
	}

	ids := gjson.Get(jsonResp, "itemRefs.#.id").Array()
	resp := make([]string, len(ids))
	for i, v := range ids {
		resp[i] = v.String()
	}

	return resp, nil
}

func (g FreshRSS) SetItemsState(ctx context.Context, token, itemID string, addAction, removeAction string) error {
	params := url.Values{}
	params.Set("T", token)
	params.Set("i", itemID)
	setOption(&params, "a", addAction)
	setOption(&params, "r", removeAction)

	err := g.postRequest(ctx, "/reader/api/0/edit-tag", params, nil)
	return err
}

type EditSubscription struct {
	// StreamID to operate on (required)
	// `feed/1` - the id
	// `feed/https:...` - or the url
	// it seems like 'feed' is required in the id
	StreamID string

	// Action can be one of those: subscribe OR unsubscribe OR edit
	Action string

	// Title, or for edit, or title for adding
	Title string

	// Add, StreamID to add the sub (generally a category)
	AddCategoryID string

	// Remove, StreamId to remove the subscription(s) from (generally a category)
	Remove string
}

func (g FreshRSS) SubscriptionEdit(ctx context.Context, token string, opts EditSubscription) (string, error) {
	// todo: action is required

	body := url.Values{}
	body.Set("T", token)
	body.Set("s", opts.StreamID)
	body.Set("ac", opts.Action)
	setOption(&body, "t", opts.Title)
	setOption(&body, "a", opts.AddCategoryID)
	setOption(&body, "r", opts.Remove)

	var resp string
	err := g.postRequest(ctx, "/reader/api/0/subscription/edit", body, &resp)
	return resp, err
}

func setOption(b *url.Values, k, v string) {
	if v != "" {
		b.Set(k, v)
	}
}

func setOptionInt(b *url.Values, k string, v int) {
	if v != 0 {
		b.Set(k, strconv.Itoa(v))
	}
}

func setOptionInt64(b *url.Values, k string, v int64) {
	if v != 0 {
		b.Set(k, strconv.FormatInt(v, 10))
	}
}

// request, makes GET request with params passed as url params
func (g *FreshRSS) request(ctx context.Context, endpoint string, params url.Values, resp any) error {
	u, err := url.Parse(g.host + endpoint)
	if err != nil {
		return err
	}

	setOptionInt64(&params, "ck", time.Now().UnixMilli())
	u.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	return g.handleResponse(req, resp)
}

// postRequest makes POST requests with parameters passed as form.
func (g *FreshRSS) postRequest(ctx context.Context, endpoint string, body url.Values, resp any) error {
	var reqBody io.Reader
	if body != nil {
		reqBody = strings.NewReader(body.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.host+endpoint, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return g.handleResponse(req, resp)
}

type apiResponse struct {
	Error string `json:"error,omitempty"`
}

func (g *FreshRSS) handleResponse(req *http.Request, out any) error {
	if g.authToken != "" {
		req.Header.Set("Authorization", "GoogleLogin auth="+g.authToken)
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return ErrUnauthorized
		}
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: status %d: %s", resp.StatusCode, string(body))
	}

	if strPtr, ok := out.(*string); ok {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
		*strPtr = string(body)

		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp, ok := out.(*apiResponse); ok && apiResp.Error != "" {
		return fmt.Errorf("%s", apiResp.Error)
	}

	return nil
}
