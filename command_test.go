// Copyright 2014 Rafael Dantas Justo. All rights reserved.
// Use of this source code is governed by a GPL
// license that can be found in the LICENSE file.

package redigomock

import (
	"fmt"
	"testing"
)

func TestCommand(t *testing.T) {
	Command("HGETALL")
	if len(commands) != 1 {
		t.Fatal("Did not registered the command")
	}

	cmd, exists := commands["HGETALL"]
	if !exists {
		t.Fatal("Wrong key defined for command")
	}

	if cmd.Response != nil {
		t.Error("Response defined without any call")
	}

	if cmd.Error != nil {
		t.Error("Error defined without any call")
	}
}

func TestExpect(t *testing.T) {
	Command("HGETALL").Expect("test")
	cmd, exists := commands["HGETALL"]
	if !exists {
		t.Fatal("Wrong key defined for command")
	}

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
	Command("HGETALL").ExpectMap(map[string]string{
		"key1": "value1",
	})

	cmd, exists := commands["HGETALL"]
	if !exists {
		t.Fatal("Wrong key defined for command")
	}

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

func TestExpectError(t *testing.T) {
	Command("HGETALL").ExpectError(fmt.Errorf("Error!"))

	cmd, exists := commands["HGETALL"]
	if !exists {
		t.Fatal("Wrong key defined for command")
	}

	if cmd.Error == nil {
		t.Fatal("Error not defined")
	}

	if cmd.Error.Error() != "Error!" {
		t.Fatal("Storing wrong error")
	}
}

func TestGenerateKey(t *testing.T) {
	data := []struct {
		CommandName string
		Args        []interface{}
		Expected    string
	}{
		{
			CommandName: "  hgetall  ",
			Args:        []interface{}{"  A  ", " B", "C "},
			Expected:    "HGETALL A B C",
		},
		{
			CommandName: "HGETALL",
			Args:        []interface{}{[]byte("A"), []byte("B"), []byte("C")},
			Expected:    "HGETALL A B C",
		},
		{
			CommandName: "HGETALL",
			Args:        []interface{}{1, 2, 3},
			Expected:    "HGETALL 1 2 3",
		},
		{
			CommandName: "HGETALL",
			Args:        []interface{}{int64(1), int64(2), int64(3)},
			Expected:    "HGETALL 1 2 3",
		},
		{
			CommandName: "HGETALL",
			Args:        []interface{}{1.1, 2.2, 3.3},
			Expected:    "HGETALL 1.1 2.2 3.3",
		},
		{
			CommandName: "HGETALL",
			Args:        []interface{}{true, false},
			Expected:    "HGETALL 1 0",
		},
		{
			CommandName: "HGETALL",
			Args:        []interface{}{nil},
			Expected:    "HGETALL ",
		},
		{
			CommandName: "HGETALL",
			Args: []interface{}{struct {
				Field string
			}{
				Field: "test",
			}},
			Expected: "HGETALL {TEST}",
		},
	}

	for _, item := range data {
		key := generateKey(item.CommandName, item.Args)
		if key != item.Expected {
			t.Errorf("Error in key generation. Expected '%s' and got '%s'", item.Expected, key)
		}
	}
}
