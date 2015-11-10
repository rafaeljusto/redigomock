package redigomock

import (
	"fmt"
	"path/filepath"
)

// NewFakeRedis returns a connection, that acts as real redis
func NewFakeRedis() *Conn {
	c := NewConn()
	c.fake()
	return c
}

type fakeRedis struct {
	keys map[string]interface{}
	sets map[string]map[string]interface{}
}

type fakeStruct struct {
	redisStruct interface{}
}

func getRedisStruct(obj interface{}) (interface{}, error) {
	s, ok := obj.(*fakeStruct)
	if !ok {
		return nil, fmt.Errorf("Wrong type of struct")
	}
	return s.redisStruct, nil
}

func newFakeRedis() *fakeRedis {
	return &fakeRedis{
		keys: make(map[string]interface{}),
	}
}

func (c *Conn) fake() {
	fake := newFakeRedis()

	c.Command("SET", NewAnyData(), NewAnyData()).ExpectCallback(func(args []interface{}) (interface{}, error) {
		key := args[0].(string)
		fake.keys[key] = args[1]
		return "OK", nil
	})

	c.Command("GET", NewAnyData()).ExpectCallback(func(args []interface{}) (interface{}, error) {
		key := args[0].(string)
		if r, exists := fake.keys[key]; exists {
			if _, ok := r.(*fakeStruct); ok {
				return nil, fmt.Errorf("Wrong type of struct")
			}
			return r, nil
		}
		return nil, nil
	})

	c.Command("KEYS", NewAnyData()).ExpectCallback(func(args []interface{}) (interface{}, error) {
		keys := make([]interface{}, 0, 64)
		pattern := args[0].(string)
		for key := range fake.keys {
			matched, err := filepath.Match(pattern, key)
			if err != nil {
				return nil, err
			}
			if matched {
				keys = append(keys, []byte(key))
			}
		}
		return keys, nil
	})

	c.Command("FLUSHDB").ExpectCallback(func(args []interface{}) (interface{}, error) {
		fake = newFakeRedis()
		return nil, nil
	})

	c.Command("SADD", NewAnyDataArray()).ExpectCallback(func(args []interface{}) (interface{}, error) {
		key := args[0].(string)
		setStruct, found := fake.keys[key]
		if !found {
			setStruct = &fakeStruct{make(map[string]interface{})}
			fake.keys[key] = setStruct
		}
		set, err := getRedisStruct(setStruct)
		if err != nil {
			return nil, err
		}
		_set := set.(map[string]interface{})
		inserted := 0
		for _, value := range args[1:] {
			if _, found := _set[value.(string)]; !found {
				inserted++
			}
			_set[value.(string)] = true
		}
		return int64(inserted), nil
	})

	c.Command("SREM", NewAnyDataArray()).ExpectCallback(func(args []interface{}) (interface{}, error) {
		key := args[0].(string)
		setStruct, found := fake.keys[key]
		if !found {
			return 0, nil
		}
		set, err := getRedisStruct(setStruct)
		if err != nil {
			return nil, err
		}
		_set := set.(map[string]interface{})
		removed := 0
		for _, value := range args[1:] {
			if _, found := _set[value.(string)]; found {
				removed++
			}
			delete(_set, value.(string))
		}
		return int64(removed), nil
	})

	c.Command("SMEMBERS", NewAnyData()).ExpectCallback(func(args []interface{}) (interface{}, error) {
		key := args[0].(string)
		setStruct, found := fake.keys[key]
		if !found {
			return []string{}, nil
		}
		set, err := getRedisStruct(setStruct)
		if err != nil {
			return nil, err
		}
		_set := set.(map[string]interface{})
		keys := make([]interface{}, 0, len(_set))
		for key := range _set {
			keys = append(keys, []byte(key))
		}
		return keys, nil
	})

}
