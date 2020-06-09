# container-do

Run project-level CLI tools in containers instead of installing them.

### Premise

 1. You have a "project directory", 
    which we take to mean a directory which contains
    all files pertaining to the task at hand, and
    nothing else.
 2. You need a certain suite of tools (at certain versions)
    to perform this task.
 3. A container image exists for this suite. 


## Install

<!-- TODO -->

## Use

Only two commands:

 - `container-do --help` -- prints usage instructions.
 - `container-do --init` -- creates config file ``.

All other calls will be passed directly to the configured container.
For instance:

```bash
container-do npm install
```

will run `npm install` inside the container and, more specifically,
through the default `SHELL` _in_ that container.

### Config File
 
```toml
[container]
runner = docker | podman
image = node:12-slim
# OR
build = Dockerfile.tooling

name = my-project-tooling
work_dir = /usr/src/app

mount = "$PWD:/usr/src/app"
keep_alive = "T15M"
keep_stopped = false
 

[container.setup]
privileged = false
script = "setup.sh"
# OR
commands = """
           npm i -g @zeit/ncc
           npm install
           """
```


## Mini ARD

 - Why Go?  
   --> Using this a learning experience. Also, Go seems be prevalent
   in the OCI space. _If_ I were to use docker/client or libpod as 
   libraries, they're written in Go.
 - Why TOML?  
   --> better trade-off between expressiveness, cleanliness and usability 
   than either of INI, JSON, YAML. 
 - Why not use docker/client resp. libpod as libraries?  
   --> would mean higher maintenance debt (security patches etc)
   (quote libpod doc)
