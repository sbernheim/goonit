package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/go-logr/logr"
	testlogr "github.com/go-logr/logr/testing"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/matchers"

	"github.com/sbernheim/goonit/mock"
)

type Test interface {
	Logf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	SetLogger(logger logr.Logger)
	Logger() logr.Logger
	TestLogr() logr.Logger
	MockLogr() *mock.MockLogger
	Mock() mock.Provider
	DoAfter(doAfterFunc func())
	SetEnv(name, val string) *BaseTest
	SetEnvs(namesAndValues ...string) *BaseTest
	SetArgs(args ...string) *BaseTest
	TempDir() string
	TempPath(filename string) string
	CopyToTempFile(srcFilepath, destFile string) string
	CopyToTemp(srcFilepath string) string
	ErrFor(errFor string) error
	Done()
}

type BaseTest struct {
	WithT
	t            *testing.T
	TestFunc     *runtime.Func
	mockProvider mock.Provider
	testLogr     testlogr.TestLogger
	mockLogr     *mock.MockLogger
	logger       logr.Logger
	captured     []interface{}
	capsFrom     map[string][]interface{}
	tempDir      string
	args         []string
	afterFunc    func()
}

func New(t *testing.T) *BaseTest {
	mockProvider := mock.NewProvider(t)
	return &BaseTest{
		WithT:        *NewWithT(t),
		t:            t,
		mockProvider: mockProvider,
		testLogr:     testlogr.TestLogger{T: t},
		mockLogr:     mockProvider.Logger(),
		afterFunc:    func() { mockProvider.Finish() },
		captured:     []interface{}{},
		capsFrom:     map[string][]interface{}{},
	}
}

func (x *BaseTest) Logf(format string, args ...interface{}) {
	x.t.Logf(format, args...)
}

func (x *BaseTest) Fatalf(format string, args ...interface{}) {
	x.t.Fatalf(format, args...)
}

func (x *BaseTest) SetLogger(logger logr.Logger) {
	x.logger = logger
}

func (x *BaseTest) Logger() logr.Logger {
	if x.logger == nil {
		x.logger = x.testLogr
	}
	return x.logger
}

func (x *BaseTest) TestLogr() logr.Logger {
	return x.testLogr
}

func (x *BaseTest) MockLogr() *mock.MockLogger {
	return x.mockLogr
}

func (x *BaseTest) Mock() mock.Provider {
	return x.mockProvider
}

func (x *BaseTest) DoAfter(doAfterFunc func()) {
	f := x.afterFunc
	x.afterFunc = func() {
		f()
		doAfterFunc()
	}
}

func (x *BaseTest) restoreExistingEnvAfter(name string) {
	currentVal, ok := os.LookupEnv(name)
	if ok {
		x.DoAfter(func() {
			os.Setenv(name, currentVal)
		})
	} else {
		x.DoAfter(func() {
			os.Unsetenv(name)
		})
	}
}

func (x *BaseTest) SetEnv(name, val string) *BaseTest {
	x.restoreExistingEnvAfter(name)
	os.Setenv(name, val)
	return x
}

func (x *BaseTest) splitPairs(namesAndValues []string) [][]string {
	pairs := make([][]string, 0, len(namesAndValues)/2)
	for i := 0; i < len(namesAndValues); i = i + 2 {
		pairs = append(pairs, namesAndValues[i:i+2])
	}
	return pairs
}

func (x *BaseTest) SetEnvs(namesAndValues ...string) *BaseTest {
	if namesAndValues != nil && len(namesAndValues) < 2 {
		x.Fatalf(fmt.Sprintf("must set at least one name and value pair! passed values %v", namesAndValues))
	}
	if len(namesAndValues)%2 == 1 {
		x.Fatalf(fmt.Sprintf("must pass names and values in even pairs! passed %d values %v", len(namesAndValues), namesAndValues))
	}
	for _, pair := range x.splitPairs(namesAndValues) {
		x.SetEnv(pair[0], pair[1])
	}
	return x
}

func (x *BaseTest) commandArg() string {
	if os.Args == nil || len(os.Args) < 1 {
		return "fakeCommand"
	}
	return os.Args[0]
}

func (x *BaseTest) restoreExistingArgsAfter() {
	currentArgs := os.Args
	if currentArgs != nil {
		x.DoAfter(func() {
			os.Args = currentArgs
		})
	}
}

func (x *BaseTest) SetArgs(args ...string) *BaseTest {
	x.restoreExistingArgsAfter()
	x.args = []string{}
	x.args = append(x.args, args...)
	os.Args = x.args
	return x
}

func (x *BaseTest) TempDir() string {
	if x.tempDir == "" {
		x.tempDir = x.t.TempDir()
	}
	return x.tempDir
}

func (x *BaseTest) TempPath(filename string) string {
	return filepath.Join(x.TempDir(), filename)
}

