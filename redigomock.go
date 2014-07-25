// Copyright 2014 Rafael Dantas Justo. All rights reserved.
// Use of this source code is governed by a GPL
// license that can be found in the LICENSE file.

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

// Conn is the struct that can be used where you inject the redigo.Conn on your project
type Conn struct {
}

// NewConn returns a new mocked connection. Obviously as we are mocking we don't need any Redis
// conneciton parameter
func NewConn() redis.Conn {
	return Conn{}
}

// Close is dummy method only to respect the redigo.Conn interface (we don't have a connection to
// close)
func (c Conn) Close() error {
	return nil
}

// Err for now is a dummy object, depending on the demand of checking this level of error we can
// start returning something useful here
func (c Conn) Err() error {
	return nil
}

// Do looks in the registered commands (via Command function) if someone matchs with the given
// command name and arguments, if so the corresponding response or error is returned. If no
// registered command is found an error is returned
func (c Conn) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	cmd := find(commandName, args)
	if cmd == nil {
		// Didn't find a specific command, try to get a generic one
		if cmd = find(commandName, nil); cmd == nil {
			return nil, fmt.Errorf("command %s with arguments %v not registered in redigomock library",
				commandName, args)
		}
	}

	return cmd.Response, cmd.Error
}

// Send stores the command and arguments to be executed later (by the Receive function) in a first-
// come first-served order
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

// Flush is a dummy method as we don't execute nothing
func (c Conn) Flush() error {
	return nil
}

// Receive will process the queue created by the Send method, only one item of the queue is
// processed by Receive call. It will work as the Do method.
func (c Conn) Receive() (reply interface{}, err error) {
	if len(queue) == 0 {
		return nil, fmt.Errorf("no more items")
	}

	reply, err = c.Do(queue[0].commandName, queue[0].args...)
	queue = queue[1:]
	return
}

// Clear remove all registered commands and empty the queue
func Clear() {
	queue = []struct {
		commandName string
		args        []interface{}
	}{}

	commands = []*Cmd{}
}
