package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func Get[T any](url string, headers map[string]string) (T, error) {
	req, err := getRequest(http.MethodGet, "application/json", url, nil, headers)
	if err != nil {
		return *new(T), err
	}
	result, err := getResponse[T](req)
	if err != nil {
		return *new(T), err
	}
	return result, nil
}

func Post[T any](url string, body interface{}, headers map[string]string) (T, error) {
	req, err := getRequest(http.MethodPost, "application/json", url, body, headers)
	if err != nil {
		return *new(T), err
	}
	result, err := getResponse[T](req)
	if err != nil {
		return *new(T), err
	}
	return result, nil
}

func Put[T any](url string, body interface{}, headers map[string]string) (T, error) {
	req, err := getRequest(http.MethodPut, "application/json", url, body, headers)
	if err != nil {
		return *new(T), err
	}
	result, err := getResponse[T](req)
	if err != nil {
		return *new(T), err
	}
	return result, nil
}

func Patch[T any](url string, body interface{}, headers map[string]string) (T, error) {
	req, err := getRequest(http.MethodPatch, "application/json-patch+json", url, body, headers)
	if err != nil {
		return *new(T), err
	}
	result, err := getResponse[T](req)
	if err != nil {
		return *new(T), err
	}
	return result, nil
}

func getRequest(method string, contentType string, url string, body interface{}, headers map[string]string) (*http.Request, error) {
	var bodyReader bytes.Reader = bytes.Reader{}
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = *bytes.NewReader(jsonBody)
	}
	req, err := http.NewRequest(method, url, &bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return req, nil
}

func getResponse[T any](req *http.Request) (T, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	var response T
	if err != nil {
		fmt.Println(err)
		return *new(T), err
	} else if resp.StatusCode > 299 {
		fmt.Println(resp.StatusCode)
		err = fmt.Errorf("Error: %v", resp.StatusCode)
		return *new(T), err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return *new(T), err
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println(err)
		return *new(T), err
	}

	return response, nil
}
