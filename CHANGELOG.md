# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and 
this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [1.2.0] - upcoming

### Features

 - Use host environment variables in `container.environment` (issue #9)
 - Kill rogue containers with `--kill` (issue #11)
 - Kill containers if `run.setup` fails (issue #7)

### Fixes

 - Avoid redundant `docker start`
 - Log `run._` properly


## [1.1.0] - 2020-06-29

### Features

 - Configure published ports ([Example](examples/nginx))
 
### Fixes

 - Partial fix for cancelling commands (issue #6):
   - keep-alive token is reset (so the container will stop as configured)
   - _but_ command in container is not killed.


## [1.0.0] - 2020-06-22

Initial release. 
