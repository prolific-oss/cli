# Prolificli

![GitHub Badge](https://github.com/benmatselby/prolificli/workflows/Go/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/benmatselby/prolificli)](https://goreportcard.com/report/github.com/benmatselby/prolificli)

CLI application for getting information out of [Prolific](https://www.prolific.co)

```text
CLI application for retrieving data from Prolific

Usage:
  prolificli [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  study       Study related commands
  user        User related commands

Flags:
      --config string   config file (default is $HOME/.benmatselby/prolificli.yaml)
  -h, --help            help for prolificli

Use "prolificli [command] --help" for more information about a command.
```

## Requirements

If you are wanting to build and develop this, you will need the following items installed. If, however, you just want to run the application I recommend using the docker container (See below).

- Go version 1.18+

## Configuration

### Environment variables

You will need the following environment variables defining:

```shell
export PROLIFIC_URL=""
export PROLIFIC_TOKEN=""
```

[Creating a Prolific token](https://www.prolific.co/developers).

## Installation via Docker

Other than requiring [docker](http://docker.com) to be installed, there are no other requirements to run the application this way.

```shell
$ docker build -t benmatselby/prolificli .
$ docker run \
  --rm \
  -t \
  -ePROLIFIC_URL \
  -ePROLIFIC_TOKEN \
  -v "${HOME}/.benmatselby":/root/.benmatselby \
  benmatselby/prolificli:latest "$@"
```

The `latest` tag mentioned above can be changed to a released version. For all releases, see [here](https://hub.docker.com/repository/docker/benmatselby/prolificli/tags). An example would then be:

```shell
benmatselby/prolificli:version-2.2.0
```

This would use the `verson-2.2.0` release in the docker command.

## Installation via Git

```shell
git clone git@github.com:benmatselby/prolificli.git
cd prolificli
make all
./prolificli
```

You can also install into your `$GOPATH/bin` by running `make build && go install`.
