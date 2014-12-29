package util

import (
	"errors"
	"fmt"
	"github.com/conformal/btcnet"
	"github.com/conformal/btcscript"
)

var _ = fmt.Println

func ExtractAddrFromScript(script []byte) (string, error) {
	//scriptClass, addresses, reqSigs, err := btcscript.ExtractPkScriptAddrs(script, &btcnet.MainNetParams)
	_, addresses, _, err := btcscript.ExtractPkScriptAddrs(script, &btcnet.MainNetParams)
	if err != nil {
		//fmt.Println(err)
		return "", err
	}
	//fmt.Println("Script Class:", scriptClass)
	//fmt.Println("Addresses:", addresses)
	//fmt.Println("Encode:", addresses[0].EncodeAddress())
	//fmt.Println("Required Signatures:", reqSigs)
	if len(addresses) == 0 {
		return "", errors.New("No address extracted")
	} else if len(addresses) > 1 {
		return "", errors.New("More than on address")
	}
	return addresses[0].EncodeAddress(), nil
}
