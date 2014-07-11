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

type Cmd struct {
	CommandName string
	Args        []interface{}
	Response    interface{}
	Err         error
}

func Command(commandName string, args ...interface{}) *Cmd {
	cmd := Cmd{
		CommandName: commandName,
		Args:        args,
	}

	commands[generateKey(commandName, args)] = &cmd
	return &cmd
}

func (c *Cmd) Expect(response interface{}) {
	c.Response = response
	c.Err = nil
}

func (c *Cmd) ExpectMap(response map[string]string) {
	var values []interface{}
	for key, value := range response {
		values = append(values, []byte(key))
		values = append(values, []byte(value))
	}

	c.Response = values
	c.Err = nil
}

func (c *Cmd) Error(err error) {
	c.Response = nil
	c.Err = err
}

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
