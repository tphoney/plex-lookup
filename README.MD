# plex-lookup

## Description

A tool to compare your plex librarys with other physical media rental / purchasing services.

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)

## Installation

Build the binary.

```bash
go build
```

## Usage

```bash
  plex-lookup [command]

Available Commands:
  cinema-paradiso compare movies in your plex library with cinema paradiso
  completion      Generate the autocompletion script for the specified shell
  help            Help about any command
  plex-libraries  list out the libraries in your plex server
  version         Print the version number of plex-lookup

Flags:
  -h, --help                   help for plex-lookup
      --plexIP string          Plex IP Address
      --plexLibraryID string   Plex Library ID
      --plexToken string       Plex Token

Use "plex-lookup [command] --help" for more information about a command.
```
