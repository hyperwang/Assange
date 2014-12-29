package explorer

import (
	. "Assange/config"
	"testing"
)

var config = Configuration{
	Db_host:           "127.0.0.1",
	Db_database:       "assange",
	Explorer_user:     "assange_explorer",
	Explorer_password: "assange_explorer_haobtc",
}

func TestInitExplorerServer01(t *testing.T) {
	InitExplorerServer(config)
}
