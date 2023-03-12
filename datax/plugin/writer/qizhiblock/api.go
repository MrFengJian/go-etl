package qizhiblock

import (
	"encoding/json"
	"fmt"
	"github.com/Breeze0806/go-etl/element"
	"github.com/pingcap/errors"
	"io/ioutil"
	"net/http"
	"strings"
)

type apiResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	Token     string `json:"token,omitempty"`
	ErrorCode int    `json:"error_code,omitempty"`
}

type apiClient struct {
	client *http.Client
	conf   *Config
	token  string
}

func NewApiClient(config *Config) (*apiClient, error) {
	client := &http.Client{}
	body := map[string]string{
		"username": config.Username,
		"password": config.Password,
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/login", config.RestApiEndpoint),
		wrapDataAsBody(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	responseBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode > 300 {
		return nil, errors.New(string(responseBytes))
	}
	res := &apiResponse{}
	if err = json.Unmarshal(responseBytes, res); err != nil {
		return nil, err
	}
	if !res.Success {
		return nil, errors.New(res.Message)
	}
	return &apiClient{
		token:  res.Token,
		client: client,
		conf:   config,
	}, nil
}

func wrapDataAsBody(data interface{}) *strings.Reader {
	jsonBytes, _ := json.Marshal(data)
	return strings.NewReader(string(jsonBytes))
}

func (c *apiClient) writeRecord(keyColumnIndex int, record element.Record) error {
	keyColumn, err := record.GetByIndex(keyColumnIndex)
	if err != nil {
		return err
	}
	key, err := keyColumn.AsString()
	if err != nil {
		return err
	}
	value := make(map[string]interface{})
	for i := 0; i < record.ColumnNumber(); i++ {
		col, err := record.GetByIndex(i)
		if err != nil {
			return err
		}
		if !col.IsNil() {
			v, err := col.AsString()
			if err != nil {
				return err
			}
			value[col.Name()] = v
		} else {
			value[col.Name()] = ""
		}

	}
	bs, err := json.Marshal(value)
	if err != nil {
		return err
	}
	body := map[string]interface{}{
		"fcn":  c.conf.ChainCodeFunction,
		"args": []interface{}{key, string(bs)},
	}
	url := fmt.Sprintf("%s/channels/%s/chaincodes/%s", c.conf.RestApiEndpoint, c.conf.Channel, c.conf.ChainCode)
	req, err := http.NewRequest("POST", url,
		wrapDataAsBody(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	response, err := c.client.Do(req)
	if err != nil {
		return err
	}
	responseBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if response.StatusCode > 300 {
		return errors.New(string(responseBytes))
	}
	fmt.Println(string(responseBytes))
	return nil
}
