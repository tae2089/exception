package main

import (
	"errors"
	"fmt"

	"github.com/tae2089/exception"
)

func main() {
	_, err := converter(0)
	if customErr, ok := err.(*exception.CustomError); ok {
		fmt.Println(customErr.PrintTrace())
	}
}

func converter(v int) (int, error) {
	_, err := checker(v)
	if err != nil {
		return 0, exception.WrapMessage(err, "converter - v is 0")
	}
	return 0, nil
}

func checker(v int) (int, error) {
	if v == 0 {
		return 0, exception.WrapMessage(errors.New("v is 0"), "v is 0")
	}
	return v, nil
}