func (x *BaseTest) copyFile(src, dest string) {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		x.Fatalf("failed to read file '%s': %s", src, err.Error())
	}
	err = ioutil.WriteFile(dest, data, 0666)
	if err != nil {
		x.Fatalf("failed to write file '%s': %s", dest, err.Error())
	}
}

// Copies the source file to a file in temp directory with the provided name
// and returns the full filepath of the copy.
func (x *BaseTest) CopyToTempFile(srcFilepath, destFile string) string {
	destFilepath := x.TempPath(destFile)
	x.copyFile(srcFilepath, destFilepath)
	return destFilepath
}

// Copies the source file to the temp directory
// and returns the full filepath of the copy.
func (x *BaseTest) CopyToTemp(srcFilepath string) string {
	return x.CopyToTempFile(srcFilepath, filepath.Base(srcFilepath))
}

func (x *BaseTest) ErrFor(errFor string) error {
	return fmt.Errorf("this is a test-generated error for '%s'", errFor)
}

func (x *BaseTest) Done() {
	x.afterFunc()
}

func (x *BaseTest) Capture(captured ...interface{}) *BaseTest {
	stack := x.BuildCallerStack()
	if stack.Mocked == nil {
		x.Logf("NO MOCK FOUND FOR CAPTURE from %s", stack.Caller.LogString())
	} else {
		caps, found := x.capsFrom[stack.MockedCall()]
		if !found {
			caps = make([]interface{}, 0, 3)
		}
		x.capsFrom[stack.MockedCall()] = append(caps, captured...)
	}
	x.captured = append(x.captured, captured...)
	return x
}

func (x *BaseTest) AllCaptured() []interface{} {
	return x.captured
}

func (x *BaseTest) Captured(index int, expectTypeOf interface{}) interface{} {
	x.Expect(x.captured).ShouldNot(BeEmpty(), "There were no captured parameter values!")
	x.Expect(len(x.captured)).Should(BeNumerically(">=", index+1), "There were only %d captured parameter values - cannot retrieve index %d", len(x.captured), index)
	x.Expect(x.captured[index]).Should(BeAssignableToTypeOf(expectTypeOf), "Captured parameter type %T at index %d is not assignable to type %T", x.captured[index], index, expectTypeOf)
	return x.captured[index]
}

func (x *BaseTest) capturedOfType(expectTypeOf interface{}, caps []interface{}) []interface{} {
	expectedType := &matchers.AssignableToTypeOfMatcher{Expected: expectTypeOf}
	results := make([]interface{}, 0, 1)
	for _, cap := range caps {
		if matched, err := expectedType.Match(cap); err == nil && matched {
			results = append(results, cap)
		}
	}
	return results
}

func (x *BaseTest) CapturedOfType(expectTypeOf interface{}) []interface{} {
	caps := make([]interface{}, 0, 1)
	for _, callCaps := range x.capsFrom {
		caps = append(caps, x.capturedOfType(expectTypeOf, callCaps)...)
	}
	if len(caps) == 0 {
		x.Fatalf("There were no captured parameters of type %T for %s", expectTypeOf, x.GetCallerInfo().LogString())
	}
	return caps
}

func (x *BaseTest) FirstCapturedOfType(expectTypeOf interface{}) interface{} {
	return x.CapturedOfType(expectTypeOf)[0]
}

func (x *BaseTest) capturedKeys() []string {
	keys := make([]string, 0, 1)
	for key := range x.capsFrom {
		keys = append(keys, key)
	}
	return keys
}

func (x *BaseTest) CapturedFrom(mockCall string) []interface{} {
	x.Expect(x.capsFrom).ShouldNot(BeEmpty(), "There were no captured parameter values!")
	caps, found := x.capsFrom[mockCall]
	if !found {
		caller := x.GetCallerInfo().LogString()
		x.Fatalf("at %s there were no captures from mock call '%s'!  keys %v", caller, mockCall, x.capturedKeys())
	}
	return caps
}

func (x *BaseTest) CapturedOfTypeFromCall(expectTypeOf interface{}, mockCall string) []interface{} {
	capsFrom := x.CapturedFrom(mockCall)
	caps := x.capturedOfType(expectTypeOf, capsFrom)
	if len(caps) == 0 {
		caller := x.GetCallerInfo().LogString()
		capTypes := make([]string, 0, len(caps))
		for _, cap := range caps {
			capTypes = append(capTypes, fmt.Sprintf("(%T)%v", cap, cap))
		}
		x.Fatalf("at %s there were no captures of type %T from mock call '%s'!  keys %v  values %v", caller, expectTypeOf, mockCall, x.capturedKeys(), capTypes)
	}
	return caps
}

type FuncInfo struct {
	*runtime.Func
	Function string
	Object   string
	Package  string
	File     string
	Line     int
}

