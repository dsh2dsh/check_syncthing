[![Go](https://github.com/dsh2dsh/check_syncthing/actions/workflows/go.yml/badge.svg)](https://github.com/dsh2dsh/check_syncthing/actions/workflows/go.yml)

-------------------------------------------------------------------------------

# Icinga2 monitoring plugin for [syncthing] daemon.

This plugin monitors syncthing daemon by using its [REST API]. Inspired by [bn8]
and [vlcty] projects.

[syncthing]:https://github.com/syncthing/syncthing
[REST API]:https://docs.syncthing.net/dev/rest.html
[bn8]:https://gitea.zionetrix.net/bn8/check_syncthing.git
[vlcty]:https://github.com/vlcty/check_syncthing

FreeBSD port [here](https://github.com/dsh2dsh/freebsd-ports/tree/master/net-mgmt/check_syncthing)

## Usage

```
$ check_syncthing
This plugin monitors syncthing daemon by using its REST API.

Requires server URL and API key using flags or environment variables
SYNCTHING_API_KEY and SYNCTHING_URL. Environment variables can be configured
inside .env file in current dir.

Usage:
  check_syncthing [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  health      Check health of syncthing server
  help        Help about any command

Flags:
  -h, --help         help for check_syncthing
  -k, --key string   syncthing REST API key
  -u, --url string   server URL

Use "check_syncthing [command] --help" for more information about a command.
```

```
$ check_syncthing health -h
Check health of syncthing server.

Checks syncthing servers handles REST API requests, has no system errors and no
folders with errors.

In case of errors, outputs last system error and last error for every folder
with errors.

Usage:
  check_syncthing health [flags]

Flags:
  -h, --help   help for health

Global Flags:
  -k, --key string   syncthing REST API key
  -u, --url string   server URL

$ check_syncthing health -u http://127.0.0.1:8384 -k XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
OK: syncthing server alive
```

## Icinga2 configuration examples

```
object CheckCommand "check_syncthing" {
  command = [ PluginDir + "/check_syncthing" ]

  arguments = {
    "--cmd" = {
      value = "$syncthing_cmd$"
      order = -1
      skip_key = true
    }
    "-u" = {
      value = "$syncthing_url$"
      required = true
    }
    "-k" = {
      value = "$syncthing_key$"
    }
  }

  env.SYNCTHING_API_KEY = SyncthingKey

  vars.syncthing_cmd = "health"
  vars.syncthing_url = "http://$address$:8384"
}
```

```
apply Service "syncthing_" for (item => cfg in host.vars.syncthing) {
  import "generic-service"

  check_command = "check_syncthing"

  vars.grafana_graph_disable = true
  vars += cfg

  assign where host.vars.syncthing
}
```

```
object Host "server" {
  vars.syncthing["health"] = {}
}
```
