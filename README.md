[![license](https://img.shields.io/github/license/reitzig/container-do.svg)](https://github.com/reitzig/container-do/blob/master/LICENSE)
[![release](https://img.shields.io/github/release/reitzig/container-do.svg)](https://github.com/reitzig/container-do/releases/latest)
[![GitHub release date](https://img.shields.io/github/release-date/reitzig/container-do.svg)](https://github.com/reitzig/container-do/releases)
[![Test](https://github.com/reitzig/container-do/workflows/Tests/badge.svg?branch=master&event=push)](https://github.com/reitzig/container-do/actions?query=workflow%3ATests+branch%3Amaster+event%3Apush++)

# container-do (WIP)

Run project-level CLI tools in (Linux) containers instead of installing them.

### Premise

 1. You have a "project directory", 
    which we take to mean a directory which contains
    all files pertaining to the task at hand, and
    nothing else.
 2. You need a certain suite of tools (at certain versions)
    to perform this task.
 3. A compatible container image exists for this suite. 


## Install

<!-- TODO -->

## Use

_Prerequisites:_

 - Docker installed and user can run commands.
 - Container has `sh`.

There are only two special commands:

 - `container-do --help` -- print usage instructions.
 - `container-do --init` -- create config file (template).

All other calls will be passed directly to the configured container.
For instance:

```bash
container-do npm install
```

will run `npm install` inside the container and, more specifically,
through the default `SHELL` _in_ that container.

Set environment variable `CONTAINER_DO_LOGGING` to `debug` to get more verbose
logging.

### Config File

At the very least, you will have to tell `container-do` which image to use.
Create a file `ContainerDo.toml` with the following content:

```toml
[container]
image = "my-image"
```

Alternatively, run `container-do --init` to get a full template.

<!-- TODO: document options and defaults -->

## FAQ

 - _Do you type `container-do` every time?_
 
   Haha, no. Even considering shell completion, that's too much for something I'll 
   use as often. On the CLI, an alias like `cdo` or `$` does wonders.


## Mini ARD

<!-- TODO: separate -->

 - Why containers?  
   --> While this is not about running services, it seemed a prudent way
   to isolate versioned tooling from the host system without too much overhead.
   Also, the approach eliminates the need for tools specific to a certain ecosystem
   such as venv, rvm, etc.
 - Why Go?  
   --> Using this a learning experience. Efficient binaries seem prudent here.
   Also, Go seems be prevalent in the OCI space.
   _If_ I were to use docker/client or libpod as libraries, they're written in Go.
 - Why TOML?  
   --> better trade-off between expressiveness, cleanliness and usability 
   than either of INI, JSON, YAML. 
 - Why not use docker/client resp. libpod as libraries?  
   --> would mean higher maintenance debt (security patches etc)
   (quote libpod doc)


## Acknowledgements

Parts of this project were created during 
    [20% time](https://en.wikipedia.org/wiki/20%25_Project) 
graciously provided by 
    [codecentric](https://codecentric.de).
Thank you!
