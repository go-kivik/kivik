package kivikmock

import (
	"encoding/json"
	"reflect"

	kivik "github.com/go-kivik/kivik/v4"
)

func meets(a, e expectation) bool {
	if reflect.TypeOf(a).Elem().Name() != reflect.TypeOf(e).Elem().Name() {
		return false
	}
	// Skip the DB test for the dbo() method
	if _, ok := e.(*ExpectedDB); !ok {
		if !dbMeetsExpectation(a.dbo(), e.dbo()) {
			return false
		}
	}
	if !optionsMeetExpectation(a.opts(), e.opts()) {
		return false
	}
	return a.met(e)
}

func dbMeetsExpectation(a, e *DB) bool {
	if e == nil {
		return true
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	a.mu.RLock()
	defer a.mu.RUnlock()
	return e.name == a.name && e.id == a.id
}

func optionsMeetExpectation(a, e kivik.Options) bool {
	if e == nil {
		return true
	}
	return reflect.DeepEqual(e, a)
}

func jsonMeets(e, a interface{}) bool {
	eJSON, _ := json.Marshal(e)
	aJSON, _ := json.Marshal(a)
	var eI, aI interface{}
	_ = json.Unmarshal(eJSON, &eI)
	_ = json.Unmarshal(aJSON, &aI)
	return reflect.DeepEqual(eI, aI)
}
