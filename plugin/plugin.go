package plugin

import (
	"context"

	"github.com/dop251/goja"
	v1 "github.com/gatewayd-io/gatewayd-plugin-sdk/plugin/v1"
	"github.com/hashicorp/go-hclog"
	goplugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

type Plugin struct {
	goplugin.GRPCPlugin
	v1.GatewayDPluginServiceServer
	Logger   hclog.Logger
	VM       *goja.Runtime
	Bindings map[string]goja.Callable
}

type JSPlugin struct {
	goplugin.NetRPCUnsupportedPlugin
	Impl Plugin
}

func (p *Plugin) RegisterFunction(name string) {
	function, ok := goja.AssertFunction(p.VM.Get(name))
	if !ok {
		p.Logger.Trace("Cannot register function, because it doesn't exist", "name", name)
		p.Bindings[name] = nil
		return
	}

	p.Logger.Trace("Registering function", "name", name)
	p.Bindings[name] = function
}

func (p *Plugin) RegisterFunctions(names []string) {
	for _, name := range names {
		p.RegisterFunction(name)
	}
}

func (p *Plugin) RunFunction(name string, ctx context.Context, req *v1.Struct) (*v1.Struct, error) {
	if p.Bindings[name] == nil {
		p.Logger.Debug("RunFunction", "err", "function not found")
		return req, nil
	}

	jsReq, err := p.Bindings[name](goja.Undefined(), p.VM.ToValue(ctx), p.VM.ToValue(req))
	if err != nil {
		p.Logger.Error("OnTrafficFromClient", "err", err)
		return req, err
	}

	return jsReq.Export().(*v1.Struct), err
}

func (p *Plugin) GetHooks() []interface{} {
	hooks := []interface{}{}
	for name := range p.Bindings {
		if p.Bindings[name] != nil {
			hooks = append(hooks, int32(Hooks[name]))
		}
	}
	return hooks
}

// GRPCServer registers the plugin with the gRPC server.
func (p *JSPlugin) GRPCServer(b *goplugin.GRPCBroker, s *grpc.Server) error {
	v1.RegisterGatewayDPluginServiceServer(s, &p.Impl)
	return nil
}

// GRPCClient returns the plugin client.
func (p *JSPlugin) GRPCClient(ctx context.Context, b *goplugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return v1.NewGatewayDPluginServiceClient(c), nil
}

// NewJSPlugin returns a new instance of the TestPlugin.
func NewJSPlugin(impl Plugin) *JSPlugin {
	return &JSPlugin{
		NetRPCUnsupportedPlugin: goplugin.NetRPCUnsupportedPlugin{},
		Impl:                    impl,
	}
}

// GetPluginConfig returns the plugin config. This is called by GatewayD
// when the plugin is loaded. The plugin config is used to configure the
// plugin.
func (p *Plugin) GetPluginConfig(
	ctx context.Context, _ *v1.Struct) (*v1.Struct, error) {
	GetPluginConfig.Inc()

	// Get plugin hooks from registered JS functions.
	// This prevents the plugin to hook to certain function if no requivalent
	// JS functions are registered.
	PluginConfig["hooks"] = p.GetHooks()
	pluginConfig, err := v1.NewStruct(PluginConfig)

	return pluginConfig, err
}

// OnConfigLoaded is called when the global config is loaded by GatewayD.
// This can be used to modify the global config. Note that the plugin config
// cannot be modified via plugins.
func (p *Plugin) OnConfigLoaded(ctx context.Context, req *v1.Struct) (*v1.Struct, error) {
	OnConfigLoaded.Inc()
	p.Logger.Debug("OnConfigLoaded", "req", req)
	// The JS function MUST return the request object, which is a *v1.Struct.
	req, err := p.RunFunction("onConfigLoaded", ctx, req)
	p.Logger.Debug("OnConfigLoaded", "req", req.AsMap(), "err", err)
	return req, err
}

// OnNewLogger is called when a new logger is created by GatewayD.
// This is a notification and the plugin cannot modify the logger.
func (p *Plugin) OnNewLogger(ctx context.Context, req *v1.Struct) (*v1.Struct, error) {
	OnNewLogger.Inc()
	p.Logger.Debug("OnNewLogger", "req", req)
	req, err := p.RunFunction("onNewLogger", ctx, req)
	p.Logger.Debug("OnNewLogger", "req", req.AsMap(), "err", err)
	return req, err
}

// OnNewPool is called when a new pool is created by GatewayD.
// This is a notification and the plugin cannot modify the pool.
func (p *Plugin) OnNewPool(ctx context.Context, req *v1.Struct) (*v1.Struct, error) {
	OnNewPool.Inc()
	p.Logger.Debug("OnNewPool", "req", req)
	req, err := p.RunFunction("onConfigLoaded", ctx, req)
	p.Logger.Debug("OnNewPool", "req", req.AsMap(), "err", err)
	return req, err
}

// OnNewClient is called when a new client is created by GatewayD.
// This is a notification and the plugin cannot modify the client.
func (p *Plugin) OnNewClient(ctx context.Context, req *v1.Struct) (*v1.Struct, error) {
	OnNewClient.Inc()
	p.Logger.Debug("OnNewClient", "req", req)
	req, err := p.RunFunction("onNewClient", ctx, req)
	p.Logger.Debug("OnNewClient", "req", req.AsMap(), "err", err)
	return req, err
}

// OnNewProxy is called when a new proxy is created by GatewayD.
// This is a notification and the plugin cannot modify the proxy.
func (p *Plugin) OnNewProxy(ctx context.Context, req *v1.Struct) (*v1.Struct, error) {
	OnNewProxy.Inc()
	p.Logger.Debug("OnNewProxy", "req", req)
	req, err := p.RunFunction("onNewProxy", ctx, req)
	p.Logger.Debug("OnNewProxy", "req", req.AsMap(), "err", err)
	return req, err
}

// OnNewServer is called when a new server is created by GatewayD.
// This is a notification and the plugin cannot modify the server.
func (p *Plugin) OnNewServer(ctx context.Context, req *v1.Struct) (*v1.Struct, error) {
	OnNewServer.Inc()
	p.Logger.Debug("OnNewServer", "req", req)
	req, err := p.RunFunction("onNewServer", ctx, req)
	p.Logger.Debug("OnNewServer", "req", req.AsMap(), "err", err)
	return req, err
}

// OnSignal is called when a signal (for example, SIGKILL) is received by GatewayD.
// This is a notification and the plugin cannot modify the signal.
func (p *Plugin) OnSignal(ctx context.Context, req *v1.Struct) (*v1.Struct, error) {
	OnSignal.Inc()
	p.Logger.Debug("OnSignal", "req", req)
	req, err := p.RunFunction("onSignal", ctx, req)
	p.Logger.Debug("OnSignal", "req", req.AsMap(), "err", err)
	return req, err
}

// OnRun is called when GatewayD is started.
func (p *Plugin) OnRun(ctx context.Context, req *v1.Struct) (*v1.Struct, error) {
	OnRun.Inc()
	p.Logger.Debug("OnRun", "req", req)
	req, err := p.RunFunction("onRun", ctx, req)
	p.Logger.Debug("OnRun", "req", req.AsMap(), "err", err)
	return req, err
}

// OnBooting is called when GatewayD is booting.
func (p *Plugin) OnBooting(ctx context.Context, req *v1.Struct) (*v1.Struct, error) {
	OnBooting.Inc()
	p.Logger.Debug("OnBooting", "req", req)
	req, err := p.RunFunction("onBooting", ctx, req)
	p.Logger.Debug("OnBooting", "req", req.AsMap(), "err", err)
	return req, err
}

// OnBooted is called when GatewayD is booted.
func (p *Plugin) OnBooted(ctx context.Context, req *v1.Struct) (*v1.Struct, error) {
	OnBooted.Inc()
	p.Logger.Debug("OnBooted", "req", req)
	req, err := p.RunFunction("onBooted", ctx, req)
	p.Logger.Debug("OnBooted", "req", req.AsMap(), "err", err)
	return req, err
}

// OnOpening is called when a new client connection is being opened.
func (p *Plugin) OnOpening(ctx context.Context, req *v1.Struct) (*v1.Struct, error) {
	OnOpening.Inc()
	p.Logger.Debug("OnOpening", "req", req)
	req, err := p.RunFunction("onOpening", ctx, req)
	p.Logger.Debug("OnOpening", "req", req.AsMap(), "err", err)
	return req, err
}

// OnOpened is called when a new client connection is opened.
func (p *Plugin) OnOpened(ctx context.Context, req *v1.Struct) (*v1.Struct, error) {
	OnOpened.Inc()
	p.Logger.Debug("OnOpened", "req", req)
	req, err := p.RunFunction("onOpened", ctx, req)
	p.Logger.Debug("OnOpened", "req", req.AsMap(), "err", err)
	return req, err
}

// OnClosing is called when a client connection is being closed.
func (p *Plugin) OnClosing(ctx context.Context, req *v1.Struct) (*v1.Struct, error) {
	OnClosing.Inc()
	p.Logger.Debug("OnClosing", "req", req)
	req, err := p.RunFunction("onClosing", ctx, req)
	p.Logger.Debug("OnClosing", "req", req.AsMap(), "err", err)
	return req, err
}

// OnClosed is called when a client connection is closed.
func (p *Plugin) OnClosed(ctx context.Context, req *v1.Struct) (*v1.Struct, error) {
	OnClosed.Inc()
	p.Logger.Debug("OnClosed", "req", req)
	req, err := p.RunFunction("onClosed", ctx, req)
	p.Logger.Debug("OnClosed", "req", req.AsMap(), "err", err)
	return req, err
}

// OnTraffic is called when a request is being received by GatewayD from the client.
// This is a notification and the plugin cannot modify the request at this point.
func (p *Plugin) OnTraffic(ctx context.Context, req *v1.Struct) (*v1.Struct, error) {
	OnTraffic.Inc()
	p.Logger.Debug("OnTraffic", "req", req)
	req, err := p.RunFunction("onTraffic", ctx, req)
	p.Logger.Debug("OnTraffic", "req", req.AsMap(), "err", err)
	return req, nil
}

// OnShutdown is called when GatewayD is shutting down.
func (p *Plugin) OnShutdown(ctx context.Context, req *v1.Struct) (*v1.Struct, error) {
	OnShutdown.Inc()
	p.Logger.Debug("OnShutdown", "req", req)
	req, err := p.RunFunction("onShutdown", ctx, req)
	p.Logger.Debug("OnShutdown", "req", req.AsMap(), "err", err)
	return req, err
}

// OnTick is called when GatewayD is ticking (if enabled).
func (p *Plugin) OnTick(ctx context.Context, req *v1.Struct) (*v1.Struct, error) {
	OnTick.Inc()
	p.Logger.Debug("OnTick", "req", req)
	req, err := p.RunFunction("onTick", ctx, req)
	p.Logger.Debug("OnTick", "req", req.AsMap(), "err", err)
	return req, err
}

// OnTrafficFromClient is called when a request is received by GatewayD from the client.
// This can be used to modify the request or terminate the connection by returning an error
// or a response.
func (p *Plugin) OnTrafficFromClient(ctx context.Context, req *v1.Struct) (*v1.Struct, error) {
	OnTrafficFromClient.Inc()
	p.Logger.Debug("OnTrafficFromClient", "req", req.AsMap())
	req, err := p.RunFunction("onTrafficFromClient", ctx, req)
	p.Logger.Debug("OnTrafficFromClient", "req", req.AsMap(), "err", err)
	return req, err
}

// OnTrafficToServer is called when a request is sent by GatewayD to the server.
// This can be used to modify the request or terminate the connection by returning an error
// or a response while also sending the request to the server.
func (p *Plugin) OnTrafficToServer(ctx context.Context, req *v1.Struct) (*v1.Struct, error) {
	OnTrafficToServer.Inc()
	p.Logger.Debug("OnTrafficToServer", "req", req)
	req, err := p.RunFunction("onTrafficToServer", ctx, req)
	p.Logger.Debug("OnTrafficToServer", "req", req.AsMap(), "err", err)
	return req, err
}

// OnTrafficFromServer is called when a response is received by GatewayD from the server.
// This can be used to modify the response or terminate the connection by returning an error
// or a response.
func (p *Plugin) OnTrafficFromServer(
	ctx context.Context, resp *v1.Struct) (*v1.Struct, error) {
	OnTrafficFromServer.Inc()
	p.Logger.Debug("OnTrafficFromServer", "resp", resp)
	resp, err := p.RunFunction("onTrafficFromServer", ctx, resp)
	p.Logger.Debug("OnTrafficFromServer", "resp", resp.AsMap(), "err", err)
	return resp, err
}

// OnTrafficToClient is called when a response is sent by GatewayD to the client.
// This can be used to modify the response or terminate the connection by returning an error
// or a response.
func (p *Plugin) OnTrafficToClient(
	ctx context.Context, resp *v1.Struct) (*v1.Struct, error) {
	OnTrafficToClient.Inc()
	p.Logger.Debug("OnTrafficToClient", "resp", resp)
	resp, err := p.RunFunction("onTrafficToClient", ctx, resp)
	p.Logger.Debug("OnTrafficToClient", "resp", resp.AsMap(), "err", err)
	return resp, err
}
