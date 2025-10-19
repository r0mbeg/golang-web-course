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
	var wg sync.WaitGroup

	for _, j := range jobs {
		out := make(chan interface{})
		wg.Add(1)
		go func(j job, in, out chan interface{}) {
			defer wg.Done()
			j(in, out)
			close(out)
		}(j, in, out)
		in = out
	}
	wg.Wait()
}

var (
	md5Mu = sync.Mutex{}
)

func SingleHash(in, out chan interface{}) {

	var stageWG sync.WaitGroup

	for v := range in {
		stageWG.Add(1)

		data := fmt.Sprint(v)

		go func(data string) {
			defer stageWG.Done()

			wg := sync.WaitGroup{}
			wg.Add(2)

			var left, right string

			go func(left *string) {
				defer wg.Done()
				*left = DataSignerCrc32(data)
			}(&left)

			go func(right *string) {
				defer wg.Done()
				md5Mu.Lock()
				rightTemp := DataSignerMd5(data)
				md5Mu.Unlock()
				*right = DataSignerCrc32(rightTemp)
			}(&right)

			wg.Wait()

			res := fmt.Sprintf("%s~%s", left, right)

			out <- res

		}(data)

	}
	stageWG.Wait()
}

func MultiHash(in, out chan interface{}) {

	var stageWG sync.WaitGroup

	for v := range in {
		data := fmt.Sprint(v)

		stageWG.Add(1)

		go func(data string) {
			defer stageWG.Done()

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

			res := strings.Join(resSlice, "")

			out <- res

		}(data)
	}
	stageWG.Wait()
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
