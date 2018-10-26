package kubernetes_infra

import (
	"io/ioutil"
	"log"
)

func GetNameSpace() string {
	nsBytes, err := ioutil.ReadFile(NameSpaceFile)
	if err != nil {
		log.Fatal("Failed to get NameSpace", err)
	}
	return string(nsBytes)
}
