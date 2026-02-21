package plugin

import (
	"context"
	"os"
	"testing"

	"github.com/dop251/goja"
	"github.com/gatewayd-io/gatewayd-plugin-sdk/logging"
	v1 "github.com/gatewayd-io/gatewayd-plugin-sdk/plugin/v1"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestPlugin(t *testing.T) *Plugin {
	t.Helper()
	logger := hclog.New(&hclog.LoggerOptions{
		Level:  logging.GetLogLevel("error"),
		Output: os.Stderr,
	})
	return &Plugin{
		Logger:   logger,
		VM:       goja.New(),
		Bindings: map[string]goja.Callable{},
	}
}

func newTestRequest(t *testing.T) *v1.Struct {
	t.Helper()
	req, err := v1.NewStruct(map[string]interface{}{
		"key": "value",
	})
	require.NoError(t, err)
	return req
}

func TestRegisterFunction_Valid(t *testing.T) {
	p := newTestPlugin(t)
	_, err := p.VM.RunString(`function onBooted(ctx, req) { return req; }`)
	require.NoError(t, err)

	p.RegisterFunction("onBooted")
	assert.NotNil(t, p.Bindings["onBooted"])
}

func TestRegisterFunction_Missing(t *testing.T) {
	p := newTestPlugin(t)
	p.RegisterFunction("nonExistent")
	assert.Nil(t, p.Bindings["nonExistent"])
}

func TestRegisterFunctions(t *testing.T) {
	p := newTestPlugin(t)
	_, err := p.VM.RunString(`
		function onBooted(ctx, req) { return req; }
		function onRun(ctx, req) { return req; }
	`)
	require.NoError(t, err)

	p.RegisterFunctions([]string{"onBooted", "onRun", "onShutdown"})
	assert.NotNil(t, p.Bindings["onBooted"])
	assert.NotNil(t, p.Bindings["onRun"])
	assert.Nil(t, p.Bindings["onShutdown"])
}

func TestRunFunction_Success(t *testing.T) {
	p := newTestPlugin(t)
	_, err := p.VM.RunString(`function onBooted(ctx, req) { return req; }`)
	require.NoError(t, err)
	p.RegisterFunction("onBooted")

	req := newTestRequest(t)
	result, err := p.RunFunction(context.Background(), "onBooted", req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "value", result.AsMap()["key"])
}

func TestRunFunction_NotFound(t *testing.T) {
	p := newTestPlugin(t)
	req := newTestRequest(t)

	result, err := p.RunFunction(context.Background(), "nonExistent", req)
	assert.NoError(t, err)
	assert.Equal(t, req, result)
}

func TestRunFunction_JSError(t *testing.T) {
	p := newTestPlugin(t)
	_, err := p.VM.RunString(`function onBooted(ctx, req) { throw new Error("test error"); }`)
	require.NoError(t, err)
	p.RegisterFunction("onBooted")

	req := newTestRequest(t)
	result, err := p.RunFunction(context.Background(), "onBooted", req)
	assert.Error(t, err)
	assert.Equal(t, req, result)
}

func TestRunFunction_WrongReturnType(t *testing.T) {
	p := newTestPlugin(t)
	_, err := p.VM.RunString(`function onBooted(ctx, req) { return 42; }`)
	require.NoError(t, err)
	p.RegisterFunction("onBooted")

	req := newTestRequest(t)
	result, err := p.RunFunction(context.Background(), "onBooted", req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected *v1.Struct")
	assert.Equal(t, req, result)
}

func TestGetHooks(t *testing.T) {
	p := newTestPlugin(t)
	_, err := p.VM.RunString(`
		function onBooted(ctx, req) { return req; }
		function onRun(ctx, req) { return req; }
	`)
	require.NoError(t, err)
	p.RegisterFunctions([]string{"onBooted", "onRun", "onShutdown"})

	hooks := p.GetHooks()
	assert.Len(t, hooks, 2)
}

func TestGetPluginConfig(t *testing.T) {
	p := newTestPlugin(t)
	config, err := p.GetPluginConfig(context.Background(), nil)
	assert.NoError(t, err)
	assert.NotNil(t, config)

	configMap := config.AsMap()
	assert.Equal(t, PluginConfig["description"], configMap["description"])
	assert.Equal(t, PluginConfig["license"], configMap["license"])
}

// hookTestCase maps a hook method to the JS function name it should dispatch to.
type hookTestCase struct {
	name   string
	jsFunc string
	call   func(*Plugin, context.Context, *v1.Struct) (*v1.Struct, error)
}

func allHookTestCases() []hookTestCase {
	return []hookTestCase{
		{"OnConfigLoaded", "onConfigLoaded", (*Plugin).OnConfigLoaded},
		{"OnNewLogger", "onNewLogger", (*Plugin).OnNewLogger},
		{"OnNewPool", "onNewPool", (*Plugin).OnNewPool},
		{"OnNewClient", "onNewClient", (*Plugin).OnNewClient},
		{"OnNewProxy", "onNewProxy", (*Plugin).OnNewProxy},
		{"OnNewServer", "onNewServer", (*Plugin).OnNewServer},
		{"OnSignal", "onSignal", (*Plugin).OnSignal},
		{"OnRun", "onRun", (*Plugin).OnRun},
		{"OnBooting", "onBooting", (*Plugin).OnBooting},
		{"OnBooted", "onBooted", (*Plugin).OnBooted},
		{"OnOpening", "onOpening", (*Plugin).OnOpening},
		{"OnOpened", "onOpened", (*Plugin).OnOpened},
		{"OnClosing", "onClosing", (*Plugin).OnClosing},
		{"OnClosed", "onClosed", (*Plugin).OnClosed},
		{"OnTraffic", "onTraffic", (*Plugin).OnTraffic},
		{"OnShutdown", "onShutdown", (*Plugin).OnShutdown},
		{"OnTick", "onTick", (*Plugin).OnTick},
		{"OnTrafficFromClient", "onTrafficFromClient", (*Plugin).OnTrafficFromClient},
		{"OnTrafficToServer", "onTrafficToServer", (*Plugin).OnTrafficToServer},
		{"OnTrafficFromServer", "onTrafficFromServer", (*Plugin).OnTrafficFromServer},
		{"OnTrafficToClient", "onTrafficToClient", (*Plugin).OnTrafficToClient},
	}
}

func TestHookMethods_PassthroughWithoutJS(t *testing.T) {
	for _, tc := range allHookTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			p := newTestPlugin(t)
			p.Bindings[tc.jsFunc] = nil

			req := newTestRequest(t)
			result, err := tc.call(p, context.Background(), req)
			assert.NoError(t, err)
			assert.Equal(t, req, result)
		})
	}
}

