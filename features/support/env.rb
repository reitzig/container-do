# frozen_string_literal: true

require 'ffi'
require 'ffi/platform'

$container_do = "container-do"
$docker = "docker"

if FFI::Platform.windows?
  #$container_do += ".exe"
  $docker += ".exe"
end

# Cross-platform way of finding an executable in the $PATH.
#
#   which('ruby') #=> /usr/bin/ruby
#
# Credits: https://stackoverflow.com/a/5471032/539599
def which(cmd)
  exts = ENV['PATHEXT'] ? ENV['PATHEXT'].split(';') : ['']
  ENV['PATH'].split(File::PATH_SEPARATOR).each do |path|
    exts.each do |ext|
      exe = File.join(path, "#{cmd}#{ext}")
      return exe if File.executable?(exe) && !File.directory?(exe)
    end
  end
  nil
end

module MockProject
end

World MockProject

Before do
    @host_workdir = Dir.pwd
    @env = {
        "CONTAINER_DO_LOGGING" => "debug"
    }
end
