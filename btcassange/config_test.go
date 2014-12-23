package btcassange

import (
	"fmt"
	"testing"
)

func TestLoadConfiguration_1(t *testing.T) {
	config, err := LoadConfiguration("./config.json.example")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(config.Db_host)
	if config.Db_host != "127.0.0.1" {
		t.Error("db_host error.")
	}
	if config.Db_user != "test_user" {
		t.Error("db_user error.")
	}
	if config.Db_password != "test_password" {
		t.Error("db_password error.")
	}
	if config.Db_database != "test_database" {
		t.Error("db_database error.")
	}
}
