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
func TestExtractAddrFromScript03(t *testing.T) {
	scriptHex := "410496b538e853519c726a2c91e61ec11600ae1390813a627c66fb8be7947be63c52da7589379515d4e0a604f8141781e62294721166bf621e73a82cbf2342c858eeac"
	script, err := hex.DecodeString(scriptHex)
	if err != nil {
		t.Error("scriptHex can not be decoded.")
	}
	_, err = ExtractAddrFromScript(script)
	if err == nil {
		t.Error("Return")
	}
}
