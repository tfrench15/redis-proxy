package main

import (
	"fmt"
	"strconv"
	"time"
)

func convertExpiryToDuration(dur string) time.Duration {
	i, err := strconv.Atoi(dur)
	if err != nil {
		fmt.Println(err)
		return 5 * time.Second // default to 5 second expiry
	}
	return time.Duration(i)
}

func convertCapacityToInt(cap string) int {
	i, err := strconv.Atoi(cap)
	if err != nil {
		fmt.Println(err)
		return 10 // default to capacity of 10
	}
	return i
}
