# 🌀 go-async

> A lightweight, generic async utility package for Go — inspired by JavaScript Promises.

This module provides utility functions (All, AllSettled, Race, Any) to run multiple asynchronous operations in parallel using goroutines — with clean, Promise-like behavior.

---

## ✨ Features

- ✅ **All**: Wait for all async tasks to succeed (like `Promise.all`)
- ⚠️ **AllSettled**: Wait for all tasks, capture both success and failure (like `Promise.allSettled`)
- 🏁 **Race**: Return the result of the first completed task (like `Promise.race`)
- 🚀 **Any**: Return the first successful result (like `Promise.any`)
- 🔁 **Generic support**: Works with any return type via Go 1.18+ generics
- 🔒 Thread-safe using `sync.Mutex` and `sync.WaitGroup`

---

## 📦 Installation

Clone the repo locally or use via `replace` in your own module:

```bash
go get github.com/ananay-nag/go-async
```


## 🧠 Usage
``` go
    import "github.com/ananay-nag/go-async"
```

## ✅ What go-async Adds Over Raw Goroutines

| Feature                          | go-async pkg                             | Raw Goroutines + Channels                 |
|-----------------------------------|-------------------------------------------|--------------------------------------------|
| 🔁 All, AllSettled, Race, Any     | ✅ Built-in                                | ❌ Manual wiring every time                 |
| 🔒 Thread-safe                    | ✅ Uses sync primitives internally         | ❌ Easy to make race conditions             |
| 🚀 Generic return types (any)     | ✅ Go 1.18+ generics                        | ❌ Need to type assert channels manually    |
| 💥 Error propagation              | ✅ Built into API                          | ❌ You manage error channels separately     |
| 🧹 Cleaner readable code          | ✅ Promise-style (`result, err := Any(...)`) | ❌ Often callback-ish, multiple selects     |
| 🧪 Tested utilities                | ✅ Covered via unit tests                  | ❌ You write tests each time                |
| 🔧 Composable                     | ✅ Chain patterns (All with Race, etc.)    | ❌ Not composable without custom code       |

## 🧠 Pro Insight
- Think of go-async as Promise-style orchestration for Go:

- great for discrete async tasks (API calls, DB queries, parallel file reads)

- less ideal for streaming and continuous processing (pipes, queues)


## 🚫 When Raw Goroutines Are Better
- Super low-level control (e.g., you need to cancel mid-way, or stream results continuously)

- Performance critical where every nanosecond counts (go-async has minor overhead for abstraction)

- Non-promise patterns like channels, fan-in, fan-out, pipelines

## ✅ Example: 
``` go

/example/main.go

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

```

## 🧪 Run Tests
All functions are covered with real-world test cases.

### 🔍 Run Unit Tests
```bash
cd go-async
go test -v or -c

```
## ✅ Test Coverage
- Here’s an overview of the test cases:

- TestAll: All tasks succeed.

- TestAllWithError: One task fails.

- TestAllSettled: Mixed results (success and failure).

- TestRace: The fastest task wins.

- TestAny: The first successful task.

- TestAnyAllFail: All tasks fail, returns an error.


## 📘 Requirements
Go 1.18+ (for generics support)

## 🔄 Improvements & Customization
- You can enhance error handling for better flexibility.

- Integrate other external services or APIs with this package.

- Extend the concurrency model to handle more complex async operations.