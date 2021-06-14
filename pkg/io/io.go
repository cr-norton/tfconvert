package io

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func ReadFile(filename string, output interface{}) error {
	path := fmt.Sprintf("%s/src/github.com/cr-norton/tfconvert/%s", os.Getenv("GOPATH"), filename)
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, &output)
}

func WriteFile(filename string, content interface{}) error {
	bytes, err := json.Marshal(content)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, bytes, os.ModePerm)
}
