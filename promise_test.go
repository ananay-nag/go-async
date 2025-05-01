package promise_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ananay-nag/go-async" //replace with your actual package path
)

func TestPromiseResolve(t *testing.T) {
	p := promise.New(func(resolve func(any), reject func(error)) {
		resolve("ok")
	})
	res, err := promise.Await(p)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res != "ok" {
		t.Fatalf("expected 'ok', got %v", res)
	}
}

func TestPromiseReject(t *testing.T) {
	p := promise.New(func(resolve func(any), reject func(error)) {
		reject(errors.New("fail"))
	})

	res, err := promise.Await(p)
	if err == nil || err.Error() != "fail" {
		t.Fatalf("expected fail error, got %v", err)
	}
	if res != nil {
		t.Fatalf("expected nil result, got %v", res)
	}
}

func TestThen(t *testing.T) {
	p := promise.New(func(resolve func(any), reject func(error)) {
		resolve(2)
	}).Then(func(val any) any {
		return val.(int) * 2
	})

	res, err := promise.Await(p)
	if err != nil {
		t.Fatal(err)
	}
	if res != 4 {
		t.Fatalf("expected 4, got %v", res)
	}
}

func TestCatch(t *testing.T) {
	p := promise.New(func(resolve func(any), reject func(error)) {
		reject(errors.New("fail"))
	}).Catch(func(e error) any {
		return "recovered"
	})

	res, err := promise.Await(p)
	if err != nil {
		t.Fatal(err)
	}
	if res != "recovered" {
		t.Fatalf("expected recovered, got %v", res)
	}
}

func TestFinally(t *testing.T) {
	called := false
	p := promise.New(func(resolve func(any), reject func(error)) {
		resolve("ok")
	}).Finally(func() {
		called = true
	})

	_, err := promise.Await(p)
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("expected finally to be called")
	}
}

func TestAsync(t *testing.T) {
	p := promise.Async(func() (any, error) {
		return 42, nil
	})

	res, err := promise.Await(p)
	if err != nil {
		t.Fatal(err)
	}
	if res != 42 {
		t.Fatalf("expected 42, got %v", res)
	}
}

func TestWithContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	p := promise.WithContext(ctx, func(resolve func(any), reject func(error)) {
		resolve("should not happen")
	})

	_, err := promise.Await(p)
	if err == nil {
		t.Fatal("expected context cancellation")
	}
}

func TestThenPanic(t *testing.T) {
	p := promise.New(func(resolve func(any), reject func(error)) {
		resolve(1)
	}).Then(func(any) any {
		panic("boom")
	})

	_, err := promise.Await(p)
	if err == nil || err.Error() != "panic in Then: boom" {
		t.Fatalf("expected panic error, got %v", err)
	}
}

func TestCatchPanic(t *testing.T) {
	p := promise.New(func(resolve func(any), reject func(error)) {
		reject(errors.New("fail"))
	}).Catch(func(e error) any {
		panic("boom")
	})

	_, err := promise.Await(p)
	if err == nil || err.Error() != "panic in Catch: boom" {
		t.Fatalf("expected panic error, got %v", err)
	}
}


// delayedResolve and delayedReject are utility functions to create promises that resolve or reject after a delay.
func delayedResolve(val any, delay time.Duration) *promise.Promise {
	return promise.New(func(resolve func(any), reject func(error)) {
		time.AfterFunc(delay, func() {
			resolve(val)
		})
	})
}


// delayedReject is a utility function to create a promise that rejects after a delay.
func delayedReject(err error, delay time.Duration) *promise.Promise {
	return promise.New(func(resolve func(any), reject func(error)) {
		time.AfterFunc(delay, func() {
			reject(err)
		})
	})
}

