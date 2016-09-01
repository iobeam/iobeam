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

You can send single data row via the CLI. Timestamps are expressed as milliseconds since
epoch.
```sh
# Send data row with temperature=72 and humidity=54 with the current time
$ iobeam import -fields=temperature,humidity -values=72,54 -labels device_id=<deviceId>

# Send data row for time 1429718512829 (epoch time).
$ iobeam import -fields=temperature,humidity -values=72,54 -labels device_id=<deviceId> \
    -time=1429718512829

# Optionally, you can specify the -projectId  (defaults to current project)
$ iobeam import -projectId <projectId> -labels device_id=<deviceId> \
    -fields=temperature,humidity -values=72,54 
```
You can also refer to our [Imports API](http://docs.iobeam.com/imports).

### Querying data

*Note: All of these commands require that you first create a valid project token with read access.
This will be created when you create the project, however if this does not work for some reason, see
the next section.*
```sh
# Query the last 10 data rows for a project
$ iobeam query -projectId=<project_id>

# You can also leave off -projectId, which will use the projectId of the
# last project you got a token for. 
# The limit on the number of rows can be changed from its default of 10
$ iobeam query -limit 1000

# Query the last 10 data rows for a given device
$ iobeam query -where "eq(device_id,<device_id>)" 

# Query the last row for each device (up to 1000 total rows)
$ iobeam query -limitBy "device_id,1" -limit 1000

# Query a specific data field under a given project and device
$ iobeam query -where "eq(device_id,<device_id>)" -field="<field_name>"

# Query the last 1000 data rows over the last day 
$ iobeam query -last="1d" -limit 1000
```
The REST API also supports richer queries with operators (e.g., `mean`, `min`, `max`), date / value
ranges, time-series rollups, and more. Please refer to our [Exports API](http://docs.iobeam.com/api/exports/)
for more information.

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
