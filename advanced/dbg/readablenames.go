package dbg

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	petname "github.com/dustinkirkland/golang-petname"
)

// This converts arbitrary strings into random readable names. It flagrantly
// leaks memory but generates the names lazily, so it's not a problem unless
// you're actually using it. This is helpful for turning pointer strings into
// something more easily distinguishable when debugging.

var memo map[interface{}]string
var memoMutex sync.RWMutex

func init() {
	memo = make(map[interface{}]string)
	// Since the ids are generated in order of demand, we make them
	// nondetemrinistic to remind the user that the same name doesn't refer to the
	// same thing between runs.
	petname.NonDeterministicMode()
}

func Name(obj interface{}) string {
	if reflect.ValueOf(obj).IsNil() {
		return "Ø"
	}

	memoMutex.RLock()
	if r, ok := memo[obj]; ok {
		memoMutex.RUnlock()
		return r
	}
	memoMutex.RUnlock()
	
	memoMutex.Lock()
	defer memoMutex.Unlock()
	// Check again in case another goroutine added it while we were waiting
	if r, ok := memo[obj]; ok {
		return r
	}
	r := fmt.Sprintf("%s%s", strings.Title(petname.Adjective()), strings.Title(petname.Name()))
	memo[obj] = r
	return r
}
