package redigomock

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"testing"
)

type Person struct {
	Name string `redis:"name"`
	Age  int    `redis:"age"`
}

func RetrievePerson(conn redis.Conn, id string) (Person, error) {
	var person Person

	values, err := redis.Values(conn.Do("HGETALL", fmt.Sprintf("person:%s", id)))
	if err != nil {
		return person, err
	}

	err = redis.ScanStruct(values, &person)
	return person, err
}

func RetrievePeople(conn redis.Conn, ids []string) ([]Person, error) {
	var people []Person

	for _, id := range ids {
		conn.Send("HGETALL", fmt.Sprintf("person:%s", id))
	}

	for i := 0; i < len(ids); i++ {
		values, err := redis.Values(conn.Receive())
		if err != nil {
			return nil, err
		}

		var person Person
		err = redis.ScanStruct(values, &person)
		if err != nil {
			return nil, err
		}

		people = append(people, person)
	}

	return people, nil
}

func TestDoCommand(t *testing.T) {
	commands = map[string]*Cmd{}

	Command("HGETALL", "person:1").ExpectMap(map[string]string{
		"name": "Mr. Johson",
		"age":  "42",
	})

	person, err := RetrievePerson(NewConn(), "1")
	if err != nil {
		t.Fatal(err)
	}

	if person.Name != "Mr. Johson" {
		t.Errorf("Invalid name. Expected 'Mr. Johson' and got '%s'", person.Name)
	}

	if person.Age != 42 {
		t.Errorf("Invalid age. Expected '42' and got '%d'")
	}
}

func TestDoGenericCommand(t *testing.T) {
	commands = map[string]*Cmd{}

	GenericCommand("HGETALL").ExpectMap(map[string]string{
		"name": "Mr. Johson",
		"age":  "42",
	})

	person, err := RetrievePerson(NewConn(), "1")
	if err != nil {
		t.Fatal(err)
	}

	if person.Name != "Mr. Johson" {
		t.Errorf("Invalid name. Expected 'Mr. Johson' and got '%s'", person.Name)
	}

	if person.Age != 42 {
		t.Errorf("Invalid age. Expected '42' and got '%d'")
	}
}

func TestDoCommandWithGeneric(t *testing.T) {
	commands = map[string]*Cmd{}

	Command("HGETALL", "person:1").ExpectMap(map[string]string{
		"name": "Mr. Johson",
		"age":  "42",
	})

	GenericCommand("HGETALL").ExpectMap(map[string]string{
		"name": "Mr. Mark",
		"age":  "32",
	})

	person, err := RetrievePerson(NewConn(), "1")
	if err != nil {
		t.Fatal(err)
	}

	if person.Name != "Mr. Johson" {
		t.Errorf("Invalid name. Expected 'Mr. Johson' and got '%s'", person.Name)
	}

	if person.Age != 42 {
		t.Errorf("Invalid age. Expected '42' and got '%d'")
	}
}

func TestDoCommandWithError(t *testing.T) {
	commands = map[string]*Cmd{}

	Command("HGETALL", "person:1").ExpectError(fmt.Errorf("simulated error"))

	_, err := RetrievePerson(NewConn(), "1")
	if err == nil {
		t.Error("Should return an error!")
		return
	}
}

func TestDoCommandWithUnexpectedCommand(t *testing.T) {
	commands = map[string]*Cmd{}

	_, err := RetrievePerson(NewConn(), "X")
	if err == nil {
		t.Error("Should detect a command not registered!")
		return
	}
}

func TestSendFlushReceive(t *testing.T) {
	commands = map[string]*Cmd{}

	Command("HGETALL", "person:1").ExpectMap(map[string]string{
		"name": "Mr. Johson",
		"age":  "42",
	})
	Command("HGETALL", "person:2").ExpectMap(map[string]string{
		"name": "Ms. Jennifer",
		"age":  "28",
	})

	conn := NewConn()
	people, err := RetrievePeople(conn, []string{"1", "2"})
	if err != nil {
		t.Fatal(err)
	}

	if len(people) != 2 {
		t.Errorf("Wrong number of people. Expected '2' and got '%d'", len(people))
	}

	if people[0].Name != "Mr. Johson" || people[1].Name != "Ms. Jennifer" {
		t.Error("People name order are wrong")
	}

	if people[0].Age != 42 || people[1].Age != 28 {
		t.Error("People age order are wrong")
	}

	if _, err := conn.Receive(); err == nil {
		t.Error("Not detecting when there's no more items to receive")
	}
}

func TestSendFlushReceiveWithError(t *testing.T) {
	commands = map[string]*Cmd{}

	Command("HGETALL", "person:1").ExpectMap(map[string]string{
		"name": "Mr. Johson",
		"age":  "42",
	})
	Command("HGETALL", "person:2").ExpectMap(map[string]string{
		"name": "Ms. Jennifer",
		"age":  "28",
	})
	Command("HGETALL", "person:2").ExpectError(fmt.Errorf("simulated error"))

	_, err := RetrievePeople(NewConn(), []string{"1", "2", "3"})
	if err == nil {
		t.Error("Not detecting error when using send/flush/receive")
	}
}

func TestDummyFunctions(t *testing.T) {
	conn := NewConn()

	if conn.Close() != nil {
		t.Error("Close is not dummy!")
	}

	if conn.Err() != nil {
		t.Error("Err is not dummy!")
	}

	if conn.Flush() != nil {
		t.Error("Flush is not dummy!")
	}
}

func TestClear(t *testing.T) {
	commands = map[string]*Cmd{}

	Command("HGETALL", "person:1").ExpectMap(map[string]string{
		"name": "Mr. Johson",
		"age":  "42",
	})
	Command("HGETALL", "person:2").ExpectMap(map[string]string{
		"name": "Ms. Jennifer",
		"age":  "28",
	})
	GenericCommand("HGETALL").ExpectMap(map[string]string{
		"name": "Ms. Mark",
		"age":  "32",
	})

	conn := NewConn()
	conn.Send("HGETALL", "person:1")
	conn.Send("HGETALL", "person:2")

	Clear()

	if len(commands) > 0 {
		t.Error("Clear function not clearing registered commands")
	}

	if len(queue) > 0 {
		t.Error("Clear function not clearing the queue")
	}
}
