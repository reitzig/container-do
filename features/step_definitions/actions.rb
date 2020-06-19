# frozen_string_literal: true

require 'open3'

When(/^container-do is called with `([^`]+)`$/) do |command|
  containers_before, _ = Open3.capture2e($docker, "ps", "-aq")
  command = command.split(/\s+/)
  @run_output, @run_err_out, status = Open3.capture3("#{@host_workdir}/#{$container_do}", *command)
  @run_status = status.exitstatus
  containers_after, _ = Open3.capture2e($docker, "ps", "-aq")

  @new_containers = containers_after.strip.split("\n") - containers_before.strip.split("\n")
  @containers = (@containers || []).concat(@new_containers)
  @command_just_ran = true
end

When("we wait for( another) {float}s") do |interval|
  sleep(interval.to_f)
end
