package util

import (
	"fmt"
	"github.com/conformal/btcnet"
	"github.com/conformal/btcscript"
)

var _ = fmt.Println

func ExtractAddrFromScript(script []byte) (string, error) {
	_, addresses, _, err := btcscript.ExtractPkScriptAddrs(script, &btcnet.MainNetParams)
	if err != nil {
		//fmt.Println(err)
		return "", err
	}
	//fmt.Println("Script Class:", scriptClass)
	//fmt.Println("Addresses:", addresses)
	//fmt.Println("Required Signatures:", reqSigs)
	return addresses[0].String(), nil
}
