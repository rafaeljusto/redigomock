package redigomock

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// NewFakeRedis returns a connection, that acts as real redis
func NewFakeRedis() *Conn {
	c := NewConn()
	c.fake()
	return c
}

type fakeRedis struct {
	keys map[string]*container
}

type container struct {
	redisStruct interface{}
	redisType   int
}

func newFakeRedis() *fakeRedis {
	return &fakeRedis{
		keys: make(map[string]*container),
	}
}

type operator func(int64, int64) bool

func gt(score, from int64) bool {
	return from < score
}

func gte(score, from int64) bool {
	return from <= score
}

func lt(score, to int64) bool {
	return score < to
}

func lte(score, to int64) bool {
	return score <= to
}

func inf(score, limit int64) bool {
	return true
}

func extractOperator(limit string, defaultOperator operator, nonEqualOperator operator) (operator, int64, error) {
	o := defaultOperator
	if limit[0] == '(' {
		o = nonEqualOperator
		score, err := strconv.ParseInt(limit[1:], 10, 64)
		if err != nil {
			return nil, 0, err
		}
		return o, score, nil
	}
	if strings.HasSuffix(limit, "inf") {
		return inf, 0, nil
	}
	score, err := strconv.ParseInt(limit, 10, 64)
	if err != nil {
		return nil, 0, err
	}
	return o, score, nil
}

func toString(value interface{}) string {
	switch value := value.(type) {
	case string:
		return string(value)
	case []byte:
		return string(value)
	case int:
		return fmt.Sprintf("%d", value)
	case int64:
		return fmt.Sprintf("%d", value)
	case float64:
		return fmt.Sprintf("%f", value)
	case bool:
		if bool(value) {
			return "1"
		}
		return "0"
	case nil:
		return ""
	default:
		panic(fmt.Sprintf("Unsupported argument type: %s", value))
	}
}

func toInt64(value interface{}) int64 {
	switch value := value.(type) {
	case int:
		return int64(value)
	case int64:
		return value
	default:
		panic(fmt.Sprintf("Unsupported argument type: %s", value))
	}
}

