package glass

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/glasslabs/looking-glass/module"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/zserge/lorca"
)

func TestNewUI(t *testing.T) {
	cfg := UIConfig{
		Width:      1024,
		Height:     764,
		Fullscreen: true,
		CustomCSS: []string{
			"testdata/custom.css",
		},
	}
	ui := &MockLorcaUI{}
	ui.On("Eval", mock.MatchedBy(func(js string) bool {
		return strings.HasPrefix(js, "loadCSS(`fonts`")
	})).Once().Return(NewValue("", nil))
	ui.On("Eval", "loadCSS(`customCSS1`, `custom css`);").Once().Return(NewValue("", nil))

	patches := ApplyFunc(lorca.New, func(url, dir string, width, height int, customArgs ...string) (lorca.UI, error) {
		assert.Equal(t, 1024, width)
		assert.Equal(t, 764, height)
		assert.Equal(t, 1024, width)
		assert.Contains(t, customArgs, "--start-fullscreen")

		return ui, nil
	})
	t.Cleanup(func() {
		patches.Reset()
	})

	got, err := NewUI(cfg)

	require.NoError(t, err)
	assert.IsType(t, (*UI)(nil), got)
	ui.AssertExpectations(t)
}

func TestNewUI_HandlesWindowError(t *testing.T) {
	cfg := UIConfig{
		Width:  1024,
		Height: 764,
	}

	patches := ApplyFunc(lorca.New, func(url, dir string, width, height int, customArgs ...string) (lorca.UI, error) {
		return nil, errors.New("test error")
	})
	t.Cleanup(func() {
		patches.Reset()
	})

	_, err := NewUI(cfg)

	assert.Error(t, err)
	assert.EqualError(t, err, "could not create window: test error")
}

func TestUI_Done(t *testing.T) {
	ch := make(chan struct{})
	t.Cleanup(func() {
		close(ch)
	})
	win := &MockLorcaUI{}
	win.On("Done").Return(ch)
	ui := &UI{win: win}

	got := ui.Done()

	if ch != got {
		assert.Fail(t, "incorrect channel")
		return
	}
	win.AssertExpectations(t)
}

func TestUI_Close(t *testing.T) {
	win := &MockLorcaUI{}
	win.On("Close").Return(nil)
	ui := &UI{win: win}

	err := ui.Close()

	assert.NoError(t, err)
	win.AssertExpectations(t)
}

func TestNewUIContext(t *testing.T) {
	emptyVal := NewValue("", nil)
	win := &MockLorcaUI{}
	win.On("Eval", `createModule("test", "top", "right");`).Return(emptyVal, nil)

	ui := &UI{win: win}
	pos := module.Position{
		Vertical:   module.Top,
		Horizontal: module.Right,
	}

	got, err := NewUIContext(ui, "test", pos)

	require.NoError(t, err)
	assert.IsType(t, &UIContext{}, got)
	win.AssertExpectations(t)
}

func TestNewUIContext_HandlesModuleError(t *testing.T) {
	errVal := NewValue("", errors.New("test err"))
	win := &MockLorcaUI{}
	win.On("Eval", `createModule("test", "top", "right");`).Return(errVal)

	ui := &UI{win: win}
	pos := module.Position{
		Vertical:   module.Top,
		Horizontal: module.Right,
	}

	_, err := NewUIContext(ui, "test", pos)

	require.Error(t, err)
	assert.EqualError(t, err, "test: could not create module ui element: test err")
	win.AssertExpectations(t)
}

func TestUIContext_LoadCSS(t *testing.T) {
	emptyVal := NewValue("", nil)
	win := &MockLorcaUI{}
	win.On("Eval", `createModule("test", "top", "right");`).Return(emptyVal)
	win.On("Eval", "loadCSS(`test`, `test css`);").Return(emptyVal)

	ui := &UI{win: win}
	pos := module.Position{
		Vertical:   module.Top,
		Horizontal: module.Right,
	}
	uiCtx, err := NewUIContext(ui, "test", pos)
	require.NoError(t, err)

	err = uiCtx.LoadCSS("test css")

	require.NoError(t, err)
	win.AssertExpectations(t)
}

func TestUIContext_LoadHTML(t *testing.T) {
	emptyVal := NewValue("", nil)
	win := &MockLorcaUI{}
	win.On("Eval", `createModule("test", "top", "right");`).Return(emptyVal)
	win.On("Eval", "loadModuleHTML(`test`, `test html`);").Return(emptyVal)

	ui := &UI{win: win}
	pos := module.Position{
		Vertical:   module.Top,
		Horizontal: module.Right,
	}
	uiCtx, err := NewUIContext(ui, "test", pos)
	require.NoError(t, err)

	err = uiCtx.LoadHTML("test html")

	require.NoError(t, err)
	win.AssertExpectations(t)
}

