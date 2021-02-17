# frozen_string_literal: true

require 'fileutils'
require 'tmpdir'

def init_project_dir(project_name)
  if @project_dir.nil?
    @project_name = project_name
    @temp_dir = Dir.mktmpdir("container-do-test-")
    @project_dir = "#{@temp_dir}/#{@project_name}"

    FileUtils.mkdir(@project_dir) unless File.exist?(@project_dir)
    Dir.chdir(@project_dir)
    log("Testing in: #{@project_dir}")
  end
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

Given(/^Docker image ([a-zA-Z0-0_-]+) exists based on$/) do |image_name, dockerfile|
  File.write("Dockerfile", dockerfile)

  out, status = Open3.capture2e($docker, "build", '-t', image_name, '.')
  unless status.success?
    log(out)
    raise "Could not build image"
  else
    log("Built image #{image_name}")
    @images = (@images || []).concat([image_name])
  end
end

Given(/^Kaniko image ([a-zA-Z0-0_-]+) exists based on$/) do |image_name, dockerfile|
  File.write("Dockerfile", dockerfile)

  out, status = Open3.capture2e($docker, 'run', '--rm',
    '-v', "#{Dir.pwd}:/workspace",
    'gcr.io/kaniko-project/executor:latest',
    '--cleanup', '--no-push',
    '--dockerfile', '/workspace/Dockerfile',
    '--tarPath', "/workspace/#{image_name}.tar",
    '--destination', "#{image_name}:latest",
    '--context', 'dir:///workspace/')
  unless status.success?
    log(out)
    raise "Could not build image"
  else
    log("Built image #{Dir.pwd}/#{image_name}.tar")
  end

  out, status = Open3.capture2e($docker, 'load', '--input', "#{image_name}.tar")
  unless status.success?
      log(out)
      raise "Could not load image"
    else
      log("Loaded image #{image_name}")
      @images = (@images || []).concat([image_name])
    end
end
