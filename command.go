// Copyright 2014 Rafael Dantas Justo. All rights reserved.
// Use of this source code is governed by a GPL
// license that can be found in the LICENSE file.

package redigomock

var (
	commands []*Cmd // global variable that store all registered commands
)

// Cmd stores the registered information about a command to return it later when request by a
// command execution
type Cmd struct {
	Name     string        // Name of the command
	Args     []interface{} // Arguments of the command
	Response interface{}   // Response to send back when this command/arguments are called
	Error    error         // Error to send back when this command/arguments are called
}

// Command register a command in the mock system using the same arguments of a Do or Send commands.
// It will return a registered command object where you can set the response or error
func Command(commandName string, args ...interface{}) *Cmd {
	cmd := &Cmd{
		Name: commandName,
		Args: args,
	}

	removeRelatedCommands(commandName, args)
	commands = append(commands, cmd)
	return cmd
}

// GenericCommand register a command without arguments. If a command with arguments doesn't match
// with any registered command, it will look for generic commands before throwing an error
func GenericCommand(commandName string) *Cmd {
	cmd := &Cmd{
		Name: commandName,
	}

	removeRelatedCommands(commandName, nil)
	commands = append(commands, cmd)
	return cmd
}

// Expect sets a response for this command. Everytime a Do or Receive methods are executed for a
// registered command this response or error will be returned. You cannot set a response and a error
// for the same command/arguments
func (c *Cmd) Expect(response interface{}) {
	c.Response = response
	c.Error = nil
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
	c.Error = nil
}

// ExpectError allows you to force an error when executing a command/arguments
func (c *Cmd) ExpectError(err error) {
	c.Response = nil
	c.Error = err
}

// find will scan the registered commands, looking for the first command with the same name and
// arguments. If the command is not found nil is returned
func find(commandName string, args []interface{}) *Cmd {
	for _, cmd := range commands {
		if cmd.Name != commandName || len(cmd.Args) != len(args) {
			continue
		}

		equal := true
		for i := range cmd.Args {
			found := false

			// Allow arguments in different order
			for j := range args {
				if cmd.Args[i] == args[j] {
					found = true
					break
				}
			}

			if !found {
				equal = false
				break
			}
		}

		if equal {
			return cmd
		}
	}

	return nil
}

// removeRelatedCommands verify if a command is already registered, removing any command already
// registered with the same name and arguments. This should avoid duplicated mocked commands
func removeRelatedCommands(commandName string, args []interface{}) {
	var unique []*Cmd

	for _, cmd := range commands {
		if cmd.Name != commandName || len(cmd.Args) != len(args) {
			unique = append(unique, cmd)
			continue
		}

		equal := true
		for i := range cmd.Args {
			found := false

			// Allow arguments in different order
			for j := range args {
				if cmd.Args[i] == args[j] {
					found = true
					break
				}
			}

			if !found {
				equal = false
				break
			}
		}

		if !equal {
			unique = append(unique, cmd)
		}
	}

	commands = unique
}
