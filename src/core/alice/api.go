package alice

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/tidwall/gjson"

	"alice-bot-go/src/core/config"
	"alice-bot-go/src/core/util"
)

type API struct {
	Url       string                 `json:"url,omitempty"`
	UrlFormat string                 `json:"url_format,omitempty"`
	UrlParams []interface{}          `json:"url_params,omitempty"`
	Method    string                 `json:"method,omitempty"`
	Params    map[string]interface{} `json:"params,omitempty"`
	Header    map[string]string      `json:"header,omitempty"`
	Body      []byte                 `json:"body,omitempty"`
	Commet    string                 `json:"commet,omitempty"`
}

func NewAPI(platform string, field string, section string) (*API, error) {
	data, err := os.ReadFile(filepath.Join(
		config.Global.Cwd, "..", "src", "core", "api",
		platform, fmt.Sprintf("%s.json", field),
	))
	if err != nil {
		return nil, err
	}
	res := gjson.GetBytes(data, section)
	api := &API{
		Url:       res.Get("url").String(),
		UrlFormat: res.Get("url_format").String(),
		Method:    res.Get("method").String(),
	}
	return api, nil
}

func (api *API) DoRequest(client *http.Client) ([]byte, error) {
	if client == nil {
		client = &http.Client{}
	}
	if api.Url == "" {
		api.Url = fmt.Sprintf(api.UrlFormat, api.UrlParams...)
	}
	params := url.Values{}
	for key, val := range api.Params {
		params.Set(key, util.Strval(val))
	}
	Url, err := url.Parse(api.Url)
	if err != nil {
		return nil, err
	}
	Url.RawQuery = params.Encode()
	request, err := http.NewRequest(api.Method, Url.String(), bytes.NewReader(api.Body))
	if err != nil {
		return nil, err
	}
	if _, ok := api.Header["User-Agent"]; !ok {
		request.Header.Set("User-Agent", config.Global.UserAgent)
	}
	for key, val := range api.Header {
		request.Header.Set(key, val)
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	err = response.Body.Close()
	if err != nil {
		return nil, err
	}
	return data, nil
}
