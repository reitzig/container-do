# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and 
this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [1.2.2] - 2021-02-09

### Fixed

 - Account for missing `container.work_dir` on subsequent runs as well (issue #18)


## [1.2.1] - 2021-02-09

### Fixed

 - Copy files into container even if `container.work_dir` is missing (issue #14)
 - Do not ignore any `[[copy._]]` sections (issue #15)
 - Always create target directory when copying files (issue #16)
 - Kill container if `copy.setup` fails (issue #17)


## [1.2.0] - 2021-02-06

### Added

 - Copy files to the container during setup and before each command (issue #12)
 - Copy files from the container after each command (issue #12)
 - Use host environment variables in `container.environment` (issue #9)
 - Kill rogue containers with `--kill` (issue #11)
 - Kill containers if `run.setup` fails (issue #7)

### Fixed

 - Avoid redundant `docker start`
 - Log `run._` properly

### Housekeeping

 - Compile with Go 1.15
 - Update dependencies


## [1.1.0] - 2020-06-29

### Added

 - Configure published ports ([Example](examples/nginx))
 
### Fixed

 - Partial fix for cancelling commands (issue #6):
   - keep-alive token is reset (so the container will stop as configured)
   - _but_ command in container is not killed.


## [1.0.0] - 2020-06-22

Initial release. 
