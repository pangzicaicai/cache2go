package main

import "fmt"

type demo1 struct {
	a int
}

func (d demo1) Count() int {
	return 1
}

type demo2 struct {
	demo1
}

type demo3 struct {
	a int
}

func (d demo3) Count() int {
	return 2
}

type demo4 struct {
	demo1
	demo3
}

//
//func (d demo2) Count() int {
//	a := d.demo1.Count()
//	fmt.Println(a)
//	return 2
//}

func main()  {
	//var demo2 demo2
	//fmt.Println(demo2.demo1.Count())
	//fmt.Println(demo2.Count())

	var demo4 demo4
	fmt.Println(demo4.demo3.Count())

}