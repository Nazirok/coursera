package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func main() {
	begin := func(in, out chan interface{}) {
		out <- 0
	}

	end := func(in, out chan interface{}) {
		data := <-in
		fmt.Println(data)
	}
	ExecutePipeline(begin, SingleHash, MultiHash, CombineResults, end)
}

func ExecutePipeline(jobs ...job) {
	in := make(chan interface{}, MaxInputDataLen)
	out := make(chan interface{}, MaxInputDataLen)
	wg := &sync.WaitGroup{}
	for _, job := range jobs {
		wg.Add(1)
		go func(wg *sync.WaitGroup, job func(in, out chan interface{}), in, out chan interface{}) {
			defer func() {
				wg.Done()
				close(out)
			}()
			job(in, out)
		}(wg, job, in, out)
		in = out
		out = make(chan interface{}, MaxInputDataLen)
	}
	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}
	for v := range in {
		wg.Add(1)
		go func(v interface{}, out chan interface{}, wg *sync.WaitGroup) {
			defer wg.Done()
			var data string
			switch v.(type) {
			case int:
				d, _ := v.(int)
				data = strconv.Itoa(d)
			case uint32:
				d, _ := v.(uint32)
				data = fmt.Sprint(d)
			}

			md5Crc32Chan := make(chan string)
			crc32Chan := make(chan string)
			go func(out chan string, data string) {
				mu.Lock()
				t := DataSignerMd5(data)
				mu.Unlock()
				out <- DataSignerCrc32(t)
			}(md5Crc32Chan, data)
			go func(out chan string, data string) {
				out <- DataSignerCrc32(data)
			}(crc32Chan, data)

			md5Crc32 := <-md5Crc32Chan
			crc32 := <-crc32Chan
			out <- crc32 + "~" + md5Crc32
		}(v, out, wg)

	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	wgg := &sync.WaitGroup{}
	for v := range in {
		wgg.Add(1)
		go func(v interface{}, out chan interface{}, wgg *sync.WaitGroup) {
			defer wgg.Done()
			data, _ := v.(string)
			var toOut string
			acc := make(map[int]string)
			wg := &sync.WaitGroup{}
			mu := &sync.Mutex{}

			for i := 0; i <= 5; i++ {
				wg.Add(1)
				go func(wg *sync.WaitGroup, i int, data string) {
					defer wg.Done()
					d := strconv.Itoa(i) + data
					crc := DataSignerCrc32(d)
					mu.Lock()
					acc[i] = crc
					mu.Unlock()
				}(wg, i, data)
			}
			wg.Wait()
			for i := 0; i <= 5; i++ {
				toOut += acc[i]
			}

			out <- toOut
		}(v, out, wgg)

	}
	wgg.Wait()
}

func CombineResults(in, out chan interface{}) {
	result := make([]string, 0)

	for v := range in {
		result = append(result, v.(string))
	}
	sort.Strings(result)
	out <- strings.Join(result, "_")
}
