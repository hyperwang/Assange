package Assange

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

var in = `{   
	"db_host" : "127.0.0.1",
	"db_user" : "test_user",
	"db_password" : "test_password"}`

type Configuration struct {
	Db_host     string
	Db_user     string
	Db_password string
}

func LoadConfiguration(fname string) (Configuration, error) {
	content, err := ioutil.ReadFile(fname)
	fmt.Println(string(content))
	if err != nil {
		fmt.Print("Error:", err)
	}
	var conf Configuration
	err = json.Unmarshal([]byte(content), &conf)
	if err != nil {
		fmt.Print("Error:", err)
	}
	return conf, nil
}

var Config Configuration

func init() {
	Config, _ = LoadConfiguration("config.json")
}
