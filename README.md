# Redis hash API for Golang

With tag of redis, setting struct into redis, getting struct from redis.

## Features

- Support tag for struct field as JSON or ORM.
- Set a struct object (with a key) into redis hash.
- Set and cover an existing hash value.
- Read a hash value without some field(s).

## Installation

redistag requires a Go version with [modules](https://github.com/golang/go/wiki/Modules) support. So make sure your project is using go module.

Then, install redistag:

```shell
go get "github.com/seedjyh/redistag/v2"
```

## Quickstart

```go
package main

import (
	"context"
	"fmt"
	goRedis "github.com/go-redis/redis/v8"
	redisTag "github.com/seedjyh/redistag/v2"
)

func main() {
	type Book struct {
		ISBN      string  `redis:"isbn"`
		Title     string  `redis:"'title'"`
		Author    string  `redis:"'author'"`
		Price     float64 `redis:"'price'"`
		IsInStock bool    `redis:"'isInStock'"`
	}

	cmdable := goRedis.NewClusterClient(&goRedis.ClusterOptions{
		Addrs: []string{
			"localhost:8001",
			"localhost:8002",
			"localhost:8003",
		},
	})

	b := &Book{
		ISBN:      "1234567890",
		Title:     "Hello, world!",
		Author:    "seedjyh",
		Price:     0.01,
		IsInStock: true,
	}

	key := "isbn2book:" + b.ISBN

	if err := redisTag.HMSet(context.Background(), cmdable, key, b); err != nil {
		panic(err)
	}

	r := new(Book)
	if err := redisTag.HMGet(context.Background(), cmdable, key, r); err != nil {
		if err == goRedis.Nil {
			fmt.Println("No such book.")
		} else {
			panic(err)
		}
	} else {
		fmt.Println("Found book!", r)
	}
	// Output: Found book! &{ Hello, world! seedjyh 0.01 true}
	return
}
```
