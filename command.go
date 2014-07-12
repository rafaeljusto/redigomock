// Copyright 2014 Rafael Dantas Justo. All rights reserved.
// Use of this source code is governed by a GPL
// license that can be found in the LICENSE file.

package redigomock

import (
	"bytes"
	"fmt"
	"strconv"
)

var (
	commands map[string]*Cmd
)

func init() {
	commands = make(map[string]*Cmd)
}

// Cmd stores the registered information about a command to return it later when request by a
// command execution
type Cmd struct {
	Response interface{} // Response to send back when this command/arguments are called
	Err      error       // Error to send back when this command/arguments are called
}

// Command register a command in the mock system using the same arguments of a Do or Send commands.
// It will return a registered command object where you can set the response or error
func Command(commandName string, args ...interface{}) *Cmd {
	var cmd Cmd
	commands[generateKey(commandName, args)] = &cmd
	return &cmd
}

// Expect sets a response for this command. Everytime a Do or Receive methods are executed for a
// registered command this response or error will be returned. You cannot set a response and a error
// for the same command/arguments
func (c *Cmd) Expect(response interface{}) {
	c.Response = response
	c.Err = nil
}

// ExpectMap works in the same way of the Expect command, but has a key/value input to make it
// easier to build test environments
func (c *Cmd) ExpectMap(response map[string]string) {
	var values []interface{}
	for key, value := range response {
		values = append(values, []byte(key))
		values = append(values, []byte(value))
	}

	c.Response = values
	c.Err = nil
}

// ExpectError allows you to force an error when executing a command/arguments
func (c *Cmd) ExpectError(err error) {
	c.Response = nil
	c.Err = err
}

// generateKey build an id for the command/arguments to make it easier to find in the registered
// commands
func generateKey(commandName string, args []interface{}) string {
	key := commandName

	for _, arg := range args {
		switch arg := arg.(type) {
		case string:
			key += " " + arg
		case []byte:
			key += " " + string(arg)
		case int:
			key += " " + strconv.Itoa(arg)
		case int64:
			key += " " + strconv.FormatInt(arg, 10)
		case float64:
			key += " " + strconv.FormatFloat(arg, 'g', -1, 64)
		case bool:
			if arg {
				key += " 1"
			} else {
				key += " 0"
			}
		case nil:
			key += " "
		default:
			var buf bytes.Buffer
			fmt.Fprint(&buf, arg)
			key += " " + buf.String()
		}
	}

	return key
}
