# frozen_string_literal: true

require 'open3'

World MockProject

AfterStep do |scenario|
  if @command_just_ran && (@run_status != 0 || !@run_error.nil?)
    log("Command run/kill error: #{@run_error}") unless @run_error.nil?
    log("Command Exit Status: #{@run_status}")
    unless @run_output.nil?
      log("Command Standard Output:")
      log("---")
      log(@run_output)
      log("---")
    end
    unless @run_err_out.nil?
      log("Command Error Output:")
      log("---")
      log(@run_err_out)
      log("---")
    end

    @command_just_ran = false
  end
end

After do |scenario|
  if scenario.failed?
    unless @config_file.nil?
      log("Config file:")
      log("---")
      log(File.read(@config_file))
      log("---")
    end
  end

  @env = {}

  Dir.chdir(@host_workdir) unless @host_workdir.nil?

  FileUtils.rm_rf(@temp_dir) unless @temp_dir.nil?
  @project_name = nil
  @temp_dir = nil
  @project_dir = nil

  @containers.each do |c|
    _, _ = Open3.capture2e($docker, "rm", '-f', c)
  end unless @containers.nil?

  # at_exit can't access "World", so we need to (ab)use a global variable:
  $images = ($images || []).concat(@images) unless @images.nil?
end

at_exit do
  $images.each do |i|
    _, _ = Open3.capture2e($docker, "rmi", i)
  end unless $images.nil?
end
