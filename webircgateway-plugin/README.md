# Gravatar Plugin for [Kiwi IRC] (https://kiwiirc.com)

This is the server side for plugin-gravatar. It can run as a webircgateway plugin or as a standalone server

## Webircgateway plugin

To run this as a webircgateway plugin copy `plugin.go` to `webircgateway/plugins/gravatar/plugin.go` before building.

Follow webircgateway instructions for building with plugins

to specifiy config loading location you can use `--config-gravatar /etc/kiwiirc/gravatar.config.json` as a run param

_note: dont forget to set "salt" in `gravatar.config.json`_

## Standalone

To build this as a standalone service

```
go build -o gravatar-service *.go
```

you will also need to add some items to gravatar.config.json

```
"listen_addr": "127.0.0.1:4646",
"allow_origins": ["*"]
```
