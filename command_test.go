// Copyright 2014 Rafael Dantas Justo. All rights reserved.
// Use of this source code is governed by a GPL
// license that can be found in the LICENSE file.

package redigomock

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"github.com/gomodule/redigo/redis"
)

func TestCommand(t *testing.T) {
	connection := NewConn()
	connection.Command("HGETALL", "a", "b", "c")
	if len(connection.commands) != 1 {
		t.Fatalf("Did not register the command. Expected '1' and got '%d'", len(connection.commands))
	}

	cmd := connection.commands[0]

	if cmd.name != "HGETALL" {
		t.Error("Wrong name defined for command")
	}

	if len(cmd.args) != 3 {
		t.Fatal("Wrong arguments defined for command")
	}

	arg := cmd.args[0].(string)
	if arg != "a" {
		t.Errorf("Wrong argument defined for command. Expected 'a' and got '%s'", arg)
	}

	arg = cmd.args[1].(string)
	if arg != "b" {
		t.Errorf("Wrong argument defined for command. Expected 'b' and got '%s'", arg)
	}

	arg = cmd.args[2].(string)
	if arg != "c" {
		t.Errorf("Wrong argument defined for command. Expected 'c' and got '%s'", arg)
	}

	if len(cmd.responses) != 0 {
		t.Error("Response defined without any call")
	}
}

func TestScript(t *testing.T) {
	connection := NewConn()
	scriptData := []byte("This should be a lua script for redis")
	h := sha1.New()
	h.Write(scriptData)
	sha1sum := hex.EncodeToString(h.Sum(nil))

	connection.Script(scriptData, 0)                           // 0
	connection.Script(scriptData, 0, "value1")                 // 1
	connection.Script(scriptData, 1, "key1")                   // 2
	connection.Script(scriptData, 1, "key1", "value1")         // 3
	connection.Script(scriptData, 2, "key1", "key2", "value1") // 4

	if len(connection.commands) != 5 {
		t.Fatalf("Did not register the commands. Expected '5' and got '%d'", len(connection.commands))
	}

	if connection.commands[0].name != "EVALSHA" {
		t.Error("Wrong name defined for command")
	}

	if len(connection.commands[0].args) != 2 {
		t.Errorf("Wrong arguments defined for command %v", connection.commands[0].args)
	}

	if len(connection.commands[1].args) != 3 {
		t.Error("Wrong arguments defined for command")
	}

	if len(connection.commands[2].args) != 3 {
		t.Error("Wrong arguments defined for command")
	}

	if len(connection.commands[3].args) != 4 {
		t.Error("Wrong arguments defined for command")
	}

	if len(connection.commands[4].args) != 5 {
		t.Error("Wrong arguments defined for command")
	}

	// Script(scriptData, 0)
	arg := connection.commands[0].args[0].(string)
	if arg != sha1sum {
		t.Errorf("Wrong argument defined for command. Expected '%s' and got '%s'", sha1sum, arg)
	}
	argInt := connection.commands[0].args[1].(int)
	if argInt != 0 {
		t.Errorf("Wrong argument defined for command. Expected '0' and got '%v'", argInt)
	}

	// Script(scriptData, 0, "value1")
	argInt = connection.commands[1].args[1].(int)
	if argInt != 0 {
		t.Errorf("Wrong argument defined for command. Expected '0' and got '%v'", argInt)
	}
	arg = connection.commands[1].args[2].(string)
	if arg != "value1" {
		t.Errorf("Wrong argument defined for command. Expected 'value1' and got '%s'", arg)
	}

	// Script(scriptData, 1, "key1")
	argInt = connection.commands[2].args[1].(int)
	if argInt != 1 {
		t.Errorf("Wrong argument defined for command. Expected '1' and got '%v'", argInt)
	}
	arg = connection.commands[2].args[2].(string)
	if arg != "key1" {
		t.Errorf("Wrong argument defined for command. Expected 'key1' and got '%s'", arg)
	}

	// Script(scriptData, 1, "key1", "value1")
	argInt = connection.commands[3].args[1].(int)
	if argInt != 1 {
		t.Errorf("Wrong argument defined for command. Expected '1' and got '%v'", argInt)
	}
	arg = connection.commands[3].args[2].(string)
	if arg != "key1" {
		t.Errorf("Wrong argument defined for command. Expected 'key1' and got '%s'", arg)
	}
	arg = connection.commands[3].args[3].(string)
	if arg != "value1" {
		t.Errorf("Wrong argument defined for command. Expected 'value1' and got '%s'", arg)
	}

	// Script(scriptData, 2, "key1", "key2", "value1")
	argInt = connection.commands[4].args[1].(int)
	if argInt != 2 {
		t.Errorf("Wrong argument defined for command. Expected '2' and got '%v'", argInt)
	}
	arg = connection.commands[4].args[2].(string)
	if arg != "key1" {
		t.Errorf("Wrong argument defined for command. Expected 'key1' and got '%s'", arg)
	}
	arg = connection.commands[4].args[3].(string)
	if arg != "key2" {
		t.Errorf("Wrong argument defined for command. Expected 'key2' and got '%s'", arg)
	}
	arg = connection.commands[4].args[4].(string)
	if arg != "value1" {
		t.Errorf("Wrong argument defined for command. Expected 'value1' and got '%s'", arg)
	}
}

