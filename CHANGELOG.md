# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and 
this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased] - TBD

### Changed

- Disable CGO to make binaries more portable


## [1.2.7] - 2024-01-29

### Housekeeping

- Bump to Go 1.21
- Bump tests to Ruby 3.3
- Update dependencies


## [1.2.6] - 2022-05-02

### Housekeeping

- Bump to Go 1.18 -- for real this time! (Sorry. 🙏)


## [1.2.5] - 2022-05-02

### Housekeeping

- Bump to Go 1.18
- Update dependencies -- thanks, Dependabot! 🧡


## [1.2.4] - 2022-03-08

### Fixed

 - Better handling of missing image (issue #21)

### Housekeeping

 - Bump to Go 1.17
 - All kinds of dependency updates -- thanks, Dependabot! 🧡


## [1.2.3] - 2021-02-17

### Fixed

 - Determine container working directory correctly for Kaniko-built images (issue #19) 


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
