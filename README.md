[![license](https://img.shields.io/github/license/reitzig/container-do.svg)](https://github.com/reitzig/container-do/blob/master/LICENSE)
[![release](https://img.shields.io/github/release/reitzig/container-do.svg)](https://github.com/reitzig/container-do/releases/latest)
[![GitHub release date](https://img.shields.io/github/release-date/reitzig/container-do.svg)](https://github.com/reitzig/container-do/releases)
[![Test](https://github.com/reitzig/container-do/workflows/Tests/badge.svg?branch=master&event=push)](https://github.com/reitzig/container-do/actions?query=workflow%3ATests+branch%3Amaster+event%3Apush++)

# container-do

Run project-level CLI tools in (Linux) containers.
In particular,

 - use tools not available on your platform,
 - avoid managing version conflicts of tooling,
 - persist and share specific setups, and 
 - minimize the blast radius of potential mishaps.

### Premise

 1. You have a "project directory", 
    which we take to mean a directory which contains
    all files pertaining to the task at hand, and
    nothing else.
 2. You need a certain suite of tools (at certain versions)
    to perform this task.
 3. A compatible container image exists for this suite. 


## Install

[Download](https://github.com/reitzig/container-do/releases/latest) 
the binary matching your OS and put it on the `PATH`.

As an alternative, you can compile from the sources like so:

```bash
go get github.com/reitzig/container-do/cmd/container-do
```

Find the binary at `$GOPATH/bin`.

## Use

_Prerequisites:_

 - Docker installed and user can run `docker`.
 - Container has `sh`.

There are two special commands:

 - `container-do --help` -- print usage instructions.
 - `container-do --init` -- create config file (template).

All other calls will be passed directly to the configured container.
For instance:

```bash
container-do npm install
```

will run `npm install` inside the container and, more specifically,
through the default `SHELL` _in_ that container.

By default, `container-do` will try to stay out of your way and 
allow you to focus on the normal command output.
However, you can enable rather more verbose logging
by setting environment variable `CONTAINER_DO_LOGGING` to `debug`.

### Config File

At the very least, you will have to tell `container-do` which image to use.
Create a file `ContainerDo.toml` with the following content:

```toml
[container]
image = "my-image"
```

Alternatively, run `container-do --init` to get a full template.
Here is a full list of the optional values:

 - `container.os_flavor` (_Default:_ auto-detect)
 
   For some commands run in the container, we need to know which flavor of Linux it runs.
   While we will attempt to detect that automatically, this induces a slight performance
   over head (and may fail).
   Set to one of `debian`, `fedora`, `alpine`, `gnu/linux`, or `busybox`.

 - `container.name` (_Default:_ `${project_dir}-do`)
 
 - `container.work_dir` (_Default:_ `$WORKDIR`)
 
   Use to override the working directory defined in the container image.
   Set to an absolute path.
 
 - `container.mounts`  (_Default:_ `[.:$WORKDIR]` / `[]`)
 
   Unless the container working directory is `/`,
   the default is a bind-mount from the host working directory to it.
   Override with entries using the
     [Docker `--volume` syntax](https://docs.docker.com/storage/bind-mounts/);
   unlike `docker`, we will resolve relative host paths.
   
   _Note:_ You also use 
     [named volumes](https://docs.docker.com/storage/volumes/#create-and-manage-volumes)!

 - `container.keep_alive` (_Default:_ `15m`)
 
   The duration to keep the container alive for after the last command was run in it.
   Set to any value compatible with [Go `time`](https://pkg.go.dev/time?tab=doc#ParseDuration).

 - `container.keep_stopped` (_Default:_ `false`)
 
   By default, we remove the container after it stops.
   Set to `true` to have the container stick around.

 - `container.environment` (_Default:_ none)
    
    Set environment variables of the container.

 - `run.setup` -- run once after container creation  
   `run.before` -- run once before each command  
   `run.after` -- run once after each (successful) command
   
    Run pre-defined shell commands, each section specified by:
    
    - `run._.attach` (_Default:_ `false`)
    
      Set `true` in order to attach the current shell to the command runs.
    
    - `run._.user` (_Default:_ `$USER`)
    
      Override the default container user to run the commands.
       
    - `run._.script_file` (_Default:_ none)   
    
      Set to a script file in (relative to `container.work_dir`).
      Run before any of the `commands` in the same section.
      
    - `run._.commands` (_Default:_ `[]`)
    
      Set to a list of shell commands run one after the other,
      so long as they are successful.

### Examples

Explore some use cases:

 - [Java with Maven](examples/java)
 - [LaTeX](examples/latex)
 - [Node.js](examples/node)


## Short ADRs

 - _Why containers?_
   
   While this is not about running services, containers seem a prudent way
   to isolate versioned tooling from the host system without too much overhead.
   Also, the approach eliminates the need for tools specific to a certain ecosystem
   such as venv, rvm, etc.
   
 - _Why Go?_
   
   Efficient binaries seem prudent here.
   Also, Go seems to be prevalent in the OCI space.
   _If_ we were to use `docker/client` or `libpod` as libraries, 
   those are written in Go.
   
 - _Why TOML?_  
   
   Comparing to the most common alternatives, 
   TOML seems to provide a better trade-off between expressiveness, cleanliness and usability 
   than either of INI, JSON, YAML.
    
 - _Why not use `docker/client` resp. `libpod` as libraries?_
 
   That would mean higher maintenance debt (security patches etc.) and
   put the duty of ensuring compatibility with user systems on us.


## Acknowledgements

Parts of this project were created during 
    [20% time](https://en.wikipedia.org/wiki/20%25_Project) 
graciously provided by 
    [codecentric](https://codecentric.de).
Thank you!
