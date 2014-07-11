package redigomock

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
)

var (
	queue []struct {
		commandName string
		args        []interface{}
	}
)

type Conn struct {
}

func NewConn() redis.Conn {
	return Conn{}
}

func (c Conn) Close() error {
	return nil
}

func (c Conn) Err() error {
	return nil
}

func (c Conn) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	cmd, found := commands[generateKey(commandName, args)]
	if !found {
		return nil, fmt.Errorf("Command %s with arguments [%v] not registered in redigomock library!",
			commandName, args)
	}

	return cmd.Response, cmd.Err
}

func (c Conn) Send(commandName string, args ...interface{}) error {
	queue = append(queue, struct {
		commandName string
		args        []interface{}
	}{
		commandName: commandName,
		args:        args,
	})

	return nil
}

func (c Conn) Flush() error {
	return nil
}

func (c Conn) Receive() (reply interface{}, err error) {
	if len(queue) == 0 {
		return nil, fmt.Errorf("No more items")
	}

	reply, err = c.Do(queue[0].commandName, queue[0].args)
	queue = queue[1:]
	return
}
