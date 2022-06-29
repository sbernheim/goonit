package match

import (
	"fmt"
	"reflect"

	"github.com/golang/mock/gomock"
)

type isType struct{ t reflect.Type }

func IsType(t interface{}) gomock.Matcher {
	return &isType{reflect.TypeOf(t)}
}

func (m *isType) Matches(param interface{}) bool {
	return reflect.TypeOf(param) == m.t
}

func (m *isType) String() string {
	return fmt.Sprintf("is of type %s", m.t)
}

func AnyString() gomock.Matcher {
	return IsType("")
}

type isFunc struct{}

func (m *isFunc) Matches(param interface{}) bool {
	return reflect.TypeOf(param).Kind() == reflect.Func
}

func (m *isFunc) String() string {
	return "is a function reference"
}

func AnyFunc() gomock.Matcher {
	return &isFunc{}
}
