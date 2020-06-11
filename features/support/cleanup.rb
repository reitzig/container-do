# frozen_string_literal: true

require 'open3'

World MockProject

After do |scenario|
  if scenario.failed?
    unless @config_file.nil?
      log("Config file:")
      log("---")
      log(File.read(@config_file))
      log("---")
    end
      unless @run_output.nil?
      log("Command Output:")
      log("---")
      log(@run_output)
      log("---")
    end
  end

  Dir.chdir(@host_workdir) unless @host_workdir.nil?
  FileUtils.rm_rf(@temp_dir) unless @temp_dir.nil?
  @containers.each do |c|
    _, _ = Open3.capture2e($docker, "rm", '-f', c)
  end
end
