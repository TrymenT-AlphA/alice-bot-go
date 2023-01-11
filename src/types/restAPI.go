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

type RestAPI struct {
	UrlFormat string                 `json:"urlFormat"`
	UrlParams []interface{}          `json:"urlParams"`
	Method    string                 `json:"method"`
	Verify    bool                   `json:"verify"`
	Params    map[string]interface{} `json:"params"`
	Comment   string                 `json:"comment"`
}

func NewRestAPI(platform string, field string, section string) (*RestAPI, error) {
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

	api := &RestAPI{
		UrlFormat: res.Get("urlFormat").String(),
		UrlParams: nil,
		Method:    res.Get("method").String(),
		Verify:    res.Get("verify").Bool(),
		Params:    make(map[string]interface{}),
		Comment:   res.Get("Comment").String(),
	}

	return api, nil
}

func (api *RestAPI) getGETRequest() (*http.Request, error) {
	params := url.Values{}

	for key, val := range api.Params {
		val, err := util.GetString(val)
		if err != nil {
			return nil, err
		}
		params.Set(key, val)
	}

	Url, err := url.Parse(fmt.Sprintf(api.UrlFormat, api.UrlParams...))
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

func (api *RestAPI) getPOSTRequest() (*http.Request, error) {
	data, err := json.Marshal(api.Params)
	if err != nil {
		return nil, err
	}

	payload := bytes.NewReader(data)

	request, err := http.NewRequest(api.Method, fmt.Sprintf(api.UrlFormat, api.UrlParams...), payload)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-type", "application/json")

	return request, nil
}

func (api *RestAPI) DoRequest(client *http.Client) ([]byte, error) {
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

func (api *RestAPI) DoRequestAuth(client *http.Client, auth string) ([]byte, error) {
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
