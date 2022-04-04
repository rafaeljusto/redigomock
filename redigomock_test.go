package redigomock

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
)

var (
	_ redis.Conn            = &Conn{}
	_ redis.ConnWithTimeout = &Conn{}
	_ redis.ConnWithContext = &Conn{}
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
	connection := NewConn()

	connection.Command("HGETALL", "person:1").ExpectMap(map[string]string{
		"name": "Mr. Johson",
		"age":  "42",
	})

	person, err := RetrievePerson(connection, "1")
	if err != nil {
		t.Fatal(err)
	}

	if person.Name != "Mr. Johson" {
		t.Errorf("Invalid name. Expected 'Mr. Johson' and got '%s'", person.Name)
	}

	if person.Age != 42 {
		t.Errorf("Invalid age. Expected '42' and got '%d'", person.Age)
	}
}

func TestPanickyDoCommand(t *testing.T) {
	panicMsg := "panic-message"

	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("Expected a panic to happen")
		}
		recoverMsg, ok := r.(string)
		if ok == false {
			t.Errorf("Expected the recovered panic value to be a string")
		}
		if recoverMsg != panicMsg {
			t.Errorf("Expected the recovered panic value to be %s", panicMsg)
		}
	}()

	connection := NewConn()

	connection.Command("HGETALL", "person:1").ExpectPanic(panicMsg)

	RetrievePerson(connection, "1")
}

func TestDoCommandMultipleReturnValues(t *testing.T) {
	connection := NewConn()

	connection.Command("HGETALL", "person:1").ExpectMap(map[string]string{
		"name": "Mr. Johson",
		"age":  "42",
	}).ExpectMap(map[string]string{
		"name": "Ms. Jennifer",
		"age":  "28",
	}).ExpectError(fmt.Errorf("simulated error"))

	person, err := RetrievePerson(connection, "1")
	if err != nil {
		t.Fatal(err)
	}
	if person.Name != "Mr. Johson" {
		t.Errorf("Invalid name. Expected 'Mr. Johson' and got '%s'", person.Name)
	}
	if person.Age != 42 {
		t.Errorf("Invalid age. Expected '42' and got '%d'", person.Age)
	}

	person, err = RetrievePerson(connection, "1")
	if err != nil {
		t.Fatal(err)
	}
	if person.Name != "Ms. Jennifer" {
		t.Errorf("Invalid name. Expected 'Mr. Johson' and got '%s'", person.Name)
	}
	if person.Age != 28 {
		t.Errorf("Invalid age. Expected '28' and got '%d'", person.Age)
	}

	_, err = RetrievePerson(connection, "1")
	if err == nil {
		t.Error("Should return an error!")
	}
}

func TestDoGenericCommand(t *testing.T) {
	connection := NewConn()

	connection.GenericCommand("HGETALL").ExpectMap(map[string]string{
		"name": "Mr. Johson",
		"age":  "42",
	})

	person, err := RetrievePerson(connection, "1")
	if err != nil {
		t.Fatal(err)
	}

	if person.Name != "Mr. Johson" {
		t.Errorf("Invalid name. Expected 'Mr. Johson' and got '%s'", person.Name)
	}

	if person.Age != 42 {
		t.Errorf("Invalid age. Expected '42' and got '%d'", person.Age)
	}
}

func TestDoCommandWithGeneric(t *testing.T) {
	connection := NewConn()

	connection.Command("HGETALL", "person:1").ExpectMap(map[string]string{
		"name": "Mr. Johson",
		"age":  "42",
	})

	connection.GenericCommand("HGETALL").ExpectMap(map[string]string{
		"name": "Mr. Mark",
		"age":  "32",
	})

	person, err := RetrievePerson(connection, "1")
	if err != nil {
		t.Fatal(err)
	}

	if person.Name != "Mr. Johson" {
		t.Errorf("Invalid name. Expected 'Mr. Johson' and got '%s'", person.Name)
	}

	if person.Age != 42 {
		t.Errorf("Invalid age. Expected '42' and got '%d'", person.Age)
	}
}

func TestDoCommandWithError(t *testing.T) {
	connection := NewConn()

	connection.Command("HGETALL", "person:1").ExpectError(fmt.Errorf("simulated error"))

	_, err := RetrievePerson(connection, "1")
	if err == nil {
		t.Error("Should return an error!")
		return
	}
}

