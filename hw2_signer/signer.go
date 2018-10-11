package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	begin := func(in, out chan interface{}) {
		out <- "0"
		out <- "1"
		close(out)
		fmt.Println("Close first")

	}

	end := func(in, out chan interface{}) {
		data := <-in
		close(out)
		fmt.Println(data)

	}

	ExecutePipeline(begin, SingleHash, MultiHash, CombineResults, end)

}

func ExecutePipeline(jobs ...job) {
	in := make(chan interface{},10)
	out := make(chan interface{}, 10)
	wg := &sync.WaitGroup{}
	for _, job := range jobs {
		wg.Add(1)
		go func(wg *sync.WaitGroup, job func(in, out chan interface{}), in, out chan interface{}) {
			defer wg.Done()
			job(in, out)
		}(wg, job, in, out)
		in = out
		out = make(chan interface{},10)
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

		md5Crc32Chan := make(chan string)
		crc32Chan := make(chan string)
		go func(out chan string, data string) {

			out <- DataSignerCrc32(DataSignerMd5(data))
		}(md5Crc32Chan, data)
		go func(out chan string, data string) {
			out <- DataSignerCrc32(data)
		}(crc32Chan, data)

		md5Crc32 := <- md5Crc32Chan
		crc32 := <- crc32Chan
		out <- crc32 + "~" + md5Crc32
	}
	close(out)
	fmt.Println("Close Single")
}

func MultiHash(in, out chan interface{}) {
	fmt.Println("MultiHash")

	for v := range in {
		data, ok := v.(string)

		if !ok {
			continue
		}

		var toOut string
		acc := make(map[int]string)
		wg := &sync.WaitGroup{}
		mu := &sync.Mutex{}
		a:= time.Now()

		for i := 0; i <= 5; i++ {
			wg.Add(1)
			go func(wg *sync.WaitGroup, i int, data string) {
				defer wg.Done()
				d := strconv.Itoa(i) + data
				mu.Lock()
				acc[i] = DataSignerCrc32(d)
				mu.Unlock()
			}(wg, i, data)
		}
		wg.Wait()
		fmt.Println(time.Since(a))
		for i := 0; i <= 5; i++ {
			toOut += acc[i]
		}

		out <- toOut
	}
	close(out)
	fmt.Println("Close Multi")

}

func CombineResults(in, out chan interface{}) {
	fmt.Println("Combine")
	result := make([]string, 0)

	for v := range in {
		result = append(result, v.(string))
	}
	sort.Strings(result)
	out <- strings.Join(result, "_")
	close(out)
	fmt.Println("Close Combine")

}
