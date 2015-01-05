package util

import (
	//"errors"
	. "Assange/logging"
	"fmt"
	"github.com/conformal/btcnet"
	"github.com/conformal/btcscript"
)

var _ = fmt.Println
var log = GetLogger("Util", DEBUG)

func ExtractAddrFromScript(script []byte) string {
	//scriptClass, addresses, reqSigs, err := btcscript.ExtractPkScriptAddrs(script, &btcnet.MainNetParams)
	_, addresses, _, err := btcscript.ExtractPkScriptAddrs(script, &btcnet.MainNetParams)
	if err != nil {
		//fmt.Println(err)
		log.Error(err.Error())
		return ""
	}
	//fmt.Println("Script Class:", scriptClass)
	//fmt.Println("Addresses:", addresses)
	//fmt.Println("Encode:", addresses[0].EncodeAddress())
	//fmt.Println("Required Signatures:", reqSigs)
	if len(addresses) == 0 {
		log.Error("No address extracted")
		return ""
	} else if len(addresses) > 1 {
		log.Error("More than one address extracted")
		return addresses[len(addresses)-1].EncodeAddress()
	}
	return addresses[0].EncodeAddress()
}
