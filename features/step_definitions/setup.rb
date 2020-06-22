# frozen_string_literal: true

require 'fileutils'
require 'tmpdir'

def init_project_dir(project_name)
  @project_name = project_name
  @temp_dir = Dir.mktmpdir("container-do-test-")
  @project_dir = "#{@temp_dir}/#{@project_name}"

  FileUtils.mkdir(@project_dir) unless File.exist?(@project_dir)
  Dir.chdir(@project_dir)
  log("Testing in: #{@project_dir}")
end

Given(/^an empty project ([^\s]+)$/) do |project_name|
  init_project_dir(project_name)
end

Given(/^config file for project ([^\s]+)$/) do |project_name, content|
  init_project_dir(project_name)

  @config_file = "ContainerDo.toml"
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

Given(/^environment variable ([A-Z_]+) is set to "([^"]+)"$/) do |key,value|
  @env[key] = value
end
