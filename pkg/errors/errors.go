package errors

import "fmt"

func CheckPanic(err error) {
	if err != nil {
		panic(err.Error())
	}
}
func CheckError(err error) {
	if err != nil {
		fmt.Printf("ERROR: %v\n", err.Error())
	}
}
