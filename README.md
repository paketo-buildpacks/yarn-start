# Paketo Buildpack for Yarn Start

## `docker.io/paketobuildpacks/yarn-start`

The Yarn Start CNB sets the start command for the given application. The start
command is generated from the contents of `package.json`. For example, given a
`package.json` with the following content:

```json
{
  "scripts": {
    "prestart": "<prestart-command>",
    "poststart": "<poststart-command>",
    "start": "<start-command>"
  }
}
```

The start command will be `<prestart-command> && <start-command> && <poststart-command>`.

## Enabling reloadable process types

You can configure this buildpack to wrap the entrypoint process of your app
such that it kills and restarts the process whenever files change in the app's working
directory in the container. With this feature enabled, copying new
verisons of source code into the running container will trigger your app's
process to restart. Set the environment variable `BP_LIVE_RELOAD_ENABLED=true`
at build time to enable this feature.

## Integration

This CNB sets a start command, so there's currently no scenario we can
imagine that you would need to require it as dependency.

## Usage

To package this buildpack for consumption:

```
$ ./scripts/package.sh --version <version-number>
```

This will create a `buildpackage.cnb` file under the `build` directory which you
can use to build your app as follows:
```
pack build <app-name> -p <path-to-app> -b <path/to/node-engine.cnb> -b \
<path/to/yarn.cnb> -b <path/to/yarn-install.cnb> -b build/buildpackage.cnb
```

## Graceful shutdown and signal handling

You can add signal handlers in your app to support graceful shutdown and
program interrupts. If running a node server is the start command, the
buildpack runs the node server as the init process, and thus it ignores any
signal with the default action. As a result, the process will not terminate on
`SIGINT` or `SIGTERM` unless it is coded to do so. You can also use docker's
`--init` flag to wrap your node process with an init system that will properly
handle signals.

## Specifying a project path

To specify a project subdirectory to be used as the root of the app, please use
the `BP_NODE_PROJECT_PATH` environment variable at build time either directly
(ex. `pack build my-app --env BP_NODE_PROJECT_PATH=./src/my-app`) or through a
[`project.toml`
file](https://github.com/buildpacks/spec/blob/main/extensions/project-descriptor.md).
This could be useful if your app is a part of a monorepo.