func TestGenericCommand(t *testing.T) {
	connection := NewConn()

	connection.GenericCommand("HGETALL")
	if len(connection.commands) != 1 {
		t.Fatalf("Did not registered the command. Expected '1' and got '%d'", len(connection.commands))
	}

	cmd := connection.commands[0]

	if cmd.name != "HGETALL" {
		t.Error("Wrong name defined for command")
	}

	if len(cmd.args) > 0 {
		t.Error("Arguments defined for command when they shouldn't")
	}

	if len(cmd.responses) != 0 {
		t.Error("Response defined without any call")
	}
}

func TestExpect(t *testing.T) {
	connection := NewConn()

	connection.Command("HGETALL").Expect("test")
	if len(connection.commands) != 1 {
		t.Fatalf("Did not registered the command. Expected '1' and got '%d'", len(connection.commands))
	}

	cmd := connection.commands[0]

	if cmd.responses[0].response == nil {
		t.Fatal("Response not defined")
	}

	value, ok := cmd.responses[0].response.(string)
	if !ok {
		t.Fatal("Not storing response in the correct type")
	}

	if value != "test" {
		t.Error("Wrong response content")
	}
}

func TestExpectMap(t *testing.T) {
	connection := NewConn()

	connection.Command("HGETALL").ExpectMap(map[string]string{
		"key1": "value1",
	})

	if len(connection.commands) != 1 {
		t.Fatalf("Did not registered the command. Expected '1' and got '%d'", len(connection.commands))
	}

	cmd := connection.commands[0]

	if cmd.responses[0].response == nil {
		t.Fatal("Response not defined")
	}

	values, ok := cmd.responses[0].response.([]interface{})
	if !ok {
		t.Fatal("Not storing response in the correct type")
	}

	expected := []string{"key1", "value1"}
	if len(values) != len(expected) {
		t.Fatal("Map values not stored properly")
	}

	for i := 0; i < len(expected); i++ {
		value, ok := values[i].([]byte)
		if ok {
			if string(value) != expected[i] {
				t.Errorf("Changing the response content. Expected '%s' and got '%s'",
					expected[i], string(value))
			}
		} else {
			t.Error("Not storing the map content in byte format")
		}
	}
}

func TestExpectMapReplace(t *testing.T) {
	connection := NewConn()

	connection.Command("HGETALL").ExpectMap(map[string]string{
		"key1": "value1",
	})

	connection.Command("HGETALL").ExpectMap(map[string]string{
		"key2": "value2",
	})

	if len(connection.commands) != 1 {
		t.Fatalf("Wrong number of registered commands. Expected '1' and got '%d'", len(connection.commands))
	}

	cmd := connection.commands[0]

	if cmd.responses[0].response == nil {
		t.Fatal("Response not defined")
	}

	values, ok := cmd.responses[0].response.([]interface{})
	if !ok {
		t.Fatal("Not storing response in the correct type")
	}

	expected := []string{"key2", "value2"}
	if len(values) != len(expected) {
		t.Fatal("Map values not stored properly")
	}

	for i := 0; i < len(expected); i++ {
		value, ok := values[i].([]byte)
		if ok {
			if string(value) != expected[i] {
				t.Errorf("Changing the response content. Expected '%s' and got '%s'",
					expected[i], string(value))
			}
		} else {
			t.Error("Not storing the map content in byte format")
		}
	}
}

