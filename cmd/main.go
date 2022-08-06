package main

import (
	"io/ioutil"
)

func main() {
	f, err := ioutil.ReadFile("/README.md")
	if err != nil {
		panic(err)
	}
	println(string(f))
}
