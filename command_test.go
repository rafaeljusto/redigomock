// Copyright 2014 Rafael Dantas Justo. All rights reserved.
// Use of this source code is governed by a GPL
// license that can be found in the LICENSE file.

package redigomock

import (
	"fmt"
	"testing"
)

func TestCommand(t *testing.T) {
	commands = []*Cmd{}

	Command("HGETALL", "a", "b", "c")
	if len(commands) != 1 {
		t.Fatalf("Did not registered the command. Expected '1' and got '%d'", len(commands))
	}

	cmd := commands[0]

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

	if cmd.Response != nil {
		t.Error("Response defined without any call")
	}

	if cmd.Error != nil {
		t.Error("Error defined without any call")
	}
}

func TestGenericCommand(t *testing.T) {
	commands = []*Cmd{}

	GenericCommand("HGETALL")
	if len(commands) != 1 {
		t.Fatalf("Did not registered the command. Expected '1' and got '%d'", len(commands))
	}

	cmd := commands[0]

	if cmd.Name != "HGETALL" {
		t.Error("Wrong name defined for command")
	}

	if len(cmd.Args) > 0 {
		t.Error("Arguments defined for command when they shouldn't")
	}

	if cmd.Response != nil {
		t.Error("Response defined without any call")
	}

	if cmd.Error != nil {
		t.Error("Error defined without any call")
	}
}

func TestExpect(t *testing.T) {
	commands = []*Cmd{}

	Command("HGETALL").Expect("test")
	if len(commands) != 1 {
		t.Fatalf("Did not registered the command. Expected '1' and got '%d'", len(commands))
	}

	cmd := commands[0]

	if cmd.Response == nil {
		t.Fatal("Response not defined")
	}

	value, ok := cmd.Response.(string)
	if !ok {
		t.Fatal("Not storing response in the correct type")
	}

	if value != "test" {
		t.Error("Wrong response content")
	}
}

func TestExpectMap(t *testing.T) {
	commands = []*Cmd{}

	Command("HGETALL").ExpectMap(map[string]string{
		"key1": "value1",
	})

	if len(commands) != 1 {
		t.Fatalf("Did not registered the command. Expected '1' and got '%d'", len(commands))
	}

	cmd := commands[0]

	if cmd.Response == nil {
		t.Fatal("Response not defined")
	}

	values, ok := cmd.Response.([]interface{})
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
	commands = []*Cmd{}

	Command("HGETALL").ExpectMap(map[string]string{
		"key1": "value1",
	})

	Command("HGETALL").ExpectMap(map[string]string{
		"key2": "value2",
	})

	if len(commands) != 1 {
		t.Fatalf("Wrong number of registered commands. Expected '1' and got '%d'", len(commands))
	}

	cmd := commands[0]

	if cmd.Response == nil {
		t.Fatal("Response not defined")
	}

	values, ok := cmd.Response.([]interface{})
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
	commands = []*Cmd{}

	Command("HGETALL").ExpectError(fmt.Errorf("error"))

	if len(commands) != 1 {
		t.Fatalf("Did not registered the command. Expected '1' and got '%d'", len(commands))
	}

	cmd := commands[0]

	if cmd.Error == nil {
		t.Fatal("Error not defined")
	}

	if cmd.Error.Error() != "error" {
		t.Fatal("Storing wrong error")
	}
}

func TestFind(t *testing.T) {
	commands = []*Cmd{}

	Command("HGETALL", "a", "b", "c")

	if find("HGETALL", []interface{}{"a"}) != nil {
		t.Error("Returning command without comparing all registered arguments")
	}

	if find("HGETALL", []interface{}{"a", "b", "c", "d"}) != nil {
		t.Error("Returning command without comparing all informed arguments")
	}

	if find("HSETALL", []interface{}{"a", "b", "c"}) != nil {
		t.Error("Returning command when the name is different")
	}

	if find("HGETALL", []interface{}{"c", "b", "a"}) == nil {
		t.Error("Could not find command with arguments in a different order")
	}

	if find("HGETALL", []interface{}{"a", "b", "c"}) == nil {
		t.Error("Could not find command with arguments in the same order")
	}
}

func TestRemoveRelatedCommands(t *testing.T) {
	commands = []*Cmd{}

	Command("HGETALL", "a", "b", "c")
	Command("HGETALL", "a", "b", "c")
	Command("HGETALL", "c", "b", "a")
	Command("HGETALL")
	Command("HSETALL", "c", "b", "a")
	Command("HSETALL")

	if len(commands) != 4 {
		t.Errorf("Not removing related commands. Expected '4' and got '%d'", len(commands))
	}
}

func TestEqual(t *testing.T) {
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
			Cmd:         &Cmd{Name: "HGETALL", Args: []interface{}{"a", "b", "c"}},
			CommandName: "HGETALL",
			Args:        []interface{}{"c", "b", "a"},
			Equal:       true,
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
		e := equal(item.CommandName, item.Args, item.Cmd)
		if e != item.Equal && item.Equal {
			t.Errorf("Expected commands to be equal for data item '%d'", i)

		} else if e != item.Equal && !item.Equal {
			t.Errorf("Expected commands to be different for data item '%d'", i)
		}
	}
}