func TestHookMethods_DispatchToCorrectJSFunction(t *testing.T) {
	for _, tc := range allHookTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			p := newTestPlugin(t)

			// Register a JS function that tags the request with the function name.
			script := `function ` + tc.jsFunc + `(ctx, req) {
				req.Fields["calledBy"] = Value("` + tc.jsFunc + `");
				return req;
			}`
			_, err := p.VM.RunString(script)
			require.NoError(t, err)

			err = p.VM.Set("Value", p.VM.ToValue(v1.NewValue))
			require.NoError(t, err)

			p.RegisterFunction(tc.jsFunc)
			assert.NotNil(t, p.Bindings[tc.jsFunc], "JS function %q should be registered", tc.jsFunc)

			req := newTestRequest(t)
			result, err := tc.call(p, context.Background(), req)
			assert.NoError(t, err)
			assert.NotNil(t, result)

			calledBy, ok := result.AsMap()["calledBy"]
			assert.True(t, ok, "result should contain calledBy field")
			assert.Equal(t, tc.jsFunc, calledBy)
		})
	}
}

func TestNewJSPlugin(t *testing.T) {
	p := newTestPlugin(t)
	jsp := NewJSPlugin(p)
	assert.NotNil(t, jsp)
	assert.Equal(t, p.Logger, jsp.Impl.Logger)
}
