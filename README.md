# merenpic
Media grouping by google metadata CLI [merenpic]

## Features
* **No dependencies (Node, Python Interpreter etc.)** - `merenpic` is a single statically linked binary. Grab the one that fits your architecture, and you're all set to save time!
* **Full Power of [Golang](https://go.dev/)** - Golang programming language.
* **Full Power of [Cobra](https://github.com/spf13/cobra)** - Cobra is a library for creating powerful modern CLI applications.

## Installation
Currently, we are only able to build this CLI locally. In the future we will build
the binaries with GitHub actions

- Build the CLI
```shell
go build .
```

- copy the CLI to your binaries
```shell
sudo cp merenpic /usr/local/bin/merenpic
```

- Run the CLI to make sure is working properly
```shell
merenpic --version
```
You should see something like: `merenpic version 0.0.1`

## Usage
Currently, the CLI is able to group media files by their metadata. The CLI will group the files by the date the media was taken.

- Run the CLI with the following command
```shell
merenpic group --folder /path/to/media
```

### Built with :yellow_heart: by Merensoft Team