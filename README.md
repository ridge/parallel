# Go library for structured concurrency
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/ridge/parallel)

[Structured concurrency](https://en.wikipedia.org/wiki/Structured_concurrency)
helps reasoning about the behaviour of parallel programs. `parallel` implements
structured concurrency for Go.

    func subtask(ctx context.Context) error {
        // to be run in parallel
    }
    
    type subtaskWithData struct { /* ... * / }
    
    func (swd *subtaskWithData) Run(ctx context.Context) error {
        // to be run in parallel
    }

    err := parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {
        swd := &subtaskWithData{}

        // do some synchronous initialization here
  
        spawn("subtask", parallel.Fail, subtask)
        spawn("subtaskWithData", parallel.Fail, swd.Run)
        return nil
    })

Runs initializaiton within `parallel.Run()`, and then waits until context is
canceled, or one of spawned tasks finishes. Panics in goroutines are captured.

See the [documentation](https://pkg.go.dev/github.com/ridge/parallel) for
additional features:
- subprocess groups without inversion of control
- tasks that may exit and keep the group running
- tasks that may exit and cause the group to stop gracefully

## Legal

Copyright Tectonic Labs Ltd.

Licensed under [Apache 2.0](LICENSE) license.

Authors:
- [Alexey Feldgendler](https://github.com/feldgendler)
- [Misha Gusarov](https://github.com/dottedmag)
- [Wojciech Małota-Wójcik](https://www.linkedin.com/in/wmalota/)
