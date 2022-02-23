package engine

import (
	"fmt"
	"reflect"
	"time"
)

type Assertion struct {
	name     string
	nodeType NodeType
	parent   *Assertion
	children []*Assertion
	details  []Detail
	result   Result
	Logger   Logger
}

type Detail struct {
	Name       string
	Message    string
	RecordTime time.Time
}

func NewAssertion(name string, nodeType NodeType, parent *Assertion, logger Logger) *Assertion {
	assertion := &Assertion{
		name:     name,
		nodeType: nodeType,
		result:   NOTRUN,
		parent:   parent,
		children: make([]*Assertion, 0),
		details:  make([]Detail, 0),
		Logger:   logger,
	}
	if assertion.parent != nil {
		assertion.parent.children = append(assertion.parent.children, assertion)
	}

	return assertion
}

func (a *Assertion) AssertEquals(expected interface{}, actual interface{}, title string) {
	if reflect.TypeOf(expected).Kind() != reflect.TypeOf(actual).Kind() {
		a.details = append(a.details, Detail{
			Name:       "Assert",
			Message:    fmt.Sprintf("%s expected %T(%v), but actual was %T(%v)", title, expected, expected, actual, actual),
			RecordTime: time.Now(),
		})
		a.fail()
		return
	}
	switch reflect.TypeOf(expected).Kind() {
	case reflect.Array, reflect.Slice:
		expectedRV := reflect.ValueOf(expected)
		actualRV := reflect.ValueOf(actual)
		if expectedRV.Len() != actualRV.Len() {
			a.details = append(a.details, Detail{
				Name:       "Assert",
				Message:    fmt.Sprintf("%s expected %T(%v), but actual was %T(%v)", title, expected, expected, actual, actual),
				RecordTime: time.Now(),
			})
			a.fail()
			return
		}
		for i := 0; i < expectedRV.Len(); i++ {
			if expectedRV.Index(i).CanInterface() {
				if expectedRV.Index(i).Interface() != actualRV.Index(i).Interface() {
					a.details = append(a.details, Detail{
						Name:       "Assert",
						Message:    fmt.Sprintf("%s expected %T(%v), but actual was %T(%v)", title, expected, expected, actual, actual),
						RecordTime: time.Now(),
					})
					a.fail()
					return
				}
			} else {
				a.details = append(a.details, Detail{
					Name:       "Assert",
					Message:    "cannot be compared",
					RecordTime: time.Now(),
				})
				a.fail()
				return
			}
		}
	default:
		if expected != actual {
			a.details = append(a.details, Detail{
				Name:       "Assert",
				Message:    fmt.Sprintf("%s expected %T(%v), but actual was %T(%v)", title, expected, expected, actual, actual),
				RecordTime: time.Now(),
			})
			a.fail()
		}
	}

}

func (a *Assertion) AssertNotEquals(expected interface{}, actual interface{}, title string) {
	if reflect.TypeOf(expected).Kind() != reflect.TypeOf(actual).Kind() {
		return
	}
	switch reflect.TypeOf(expected).Kind() {
	case reflect.Array, reflect.Slice:
		expectedRV := reflect.ValueOf(expected)
		actualRV := reflect.ValueOf(actual)
		if expectedRV.Len() != actualRV.Len() {
			return
		}
		for i := 0; i < expectedRV.Len(); i++ {
			if expectedRV.Index(i).CanInterface() {
				if expectedRV.Index(i).Interface() != actualRV.Index(i).Interface() {
					return
				}
			} else {
				a.details = append(a.details, Detail{
					Name:       "Assert",
					Message:    "cannot be compared",
					RecordTime: time.Now(),
				})
				a.fail()
			}
		}
		a.details = append(a.details, Detail{
			Name:       "Assert",
			Message:    fmt.Sprintf("%s expected not %T(%v), but actual was %T(%v)", title, expected, expected, actual, actual),
			RecordTime: time.Now(),
		})
		a.fail()
	default:
		if expected == actual {
			a.details = append(a.details, Detail{
				Name:       "Assert",
				Message:    fmt.Sprintf("%s expected not %T(%v), but actual was %T(%v)", title, expected, expected, actual, actual),
				RecordTime: time.Now(),
			})
			a.fail()
		}
	}
}

func (a *Assertion) AssertFail(title string) {
	a.details = append(a.details, Detail{
		Name:       "Assert",
		Message:    fmt.Sprintf("Fail, because %s", title),
		RecordTime: time.Now(),
	})
	a.fail()
}

func (a *Assertion) fail() {
	if a.parent != nil {
		a.parent.fail()
	}
	a.result = FAIL
}

func (a *Assertion) Result() Result {
	return a.result
}

func (a *Assertion) AddDetail(name string, message string, args ...interface{}) {
	a.details = append(a.details, Detail{Name: name, Message: fmt.Sprintf(message, args...), RecordTime: time.Now()})
}

func (a *Assertion) GetDetails() []Detail {
	return a.details
}
