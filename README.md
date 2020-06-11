# container-do

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
 
   Haha, no. Even considering shell completion, that's too much something I'll 
   use as often. On the CLI, an alias like `cdo` or `$` does wonders.


## Mini ARD

<!-- TODO: separate -->

 - Why containers?  
   --> While this is not about running services, it seemed a prudent way
   to isolate versioned tooling from the host system without too much overhead.
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
