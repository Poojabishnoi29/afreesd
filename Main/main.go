package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

func FetchLength(url string) (int, error) {
	Client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := Client.Get(url)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return -1, err
	}
	return len(body), nil
}

type Job struct {
	url   string
	index int
}
type Result struct {
	index  int
	length int
}

func ProcessUrl(urls []string, worker int) []int {
	wg := sync.WaitGroup{}
	job := make(chan Job)
	result := make(chan Result)
	for i := 0; i < worker; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range job {
				length := func() int {
					defer func() {
						if r := recover(); r != nil {
							fmt.Printf("Panic fetching %s: %v\n", j.url, r)
						}
					}()

					length, err := FetchLength(j.url)
					if err != nil {
						return -1
					}
					return length
				}()
				result <- Result{index: j.index, length: length}
			}
		}()
	}
	go func() {
		for i, url := range urls {
			job <- Job{url: url, index: i}
		}
		close(job)
	}()

	results := make([]int, len(urls))
	done := make(chan bool)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Consumer panic: %v\n", r)
			}
			done <- true
		}()

		for r := range result {
			results[r.index] = r.length
		}
	}()

	wg.Wait()
	close(result)
	<-done

	return results
}
func main() {
	urls := []string{
		"http://www.google.com",
		"http://www.facebook.com",
		"http://www.twitter.com",
		"http://www.instagram.com",
		"http://www.youtube.com",
	}
	worker := 3
	results := ProcessUrl(urls, worker)

	for i, length := range results {
		fmt.Printf("URL %d (%s): Length = %d\n", i, urls[i], length)
	}
}
