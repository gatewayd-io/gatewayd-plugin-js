package plugin

import (
	sdkConfig "github.com/gatewayd-io/gatewayd-plugin-sdk/config"
	v1 "github.com/gatewayd-io/gatewayd-plugin-sdk/plugin/v1"
	goplugin "github.com/hashicorp/go-plugin"
)

var (
	Version  = "0.0.1"
	PluginID = v1.PluginID{
		Name:      "gatewayd-plugin-js",
		Version:   Version,
		RemoteUrl: "github.com/gatewayd-io/gatewayd-plugin-js",
	}
	PluginMap = map[string]goplugin.Plugin{
		"gatewayd-plugin-js": &JSPlugin{},
	}
	// TODO: Handle this in a better way
	// https://github.com/gatewayd-io/gatewayd-plugin-sdk/issues/3
	PluginConfig = map[string]interface{}{
		"id": map[string]interface{}{
			"name":      PluginID.GetName(),
			"version":   PluginID.GetVersion(),
			"remoteUrl": PluginID.GetRemoteUrl(),
		},
		"description": "GatewayD plugin for running JavaScript functions as hooks",
		"authors": []interface{}{
			"Mostafa Moradian <mostafa@gatewayd.io>",
		},
		"license":    "Apache 2.0",
		"projectUrl": "https://github.com/gatewayd-io/gatewayd-plugin-js",
		// Compile-time configuration
		"config": map[string]interface{}{
			"metricsEnabled": sdkConfig.GetEnv("METRICS_ENABLED", "true"),
			"metricsUnixDomainSocket": sdkConfig.GetEnv(
				"METRICS_UNIX_DOMAIN_SOCKET", "/tmp/gatewayd-plugin-js.sock"),
			"metricsEndpoint": sdkConfig.GetEnv("METRICS_ENDPOINT", "/metrics"),
			"scriptPath":      sdkConfig.GetEnv("SCRIPT_PATH", "./scripts/index.js"),
		},
		"hooks":      []interface{}{},
		"tags":       []interface{}{"plugin", "javascript", "js"},
		"categories": []interface{}{"builtin"},
	}
	Hooks = map[string]v1.HookName{
		"onConfigLoaded":      v1.HookName_HOOK_NAME_ON_CONFIG_LOADED,
		"onNewLogger":         v1.HookName_HOOK_NAME_ON_NEW_LOGGER,
		"onNewPool":           v1.HookName_HOOK_NAME_ON_NEW_POOL,
		"onNewClient":         v1.HookName_HOOK_NAME_ON_NEW_CLIENT,
		"onNewProxy":          v1.HookName_HOOK_NAME_ON_NEW_PROXY,
		"onNewServer":         v1.HookName_HOOK_NAME_ON_NEW_SERVER,
		"onSignal":            v1.HookName_HOOK_NAME_ON_SIGNAL,
		"onRun":               v1.HookName_HOOK_NAME_ON_RUN,
		"onBooting":           v1.HookName_HOOK_NAME_ON_BOOTING,
		"onBooted":            v1.HookName_HOOK_NAME_ON_BOOTED,
		"onOpening":           v1.HookName_HOOK_NAME_ON_OPENING,
		"onOpened":            v1.HookName_HOOK_NAME_ON_OPENED,
		"onClosing":           v1.HookName_HOOK_NAME_ON_CLOSING,
		"onClosed":            v1.HookName_HOOK_NAME_ON_CLOSED,
		"onTraffic":           v1.HookName_HOOK_NAME_ON_TRAFFIC,
		"onTrafficFromClient": v1.HookName_HOOK_NAME_ON_TRAFFIC_FROM_CLIENT,
		"onTrafficToServer":   v1.HookName_HOOK_NAME_ON_TRAFFIC_TO_SERVER,
		"onTrafficFromServer": v1.HookName_HOOK_NAME_ON_TRAFFIC_FROM_SERVER,
		"onTrafficToClient":   v1.HookName_HOOK_NAME_ON_TRAFFIC_TO_CLIENT,
		"onShutdown":          v1.HookName_HOOK_NAME_ON_SHUTDOWN,
		"onTick":              v1.HookName_HOOK_NAME_ON_TICK,
	}
)