func TestDoCommandWithUnexpectedCommand(t *testing.T) {
	connection := NewConn()

	_, err := RetrievePerson(connection, "X")
	if err == nil {
		t.Error("Should detect a command not registered!")
		return
	}
}

func TestDoCommandWithUnexpectedCommandWithSuggestions(t *testing.T) {
	connection := NewConn()
	connection.Command("HGETALL", "person:1").ExpectError(fmt.Errorf("simulated error"))

	_, err := RetrievePerson(connection, "X")
	if err == nil {
		t.Fatal("Should detect a command not registered!")
	}

	msg := `command HGETALL with arguments []interface {}{"person:X"} not registered in redigomock library. Possible matches are with the arguments:
* []interface {}{"person:1"}`
	if err.Error() != msg {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestDoCommandWithoutResponse(t *testing.T) {
	connection := NewConn()

	connection.Command("HGETALL", "person:1")

	_, err := RetrievePerson(connection, "1")
	if err == nil {
		t.Fatal("Returning an information when it shoudn't")
	}
}

func TestDoCommandWithNilArgument(t *testing.T) {
	connection := NewConn()
	connection.Command("MGET", "1", nil, "2").Expect([]interface{}{"result1", nil, "result2"})

	rawValues, err := connection.Do("MGET", "1", nil, "2")
	if err != nil {
		t.Fatal(err)
	}

	values, ok := rawValues.([]interface{})
	if !ok {
		t.Fatalf("unexpected returned type %T", rawValues)
	}

	if len(values) != 3 {
		t.Fatalf("unexpected number of values (%d)", len(values))
	}

	if values[0] != "result1" {
		t.Errorf("unexpected value[0]: %v", values[0])
	}

	if values[1] != nil {
		t.Errorf("unexpected value[1]: %v", values[1])
	}

	if values[2] != "result2" {
		t.Errorf("unexpected value[2]: %v", values[2])
	}
}

func TestDoEmptyCommand(t *testing.T) {
	c := NewConn()
	reply, err := c.Do("")
	if reply != nil || err != nil {
		t.Errorf("Expected nil values received: %v, %v", reply, err)
	}
}

func TestDoEmptyFlushPipeline(t *testing.T) {
	c := NewConn()

	cmds := []struct {
		cmd []string
		val string
	}{
		{
			cmd: []string{"HGET", "somekey"},
			val: "someval",
		},
		{
			cmd: []string{"HGET", "anotherkey"},
			val: "anotherval",
		},
	}

	for _, cmd := range cmds {
		args := []interface{}{}
		for _, arg := range cmd.cmd[1:] {
			args = append(args, arg)
		}
		c.Command(cmd.cmd[0], args...).Expect(cmd.val)
		c.Send(cmd.cmd[0], args...)
	}

	reply, err := c.Do("")
	if err != nil {
		t.Errorf("Error received when trying to flush pipeline")
	}

	replies, ok := reply.([]interface{})
	if !ok {
		t.Errorf("Didn't receive slice of replies received type '%T'", reply)
	}

	if len(replies) != len(cmds) {
		t.Errorf("Expected %d replies and received %d", len(cmds), len(replies))
	}

	for i, r := range replies {
		val, ok := r.(string)
		if !ok {
			t.Errorf("Didn't receive string")
		}
		if cmds[i].val != val {
			t.Errorf("Didn't receive expected: %s != %s", cmds[i].val, val)
		}
	}
}

func TestSendFlushReceive(t *testing.T) {
	connection := NewConn()

	connection.Command("HGETALL", "person:1").ExpectMap(map[string]string{
		"name": "Mr. Johson",
		"age":  "42",
	})
	connection.Command("HGETALL", "person:2").ExpectMap(map[string]string{
		"name": "Ms. Jennifer",
		"age":  "28",
	})

	people, err := RetrievePeople(connection, []string{"1", "2"})
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

	if _, err := connection.Receive(); err == nil {
		t.Error("Not detecting when there's no more items to receive")
	}
}

func TestSendReceiveWithWait(t *testing.T) {
	conn := NewConn()
	conn.ReceiveWait = true

	conn.Command("HGETALL", "person:1").ExpectMap(map[string]string{
		"name": "Mr. Johson",
		"age":  "42",
	})
	conn.Command("HGETALL", "person:2").ExpectMap(map[string]string{
		"name": "Ms. Jennifer",
		"age":  "28",
	})

	ids := []string{"1", "2"}
	for _, id := range ids {
		conn.Send("HGETALL", fmt.Sprintf("person:%s", id))
	}

	var people []Person
	var peopleLock sync.RWMutex

	go func() {
		for i := 0; i < len(ids); i++ {
			values, err := redis.Values(conn.Receive())
			if err != nil {
				t.Fatal(err)
			}

			var person Person
			err = redis.ScanStruct(values, &person)
			if err != nil {
				t.Fatal(err)
			}

			peopleLock.Lock()
			people = append(people, person)
			peopleLock.Unlock()
		}
	}()

	for i := 0; i < len(ids); i++ {
		conn.ReceiveNow <- true
	}
	time.Sleep(10 * time.Millisecond)

	peopleLock.RLock()
	defer peopleLock.RUnlock()

	if len(people) != 2 {
		t.Fatalf("Wrong number of people. Expected '2' and got '%d'", len(people))
	}

	if people[0].Name != "Mr. Johson" || people[1].Name != "Ms. Jennifer" {
		t.Error("People name order are wrong")
	}

	if people[0].Age != 42 || people[1].Age != 28 {
		t.Error("People age order are wrong")
	}
}

func TestPubSub(t *testing.T) {
	channel := "subchannel"
	conn := NewConn()
	conn.ReceiveWait = true

	conn.Command("SUBSCRIBE", channel).Expect([]interface{}{
		[]byte("subscribe"),
		[]byte(channel),
		[]byte("1"),
	})

	// check some values are correct
	if len(conn.commands) != 1 {
		t.Errorf("unexpected number of commands, expected '%d', got '%d'", 1, len(conn.commands))
	}

	psc := redis.PubSubConn{Conn: conn}
	if err := psc.Subscribe(channel); err != nil {
		t.Error(err)
	}
	defer psc.Unsubscribe(channel)

	if err := conn.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}

	messages := [][]byte{
		[]byte("value1"),
		[]byte("value2"),
		[]byte("value3"),
		[]byte("finished"),
	}

	for _, message := range messages {
		conn.AddSubscriptionMessage([]interface{}{
			[]byte("message"),
			[]byte(channel),
			message,
		})
	}

	if len(conn.subResponses) != 4 {
		t.Errorf("unexpected number of sub-responses, expected '%d', got '%d'", 4, len(conn.subResponses))
	}

	// function to trigger a new pub/sub message received from the redis server
	nextMessage := func() {
		conn.ReceiveNow <- true
	}

	// Should receive the subscription message first
	go nextMessage()
	switch msg := psc.Receive().(type) {
	case redis.Subscription:
		break
	default:
		t.Errorf("Expected subscribe message type but received '%T'", msg)
	}

	for _, expectedMessage := range messages {
		go nextMessage()

		switch msg := psc.Receive().(type) {
		case redis.Message:
			if !reflect.DeepEqual(msg.Data, expectedMessage) {
				t.Errorf("expected message '%s'and got '%s'", string(expectedMessage), string(msg.Data))
			}
		default:
			t.Errorf("got wrong message type '%T' for message", msg)
			break
		}
	}
}

func TestSendFlushReceiveWithError(t *testing.T) {
	connection := NewConn()

	connection.Command("HGETALL", "person:1").ExpectMap(map[string]string{
		"name": "Mr. Johson",
		"age":  "42",
	})
	connection.Command("HGETALL", "person:2").ExpectMap(map[string]string{
		"name": "Ms. Jennifer",
		"age":  "28",
	})
	connection.Command("HGETALL", "person:2").ExpectError(fmt.Errorf("simulated error"))

	_, err := RetrievePeople(connection, []string{"1", "2", "3"})
	if err == nil {
		t.Error("Not detecting error when using send/flush/receive")
	}
}

func TestDummyFunctions(t *testing.T) {
	var conn Conn

	if conn.Close() != nil {
		t.Error("Close is not dummy!")
	}

	conn.CloseMock = func() error {
		return fmt.Errorf("close error")
	}

	if err := conn.Close(); err == nil || err.Error() != "close error" {
		t.Errorf("Not mocking Close method correctly. Expected “close error” and got “%v”", err)
	}

	if conn.Err() != nil {
		t.Error("Err is not dummy!")
	}

	conn.ErrMock = func() error {
		return fmt.Errorf("err error")
	}

	if err := conn.Err(); err == nil || err.Error() != "err error" {
		t.Errorf("Not mocking Err method correctly. Expected “err error” and got “%v”", err)
	}

	if conn.Flush() != nil {
		t.Error("Flush is not dummy!")
	}

	conn.FlushMock = func() error {
		return fmt.Errorf("flush error")
	}

	if err := conn.Flush(); err == nil || err.Error() != "flush error" {
		t.Errorf("Not mocking Flush method correctly. Expected “flush error” and got “%v”", err)
	}
}

func TestSkipDummyFlushFunction(t *testing.T) {
	connection := NewConn()

	errResult := false
	connection.Command("PING").Expect("PONG")
	connection.FlushSkippableMock = func() error {
		if errResult {
			return fmt.Errorf("flush error")
		}
		return nil
	}

	connection.Send("PING")
	err := connection.Flush()
	if err != nil {
		t.Errorf("Not mocking Flush method correctly. Got unexpected error “%v”", err)
	}
	s, err := connection.Receive()
	if err != nil || s != "PONG" {
		t.Errorf("Not mocking Flush method correctly. Got unexpected result “%v” and error “%v”", s, err)
	}

	errResult = true
	connection.Clear()
	connection.Command("PING").Expect("PONG")
	connection.Send("PING")
	err = connection.Flush()
	if err == nil || err.Error() != "flush error" {
		t.Errorf("Not mocking Flush method correctly. Expected “flush error” and got “%v”", err)
	}
}

func TestClear(t *testing.T) {
	connection := NewConn()

	connection.Command("HGETALL", "person:1").ExpectMap(map[string]string{
		"name": "Mr. Johson",
		"age":  "42",
	})
	connection.Command("HGETALL", "person:2").ExpectMap(map[string]string{
		"name": "Ms. Jennifer",
		"age":  "28",
	})
	connection.GenericCommand("HGETALL").ExpectMap(map[string]string{
		"name": "Ms. Mark",
		"age":  "32",
	})

	connection.Do("HGETALL", "person:1")
	connection.Do("HGETALL", "person:2")

	connection.Clear()

	if len(connection.commands) > 0 {
		t.Error("Clear function not clearing registered commands")
	}

	if len(connection.queue) > 0 {
		t.Error("Clear function not clearing the queue")
	}

	if len(connection.stats) > 0 {
		t.Error("Clear function not clearing stats")
	}
}

func TestStats(t *testing.T) {
	connection := NewConn()

	cmd1 := connection.Command("HGETALL", "person:1").ExpectMap(map[string]string{
		"name": "Mr. Johson",
		"age":  "42",
	}).ExpectMap(map[string]string{
		"name": "Mr. Johson",
		"age":  "42",
	})

	cmd2 := connection.Command("HGETALL", "person:2").ExpectMap(map[string]string{
		"name": "Mr. Larry",
		"age":  "27",
	})

	if _, err := RetrievePerson(connection, "1"); err != nil {
		t.Fatal(err)
	}

	if _, err := RetrievePerson(connection, "1"); err != nil {
		t.Fatal(err)
	}

	if counter := connection.Stats(cmd1); counter != 2 {
		t.Errorf("Expected command cmd1 to be called 2 times, but it was called %d times", counter)
	}

	if counter := connection.Stats(cmd2); counter != 0 {
		t.Errorf("Expected command cmd2 to don't be called, but it was called %d times", counter)
	}
}

func TestDoFlushesQueue(t *testing.T) {
	connection := NewConn()

	cmd1 := connection.Command("MULTI")
	cmd2 := connection.Command("SET", "person-123", 123456)
	cmd3 := connection.Command("EXPIRE", "person-123", 1000)
	cmd4 := connection.Command("EXEC").Expect([]interface{}{"OK", "OK"})

	connection.Send("MULTI")
	connection.Send("SET", "person-123", 123456)
	connection.Send("EXPIRE", "person-123", 1000)

	if _, err := connection.Do("EXEC"); err != nil {
		t.Fatal(err)
	}

	if counter := connection.Stats(cmd1); counter != 1 {
		t.Errorf("Expected cmd1 to be called once but was called %d times", counter)
	}

	if counter := connection.Stats(cmd2); counter != 1 {
		t.Errorf("Expected cmd2 to be called once but was called %d times", counter)
	}

	if counter := connection.Stats(cmd3); counter != 1 {
		t.Errorf("Expected cmd3 to be called once but was called %d times", counter)
	}

	if counter := connection.Stats(cmd4); counter != 1 {
		t.Errorf("Expected cmd4 to be called once but was called %d times", counter)
	}
}

func TestReceiveFallsBackOnGenericCommands(t *testing.T) {
	connection := NewConn()

	cmd1 := connection.Command("MULTI")
	cmd2 := connection.GenericCommand("SET")
	cmd3 := connection.GenericCommand("EXPIRE")
	cmd4 := connection.Command("EXEC")

	connection.Send("MULTI")
	connection.Send("SET", "person-123", 123456)
	connection.Send("EXPIRE", "person-123", 1000)
	connection.Send("EXEC")

	connection.Flush()

	connection.Receive()
	connection.Receive()
	connection.Receive()
	connection.Receive()

	if counter := connection.Stats(cmd1); counter != 1 {
		t.Errorf("Expected cmd1 to be called once but was called %d times", counter)
	}

	if counter := connection.Stats(cmd2); counter != 1 {
		t.Errorf("Expected cmd2 to be called once but was called %d times", counter)
	}

	if counter := connection.Stats(cmd3); counter != 1 {
		t.Errorf("Expected cmd3 to be called once but was called %d times", counter)
	}

	if counter := connection.Stats(cmd4); counter != 1 {
		t.Errorf("Expected cmd4 to be called once but was called %d times", counter)
	}
}

func TestReceiveReturnsErrorWithNoRegisteredCommand(t *testing.T) {
	connection := NewConn()

	connection.Command("SET", "person-123", "Councilman Jamm")

	connection.Send("GET", "person-123")

	connection.Flush()

	resp, err := connection.Receive()

	if err == nil {
		t.Errorf("Should have received an error when calling Receive with a command in the queue that was not registered")
	}

	if resp != nil {
		t.Errorf("Should have returned a nil response when calling Receive with a command in the queue that was not registered")
	}
}

func TestFailCommandOnTransaction(t *testing.T) {
	connection := NewConn()
	connection.Command("EXEC").Expect([]interface{}{"OK", "OK"})

	connection.Send("MULTI")

	if _, err := connection.Do("EXEC"); err == nil {
		t.Errorf("Should have received an error when calling EXEC with a transaction with a command that was not registered")
	}
}

func TestAllCommandsCalled(t *testing.T) {
	connection := NewConn()

	connection.Command("GET", "hello").Expect("world")
	connection.Command("SET", "hello", "world")

	connection.Do("GET", "hello")
	// this error is expected
	err := connection.ExpectationsWereMet()
	if err == nil {
		t.Fatal("Should have received and error because SET command not called yet")
	}

	connection.Do("SET", "hello", "world")
	err = connection.ExpectationsWereMet()
	if err != nil {
		t.Fatal("Should have no error due to SET already called")
	}

	connection.Do("DEL", "hello")
	err = connection.ExpectationsWereMet()
	if err == nil {
		t.Fatal("Should have error due to DEL is unexpected")
	}
}

func TestDoRace(t *testing.T) {
	connection := NewConn()
	n := 100

	connection.Command("GET", "hello").Expect("world")

	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			connection.Do("GET", "hello")
			wg.Done()
		}()
	}

	wg.Wait()

	err := connection.ExpectationsWereMet()
	if err != nil {
		t.Error("Called was not correctly set")
	}
}

func TestDoCommandWithHandler(t *testing.T) {
	connection := NewConn()
	connection.GenericCommand("PUBLISH").Handle(ResponseHandler(func(args []interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("unexpected number of arguments: %d", len(args))
		}
		v, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("unexpected type %T", args[1])
		}
		if v != "stuff" {
			return nil, fmt.Errorf("unexpected value '%s'", v)
		}
		return int64(1), nil
	}))

	clients, err := redis.Int64(connection.Do("PUBLISH", "ch", "stuff"))
	if err != nil {
		t.Fatal(err)
	}
	if clients != 1 {
		t.Errorf("unexpected number of notified clients '%d'", clients)
	}
}

func TestErrorRace(t *testing.T) {
	connection := NewConn()
	n := 100

	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			connection.Do("GET", "hello")
			wg.Done()
		}()
	}

	wg.Wait()

	if len(connection.Errors()) != n {
		t.Errorf("wanted %v errors, got %v", n, len(connection.Errors()))
	}
}
