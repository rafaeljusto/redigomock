// Copyright 2014 Rafael Dantas Justo. All rights reserved.
// Use of this source code is governed by a GPL
// license that can be found in the LICENSE file.

package redigomock

import (
	"crypto/sha1"
	"encoding/hex"
	"reflect"
)

// Response struct that represents single response from `Do` call
type Response struct {
	Response interface{} // Response to send back when this command/arguments are called
	Error    error       // Error to send back when this command/arguments are called
}

// Cmd stores the registered information about a command to return it later when request by a
// command execution
type Cmd struct {
	Name      string        // Name of the command
	Args      []interface{} // Arguments of the command
	Responses []Response    // Slice of returned responses
}

// Command register a command in the mock system using the same arguments of a Do or Send commands.
// It will return a registered command object where you can set the response or error
func (c *Conn) Command(commandName string, args ...interface{}) *Cmd {
	cmd := &Cmd{
		Name: commandName,
		Args: args,
	}
	c.removeRelatedCommands(commandName, args)
	c.commands = append(c.commands, cmd)
	return cmd
}

// Script registers a command in the mock system just like Command method would do
// The first argument is a byte array with the script text, next ones are the ones
// you would pass to redis Script.Do() method
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

// GenericCommand register a command without arguments. If a command with arguments doesn't match
// with any registered command, it will look for generic commands before throwing an error
func (c *Conn) GenericCommand(commandName string) *Cmd {
	cmd := &Cmd{
		Name: commandName,
	}

	c.removeRelatedCommands(commandName, nil)
	c.commands = append(c.commands, cmd)
	return cmd
}

//find will scan the registered commands, looking for the first command with the same name and
//arguments. If the command is not found nil is returned
func (c *Conn) find(commandName string, args []interface{}) *Cmd {
	for _, cmd := range c.commands {
		if match(commandName, args, cmd) {
			return cmd
		}
	}
	return nil
}

// removeRelatedCommands verify if a command is already registered, removing any command already
// registered with the same name and arguments. This should avoid duplicated mocked commands
func (c *Conn) removeRelatedCommands(commandName string, args []interface{}) {
	var unique []*Cmd

	for _, cmd := range c.commands {
		// New array will contain only commands that are not related to the given one
		if !equal(commandName, args, cmd) {
			unique = append(unique, cmd)
		}
	}
	c.commands = unique
}

func (c *Conn) Clear() {
	c.commands = []*Cmd{}
}

// match verify if a command/argumets is related to a registered command.
func equal(commandName string, args []interface{}, cmd *Cmd) bool {
	if commandName != cmd.Name || len(args) != len(cmd.Args) {
		return false
	}

	for pos := range cmd.Args {
		if implementsFuzzy(cmd.Args[pos]) && implementsFuzzy(args[pos]) {
			if reflect.TypeOf(cmd.Args[pos]) != reflect.TypeOf(args[pos]) {
				return false
			}
		} else if implementsFuzzy(cmd.Args[pos]) || implementsFuzzy(args[pos]) {
			return false
		} else {
			if reflect.DeepEqual(cmd.Args[pos], args[pos]) == false {
				return false
			}
		}
	}
	return true
}

//match check if provided arguments can be matched with any registered commands
func match(commandName string, args []interface{}, cmd *Cmd) bool {
	if commandName != cmd.Name || len(args) != len(cmd.Args) {
		return false
	}

	for pos := range cmd.Args {
		if implementsFuzzy(cmd.Args[pos]) {
			if cmd.Args[pos].(FuzzyMatcher).Match(args[pos]) == false {
				return false
			}
		} else if reflect.DeepEqual(cmd.Args[pos], args[pos]) == false {
			return false
		}

	}
	return true
}

// Expect sets a response for this command. Everytime a Do or Receive methods are executed for a
// registered command this response or error will be returned. Expect call returns a pointer to Cmd struct,
// so you can chain Expect calls. Chained responses will be returned on subsequend calls matching this commands arguments
// in FIFO order.
func (c *Cmd) Expect(response interface{}) *Cmd {
	c.Responses = append(c.Responses, Response{response, nil})
	return c
}

// ExpectMap works in the same way of the Expect command, but has a key/value input to make it
// easier to build test environments
func (c *Cmd) ExpectMap(response map[string]string) *Cmd {
	var values []interface{}
	for key, value := range response {
		values = append(values, []byte(key))
		values = append(values, []byte(value))
	}
	c.Responses = append(c.Responses, Response{values, nil})
	return c
}

// ExpectError allows you to force an error when executing a command/arguments
func (c *Cmd) ExpectError(err error) *Cmd {
	c.Responses = append(c.Responses, Response{nil, err})
	return c
}
