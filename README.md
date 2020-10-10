![Logo](http://svg.wiersma.co.za/glasslabs/looking-glass)

[![Go Report Card](https://goreportcard.com/badge/github.com/glasslabs/looking-glass)](https://goreportcard.com/report/github.com/glasslabs/looking-glass)
[![Build Status](https://travis-ci.com/glasslabs/looking-glass.svg?branch=master)](https://travis-ci.com/glasslabs/looking-glass)
[![Coverage Status](https://coveralls.io/repos/github/glasslabs/looking-glass/badge.svg?branch=master)](https://coveralls.io/github/glasslabs/looking-glass?branch=master)
[![GitHub release](https://img.shields.io/github/release/glasslabs/looking-glass.svg)](https://github.com/glasslabs/looking-glass/releases)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/glasslabs/looking-glass/master/LICENSE)

Smart mirror platform written in Go leveraging Yaegi.

## Usage

### Run

Runs looking glass using the specified config and modules path.

```bash
glass run -c /path/to/config.yaml -m /path/to/modules
```

| Command               | Env        | Description                                                                  |
|-----------------------|------------|------------------------------------------------------------------------------|
|`--secrets`/`-s` FILE  | SECRETS    | The path to the secrets file.                                                |
| `--config`/`-c` FILE  | CONFIG     | The path to the configuration file.                                          |
| `--modules`/`-m` PATH | MODULES    | The path to the modules. Modules will be located under `src` in this folder. |
| `--log.format` FORMAT | LOG_FORMAT | Specify the format of logs. Supported formats: 'logfmt', 'json'              |
| `--log.level` LEVEL   | LOG_LEVEL  | Specify the log level. E.g. 'debug', 'warning'. (default: "info")            |

## Configuration

```yaml
ui:
  width:  640           # The width of the chrome window
  height: 480           # The height of the chrome window
  fullscreen: true      # If the chrome window should start fullscreen
modules:
  - name: simple-clock      # The name of the module (must be unique)
    path: clock             # The path to the module
    position: top:right     # The position of the module
  - name: simple-weather
    path: weather
    position: top:left
    config:
      locationId: 996506
      appId: {{ .Secrets.weather.appId }}
      units: metric
```

The module configuration can contain secrets from the secrets yaml prefixed with `.Secrets`
as shown in the example above. 

## Modules

### Package Naming

If your module uses a hyphen, which is not supported by Go, it will be assumed that the package name is the
last path of the hyphenated name (e.g. `looking-glass` would result in a package name `glass`). If this is not
the case for your module, the `Package` should be set in the module cofiguration to the correct package name and
should be documented in your module.

### Development

Modules are parsed in [yaegi](http://github.com/traefik/yaegi) and must expose two functions to be loaded:

#### NewConfig

`NewConfig` exposes your configuration structure to looking glass. The function must return
a single structure with default values set. The yaml configuration will be decoded into
the returned structure so it should contain `yaml` tags for configuration to be decoded
properly.  

```go
func NewConfig() *Config 
```

#### New

`New` creates an instance of your module. It must return an `io.Closer` and an `error.
The function takes a `context.Context`, the configuration structure returned by `NewConfig`,
`Info` and `UI` objects. `Info` and `UI` are located in `github.com/glasslabs/looking-glass/module/types`.

```go
func New(_ context.Context, cfg *Config, info types.Info, ui types.UI) (io.Closer, error)
```

## TODO

This is very much a work in progress and under active development. The immediate list of
things to do is below:

* Better documentation
* Module download from config
* Better test coverage