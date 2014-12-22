package Assange

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Configuration struct {
	Db_host     string
	Db_user     string
	Db_password string
	Db_database string
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
