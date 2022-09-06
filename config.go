package main

import (
	"os"
)

var mongoPath = (func() string {
	value, exist := os.LookupEnv("MONGODB")
	if !exist {
		value = "mongodb://localhost:27017"
	}
	return value
})()
