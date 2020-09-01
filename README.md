# Yarn Start Cloud Native Buildpack

## `gcr.io/paketo-buildpacks/yarn-start`

The Yarn Start CNB sets the start command for the given application.

## Integration

This CNB writes a command, so there's currently no scenario we can
imagine that you would need to require it as dependency.

To package this buildpack for consumption:
```
$ ./scripts/package.sh
```
This builds the buildpack's source using GOOS=linux by default. You can supply another value as the first argument to package.sh.

## `buildpack.yml` Configurations

There are no extra configurations for this buildpack based on `buildpack.yml`.
