package main

import (
	"context"
	"fmt"
	"log"
	"time"

	promise "github.com/ananay-nag/go-async"
)

func fetchUserData(userID int) *promise.Promise {
	return promise.New(func(resolve func(interface{}), reject func(error)) {
		// Simulating an API call with a delay
		time.Sleep(2 * time.Second)
		if userID == 1 {
			resolve(map[string]string{"id": "1", "name": "John Doe"})
		} else {
			reject(fmt.Errorf("user not found"))
		}
	})
}

func fetchUserPosts(userID int) *promise.Promise {
	return promise.New(func(resolve func(interface{}), reject func(error)) {
		// Simulating an API call with a delay
		time.Sleep(1 * time.Second)
		if userID == 1 {
			resolve([]string{"Post 1", "Post 2"})
		} else {
			reject(fmt.Errorf("posts not found"))
		}
	})
}

func main() {
	// Example 1: Using All to fetch multiple resources
	p1 := fetchUserData(1)
	p2 := fetchUserPosts(1)

	all := promise.All([]*promise.Promise{p1, p2})

	res, err := promise.Await(all)
	if err != nil {
		log.Fatalf("All failed: %v", err)
	}
	fmt.Println("All resolved:", res)

	// Example 2: Using Race to fetch the first resolved resource
	race := promise.Race([]*promise.Promise{p1, p2})

	res, err = promise.Await(race)
	if err != nil {
		log.Fatalf("Race failed: %v", err)
	}
	fmt.Println("Race resolved:", res)

	// Example 3: Using Any to fetch the first successful resource
	any := promise.Any([]*promise.Promise{p1, p2})

	res, err = promise.Await(any)
	if err != nil {
		log.Fatalf("Any failed: %v", err)
	}
	fmt.Println("Any resolved:", res)

	// Example 4: Using WithContext to handle cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	withCtx := promise.WithContext(ctx, func(resolve func(interface{}), reject func(error)) {
		// Simulating a long-running operation
		time.Sleep(4 * time.Second)
		resolve("Operation completed")
	})

	_, err = promise.Await(withCtx)
	if err != nil {
		log.Printf("WithContext failed: %v", err)
	}
}
