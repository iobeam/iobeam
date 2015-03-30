# iobeam Command-line Interface #

A command-line interface (CLI) for the iobeam API.


## Installation ##

Install Go, if not already installed (e.g., 'brew install go').

    $ mkdir $GOPATH/src/iobeam.com
    $ cd $GOPATH/src/iobeam.com
    $ git clone git@bitbucket.org:440-labs/beam.git
    $ cd beam && go install
    
## Usage ##

`iobeam` allows you to manage your projects, devices, and data in the iobeam
cloud. A dot directory, `.iobeam`, is created under your user's home directory
to store state such as user and project tokens which authenticate you to the
iobeam cloud.