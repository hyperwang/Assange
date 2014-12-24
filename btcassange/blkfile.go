package btcassange

import (
	"fmt"
	"io/ioutil"
	"path"
)

func GetBlkFileList(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	var list []string
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	for _, f := range files {
		match, _ := path.Match("blk?????.dat", f.Name())
		if match {
			//fmt.Println(f.Name())
			list = append(list, f.Name())
		}
	}
	//fmt.Println(list)
	return list, nil
}
