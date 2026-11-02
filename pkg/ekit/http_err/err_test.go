package httperr

import (
	"fmt"
	"testing"
)

//	func TestErrIsZero(t *testing.T) {
//		abc := ConsumeData[int]{}
//		ierrTest(abc.Err)
//	}
func ierrTest(ierr IError) {
	if ierr.IsZero() {
		fmt.Println("ok")
	}
}

func TestAbc(t *testing.T) {
	abc := make([]int, 0, 10)
	_ = append(abc, 1, 2, 3)
	fmt.Println(abc)
}
