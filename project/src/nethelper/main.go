// nethelper project main.go
package main

/*
#include <stdio.h>
#include <stdlib.h>

void test_c(void)
{
	printf("go call c\n");
}
*/
import "C"

import (
	"fmt"
	"time"
)

func randomDept() string {
	ch := make(chan string, 1)
	select {
	case ch <- "内科":
	case ch <- "外科":
	case ch <- "妇科":
	case ch <- "儿科":
	case ch <- "急诊":
	}
	str := <-ch
	fmt.Println("P", str)
	return str
}

var ch chan string

func p() {
	for {
		ch <- randomDept()
	}
}
func c() {
	for {
		fmt.Println("D", <-ch)
		time.Sleep(200 * time.Millisecond)
	}
}

func main() {
	C.test_c()
	fmt.Println("okkkk")
	//ch = make(chan string, 10)

	//go p()
	//go c()

	//select {}

}
