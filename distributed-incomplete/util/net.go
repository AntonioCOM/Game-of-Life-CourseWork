package util

import (
	"io/ioutil"
	"net/http"
)

// GetPublicIp fetches our public IP and returns it as a string
func GetPublicIP() string {
	url := "https://api.ipify.org?format=text"
	println("Getting IP address from  ipify")
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return string(ip)
}
