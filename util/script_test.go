package util

import (
	"encoding/hex"
	"testing"
)

func TestExtractAddrFromScript01(t *testing.T) {
	scriptHex := "76a914128004ff2fcaf13b2b91eb654b1dc2b674f7ec6188ac"
	script, err := hex.DecodeString(scriptHex)
	if err != nil {
		t.Error("scriptHex can not be decoded.")
	}
	_, err = ExtractAddrFromScript(script)
	if err == nil {
		t.Error("Return")
	}
}

func TestExtractAddrFromScript02(t *testing.T) {
	scriptHex := "a914e895dfe6ca7affb5f066685ac93eaa138297e13887"
	script, err := hex.DecodeString(scriptHex)
	if err != nil {
		t.Error("scriptHex can not be decoded.")
	}
	_, err = ExtractAddrFromScript(script)
	if err == nil {
		t.Error("Return")
	}
}
