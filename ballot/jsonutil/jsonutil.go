/*
 * The MIT License
 *
 * Copyright (c) 2019,  Andrei Taranchenko
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

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

