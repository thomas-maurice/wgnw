# wgnw [![Build Status](https://travis-ci.org/thomas-maurice/wgnw.svg?branch=master)](https://travis-ci.org/thomas-maurice/wgnw)

wgnw is a toy project I implemented resulting from a discussion with a bunch of friends. The point is to make easy and somewhat
automatic the configuration of a meshed WireGuard network. The project works as follows:

* A central controller keeps tracks of the networks that exist, and allocate leases to clients
* Clients request and maintain leases within a certain network from the controller
* Clients poll active leases from their respective networks and configure their mesh interface accordingly
* Shit ain't working on Windows or OSX
* A CLI is used to define networks and inspect a few things

## Building

### RPC

Run `make gen` to regen the protobug and grpc thingies

### Build
Running `make` should do it, you will have 3 binaries pop up in `bin/`.

## Controller
Start the controller running `./bin/wgnw-server --listen 0.0.0.0:10000`, it will start a controller with an SQLite backend.
Normally SQL backends supported by `gorm` should work, I tested it with CockroachDB and SQLite for now. Let me know if that
does not work elsewhere.

## Admin CLI
Run the cli with `./bin/wgnw --controller localhost:10000 --help` to know how to use it. You probably want to create a network first,
to do that, run `./bin/wgnw network create mynet 10.42.0.0/16 --subnets 32` to create a network that will allocate up to `32` sub-ranges
that the clients will be able to use.

## Agent
You will need 2 nodes, on each one run `./bin/wgnwd -net mynet -controller <your controller addr:port> -iface <iface name>`. This assumes
that the nodes are behind a NAT. If the node is accessible from somewhere (i.e. if all the nodes are in the same LAN or reachable on the internet)
you can also add a flag `-public <addr>` specifying the public address (or LAN address) your nodes can be talked to. This will be used when the peers
fetch their conf and heart beat each other.

## Authentication
lol what ?
