package model

import (
	"bot-go/src/util"
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

func (api *API) GetAPI(field string, section string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(filepath.Join(cwd, "..", "data", "api", "bilibili", fmt.Sprintf("%s.json", field)))
	if err != nil {
		return err
	}

	res := gjson.GetBytes(data, section)
	api.Url = res.Get("url").String()
	api.Method = res.Get("method").String()
	api.Verify = res.Get("verify").Bool()
	api.Comment = res.Get("Comment").String()

	return nil
}

func (api *API) Request() ([]byte, error) {
	client := &http.Client{}

	request := &http.Request{}

	if api.Method == "GET" {
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

		request, err = http.NewRequest(api.Method, Url.String(), nil)
		if err != nil {
			return nil, err
		}
	} else if api.Method == "POST" {
		data, err := json.Marshal(api.Params)
		if err != nil {
			return nil, err
		}

		payload := bytes.NewReader(data)

		request, err = http.NewRequest(api.Method, api.Url, payload)
		if err != nil {
			return nil, err
		}

		request.Header.Add("Content-type", "application/json")
	}

	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 Edg/108.0.1462.54")

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if err := response.Body.Close(); err != nil {
		return nil, err
	}

	return data, nil
}
