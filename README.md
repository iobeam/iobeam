# iobeam Command-line Interface #

**[iobeam](http://iobeam.com)** is a data platform for connected devices. 

This is a Command-line Interface (CLI) for the **iobeam API**. For more 
information on the iobeam API, please read our [full API documentation](http://docs.iobeam.com).

*Please note that we are currently invite-only. You will need an invite 
to generate a valid token and use our APIs. (Want an invite? Sign up [here](http://iobeam.com).)*

## Installation ##

Install `Go` if not already installed (e.g., `brew install go` on Mac OS X). Make sure you
set your `GOPATH` (e.g., export GOPATH=~/go) and that the `GOPATH` exists.

Then,

    $ go get github.com/iobeam/iobeam
    
Note: A dot directory, `.iobeam`, is created under your user's home directory to
store state such as user and project tokens which authenticate you to the iobeam cloud.

**Note: If you had this installed previously, you may need to move it. Its old path was
`$GOPATH/src/iobeam.com/iobeam`, you will need to move this to `$GOPATH/src/github.com/iobeam/iobeam`.**

## Getting Started ##

### Creating your first project and device ###

    # Register as a new user, this will automatically log you in.
    $ iobeam user create -email="<email>" -password="<password>"

    # Optionally, if you have an account already, you can just login
    $ iobeam user login -email="<email>" -password="<password>"

    # Create a new project. Project name must be globally unique.
    # You will be given a project ID (keep track of it), and a project token will be stored.
    $ iobeam project create -name="<project_name>"

    # Create a new device. (Keep track of the device_id that the API returns.)
    $ iobeam device create -projectId=<project_id>
    
### Sending data ###

Please refer to our [Imports API](http://docs.iobeam.com/imports).

### Querying data ###

*Note: All of these commands require that you first create a valid project token with read access.
This will be created when you create the project, however if this does not work for some reason, see
the next section.*

    # Query all device data under a given project
    $ iobeam export -projectId=<project_id>

    # Query all data streams under a given project and device
    $ iobeam export -projectId=<project_id> -deviceId="<device_id>"
    
    # Query a specific data stream under a given project and device
    $ iobeam export -projectId=<project_id> -deviceId="<device_id>" -series="<series_name>"

The REST API also supports richer queries with operators (e.g., `mean`, `min`, `max`), date / value
ranges, time-series rollups, and more. Please refer to our [Exports API](http://docs.iobeam.com/exports/) 
for more information.

### Creating additional project tokens ###

When you create a project, the token you are given has admin privileges, which you will not want to
distribute with devices going to third parties. Instead, you can generate a new token that
allows you to upload data (i.e. has write permissions) but not admin permissions.

    # Generating a read/write only token
    $ iobeam token project -id=<project_id> -read=true -write=true -admin=false

    # Generating a write only token
    $ iobeam token project -id=<project_id> -read=false -write=true -admin=false

