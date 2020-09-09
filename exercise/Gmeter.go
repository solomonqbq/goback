package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
)

func main() {
	fmt.Println("命令行参数数量:", len(os.Args))
	flag.Parse()
	url := flag.Arg(0)
	if url == "" {
		url = "http://www.jd.com"
	}
	var concurrent int
	if flag.Arg(1) == "" {
		concurrent = 10
	} else {
		concurrent, _ = strconv.Atoi(flag.Arg(1))
	}
	var requestCount = 100
	if flag.Arg(2) != "" {
		requestCount, _ = strconv.Atoi(flag.Arg(2))
	}
	c := make(chan int, concurrent)
	r := make(chan int, requestCount)
	for i := 0; i < concurrent; i++ {
		c <- 1
	}
	for i := 0; i < requestCount; i++ {
		go func() {
			<-c
			b := time.Now()
			http.Get(url)
			e := time.Now()
			c <- 1
			r <- int(e.Sub(b).Milliseconds())
		}()
	}
	times := make([]int, 0)
	sum := 0
	for len(times) < requestCount {
		t := <-r
		//fmt.Println(t)
		sum += t
		times = append(times, t)
	}
	sort.Sort(sort.IntSlice(times))
	fmt.Println(fmt.Sprintf("avg:%d", sum/requestCount))
	fmt.Println(fmt.Sprintf("tp95:%d", times[95]))
}
