# iobeam Command-line Interface #

**[iobeam](http://iobeam.com)** is a data platform for connected devices. 

This is a Command-line Interface (CLI) for the **iobeam API**. For more 
information on the iobeam API, please read our [full API documentation](http://docs.iobeam.com).

*Please note that we are currently invite-only. You will need an invite 
to generate a valid token and use our APIs. (Want an invite? Sign up [here](http://iobeam.com).)*

## Installation ##

Install `Go` if not already installed (e.g., `brew install go`)

Then,

    $ mkdir $GOPATH/src/iobeam.com
    $ cd $GOPATH/src/iobeam.com
    $ git clone https://github.com/iobeam/iobeam.git
    $ cd iobeam && go install
    
Note: A dot directory, `.iobeam`, is created under your user's home directory to
store state such as user and project tokens which authenticate you to the iobeam cloud.

## Getting Started ##

### Creating your first project and device ###

    # Register as a new user
    $ iobeam user create -email="<email>" -password="<password>"

    # Create a new project. Project name must be globally unique. (Keep track of the project_id that the API returns.)
    $ iobeam project create -name="<project_name>"
    
    # Get a project token with read/write/admin access
    $ iobeam token proj -id=<project_id> -admin=true -read=true -write=true

    # Create a new device. (Keep track of the device_id that the API returns.)
    $ iobeam device create -projectId=<project_id>
    
### Sending data ###

Please refer to our [Imports API](http://docs.iobeam.com/imports).

### Querying data ###

(Note: All of these commands require that you first create a valid project token with read access.)

    # Query all device data under a given project
    $ iobeam export -projectId=<project_id>

    # Query all data streams under a given project and device
    $ iobeam export -projectId=<project_id> -deviceId="<device_id>"
    
    # Query a specific data stream under a given project and device
    $ iobeam export -projectId=<project_id> -deviceId="<device_id>" -series="<series_name>"

The REST API also supports richer queries with operators (e.g., `mean`, `min`, `max`), date / value
ranges, time-series rollups, and more. Please refer to our [Exports API](http://docs.iobeam.com/exports/) 
for more information.