// Copyright 2014 Rafael Dantas Justo. All rights reserved.
// Use of this source code is governed by a GPL
// license that can be found in the LICENSE file.

package redigomock

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sync"
)

var statMutex sync.Mutex

type queueElement struct {
	commandName string
	args        []interface{}
}

// Conn is the struct that can be used where you inject the redigo.Conn on
// your project
type Conn struct {
	ReceiveWait    bool            // When set to true, Receive method will wait for a value in ReceiveNow channel to proceed, this is useful in a PubSub scenario
	ReceiveNow     chan bool       // Used to lock Receive method to simulate a PubSub scenario
	CloseMock      func() error    // Mock the redigo Close method
	ErrMock        func() error    // Mock the redigo Err method
	FlushMock      func() error    // Mock the redigo Flush method
	commands       []*Cmd          // Slice that stores all registered commands for each connection
	queue          []queueElement  // Slice that stores all queued commands for each connection
	stats          map[cmdHash]int // Command calls counter
	pendingResults []interface{}
}

// NewConn returns a new mocked connection. Obviously as we are mocking we
// don't need any Redis connection parameter
func NewConn() *Conn {
	return &Conn{
		ReceiveNow:     make(chan bool),
		stats:          make(map[cmdHash]int),
		pendingResults: make([]interface{}, 0),
	}
}

// Close can be mocked using the Conn struct attributes
func (c *Conn) Close() error {
	if c.CloseMock == nil {
		return nil
	}

	return c.CloseMock()
}

// Err can be mocked using the Conn struct attributes
func (c *Conn) Err() error {
	if c.ErrMock == nil {
		return nil
	}

	return c.ErrMock()
}

// Command register a command in the mock system using the same arguments of
// a Do or Send commands. It will return a registered command object where
// you can set the response or error
func (c *Conn) Command(commandName string, args ...interface{}) *Cmd {
	cmd := &Cmd{
		Name: commandName,
		Args: args,
	}
	for _, a := range args {
		if any, ok := a.(anyData); ok {
			cmd.ignoreArgsLength = any.ignoreArgsLength
		}
	}
	c.removeRelatedCommands(commandName, args)
	c.commands = append(c.commands, cmd)
	return cmd
}

// Script registers a command in the mock system just like Command method
// would do. The first argument is a byte array with the script text, next
// ones are the ones you would pass to redis Script.Do() method
func (c *Conn) Script(scriptData []byte, keyCount int, args ...interface{}) *Cmd {
	h := sha1.New()
	h.Write(scriptData)
	sha1sum := hex.EncodeToString(h.Sum(nil))

	newArgs := make([]interface{}, 2+len(args))
	newArgs[0] = sha1sum
	newArgs[1] = keyCount
	copy(newArgs[2:], args)

	return c.Command("EVALSHA", newArgs...)
}

// GenericCommand register a command without arguments. If a command with
// arguments doesn't match with any registered command, it will look for
// generic commands before throwing an error
func (c *Conn) GenericCommand(commandName string) *Cmd {
	cmd := &Cmd{
		Name: commandName,
	}

	c.removeRelatedCommands(commandName, nil)
	c.commands = append(c.commands, cmd)
	return cmd
}

// find will scan the registered commands, looking for the first command with
// the same name and arguments. If the command is not found nil is returned
func (c *Conn) find(commandName string, args []interface{}) *Cmd {
	for _, cmd := range c.commands {
		if match(commandName, args, cmd) {
			return cmd
		}
	}
	return nil
}

// removeRelatedCommands verify if a command is already registered, removing
// any command already registered with the same name and arguments. This
// should avoid duplicated mocked commands
func (c *Conn) removeRelatedCommands(commandName string, args []interface{}) {
	var unique []*Cmd

	for _, cmd := range c.commands {
		// new array will contain only commands that are not related to the given
		// one
		if !equal(commandName, args, cmd) {
			unique = append(unique, cmd)
		}
	}
	c.commands = unique
}

// Clear removes all registered commands. Useful for connection reuse in test
// scenarios
func (c *Conn) Clear() {
	c.commands = []*Cmd{}
	c.queue = []queueElement{}
}

// Do looks in the registered commands (via Command function) if someone
// matches with the given command name and arguments, if so the corresponding
// response or error is returned. If no registered command is found an error
// is returned
func (c *Conn) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	queueLength := len(c.queue)
	if queueLength > 0 {
		// Process the queued commands first
		cmd := c.queue[queueLength-1]
		c.queue = c.queue[:queueLength-1]
		reply, err = c.Do(cmd.commandName, cmd.args...)
		if err != nil {
			return nil, err
		}
		if reply != nil {
			c.pendingResults = append(c.pendingResults, reply)
		}
	}

	cmd := c.find(commandName, args)
	if cmd == nil {
		// Didn't find a specific command, try to get a generic one
		if cmd = c.find(commandName, nil); cmd == nil {
			return nil, fmt.Errorf("command %s with arguments %#v not registered in redigomock library",
				commandName, args)
		}
	}

	statMutex.Lock()
	c.stats[cmd.hash()]++
	statMutex.Unlock()

	if cmd.Callback != nil {
		return cmd.invokeCallback(args)
	}

	if len(cmd.Responses) == 0 {
		return nil, nil
	}
	response := cmd.Responses[0]
	cmd.Responses = cmd.Responses[1:]
	return response.Response, response.Error
}

// Send stores the command and arguments to be executed later (by the Receive
// function) in a first-come first-served order
func (c *Conn) Send(commandName string, args ...interface{}) error {
	c.queue = append(c.queue, queueElement{
		commandName: commandName,
		args:        args,
	})
	return nil
}

// Flush can be mocked using the Conn struct attributes
func (c *Conn) Flush() error {
	if c.FlushMock == nil {
		return nil
	}

	return c.FlushMock()
}

// Receive will process the queue created by the Send method, only one item
// of the queue is processed by Receive call. It will work as the Do method
func (c *Conn) Receive() (reply interface{}, err error) {
	if c.ReceiveWait {
		<-c.ReceiveNow
	}

	if len(c.queue) == 0 {
		return nil, fmt.Errorf("no more items")
	}

	commandName, args := c.queue[0].commandName, c.queue[0].args
	cmd := c.find(commandName, args)
	if cmd == nil {
		// Didn't find a specific command, try to get a generic one
		if cmd = c.find(commandName, nil); cmd == nil {
			return nil, fmt.Errorf("command %s with arguments %#v not registered in redigomock library",
				commandName, args)
		}
	}

	statMutex.Lock()
	c.stats[cmd.hash()]++
	statMutex.Unlock()

	if cmd.Callback != nil {
		return cmd.invokeCallback(args)
	}

	if len(cmd.Responses) == 0 {
		reply, err = nil, nil
	} else {
		response := cmd.Responses[0]
		cmd.Responses = cmd.Responses[1:]
		reply, err = response.Response, response.Error
	}

	c.queue = c.queue[1:]
	return
}

// Stats returns the number of times that a command was called in the current
// connection
func (c Conn) Stats(cmd *Cmd) int {
	return c.stats[cmd.hash()]
}
