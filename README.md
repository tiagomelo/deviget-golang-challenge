# Golang-Challenge

## The Challenge
Finish the implementation of the provided Transparent Cache package.

## Key decisions
- running tests with `-race` option to detect race conditions
- using [sync.Map](https://pkg.go.dev/sync#Map) instead of a plain map to avoid race conditions; this allow us to not use a mutex in `GetPriceFor()` and thus keeping faster response times
- using [sync.Errgroup](https://pkg.go.dev/golang.org/x/sync/errgroup) to paralelize `GetPricesFor()`; thus it keeps the original implementation decision of stoping execution if any error is raised when getting the price for any given item
- keeping up with good practices such as variable declaration: use `var x` when its desirable to initialize it with its zero value, and use `x := some_value` when its desirable to have an initial value

## Running tests

```
make test
```

## Getting help about available targets in [Makefile](./Makefile)

```
make help
```