redigomock
==========

[![Build Status](https://travis-ci.org/gmlexx/redigomock.svg)](https://travis-ci.org/gmlexx/redigomock)

[![GoDoc](https://godoc.org/github.com/gmlexx/redigomock?status.png)](https://godoc.org/github.com/gmlexx/redigomock)

Easy way to unit test projects using [redigo library](https://github.com/garyburd/redigo) (Redis client in go).

install
-------

```
go get -u github.com/gmlexx/redigomock
```

usage
-----

Here is an example of using redigomock, for more information please check the [API documentation](https://godoc.org/github.com/gmlexx/redigomock).

```go
package main

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/gmlexx/redigomock"
)

type Person struct {
	Name string `redis:"name"`
	Age  int    `redis:"age"`
}

func RetrievePerson(conn redis.Conn, id string) (Person, error) {
	var person Person

	values, err := redis.Values(conn.Do("HGETALL", fmt.Sprintf("person:%s", id)))
	if err != nil {
		return person, err
	}

	err = redis.ScanStruct(values, &person)
	return person, err
}

func main() {
	// Simulate command result

	conn := redigomock.NewConn()
	cmd := conn.Command("HGETALL", "person:1").ExpectMap(map[string]string{
		"name": "Mr. Johson",
		"age":  "42",
	})

	person, err := RetrievePerson(conn, "1")
	if err != nil {
		fmt.Println(err)
		return
	}

	if conn.Stats(cmd) != 1 {
		fmt.Println("Command was not used")
		return
	}

	if person.Name != "Mr. Johson" {
		fmt.Printf("Invalid name. Expected 'Mr. Johson' and got '%s'\n", person.Name)
		return
	}

	if person.Age != 42 {
		fmt.Printf("Invalid age. Expected '42' and got '%d'\n", person.Age)
		return
	}

	// Simulate command error

	conn.Clear()
	cmd = conn.Command("HGETALL", "person:1").ExpectError(fmt.Errorf("Simulate error!"))

	person, err = RetrievePerson(conn, "1")
	if err == nil {
		fmt.Println("Should return an error!")
		return
	}

	if conn.Stats(cmd) != 1 {
		fmt.Println("Command was not used")
		return
	}

	fmt.Println("Success!")
}
```

fakeredis
---------

Inspired by [python fakeredis](https://github.com/jamesls/fakeredis), this library can mock redis connection to act as real redis.
This work is at the beginning and only the following commands are supported:
- GET
- SET
- MULTI, EXEC
- KEYS
- FLUSHDB
- SADD
- SMEMBERS
- SREM
- SUNION
- ZADD
- ZRANGE
- ZCOUNT
- ZREMRANGEBYSCORE
- ZRANGEBYSCORE
- HGET
- HSET
- HLEN
- HGETALL

Note: commands implementations are not optimized for performance.

Fakeredis example

```go
	import (
		"github.com/gmlexx/redigomock"
		"github.com/garyburd/redigo/redis"
	)

	// Fake connection is thread-safe
	c := redigomock.NewFakeRedis()
	pool := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return c, nil
		},
	}
```

