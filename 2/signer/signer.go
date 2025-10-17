package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func ExecutePipeline(jobs ...job) {
	in := make(chan interface{})
	for _, j := range jobs {
		out := make(chan interface{})
		go func(j job, in, out chan interface{}) {
			j(in, out)
			close(out)
		}(j, in, out)
		in = out
	}
}

var (
	md5Mu = sync.Mutex{}
)

func SingleHash(in, out chan interface{}) {
	for v := range in {
		data := fmt.Sprint(v)

		d := data

		wg := sync.WaitGroup{}
		wg.Add(2)

		var left, right string

		go func(left *string) {
			defer wg.Done()
			*left = DataSignerCrc32(d)
		}(&left)

		go func(right *string) {
			defer wg.Done()
			md5Mu.Lock()
			rightTemp := DataSignerMd5(d)
			md5Mu.Unlock()
			*right = DataSignerCrc32(rightTemp)
		}(&right)

		wg.Wait()

		res := fmt.Sprintf("%s~%s", left, right)

		out <- res
	}
}

func MultiHash(in, out chan interface{}) {
	for v := range in {
		data := fmt.Sprint(v)

		//d := data

		resSlice := make([]string, 6)
		wg := sync.WaitGroup{}
		wg.Add(6)
		for i := 0; i < 6; i++ {
			it := i

			go func(data string) {
				defer wg.Done()
				resSlice[it] = DataSignerCrc32(strconv.Itoa(it) + data)
			}(data)

		}

		wg.Wait()

		var res string

		for i := 0; i < len(resSlice); i++ {
			res += resSlice[i]
		}

		out <- res
	}
}

func CombineResults(in, out chan interface{}) {

	slice := make([]string, 0, 100)

	for v := range in {
		slice = append(slice, fmt.Sprint(v))
	}

	sort.Strings(slice)

	res := strings.Join(slice, "_")

	out <- res
}
