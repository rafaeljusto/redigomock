package redigomock

import "reflect"

var (
	fuzzyCommands []*Cmd // Global variable that stores all registered  fuzzy commands
)

// FuzzyMatcher is an interface that exports exports one function. It can be passed to the Command as a argument.
// When the command is evaluated agains data provided in mock connection Do call, FuzzyMatcher will call Match on the
//argument and returns true if argument fulfils constraints set in concrete implementation .
type FuzzyMatcher interface {

	//Func Match takes an argument passed to mock connection Do call and check if it fulfulls constraints set in concrete implementation of this interface
	Match(interface{}) bool
}

// NewAnyInt retunrs a FuzzyMatcher instance maching any integer passed as an argument
func NewAnyInt() FuzzyMatcher {
	return anyInt{}
}

// NewAnyDouble returns a FuzzyMatcher instance mathing any double passed as an argument
func NewAnyDouble() FuzzyMatcher {
	return anyDouble{}
}

type anyInt struct{}

func (matcher anyInt) Match(input interface{}) bool {
	switch input.(type) {
	case int, int8, int16, int32, int64, uint8, uint16, uint32, uint64:
		return true
	default:
		return false
	}
}

type anyDouble struct{}

func (matcher anyDouble) Match(input interface{}) bool {
	switch input.(type) {
	case float32, float64:
		return true
	default:
		return false
	}
}

func fuzzyCommandMatch(commandName string, args []interface{}, cmd *Cmd) bool {
	if commandName != cmd.Name || len(args) != len(cmd.Args) {
		return false
	}

	for pos := range cmd.Args {
		if implementsFuzzy(cmd.Args[pos]) {
			if cmd.Args[pos].(FuzzyMatcher).Match(args[pos]) == false {
				return false
			}
		} else {
			if reflect.DeepEqual(cmd.Args[pos], args[pos]) == false {
				return false
			}
		}
	}
	return true
}

func implementsFuzzy(input interface{}) bool {
	return reflect.TypeOf(input).Implements(reflect.TypeOf((*FuzzyMatcher)(nil)).Elem())
}

func fuzzyCommandEqual(commandName string, args []interface{}, cmd *Cmd) bool {
	if commandName != cmd.Name || len(args) != len(cmd.Args) {
		return false
	}

	for pos := range cmd.Args {
		if reflect.DeepEqual(cmd.Args[pos], args[pos]) == false {
			return false
		}
	}
	return true
}