func TestExpectError(t *testing.T) {
	connection := NewConn()

	connection.Command("HGETALL").ExpectError(fmt.Errorf("error"))

	if len(connection.commands) != 1 {
		t.Fatalf("Did not registered the command. Expected '1' and got '%d'", len(connection.commands))
	}

	cmd := connection.commands[0]

	if cmd.responses[0].err == nil {
		t.Fatal("Error not defined")
	}

	if cmd.responses[0].err.Error() != "error" {
		t.Fatal("Storing wrong error")
	}
}

func TestExpectPanic(t *testing.T) {
	connection := NewConn()

	connection.Command("HGETALL").ExpectPanic("panic")

	if len(connection.commands) != 1 {
		t.Fatalf("Did not registered the command. Expected '1' and got '%d'", len(connection.commands))
	}

	cmd := connection.commands[0]

	if cmd.responses[0].panicVal == nil {
		t.Fatal("Panic not defined")
	}

	if cmd.responses[0].panicVal != "panic" {
		t.Fatal("Storing wrong panic message")
	}
}

func TestExpectSlice(t *testing.T) {
	connection := NewConn()

	field1 := []byte("hello")
	connection.Command("HMGET", "key", "field1", "field2").ExpectSlice(field1, nil)
	if len(connection.commands) != 1 {
		t.Fatalf("Did not registered the command. Expected '1' and got '%d'", len(connection.commands))
	}

	reply, err := redis.ByteSlices(connection.Do("HMGET", "key", "field1", "field2"))
	if err != nil {
		t.Fatal(err)
	}

	if string(reply[0]) != string(field1) {
		t.Fatalf("reply[0] not hello but %s", string(reply[0]))
	}

	if reply[1] != nil {
		t.Fatal("reply[1] not nil")
	}
}

func TestExpectSliceFromStrings(t *testing.T) {
	connection := NewConn()

	field1 := "hello"
	field2 := "redigo"
	connection.Command("HMGET", "key", "field1", "field2").ExpectStringSlice(field1, field2)
	if len(connection.commands) != 1 {
		t.Fatalf("Did not registered the command. Expected '1' and got '%d'", len(connection.commands))
	}

	reply, err := redis.Strings(connection.Do("HMGET", "key", "field1", "field2"))
	if err != nil {
		t.Fatal(err)
	}

	if reply[0] != field1 {
		t.Fatalf("reply[0] not %s but %s", field1, reply[0])
	}

	if reply[1] != field2 {
		t.Fatalf("reply[1] not %s but %s", field2, reply[1])
	}
}

func TestFind(t *testing.T) {
	connection := NewConn()

	connection.Command("HGETALL", "a", "b", "c")

	if connection.find("HGETALL", []interface{}{"a"}) != nil {
		t.Error("Returning command without comparing all registered arguments")
	}

	if connection.find("HGETALL", []interface{}{"a", "b", "c", "d"}) != nil {
		t.Error("Returning command without comparing all informed arguments")
	}

	if connection.find("HSETALL", []interface{}{"a", "b", "c"}) != nil {
		t.Error("Returning command when the name is different")
	}

	if connection.find("HGETALL", []interface{}{"a", "b", "c"}) == nil {
		t.Error("Could not find command with arguments in the same order")
	}
}

func TestRemoveRelatedCommands(t *testing.T) {
	connection := NewConn()

	connection.Command("HGETALL", "a", "b", "c") // 1
	connection.Command("HGETALL", "a", "b", "c") // omit
	connection.Command("HGETALL", "c", "b", "a") // 2
	connection.Command("HGETALL")                // 3
	connection.Command("HSETALL", "c", "b", "a") // 4
	connection.Command("HSETALL")                // 5

	if len(connection.commands) != 5 {
		t.Errorf("Not removing related commands. Expected '5' and got '%d'", len(connection.commands))
	}
}

