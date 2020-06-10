# frozen_string_literal: true

require 'open3'

When(/^container\-do is called with `([^`]+)`$/) do |command|
  containers_before, _ = Open3.capture2e("docker", "ps", "-aq")
  @run_output, @run_status = Open3.capture2e("#{@host_workdir}/container-do", command)
  containers_after, _ = Open3.capture2e("docker", "ps", "-aq")

  @new_containers = containers_after.strip.split("\n") - containers_before.strip.split("\n")
end
