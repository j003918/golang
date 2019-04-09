// test project main.go
package main

import (
	"encoding/json"
	"fmt"
)

type Student struct {
	Name    string   `json:"name" xml:"name"`
	Age     int      `json:"age" xml:"age"`
	Guake   bool     `json:"guake" xml:"guake"`
	Classes []string `json:"classes" xml:"classes"`
	Price   float32  `json:"price" xml:"price"`
}

func main() {

	array := []int{10, 11, 12, 13, 14}
	slice := array[0:4]                                       // slice是对array的引用
	fmt.Println("array: ", array)                             // array:  [20 21 12 13 14]
	fmt.Println("slice: cap=", cap(slice), ", value=", slice) // slice: cap= 5 , value= [10 11 12 13]

	slice = append(slice, 1)
	slice[0] = 234

	fmt.Println("slice: cap=", cap(slice), ", value=", slice, array)

	st := &Student{
		"Xiao Ming",
		16,
		true,
		[]string{"Math", "English", "Chinese"},
		9.99,
	}

	st.Classes = append(st.Classes, "n1", "n2")

	b, _ := json.Marshal(st)
	fmt.Println(string(b))
}
