# frozen_string_literal: true

require 'fileutils'
require 'tmpdir'

Given(/^config file for project ([^\s]+)$/) do |project_name, content|
  @host_workdir = Dir.pwd
  @project_name = project_name
  @temp_dir = Dir.mktmpdir("container-do-test-")
  @project_dir = "#{@temp_dir}/#{@project_name}"
  @config_file = "ContainerDo.toml"

  FileUtils.mkdir(@project_dir) unless File.exist?(@project_dir)
  Dir.chdir(@project_dir)
  log("Testing in: #{@project_dir}")
  File.write(@config_file, content)
end

Given(/^the config file also contains$/) do |content|
  File.write(@config_file, "\n#{content}", mode: 'a')
end

Given(/^temporary folders? ([^\s,]+(?:,\s*[^\s,]+))$/) do |folders|
  folders.split(",") \
         .map { |s| s.strip } \
         .each { |f| FileUtils.mkdir(f) }
end

Given(/^the project contains a file ([a-zA-Z0-9\-_.]+) with content$/) do |file, content|
  File.write(file, content)
end

Given(/^([a-zA-Z0-9\-_.]+) is executable$/) do |file|
  FileUtils.chmod("+x", file)
end
