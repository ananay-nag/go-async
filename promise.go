package promise

import (
	"context"
	"sync"
	"fmt"
)



func New(executor func(resolve func(any), reject func(error))) *Promise {
	p := &Promise{done: make(chan struct{})}
	go func() {
		executor(
			func(v any) {
				p.resolve(v)
			},
			func(err error) {
				p.reject(err)
			},
		)
	}()
	return p
}

func Async(f func() (any, error)) *Promise {
	return New(func(resolve func(any), reject func(error)) {
		v, err := f()
		if err != nil {
			reject(err)
		} else {
			resolve(v)
		}
	})
}

func (p *Promise) resolve(v any) {
	p.once.Do(func() {
		p.result = v
		close(p.done)
	})
}

func (p *Promise) reject(err error) {
	p.once.Do(func() {
		p.err = err
		close(p.done)
	})
}

func Await(p *Promise) (any, error) {
	<-p.done
	return p.result, p.err
}

func (p *Promise) Then(onFulfilled func(any) any) *Promise {
	return New(func(resolve func(any), reject func(error)) {
		go func() {
			val, err := Await(p)
			if err != nil {
				reject(err)
				return
			}
			// Recover panic inside Then
			defer func() {
				if r := recover(); r != nil {
					reject(fmt.Errorf("panic in Then: %v", r))
				}
			}()
			res := onFulfilled(val)
			resolve(res)
		}()
	})
}


func (p *Promise) Catch(onRejected func(error) any) *Promise {
	return New(func(resolve func(any), reject func(error)) {
		go func() {
			val, err := Await(p)
			if err != nil {
				// Recover panic inside Catch
				defer func() {
					if r := recover(); r != nil {
						reject(fmt.Errorf("panic in Catch: %v", r))
					}
				}()
				res := onRejected(err)
				resolve(res)
				return
			}
			resolve(val)
		}()
	})
}


func (p *Promise) Finally(f func()) *Promise {
	return New(func(resolve func(any), reject func(error)) {
		res, err := Await(p)
		f()
		if err != nil {
			reject(err)
		} else {
			resolve(res)
		}
	})
}


func All(promises []*Promise) *Promise {
	return Async(func() (any, error) {
		results := make([]any, len(promises))
		for i, p := range promises {
			res, err := Await(p)
			if err != nil {
				return nil, err
			}
			results[i] = res
		}
		return results, nil
	})
}

func Race(promises []*Promise) *Promise {
	return New(func(resolve func(any), reject func(error)) {
		var once sync.Once
		for _, p := range promises {
			go func(p *Promise) {
				res, err := Await(p)
				once.Do(func() {
					if err != nil {
						reject(err)
					} else {
						resolve(res)
					}
				})
			}(p)
		}
	})
}

func AllSettled(promises []*Promise) *Promise {
	return Async(func() (any, error) {
		results := make([]Settlement, len(promises))
		for i, p := range promises {
			res, err := Await(p)
			if err != nil {
				results[i] = Settlement{Status: "rejected", Reason: err}
			} else {
				results[i] = Settlement{Status: "fulfilled", Value: res}
			}
		}
		return results, nil
	})
}

func Any(promises []*Promise) *Promise {
	return New(func(resolve func(any), reject func(error)) {
		var (
			errs     = make([]error, len(promises))
			errCount = 0
			mu       sync.Mutex
		)
		for i, p := range promises {
			go func(index int, prom *Promise) {
				val, err := Await(prom)
				if err == nil {
					resolve(val)
					return
				}
				mu.Lock()
				defer mu.Unlock()
				errs[index] = err
				errCount++
				if errCount == len(promises) {
					reject(fmt.Errorf("all promises rejected: %v", errs))
				}
			}(i, p)
		}
	})
}


func WithContext(ctx context.Context, executor func(resolve func(any), reject func(error))) *Promise {
	p := &Promise{done: make(chan struct{})}
	go func() {
		executor(
			func(v any) {
				select {
				case <-ctx.Done():
					p.reject(ctx.Err())
				case <-p.done:
					// already done
				default:
					p.resolve(v)
				}
			},
			func(err error) {
				select {
				case <-ctx.Done():
					p.reject(ctx.Err())
				case <-p.done:
					// already done
				default:
					p.reject(err)
				}
			},
		)
	}()
	go func() {
		<-ctx.Done()
		p.reject(ctx.Err())
	}()
	return p
}
