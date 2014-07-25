// Copyright 2014 Rafael Dantas Justo. All rights reserved.
// Use of this source code is governed by a GPL
// license that can be found in the LICENSE file.

// Package redigomock is a mock for redigo library (redis client)
//
// Redigomock basically register the commands with the expected results in a internal global
// variable. When the command is executed via Conn interface, the mock will look to this global
// variable to retrieve the corresponding result.
//
// To start a mocked connection just do the following:
//  c := redigomock.NewConn()
//
// Now you can inject it whenever your system needs a redigo.Conn because it satisfies all interface
// requirements. Before running your tests you need beyond of mocking the connection, registering
// the expected results. For that you can generate commands with the expected results.
//  redigomock.Command("HGETALL", "person:1").Expect("Person!")
//  redigomock.Command(
//    "HMSET", []string{"person:1", "name", "John"},
//  ).Expect("ok")
//
// As the Expect method from Command receives anything (interface{}), another method was created to
// easy map the result to your structure. For that use ExpectMap:
//  redigomock.Command("HGETALL", "person:1").ExpectMap(map[string]string{
//    "name": "John",
//    "age": 42,
//  })
//
// You should also test the error cases, and you can do it in the same way of a normal result.
//  redigomock.Command("HGETALL", "person:1").ExpectError(fmt.Errorf("Low level error!"))
//
// Sometimes you will want to register a command regardless the arguments, and you can do it with
// the method GenericCommand (mainly with the HMSET).
//  redigomock.GenericCommand("HMSET").Expect("ok")
//
// All commands are registered in a global variable, so they will be there until all your test cases
// ends. So for good practice in test writing you should in the beginning of each test case clear
// the mock states.
//  redigomock.Clear()
//
// Let's see a full test example. Imagine a Person structure and a function that pick up this
// person in Redis using redigo library (file person.go):
//  package person
//
//  import (
// 	 "fmt"
//  	"github.com/garyburd/redigo/redis"
//  )
//
//  type Person struct {
//  	Name string `redis:"name"`
//  	Age  int    `redis:"age"`
//  }
//
//  func RetrievePerson(conn redis.Conn, id string) (Person, error) {
//  	var person Person
//
//  	values, err := redis.Values(conn.Do("HGETALL", fmt.Sprintf("person:%s", id)))
//  	if err != nil {
//  		return person, err
//  	}
//
//  	err = redis.ScanStruct(values, &person)
//  	return person, err
//  }
//
// Now we need to test it, so let's create the corresponding test with redigomock
// (fileperson_test.go):
//  package person
//
//  import (
//    "github.com/rafaeljusto/redigomock"
//    "testing"
//  )
//
//  func TestRetrievePerson(t *testing.T) {
//    redigomock.Clear()
//	  redigomock.Command("HGETALL", "person:1").ExpectMap(map[string]string{
// 	    "name": "Mr. Johson",
//      "age":  "42",
//    })
//
//    person, err := RetrievePerson(redigomock.NewConn(), "1")
//    if err != nil {
// 	    t.Fatal(err)
//    }
//
//    if person.Name != "Mr. Johson" {
// 	    t.Errorf("Invalid name. Expected 'Mr. Johson' and got '%s'", person.Name)
//    }
//
//    if person.Age != 42 {
// 	    t.Errorf("Invalid age. Expected '42' and got '%d'", person.Age)
//    }
//  }
//
//  func TestRetrievePersonError(t *testing.T) {
//    redigomock.Clear()
//    redigomock.Command("HGETALL", "person:1").ExpectError(fmt.Errorf("Simulate error!"))
//
//    person, err = RetrievePerson(redigomock.NewConn(), "1")
//    if err == nil {
// 	    t.Error("Should return an error!")
//    }
//  }
package redigomock
