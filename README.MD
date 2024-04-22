# plex-lookup

## Description

A tool to compare your plex libraries with other physical media rental / purchasing services.

![demo](demo.gif)

## Table of Contents

- [Docker](#docker)
- [Usage](#usage)
  - [Amazon](#amazon) uses blu-ray.com
  - [Cinema Paradiso](#cinema-paradiso)
  - [plex libraries](#plex-libraries)
  - [web](#web)
- [Building](#building)

## Docker

You can get the docker image here [tphoney/plex-lookup](https://hub.docker.com/r/tphoney/plex-lookup)

## **TLDR run the website already**

```bash
docker run --rm  -p 9090:9090 tphoney/plex-lookup
```

## Usage

```bash
plex-lookup [command]

Available Commands:
  amazon          Compare movies in your plex library with amazon
  cinema-paradiso Compare movies in your plex library with cinema paradiso
  completion      Generate the autocompletion script for the specified shell
  help            Help about any command
  plex-libraries  List out the libraries in your plex server
  version         Print the version number of plex-lookup
  web             Starts the web server

Flags:
  -h, --help                        help for plex-lookup
      --plexIP string               Plex IP Address
      --plexMovieLibraryID string   Plex Library ID
      --plexToken string            Plex Token
      --type string                 Library Type (Movie, TV) (default "Movie")

Use "plex-lookup [command] --help" for more information about a command.
```

**You will need to find the ip of your plex server under `https://plex.tv/web` under settings->remote-access. It is the private IP address.**

**You can find your plex token by following the instructions in this [plex support article](https://support.plex.tv/articles/204059436-finding-an-authentication-token-x-plex-token/)**

### plex-libraries

```bash
docker run --rm tphoney/plex-lookup /plex-lookup plex-libraries --plexIP=192.168.1.150 --plexToken=secret
```

gives the output:

```bash
Title: Films
Type: movie
ID: 3

Title: TV shows
Type: show
ID: 2

Title: Music
Type: artist
ID: 1
```

### Amazon

**NB** You will need to get the library id using the library command above.

```bash
docker run --rm tphoney/plex-lookup /plex-lookup amazon --plexIP=192.168.1.3 --plexMovieLibraryID=3 --plexToken=<secrets>
```

gives the output:

```bash
```

### Cinema Paradiso

**NB** You will need to get the library id using the library command above.

```bash
docker run --rm tphoney/plex-lookup /plex-lookup cinema-paradiso --plexIP=192.168.1.3 --plexMovieLibraryID=3 --plexToken=<secrets>
```

gives the output:

```bash
Movies at resolution 480: [{Anchorman: The Legend of Ron Burgundy 2004} {Bombshell: The Hedy Lamarr Story 2018} {Boycie in Belgrade 2021} {Elvira's Haunted Hills 2002} {The Great Outdoors 1988} {Heist 2001} {Hitchcock/Truffaut 2016} {Kingpin 1996} {Napoleon Dynamite 2004} {Ong-Bak 2003} {Pink Floyd: Meddle - A Classic Album Under Review 2007} {Pink Floyd: The Wall 1982} {Radiohead: OK Computer - A Classic Album Under Review 2006} {School of Rock 2003} {Thin Lizzy Rock Review 2020}]

There are a total of 15 movies in the library.

Movies available:
Anchorman: The Legend of Ron Burgundy [DVD Blu-ray]: https://www.cinemaparadiso.co.uk/rentals/anchorman-the-legend-of-ron-burgundy-10478.html
School of Rock [DVD Blu-ray]: https://www.cinemaparadiso.co.uk/rentals/school-of-rock-2585.html
```

### web

Runs the webui on port 9090 by default.

```bash
docker run --rm  -p 9090:9090 tphoney/plex-lookup
```

## Building

Build the binary.

```bash
go build
```