func TestMatch(t *testing.T) {
	data := []struct {
		cmd         *Cmd
		commandName string
		args        []interface{}
		equal       bool
	}{
		{
			cmd:         &Cmd{name: "HGETALL", args: []interface{}{"a", "b", "c"}},
			commandName: "HGETALL",
			args:        []interface{}{"a", "b", "c"},
			equal:       true,
		},
		{
			cmd:         &Cmd{name: "HGETALL", args: []interface{}{"a", []byte("abcdef"), "c"}},
			commandName: "HGETALL",
			args:        []interface{}{"a", []byte("abcdef"), "c"},
			equal:       true,
		},
		{
			cmd:         &Cmd{name: "HGETALL", args: []interface{}{"a", "b", "c"}},
			commandName: "HGETALL",
			args:        []interface{}{"c", "b", "a"},
			equal:       false,
		},
		{
			cmd:         &Cmd{name: "HGETALL", args: []interface{}{"a", "b", "c"}},
			commandName: "HGETALL",
			args:        []interface{}{"a", "b"},
			equal:       false,
		},
		{
			cmd:         &Cmd{name: "HGETALL", args: []interface{}{"a", "b"}},
			commandName: "HGETALL",
			args:        []interface{}{"a", "b", "c"},
			equal:       false,
		},
		{
			cmd:         &Cmd{name: "HGETALL", args: []interface{}{"a", "b", "c"}},
			commandName: "HSETALL",
			args:        []interface{}{"a", "b", "c"},
			equal:       false,
		},
		{
			cmd:         &Cmd{name: "HSETALL", args: nil},
			commandName: "HSETALL",
			args:        nil,
			equal:       true,
		},
	}

	for i, item := range data {
		e := match(item.commandName, item.args, item.cmd)
		if e != item.equal && item.equal {
			t.Errorf("Expected commands to be equal for data item '%d'", i)
		} else if e != item.equal && !item.equal {
			t.Errorf("Expected commands to be different for data item '%d'", i)
		}
	}
}

func TestHash(t *testing.T) {
	data := []struct {
		cmd      *Cmd
		expected cmdHash
	}{
		{
			cmd:      &Cmd{name: "HGETALL", args: []interface{}{"a", "b", "c"}},
			expected: cmdHash("HGETALLabc"),
		},
		{
			cmd:      &Cmd{name: "HGETALL", args: []interface{}{"a", []byte("abcdef"), "c"}},
			expected: "HGETALLa[97 98 99 100 101 102]c",
		},
	}

	for i, item := range data {
		if hash := item.cmd.hash(); hash != item.expected {
			t.Errorf("Expected “%s” and got “%s” for data item “%d”", item.expected, hash, i)
		}
	}
}

func TestRace(t *testing.T) {
	funcs := []func(*Cmd){
		func(c *Cmd) { _ = equal("GET", []interface{}{[]byte("hello")}, c) },
		func(c *Cmd) { _ = match("GET", []interface{}{[]byte("hello")}, c) },
		func(c *Cmd) { _ = c.hash() },
		func(c *Cmd) { _ = c.Called() },
		func(c *Cmd) { _ = c.getResponse() },
		func(c *Cmd) { c.Expect([]byte("OK")) },
		func(c *Cmd) { c.ExpectMap(map[string]string{"hello": "world"}) },
		func(c *Cmd) { c.ExpectError(fmt.Errorf("oh no")) },
		func(c *Cmd) { c.ExpectPanic(fmt.Errorf("oh no")) },
		func(c *Cmd) { c.ExpectSlice([]byte("hello"), []byte("world")) },
		func(c *Cmd) { c.ExpectStringSlice("hello", "world") },
		func(c *Cmd) { c.Handle(func(args []interface{}) (interface{}, error) { return nil, nil }) },
	}

	// include two copies of each function, in case a function races with
	// itself
	funcs = append(funcs, funcs...)

	// run the test several times, with shuffled order, to make sure it's not
	// too order-dependent
	rand := rand.New(rand.NewSource(0))
	for i := 0; i < 10; i++ {
		rand.Shuffle(len(funcs), func(i, j int) { funcs[i], funcs[j] = funcs[j], funcs[i] })

		cmd := Cmd{name: "GET", args: []interface{}{[]byte("hello")}}

		var wg sync.WaitGroup
		wg.Add(len(funcs))
		for _, f := range funcs {
			f := f
			go func() {
				f(&cmd)
				wg.Done()
			}()
		}

		wg.Wait()

		// we should have 12 to 14 responses, depending on how many of the
		// getResponse calls ran before at least two Expects, since getResponse
		// pops a response only if there are at least two responses

		l := len(cmd.responses)
		if l < 12 || l > 14 {
			t.Errorf("wanted 12-14 responses, got %v: %v", l, cmd.responses)
		}
	}
}
