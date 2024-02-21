# plex-lookup

## Description

A tool to compare your plex libraries with other physical media rental / purchasing services.

## Table of Contents

- [Docker](#Docker)
- [Usage](#usage)
  - [plex libraries](#plex-libraries)
  - [cinema paradiso](#cinema-paradiso)
- [Building](#building)

## Docker

You can get the docker image here [tphoney/plex-lookup](https://hub.docker.com/r/tphoney/plex-lookup)

## Usage

```bash
plex-lookup [command]

Available Commands:
  cinema-paradiso Compare movies in your plex library with cinema paradiso
  completion      Generate the autocompletion script for the specified shell
  help            Help about any command
  plex-libraries  List out the libraries in your plex server
  version         Print the version number of plex-lookup

Flags:
  -h, --help                   help for plex-lookup
      --plexIP string          Plex IP Address
      --plexLibraryID string   Plex Library ID
      --plexToken string       Plex Token

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

### cinema-paradiso

**You will need to get the library id using the library command above**

```bash
docker run --rm tphoney/plex-lookup /plex-lookup cinema-paradiso --plexIP=192.168.1.3 --plexLibraryID=3 --plexToken=secrets
```

gives the output:

```bash
Movies at resolution 480: [{Anchorman: The Legend of Ron Burgundy 2004} {Bombshell: The Hedy Lamarr Story 2018} {Boycie in Belgrade 2021} {Elvira's Haunted Hills 2002} {The Great Outdoors 1988} {Heist 2001} {Hitchcock/Truffaut 2016} {Kingpin 1996} {Napoleon Dynamite 2004} {Ong-Bak 2003} {Pink Floyd: Meddle - A Classic Album Under Review 2007} {Pink Floyd: The Wall 1982} {Radiohead: OK Computer - A Classic Album Under Review 2006} {School of Rock 2003} {Thin Lizzy Rock Review 2020}]

There are a total of 15 movies in the library.

Movies available:
Anchorman: The Legend of Ron Burgundy [DVD Blu-ray]: https://www.cinemaparadiso.co.uk/rentals/anchorman-the-legend-of-ron-burgundy-10478.html
School of Rock [DVD Blu-ray]: https://www.cinemaparadiso.co.uk/rentals/school-of-rock-2585.html
```

## Building

Build the binary.

```bash
go build
```
