# Gravatar Plugin for [Kiwi IRC] (https://kiwiirc.com)

This plugin adds gravatars to kiwiirc using a webircgateway plugin to make server side sql queries

Its been designed to work with Anope but will likely work with other services databases too

#### Dependencies
* node (https://nodejs.org/)
* yarn (https://yarnpkg.com/)

#### Building and installing

1. Build the plugin

   ```console
   $ yarn
   $ yarn build
   ```

   The plugin will then be created at `dist/plugin-gravatar.js`

2. Copy the plugin to your Kiwi webserver

   The plugin file must be loadable from a webserver. Creating a `plugins/` folder with your KiwiIRC files is a good place to put it.

3. Add the plugin to KiwiIRC

   In your kiwi `config.json` file, find the `plugins` section and add:
   ```json
   {"name": "gravatar", "url": "/plugins/plugin-gravatar.js"}
   ```

#### Configuration

```
"plugin-gravatar": {
    "gatewayURL", "//localhost:8001",
    "gravatarURL", "https://www.gravatar.com/avatar/",
    "gravatarRating", "g",
    "gravatarFallback": "robohash"
}
```

## License

[Licensed under the Apache License, Version 2.0](LICENSE).
