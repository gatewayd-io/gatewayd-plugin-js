<p align="center">
  <a href="https://docs.gatewayd.io/plugins/gatewayd-plugin-js">
    <picture>
      <img alt="gatewayd-plugin-js-logo" src="https://github.com/gatewayd-io/gatewayd-plugin-js/blob/main/assets/gatewayd-plugin-js-logo.png" width="96" />
    </picture>
  </a>
  <h3 align="center">gatewayd-plugin-js</h3>
  <p align="center">GatewayD plugin for running JS functions as hooks.</p>
</p>

<p align="center">
    <a href="https://github.com/gatewayd-io/gatewayd-plugin-js/releases">Download</a> ·
    <a href="https://docs.gatewayd.io/plugins/gatewayd-plugin-js">Documentation</a>
</p>

## Features

- Run JS functions as hooks
- Helper functions for common tasks such as parsing incoming queries
- Support for running multiple JS functions as hooks
- Prometheus metrics for monitoring

## Build for testing

To build the plugin for development and testing, run the following command:

```bash
make build-dev
```

Running the above command causes the `go mod tidy` and `go build` to run for compiling and generating the plugin binary in the current directory, named `gatewayd-plugin-js`.

> [!WARNING]
> This plugin is experimental and is not recommended for production use. This is unless you know what you are doing.

