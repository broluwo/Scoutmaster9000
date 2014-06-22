Scoutmaster 9000 : Utility
======
##Overview
Scoutmaster 9000 is a scouting system for the [FIRST Robotics Competition](http://www.usfirst.org/roboticsprograms/frc) created by [Team 449, The Blair Robot Project](http://robot.mbhs.edu/). It is a replacement for Scout449 (our old scouting program), which was not flexible enough and had become difficult to maintain. The name is a reference to the Gruntmaster 9000, a theoretical upgraded version of the flagship product from the [Dilbert TV series](http://en.wikipedia.org/wiki/Dilbert_%28TV_series%29).
>Dilmom: Why don't you call your product the Gruntmaster 6000?  
>Dilbert: What kind of product do you see when you imagine a Gruntmaster 6000?  
>Dilmom: Well it's a stripped-down version of the Gruntmaster 9000 of course. But it's software-upgradeable.

This repository contains utility script(s) for adding data to the Scoutmaster 9000 server.

##Setup
Make sure you have a working Go environment (go 1.1 is *required*, the higher the version the better). [See the install instructions](http://golang.org/doc/install.html).
The one outside dependency is codegangsta's [cli](https://github.com/codegangsta/cli)
To install it, simply run:  
```
$ go get github.com/codegangsta/cli
```
Make sure your PATH includes to the `$GOPATH/bin` directory so your commands can be easily used:
```
$ export PATH=$PATH:$GOPATH/bin
```
This dependency will most likely be resolved with a static binary
##Usage
There is a built-in help command. Run it with:
```
$ go run scoutingUtilities.go h
```
To get subcommand specific help, Run:
```
$ go run scoutingUtilities.go <command> -h
```
For the scrapeTeam subcommand that would end up being:
```
$ go run scoutingUtilities.go t -h
```
