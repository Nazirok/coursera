package main

import (
	"fmt"
	"strconv"
)

func main() {
	DataSignerCrc32("sdf")
	first := (SingleHash(strconv.Itoa(0)))
	fmt.Println(first)
	MultiHash(first)
}

func ExecutePipeline(job...func(in, out chan interface{})) {

}

func SingleHash(in, out chan interface{}) string {
	md5Crc32 := DataSignerCrc32(DataSignerMd5(data))
	crc32 := DataSignerCrc32(data)
	return crc32 + "~" + md5Crc32
}

func MultiHash(in, out chan interface{}) string {
	for i:=0; i<=5; i++ {
		d := strconv.Itoa(i) + data
		fmt.Println(i, DataSignerCrc32(d))
	}
	return ""
}