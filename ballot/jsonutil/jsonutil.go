package jsonutil

import (
	"encoding/json"
	"github.com/joomcode/errorx"
	"io/ioutil"
	"net/http"
)

func GetRequestJson(r *http.Request) (map[string]interface{}, error) {
	reqBody, err := GetRequestBody(r)
	if err != nil {return nil, errorx.EnsureStackTrace(err)}

	jsonData, err := GetJsonFromString(string(reqBody))
	if err != nil {return nil, errorx.EnsureStackTrace(err)}

	return jsonData, nil
}

func GetRequestBody(r *http.Request) (string, error)  {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {return "", errorx.EnsureStackTrace(err)}

	return string(reqBody), nil
}

func GetJsonFromString(data string) (map[string]interface{}, error) {
	var f interface{}
	err := json.Unmarshal([]byte(data), &f)

	if err != nil {return nil, errorx.EnsureStackTrace(err)}

	return f.(map[string]interface{}), nil
}

