package main

import "fmt"

func demo()  {

}

type a interface {}

func main()  {
	mapDemo := make(map[interface{}]interface{})

	mapDemo[demo] = "b"

	fmt.Println(mapDemo)
}