func TestUIContext_Bind(t *testing.T) {
	emptyVal := NewValue("", nil)
	win := &MockLorcaUI{}
	win.On("Eval", `createModule("test", "top", "right");`).Return(emptyVal)
	win.On("Bind", "testfunc", mock.AnythingOfType("func(string, string) string")).Return(nil)

	ui := &UI{win: win}
	pos := module.Position{
		Vertical:   module.Top,
		Horizontal: module.Right,
	}
	uiCtx, err := NewUIContext(ui, "test", pos)
	require.NoError(t, err)

	err = uiCtx.Bind("testfunc", func(a, b string) string { return "test" })

	require.NoError(t, err)
	win.AssertExpectations(t)
}

func TestUIContext_Eval(t *testing.T) {
	emptyVal := NewValue("", nil)
	mapVal := NewValue(`{"test": "return"}`, nil)
	win := &MockLorcaUI{}
	win.On("Eval", `createModule("test", "top", "right");`).Return(emptyVal)
	win.On("Eval", "some js test").Return(mapVal)

	ui := &UI{win: win}
	pos := module.Position{
		Vertical:   module.Top,
		Horizontal: module.Right,
	}
	uiCtx, err := NewUIContext(ui, "test", pos)
	require.NoError(t, err)

	got, err := uiCtx.Eval("some js %s", "test")

	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"test": "return"}, got)
	win.AssertExpectations(t)
}

func TestUIContext_EvalHandlesEmptyValue(t *testing.T) {
	emptyVal := NewValue("", nil)
	win := &MockLorcaUI{}
	win.On("Eval", `createModule("test", "top", "right");`).Return(emptyVal)
	win.On("Eval", "some js test").Return(emptyVal)

	ui := &UI{win: win}
	pos := module.Position{
		Vertical:   module.Top,
		Horizontal: module.Right,
	}
	uiCtx, err := NewUIContext(ui, "test", pos)
	require.NoError(t, err)

	got, err := uiCtx.Eval("some js %s", "test")

	require.NoError(t, err)
	assert.Nil(t, got)
	win.AssertExpectations(t)
}

func TestUIContext_EvalHandlesError(t *testing.T) {
	emptyVal := NewValue("", nil)
	errorVal := NewValue("", errors.New("test"))
	win := &MockLorcaUI{}
	win.On("Eval", `createModule("test", "top", "right");`).Return(emptyVal)
	win.On("Eval", "some js test").Return(errorVal)

	ui := &UI{win: win}
	pos := module.Position{
		Vertical:   module.Top,
		Horizontal: module.Right,
	}
	uiCtx, err := NewUIContext(ui, "test", pos)
	require.NoError(t, err)

	got, err := uiCtx.Eval("some js %s", "test")

	require.Error(t, err)
	assert.Nil(t, got)
	win.AssertExpectations(t)
}

type MockLorcaUI struct {
	mock.Mock
}

func (m *MockLorcaUI) Load(url string) error {
	args := m.Called(url)
	return args.Error(0)
}

func (m *MockLorcaUI) Bounds() (lorca.Bounds, error) {
	args := m.Called()
	return args.Get(0).(lorca.Bounds), args.Error(1)
}

func (m *MockLorcaUI) SetBounds(bounds lorca.Bounds) error {
	args := m.Called(bounds)
	return args.Error(0)
}

func (m *MockLorcaUI) Bind(name string, f interface{}) error {
	args := m.Called(name, f)
	return args.Error(0)
}

func (m *MockLorcaUI) Eval(js string) lorca.Value {
	args := m.Called(js)
	return args.Get(0).(lorca.Value)
}

func (m *MockLorcaUI) Done() <-chan struct{} {
	args := m.Called()
	return args.Get(0).(chan struct{})
}

func (m *MockLorcaUI) Close() error {
	args := m.Called()
	return args.Error(0)
}

type Value struct {
	err error
	raw json.RawMessage
}

func NewValue(val string, err error) Value {
	v := json.RawMessage{}
	if val != "" {
		v = json.RawMessage(val)
	}

	return Value{
		raw: v,
		err: err,
	}
}

func (v Value) Err() error             { return v.err }
func (v Value) To(x interface{}) error { return json.Unmarshal(v.raw, x) }
func (v Value) Float() (f float32)     { v.To(&f); return f }
func (v Value) Int() (i int)           { v.To(&i); return i }
func (v Value) String() (s string)     { v.To(&s); return s }
func (v Value) Bool() (b bool)         { v.To(&b); return b }
func (v Value) Bytes() []byte          { return v.raw }
func (v Value) Array() (values []lorca.Value) {
	array := []json.RawMessage{}
	v.To(&array)
	for _, el := range array {
		values = append(values, Value{raw: el})
	}
	return values
}
func (v Value) Object() (object map[string]lorca.Value) {
	object = map[string]lorca.Value{}
	kv := map[string]json.RawMessage{}
	v.To(&kv)
	for k, v := range kv {
		object[k] = Value{raw: v}
	}
	return object
}