func (c *Conn) fake() {
	fake := newFakeRedis()

	c.Command("MULTI")

	c.Command("EXEC").ExpectCallback(func(args []interface{}) (interface{}, error) {
		result := make([]interface{}, len(c.pendingResults))
		copy(result, c.pendingResults)
		c.pendingResults = make([]interface{}, 0)
		return result, nil
	})

	c.Command("SET", NewAnyDataArray()).ExpectCallback(func(args []interface{}) (interface{}, error) {
		key := toString(args[0])
		fake.keys[key] = &container{args[1], _redisKey}
		return "OK", nil
	})

	c.Command("GET", NewAnyData()).ExpectCallback(func(args []interface{}) (interface{}, error) {
		key := toString(args[0])
		return fake.getRedisStruct(key, _redisKey)
	})

	c.Command("KEYS", NewAnyData()).ExpectCallback(func(args []interface{}) (interface{}, error) {
		keys := make([]interface{}, 0, 64)
		pattern := toString(args[0])
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
		key := toString(args[0])
		set, err := fake.getSet(key)
		if err != nil {
			return nil, err
		}
		if set == nil {
			set = make(map[string]interface{})
			fake.keys[key] = &container{set, _redisSet}
		}
		inserted := 0
		for _, value := range args[1:] {
			if _, found := set[toString(value)]; !found {
				inserted++
			}
			set[toString(value)] = value
		}
		return int64(inserted), nil
	})

	c.Command("SREM", NewAnyDataArray()).ExpectCallback(func(args []interface{}) (interface{}, error) {
		key := toString(args[0])
		set, err := fake.getSet(key)
		if err != nil {
			return nil, err
		}
		if set == nil {
			return 0, nil
		}
		removed := 0
		for _, value := range args[1:] {
			if _, found := set[toString(value)]; found {
				removed++
			}
			delete(set, toString(value))
		}
		return int64(removed), nil
	})

	c.Command("SMEMBERS", NewAnyData()).ExpectCallback(func(args []interface{}) (interface{}, error) {
		key := toString(args[0])
		set, err := fake.getSet(key)
		if err != nil {
			return nil, err
		}
		if set == nil {
			return [][]byte{}, nil
		}
		keys := make([]interface{}, 0, len(set))
		for key := range set {
			keys = append(keys, []byte(key))
		}
		return keys, nil
	})

	c.Command("SUNION", NewAnyDataArray()).ExpectCallback(func(args []interface{}) (interface{}, error) {
		keys := make(map[string]bool)
		for _, arg := range args {
			key := toString(arg)
			set, err := fake.getSet(key)
			if err != nil {
				return nil, err
			}
			if set == nil {
				continue
			}
			for key := range set {
				keys[key] = true
			}
		}
		a := make([]interface{}, 0, len(keys))
		for key := range keys {
			a = append(a, []byte(key))
		}
		return a, nil
	})

	c.Command("ZADD", NewAnyDataArray()).ExpectCallback(func(args []interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("Wrong number of arguments passed")
		}
		key := toString(args[0])
		set, err := fake.getSortedSet(key)
		if err != nil {
			return nil, err
		}
		if set == nil {
			set = make(map[string]*scoredValue)
			fake.keys[key] = &container{set, _redisSortedSet}
		}
		inserted := 0
		for i := range args[1:] {
			if i%2 == 0 && i < len(args)-1 {
				score := args[1+i]
				value := args[2+i]
				if _, found := set[toString(value)]; !found {
					inserted++
				}
				set[toString(value)] = &scoredValue{value, toInt64(score)}
			}
		}
		return int64(inserted), nil
	})

	c.Command("ZRANGE", NewAnyDataArray()).ExpectCallback(func(args []interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("Wrong number of arguments passed")
		}
		key := toString(args[0])
		from := args[1].(int)
		to := args[2].(int)
		withScores := strings.ToLower(fmt.Sprintf("%s", args[len(args)-1])) == "withscores"
		set, err := fake.getSortedSet(key)
		if err != nil {
			return nil, err
		}
		if set == nil {
			return []string{}, nil
		}
		values := make(scoredValueArray, 0, len(set))
		for _, value := range set {
			values = append(values, value)
		}
		sort.Sort(values)
		if to < 0 {
			to = len(values) + to
		}
		if to > len(values)-1 {
			to = len(values) - 1
		}

		result := make([]interface{}, 0, len(values))
		for _, v := range values[from : to+1] {
			result = append(result, []byte(toString(v.value)))
			if withScores {
				result = append(result, []byte(fmt.Sprintf("%d", v.score)))
			}
		}
		return result, nil
	})

	c.Command("ZCOUNT", NewAnyDataArray()).ExpectCallback(func(args []interface{}) (interface{}, error) {
		count := int64(0)
		if err := fake.sortedSetEnum(args, func(set map[string]*scoredValue, v *scoredValue) {
			count++
		}); err != nil {
			return 0, err
		}
		return count, nil
	})

	c.Command("ZREMRANGEBYSCORE", NewAnyDataArray()).ExpectCallback(func(args []interface{}) (interface{}, error) {
		count := int64(0)
		if err := fake.sortedSetEnum(args, func(set map[string]*scoredValue, v *scoredValue) {
			count++
			delete(set, toString(v.value))
		}); err != nil {
			return 0, err
		}
		return count, nil
	})

	c.Command("ZRANGEBYSCORE", NewAnyDataArray()).ExpectCallback(func(args []interface{}) (interface{}, error) {
		var result []interface{}
		withScores := strings.ToLower(fmt.Sprintf("%s", args[len(args)-1])) == "withscores"

		result = make([]interface{}, 0)
		if err := fake.sortedSetEnum(args, func(set map[string]*scoredValue, v *scoredValue) {
			result = append(result, []byte(toString(v.value)))
			if withScores {
				result = append(result, []byte(fmt.Sprintf("%d", v.score)))
			}
		}); err != nil {
			return result, err
		}

		return result, nil
	})
}
