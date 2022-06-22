package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func GetAddr(address string) (res *http.Response) {
	url := "http://" + address + ":8545"
	data := `{ "jsonrpc":"2.0", "method":"eth_blockNumber", "params":[], "id":67 }`
	resp, err := http.Post(url, "application/json", strings.NewReader(data))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	return resp
}

func RpcRequest(address string) (string, error) {
	addr := GetAddr(address)
	b, err := ioutil.ReadAll(addr.Body)
	if err != nil {
		log.Fatal(err)
	}
	var str Student
	_ = json.Unmarshal(b, &str)
	return str.Result, nil
}

func main2() {

}
