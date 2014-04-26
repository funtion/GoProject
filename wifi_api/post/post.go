package main

import (
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
)

func main() {
	response, _ := http.Get("http://www.baidu.com")
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	result := string(body)
	//fmt.Println(result)
}
