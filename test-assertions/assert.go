package assert

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TODO: create tests for assertions

func Error(t *testing.T, err error, msg ...string) {
	if err == nil {
		t.Error("expected an error, got nil", msg)
	}
}

// NoError assert that error is nil
func NoError(t *testing.T, err error, msg ...string) {
	if err != nil {
		t.Error("err not expected", msg)
	}
}

// Equal compare if two values are exactly equals
func Equal(t *testing.T, notExpected, actual interface{}, msg ...string) {
	if equal := cmp.Equal(notExpected, actual); equal {
		return
	}
	errorWithDiffMsg(t, notExpected, actual, msg...)
}

// EqualValues compare the underlying values of two types
func EqualValues(t *testing.T, expected, actual interface{}, msg ...string) {
	defer func() {
		if r := recover(); r != nil {
			t.Error("Error on comparing values.\nError message: ", r)
		}
	}()

	// get the type of actual value
	actualType := reflect.TypeOf(actual)

	// get the values of expected values
	expectedValue := reflect.ValueOf(expected)

	// convert the expected values to the actual type and get it as an interface
	expectedAsInterface := expectedValue.Convert(actualType).Interface()

	if reflect.DeepEqual(expectedAsInterface, actual) {
		return
	}

	errorWithDiffMsg(t, expected, actual, msg...)
}

// NotEqual asserts that two values are not equals
func NotEqual(t *testing.T, expected, actual interface{}, msg ...string) {
	if equal := cmp.Equal(expected, actual); !equal {
		return
	}
	errorWithDiffMsg(t, expected, actual)
}

func errorWithDiffMsg(t *testing.T, expected, actual interface{}, msg ...string) {
	diff := cmp.Diff(expected, actual)
	diffMsg := fmt.Sprintf("\n- Expected\n+ Actual\n%s", diff)
	if len(msg) > 0 {
		t.Error("Not equal\n", msg, diffMsg)
		return
	}
	t.Error("Not equal", diffMsg)
}
