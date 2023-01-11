package types

import (
	"alice-bot-go/src/util"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

type API struct {
	Url     string                 `json:"url"`
	Method  string                 `json:"method"`
	Verify  bool                   `json:"verify"`
	Params  map[string]interface{} `json:"params"`
	Comment string                 `json:"comment"`
}

func NewAPI(platform string, field string, section string) (*API, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(filepath.Join(
		cwd,
		"..",
		"src",
		"api",
		platform,
		fmt.Sprintf("%s.json", field),
	))
	if err != nil {
		return nil, err
	}

	res := gjson.GetBytes(data, section)

	api := &API{
		Url:     res.Get("url").String(),
		Method:  res.Get("method").String(),
		Verify:  res.Get("verify").Bool(),
		Params:  make(map[string]interface{}),
		Comment: res.Get("Comment").String(),
	}

	return api, nil
}

func (api *API) getGETRequest() (*http.Request, error) {
	params := url.Values{}

	for key, val := range api.Params {
		val, err := util.GetString(val)
		if err != nil {
			return nil, err
		}
		params.Set(key, val)
	}

	Url, err := url.Parse(api.Url)
	if err != nil {
		return nil, err
	}

	Url.RawQuery = params.Encode()

	request, err := http.NewRequest(api.Method, Url.String(), nil)
	if err != nil {
		return nil, err
	}

	return request, nil
}

func (api *API) getPOSTRequest() (*http.Request, error) {
	data, err := json.Marshal(api.Params)
	if err != nil {
		return nil, err
	}

	payload := bytes.NewReader(data)

	request, err := http.NewRequest(api.Method, api.Url, payload)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-type", "application/json")

	return request, nil
}

func (api *API) DoRequest(client *http.Client) ([]byte, error) {
	request := &http.Request{}

	var err error
	if api.Method == "GET" {
		request, err = api.getGETRequest()
		if err != nil {
			return nil, err
		}
	} else if api.Method == "POST" {
		request, err = api.getPOSTRequest()
		if err != nil {
			return nil, err
		}
	}

	request.Header.Set(
		"User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 Edg/108.0.1462.54",
	)

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

func (api *API) DoRequestAuth(client *http.Client, auth string) ([]byte, error) {
	request := &http.Request{}

	var err error
	if api.Method == "GET" {
		request, err = api.getGETRequest()
		if err != nil {
			return nil, err
		}
	} else if api.Method == "POST" {
		request, err = api.getPOSTRequest()
		if err != nil {
			return nil, err
		}
	}

	request.Header.Set(
		"User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 Edg/108.0.1462.54",
	)

	request.Header.Add("Authorization", auth)

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
