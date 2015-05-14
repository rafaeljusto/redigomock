// Copyright 2014 Rafael Dantas Justo. All rights reserved.
// Use of this source code is governed by a GPL
// license that can be found in the LICENSE file.

package redigomock

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"testing"
)

func TestCommand(t *testing.T) {
	connection := NewConn()
	connection.Command("HGETALL", "a", "b", "c")
	if len(connection.commands) != 1 {
		t.Fatalf("Did not register the command. Expected '1' and got '%d'", len(connection.commands))
	}

	cmd := connection.commands[0]

	if cmd.Name != "HGETALL" {
		t.Error("Wrong name defined for command")
	}

	if len(cmd.Args) != 3 {
		t.Fatal("Wrong arguments defined for command")
	}

	arg := cmd.Args[0].(string)
	if arg != "a" {
		t.Errorf("Wrong argument defined for command. Expected 'a' and got '%s'", arg)
	}

	arg = cmd.Args[1].(string)
	if arg != "b" {
		t.Errorf("Wrong argument defined for command. Expected 'b' and got '%s'", arg)
	}

	arg = cmd.Args[2].(string)
	if arg != "c" {
		t.Errorf("Wrong argument defined for command. Expected 'c' and got '%s'", arg)
	}

	if len(cmd.Responses) != 0 {
		t.Error("Response defined without any call")
	}
}

