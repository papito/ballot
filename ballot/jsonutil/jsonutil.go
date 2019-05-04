package jsonutil

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

func GetRequestJson(r *http.Request) (map[string]interface{}, error) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	jsonData, err := GetJsonFromString(string(reqBody))
	if err != nil {
		log.Print(err)
		return nil, err
	}

	return jsonData, nil
}

func GetJsonFromString(data string) (map[string]interface{}, error) {
	var f interface{}
	err := json.Unmarshal([]byte(data), &f)

	if err != nil {
		return nil, err
	}

	return f.(map[string]interface{}), nil
}

