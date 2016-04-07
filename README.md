# iobeam Command-line Interface

**[iobeam](http://iobeam.com)** is a data platform for connected devices.

This is a Command-line Interface (CLI) for the **iobeam API**. For more
information on the iobeam API, please read our [full API documentation](http://docs.iobeam.com).

*Please note that we are currently invite-only. You will need an invite
to generate a valid token and use our APIs. (Want an invite? Sign up [here](http://iobeam.com).)*

## Installation

Included with each release are [binary executables](https://github.com/iobeam/iobeam/releases)
for OSX (darwin), Linux, and Windows. **This is the easiest way** to use the
CLI: simply download the one that corresponds to your platform, rename it
`iobeam`, and make sure it is executable.

For OSX users, it is also available via Homebrew:
```sh
$ brew tap iobeam/tap
$ brew install iobeam
```

### Building and installing from source

You'll need to install [Go](https://golang.org/) (e.g., `brew install go` on
Mac OSX). Make sure your `GOPATH` is set (e.g., `export GOPATH=~/go`).
We recommend using the latest version of Go, but at least version 1.4.3.
Older versions may have problems.

Then,
```sh
$ go get github.com/iobeam/iobeam
```

## Getting Started

On first run, a dot directory, `.iobeam`, is created in your home
directory to store state such as user and project tokens which authenticate you to the iobeam cloud.

### Creating your first project and device
```sh
# Register as a new user, this will automatically log you in.
$ iobeam user create -email="<email>" -password="<password>" -invite="<invite_code>"

# Optionally, if you have an account already, you can just login.
# You will be prompted for your credentials
$ iobeam user login

# Create a new project. Project name must be globally unique.
# You will be given a project ID (keep track of it), and a
# project token will be stored.
$ iobeam project create -name="<project_name>"

# Create a new device. (Keep track of the device_id that the API returns.)
$ iobeam device create -projectId=<project_id>
```

### Sending data

You can send single data points via the CLI. Timestamps are expressed as milliseconds since
epoch.
```sh
# Send data point of value 12.5 with the current time
$ iobeam import -projectId=<projectId> -deviceId=<deviceId> -series=<series name> -value=12.5

# Send data point with value 12.5 at time 1429718512829
$ iobeam import -projectId=<projectId> -deviceId=<deviceId> -series=<series name> \
    -time=1429718512829 -value=12.5

# Optionally, you can leave the -projectId off and it will default to the
# last project you got a token for
$ iobeam import -deviceId=<deviceId> -series=<series name> -value=12.5
```
You can also refer to our [Imports API](http://docs.iobeam.com/imports).

### Querying data

*Note: All of these commands require that you first create a valid project token with read access.
This will be created when you create the project, however if this does not work for some reason, see
the next section.*
```sh
# Query all device data under a given project
$ iobeam query -projectId=<project_id>

# You can also leave off -projectId, which will use the projectId of the
# last project you got a token for.
$ iobeam query

# Query all data streams under a given project and device
$ iobeam query -projectId=<project_id> -deviceId="<device_id>"

# Query a specific data stream under a given project and device
$ iobeam query -projectId=<project_id> -deviceId="<device_id>" -series="<series_name>"

# Query multiple series under a given project
$ iobeam query -projectId=<project_id> -series="<series_name1>" -series="<series_name2>"

# Query a specific data stream over the last day
$ iobeam query -projectId=<project_id> -series="<series_name>" -last="1d"
```
The REST API also supports richer queries with operators (e.g., `mean`, `min`, `max`), date / value
ranges, time-series rollups, and more. Please refer to our [Exports API](http://docs.iobeam.com/api/exports/)
for more information.

### Testing triggers

If you want to make sure you have set up a trigger correctly with iobeam,
you can test it by firing a test event. This allows you to verify your
trigger is working correctly independent of any application logic. For
the simplest triggers that don't use parameters:
```sh
$ iobeam trigger test -name=<trigger_name>
```
This will cause the trigger to fire, e.g., POST to the endpoint you set up when
you made the trigger.

If your trigger uses parameters in the payload, all of them must be specified
using the `-param` flag in the form of `parameter_key,parameter_value`. So
if you have a paramter called `name` and you want to test it with a value of `Bob`:
```sh
$ iobeam trigger test -name=<trigger_name> -param="name,Bob"
```

Multiple parameters can also be specified:
```sh
$ iobeam trigger test -name=<trigger_name> -param="name,Bob" -param="age,20"
```

### Creating additional project tokens

When you create a project, the token you are given has admin privileges, which you will not want to
distribute with devices going to third parties. Instead, you can generate a new token that
allows you to upload data (i.e. has write permissions) but not admin permissions.
```sh
# Generating a read/write only token
$ iobeam project token -id=<project_id> -read=true -write=true -admin=false

# Generating a write only token
$ iobeam project token -id=<project_id> -read=false -write=true -admin=false
```