func TestScript(t *testing.T) {
	connection := NewConn()
	scriptData := []byte("This should be a lua script for redis")
	h := sha1.New()
	h.Write(scriptData)
	sha1sum := hex.EncodeToString(h.Sum(nil))

	connection.Script(scriptData, 0)                           //0
	connection.Script(scriptData, 0, "value1")                 //1
	connection.Script(scriptData, 1, "key1")                   //2
	connection.Script(scriptData, 1, "key1", "value1")         //3
	connection.Script(scriptData, 2, "key1", "key2", "value1") //4

	if len(connection.commands) != 5 {
		t.Fatalf("Did not register the commands. Expected '5' and got '%d'", len(connection.commands))
	}

	if connection.commands[0].Name != "EVALSHA" {
		t.Error("Wrong name defined for command")
	}

	if len(connection.commands[0].Args) != 2 {
		t.Errorf("Wrong arguments defined for command %v", connection.commands[0].Args)
	}

	if len(connection.commands[1].Args) != 3 {
		t.Error("Wrong arguments defined for command")
	}

	if len(connection.commands[2].Args) != 3 {
		t.Error("Wrong arguments defined for command")
	}

	if len(connection.commands[3].Args) != 4 {
		t.Error("Wrong arguments defined for command")
	}

	if len(connection.commands[4].Args) != 5 {
		t.Error("Wrong arguments defined for command")
	}

	//Script(scriptData, 0)
	arg := connection.commands[0].Args[0].(string)
	if arg != sha1sum {
		t.Errorf("Wrong argument defined for command. Expected '%s' and got '%s'", sha1sum, arg)
	}
	argInt := connection.commands[0].Args[1].(int)
	if argInt != 0 {
		t.Errorf("Wrong argument defined for command. Expected '0' and got '%v'", argInt)
	}

	//Script(scriptData, 0, "value1")
	argInt = connection.commands[1].Args[1].(int)
	if argInt != 0 {
		t.Errorf("Wrong argument defined for command. Expected '0' and got '%v'", argInt)
	}
	arg = connection.commands[1].Args[2].(string)
	if arg != "value1" {
		t.Errorf("Wrong argument defined for command. Expected 'value1' and got '%s'", arg)
	}

	//Script(scriptData, 1, "key1")
	argInt = connection.commands[2].Args[1].(int)
	if argInt != 1 {
		t.Errorf("Wrong argument defined for command. Expected '1' and got '%v'", argInt)
	}
	arg = connection.commands[2].Args[2].(string)
	if arg != "key1" {
		t.Errorf("Wrong argument defined for command. Expected 'key1' and got '%s'", arg)
	}

	//Script(scriptData, 1, "key1", "value1")
	argInt = connection.commands[3].Args[1].(int)
	if argInt != 1 {
		t.Errorf("Wrong argument defined for command. Expected '1' and got '%v'", argInt)
	}
	arg = connection.commands[3].Args[2].(string)
	if arg != "key1" {
		t.Errorf("Wrong argument defined for command. Expected 'key1' and got '%s'", arg)
	}
	arg = connection.commands[3].Args[3].(string)
	if arg != "value1" {
		t.Errorf("Wrong argument defined for command. Expected 'value1' and got '%s'", arg)
	}

	//Script(scriptData, 2, "key1", "key2", "value1")
	argInt = connection.commands[4].Args[1].(int)
	if argInt != 2 {
		t.Errorf("Wrong argument defined for command. Expected '2' and got '%v'", argInt)
	}
	arg = connection.commands[4].Args[2].(string)
	if arg != "key1" {
		t.Errorf("Wrong argument defined for command. Expected 'key1' and got '%s'", arg)
	}
	arg = connection.commands[4].Args[3].(string)
	if arg != "key2" {
		t.Errorf("Wrong argument defined for command. Expected 'key2' and got '%s'", arg)
	}
	arg = connection.commands[4].Args[4].(string)
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

	if cmd.Name != "HGETALL" {
		t.Error("Wrong name defined for command")
	}

	if len(cmd.Args) > 0 {
		t.Error("Arguments defined for command when they shouldn't")
	}

	if len(cmd.Responses) != 0 {
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

	if cmd.Responses[0].Response == nil {
		t.Fatal("Response not defined")
	}

	value, ok := cmd.Responses[0].Response.(string)
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

	if cmd.Responses[0].Response == nil {
		t.Fatal("Response not defined")
	}

	values, ok := cmd.Responses[0].Response.([]interface{})
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

	if cmd.Responses[0].Response == nil {
		t.Fatal("Response not defined")
	}

	values, ok := cmd.Responses[0].Response.([]interface{})
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

	if cmd.Responses[0].Error == nil {
		t.Fatal("Error not defined")
	}

	if cmd.Responses[0].Error.Error() != "error" {
		t.Fatal("Storing wrong error")
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
	connection.Command("HGETALL")                //3
	connection.Command("HSETALL", "c", "b", "a") // 4
	connection.Command("HSETALL")                //5

	if len(connection.commands) != 5 {
		t.Errorf("Not removing related commands. Expected '5' and got '%d'", len(connection.commands))
	}
}

func TestMatch(t *testing.T) {
	data := []struct {
		Cmd         *Cmd
		CommandName string
		Args        []interface{}
		Equal       bool
	}{
		{
			Cmd:         &Cmd{Name: "HGETALL", Args: []interface{}{"a", "b", "c"}},
			CommandName: "HGETALL",
			Args:        []interface{}{"a", "b", "c"},
			Equal:       true,
		},
		{
			Cmd:         &Cmd{Name: "HGETALL", Args: []interface{}{"a", []byte("abcdef"), "c"}},
			CommandName: "HGETALL",
			Args:        []interface{}{"a", []byte("abcdef"), "c"},
			Equal:       true,
		},
		{
			Cmd:         &Cmd{Name: "HGETALL", Args: []interface{}{"a", "b", "c"}},
			CommandName: "HGETALL",
			Args:        []interface{}{"c", "b", "a"},
			Equal:       false,
		},
		{
			Cmd:         &Cmd{Name: "HGETALL", Args: []interface{}{"a", "b", "c"}},
			CommandName: "HGETALL",
			Args:        []interface{}{"a", "b"},
			Equal:       false,
		},
		{
			Cmd:         &Cmd{Name: "HGETALL", Args: []interface{}{"a", "b"}},
			CommandName: "HGETALL",
			Args:        []interface{}{"a", "b", "c"},
			Equal:       false,
		},
		{
			Cmd:         &Cmd{Name: "HGETALL", Args: []interface{}{"a", "b", "c"}},
			CommandName: "HSETALL",
			Args:        []interface{}{"a", "b", "c"},
			Equal:       false,
		},
		{
			Cmd:         &Cmd{Name: "HSETALL", Args: nil},
			CommandName: "HSETALL",
			Args:        nil,
			Equal:       true,
		},
	}

	for i, item := range data {
		e := match(item.CommandName, item.Args, item.Cmd)
		if e != item.Equal && item.Equal {
			t.Errorf("Expected commands to be equal for data item '%d'", i)

		} else if e != item.Equal && !item.Equal {
			t.Errorf("Expected commands to be different for data item '%d'", i)
		}
	}
}
