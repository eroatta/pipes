package main

import (
	"fmt"
	"sync"
)

func main() {
	fmt.Println("Pipelines...")

	fmt.Println("Basic Pipelines...")
	// set up the pipeline
	c := gen(2, 3)
	out := sq(c)

	// consume the output
	fmt.Println(<-out)
	fmt.Println(<-out)

	fmt.Println("Chaining stages...")
	// composition through same type for inbound and outbound channels
	for n := range sq(sq(gen(2, 3))) {
		fmt.Println(n)
	}

	fmt.Println("Fan-out, Fan-in...")
	in := gen(2, 3)

	// distribute the sq work across two goroutines that both read from in
	c1 := sq(in)
	c2 := sq(in)

	// consume the merged output from c1 and c2
	for n := range merge(c1, c2) {
		fmt.Println(n)
	}

	fmt.Println("Explicit cancellation...")
	// set up a done channel that's shared by the whole pipeline, and close
	// that channel when this pipeline exits, as a signal for all the goroutines we started to exit
	done := make(chan struct{})
	defer close(done)

	inp := genWithCancel(done, 2, 3)
	cd1 := sqWithCancel(done, inp)
	cd2 := sqWithCancel(done, inp)

	// consume the first value from output
	outp := mergeWithCancel(done, cd1, cd2)
	fmt.Println(<-outp)
}

func gen(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		for _, n := range nums {
			out <- n
		}
		close(out)
	}()

	return out
}

func sq(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			out <- n * n
		}
		close(out)
	}()

	return out
}

func merge(cs ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	out := make(chan int)

	// start and output goroutine for each input channel in cs
	// output copies values from c to out until c is closed, then calls wg.Done
	output := func(c <-chan int) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}

	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// start a goroutine to close out once all the output goroutines are done
	// this must start after the wg.Add call
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func genWithCancel(done <-chan struct{}, nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			select {
			case out <- n:
			case <-done:
				return
			}
		}

	}()

	return out
}

func sqWithCancel(done <-chan struct{}, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			select {
			case out <- n * n:
			case <-done:
				return
			}
		}
	}()

	return out
}

func mergeWithCancel(done <-chan struct{}, cs ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	out := make(chan int)

	// start and output goroutine for each input channel in cs
	// output copies values from c to out until c is closed, then calls wg.Done
	output := func(c <-chan int) {
		defer wg.Done()
		for n := range c {
			select {
			case out <- n:
			case <-done:
				return
			}
		}
	}

	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// start a goroutine to close out once all the output goroutines are done
	// this must start after the wg.Add call
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
