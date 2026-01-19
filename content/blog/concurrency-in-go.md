# Concurrency in Go: Practical Patterns

Go's concurrency model is built around two ideas: lightweight goroutines and typed channels for communication. The goal is not "parallelism everywhere" but clean, testable structure for work that overlaps or waits (I/O, timers, user events, background tasks). This post is a quick tour of patterns you can drop into real programs.

## 1) Fan-out / fan-in with workers

Use a worker pool to bound concurrency, then combine results on a single channel. This keeps memory and CPU use predictable.

```go
package main

import (
    "fmt"
    "sync"
)

type Job struct {
    ID int
}

type Result struct {
    ID  int
    Msg string
}

func worker(id int, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
    defer wg.Done()
    for job := range jobs {
        results <- Result{ID: job.ID, Msg: fmt.Sprintf("worker %d handled job %d", id, job.ID)}
    }
}

func main() {
    const workerCount = 3
    const jobCount = 10

    jobs := make(chan Job)
    results := make(chan Result)

    var wg sync.WaitGroup
    wg.Add(workerCount)
    for i := 1; i <= workerCount; i++ {
        go worker(i, jobs, results, &wg)
    }

    go func() {
        for i := 1; i <= jobCount; i++ {
            jobs <- Job{ID: i}
        }
        close(jobs)
        wg.Wait()
        close(results)
    }()

    for result := range results {
        fmt.Println(result.Msg)
    }
}
```

## 2) Timeouts with select

Use `select` to avoid blocking forever. This is crucial when an API call might hang.

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    done := make(chan string)

    go func() {
        time.Sleep(200 * time.Millisecond)
        done <- "finished"
    }()

    select {
    case msg := <-done:
        fmt.Println(msg)
    case <-time.After(100 * time.Millisecond):
        fmt.Println("timed out")
    }
}
```

## 3) Cancellation with context

When you kick off work, pass a `context.Context` so callers can cancel. This prevents goroutine leaks when a request ends early.

```go
package main

import (
    "context"
    "fmt"
    "time"
)

func slowOperation(ctx context.Context) error {
    select {
    case <-time.After(2 * time.Second):
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
    defer cancel()

    if err := slowOperation(ctx); err != nil {
        fmt.Println("cancelled:", err)
        return
    }

    fmt.Println("success")
}
```

## 4) Pipelines for streaming

Pipelines keep your code modular and composable. Each stage only cares about input/output channels.

```go
package main

import "fmt"

func generate(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for _, n := range nums {
            out <- n
        }
    }()
    return out
}

func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            out <- n * n
        }
    }()
    return out
}

func main() {
    for n := range square(generate(2, 3, 4)) {
        fmt.Println(n)
    }
}
```

## 5) Avoid common traps

- Avoid shared mutable state. Prefer message passing.
- If you must share, guard with `sync.Mutex` and keep critical sections small.
- Always consider shutdown: who closes channels, and when?
- Limit concurrency for I/O heavy work (databases, APIs) to avoid overload.

## Closing thoughts

Concurrency is not about doing everything at once. It is about structuring code to handle waiting and overlapping work without getting tangled. Start with small goroutines, wire them together with channels, and always build in cancellation and timeouts.