// TestAll tests the All function of the promise package.
func TestAll(t *testing.T) {
	p1 := delayedResolve("one", 50*time.Millisecond)
	p2 := delayedResolve("two", 100*time.Millisecond)

	all := promise.All([]*promise.Promise{p1, p2})

	res, err := promise.Await(all)
	if err != nil {
		t.Fatalf("All rejected: %v", err)
	}

	results, ok := res.([]any)
	if !ok || len(results) != 2 {
		t.Fatalf("Unexpected result: %v", res)
	}
	if results[0] != "one" || results[1] != "two" {
		t.Errorf("Unexpected All result values: %v", results)
	}
}

// TestAllReject tests the All function with a rejection.
func TestAllReject(t *testing.T) {
	p1 := delayedResolve("ok", 50*time.Millisecond)
	p2 := delayedReject(errors.New("fail"), 20*time.Millisecond)

	all := promise.All([]*promise.Promise{p1, p2})

	_, err := promise.Await(all)
	if err == nil || err.Error() != "fail" {
		t.Fatalf("Expected rejection with 'fail', got: %v", err)
	}
}

// TestAllSettled tests the AllSettled function of the promise package.
func TestAllSettled(t *testing.T) {
	p1 := delayedResolve("ok", 30*time.Millisecond)
	p2 := delayedReject(errors.New("fail"), 50*time.Millisecond)

	allSettled := promise.AllSettled([]*promise.Promise{p1, p2})

	res, err := promise.Await(allSettled)
	if err != nil {
		t.Fatalf("AllSettled rejected: %v", err)
	}

	settlements, ok := res.([]promise.Settlement)
	if !ok || len(settlements) != 2 {
		t.Fatalf("Unexpected settlements: %v", res)
	}

	if settlements[0].Status != "fulfilled" || settlements[0].Value != "ok" {
		t.Errorf("Unexpected first settlement: %v", settlements[0])
	}
	if settlements[1].Status != "rejected" || settlements[1].Reason.Error() != "fail" {
		t.Errorf("Unexpected second settlement: %v", settlements[1])
	}
}

// TestRace tests
func TestRace(t *testing.T) {
	p1 := delayedResolve("slow", 100*time.Millisecond)
	p2 := delayedResolve("fast", 20*time.Millisecond)

	racer := promise.Race([]*promise.Promise{p1, p2})

	res, err := promise.Await(racer)
	if err != nil {
		t.Fatalf("Race rejected: %v", err)
	}
	if res != "fast" {
		t.Errorf("Race should resolve with 'fast', got: %v", res)
	}
}

// TestAny tests the Any function of the promise package.
func TestAny(t *testing.T) {
	p1 := delayedReject(errors.New("err1"), 10*time.Millisecond)
	p2 := delayedResolve("yay", 30*time.Millisecond)

	anyP := promise.Any([]*promise.Promise{p1, p2})

	res, err := promise.Await(anyP)
	if err != nil {
		t.Fatalf("Any rejected: %v", err)
	}
	if res != "yay" {
		t.Errorf("Any should resolve with 'yay', got: %v", res)
	}
}

// TestAnyRejectsAll tests the Any function when all promises reject.
func TestAnyRejectsAll(t *testing.T) {
	p1 := delayedReject(errors.New("fail1"), 10*time.Millisecond)
	p2 := delayedReject(errors.New("fail2"), 20*time.Millisecond)

	anyP := promise.Any([]*promise.Promise{p1, p2})

	_, err := promise.Await(anyP)
	if err == nil || !strings.Contains(err.Error(), "all promises rejected") {
		t.Fatalf("Unexpected error message: %v", err)
	}
	if got := err.Error(); got == "all promises rejected: [fail1 fail2]" {
		// correct
	} else {
		t.Errorf("Unexpected error message: %v", got)
	}
}

// TestAggregateError tests the AggregateError type.
func TestWithContextCancel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	p := promise.WithContext(ctx, func(resolve func(any), reject func(error)) {
		// Never resolve or reject, simulating a hang
		time.Sleep(100 * time.Millisecond)
		resolve("too late")
	})

	_, err := promise.Await(p)
	if err == nil {
		t.Fatalf("Expected context cancel error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected DeadlineExceeded, got: %v", err)
	}
}