func NewFuncInfo(f *runtime.Func, file string, line int) *FuncInfo {
	return (&FuncInfo{
		Func: f,
		File: file,
		Line: line}).parseName()
}

func (f *FuncInfo) isObjectRef(s string) bool {
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		return true
	}
	return false
}

func (f *FuncInfo) findObjectRef(refs []string) (int, bool) {
	for i, s := range refs {
		if f.isObjectRef(s) {
			return i, true
		}
	}
	return -1, false
}

func (f *FuncInfo) trimObjectRef(s string) string {
	return strings.TrimPrefix(strings.TrimPrefix(strings.TrimSuffix(s, ")"), "("), "*")
}

func (f *FuncInfo) parseName() *FuncInfo {
	nameElems := strings.Split(f.Name(), ".")
	if len(nameElems) < 2 {
		f.Function = f.Name()
		return f
	}
	if objIndex, found := f.findObjectRef(nameElems); found {
		f.Function = strings.Join(nameElems[objIndex+1:], ".")
		f.Object = f.trimObjectRef(nameElems[objIndex])
		f.Package = strings.Join(nameElems[:objIndex], ".")
	} else {
		f.Function = nameElems[len(nameElems)-1]
		f.Package = strings.Join(nameElems[:len(nameElems)-1], ".")
	}
	return f
}

func CallerFuncInfo(callerIndex int) (*FuncInfo, bool) {
	pc, file, line, ok := runtime.Caller(callerIndex)
	if !ok {
		return nil, false
	}
	if file == "<autogenerated>" {
		return nil, false
	}
	f := runtime.FuncForPC(pc)
	if f == nil {
		return nil, false
	}
	return NewFuncInfo(f, file, line), true
}

// Adapted from the `go test` tool.
// See https://github.com/stretchr/testify/blob/f390dcf405f7b83c997eac1b06768bb9f44dec18/assert/assertions.go#L118-L131
// Returns true if the function name name looks like a test (or benchmark, according to prefix).
// It is a Test (say) if there is a character after Test that is not a lower-case letter.
// We don't want TesticularCancer. [ Code comment gold right here! Good on ya, @paulbellamy! ]
func startsWith(name, prefix string) bool {
	if !strings.HasPrefix(name, prefix) {
		return false
	}
	if len(name) == len(prefix) { // "Test" is ok
		return true
	}
	nextChar, _ := utf8.DecodeRuneInString(name[len(prefix):])
	return !unicode.IsLower(nextChar)
}

// Returns true if this FuncInfo is a mock object function call.
func (f *FuncInfo) IsMock() bool {
	return startsWith(f.Object, "Mock")
}

// Returns true if this FuncInfo is a test function call.
func (f *FuncInfo) IsTest() bool {
	return startsWith(f.Function, "Test") ||
		startsWith(f.Function, "Benchmark") ||
		startsWith(f.Function, "Example")
}

func (f *FuncInfo) FileLine() string {
	return fmt.Sprintf("%s:%d", f.File, f.Line)
}

func (f *FuncInfo) LogString() string {
	if f.Object == "" {
		return fmt.Sprintf("%s.%s at %s", f.Package, f.Function, f.FileLine())
	}
	return fmt.Sprintf("%s.%s at %s", f.Object, f.Function, f.FileLine())
}

type FuncInfoStack struct {
	Caller *FuncInfo
	Stack  []*FuncInfo
	Test   *FuncInfo
	Tested *FuncInfo
	Mocked *FuncInfo
	Mocker *FuncInfo
	last   *FuncInfo
	this   *FuncInfo
}

func (s *FuncInfoStack) MockedCall() string {
	return strings.Join([]string{s.Mocker.Object, s.Mocker.Function, s.Mocked.Object, s.Mocked.Function}, ".")
}

func (x *BaseTest) BuildCallerStack() *FuncInfoStack {
	s := &FuncInfoStack{Stack: []*FuncInfo{}}
	this, ok := CallerFuncInfo(1)
	if !ok {
		return s
	}
	s.this = this
	for i := 3; ; i++ {
		f, ok := CallerFuncInfo(i)
		if !ok {
			return s
		}
		s.Stack = append(s.Stack, f)
		if s.Caller == nil && f.Package != s.this.Package {
			s.Caller = f
		}
		if s.last == s.Mocked {
			s.Mocker = f
		}
		if f.IsMock() {
			s.Mocked = f
		}
		if f.IsTest() {
			s.Test = f
			s.Tested = s.last
		}
		s.last = f
	}
}

func (x *BaseTest) GetMockedCallName() string {
	s := x.BuildCallerStack()
	return s.MockedCall()
}

func (x *BaseTest) GetCallerInfo() *FuncInfo {
	s := x.BuildCallerStack()
	return s.Caller
}
