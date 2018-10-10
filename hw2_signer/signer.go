package main

import (
	"fmt"
	"strconv"
	"sync"
)

func main() {
	f := func(in, out chan interface{}) {
		out <- "0"
	}
	ExecutePipeline(f, SingleHash, MultiHash, CombineResults)
	//time.Sleep(10 * time.Second)
}

func ExecutePipeline(jobs ...func(in, out chan interface{})) {
	in := make(chan interface{})
	out := make(chan interface{})
	wg := &sync.WaitGroup{}
	for _, job := range jobs {
		wg.Add(1)
		go func(wg *sync.WaitGroup, in, out chan interface{}) {
			defer wg.Done()
			job(in, out)
		}(wg, in, out)
		in = out
		out = make(chan interface{})
	}
	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	fmt.Println("SingleHash")
	for v := range in {
		data, ok := v.(string)

		if !ok {
			continue
		}

		md5Crc32 := DataSignerCrc32(DataSignerMd5(data))
		crc32 := DataSignerCrc32(data)
		out <- crc32 + "~" + md5Crc32

	}
	//close(out)
}

func MultiHash(in, out chan interface{}) {
	fmt.Println("MultiHash")

	for v := range in {
		data, ok := v.(string)

		if !ok {
			continue
		}

		var toOut string
		for i := 0; i <= 5; i++ {
			d := strconv.Itoa(i) + data
			toOut += DataSignerCrc32(d)

		}
		out <- toOut
	}
	//close(out)
}

func CombineResults(in, out chan interface{}) {
	fmt.Println("Combine")

	for v := range in {
		fmt.Println(v)
		out <- v
	}
	//close(out)
}
