# frozen_string_literal: true

require 'open3'
require 'json'

Given(/^(\w+) is installed$/) do |command|
  raise "#{command} not installed" if which(command).nil?
end

Then(/^a container is started with name ([^\s]+)$/) do |name|
  expect(@new_containers.count).to eq(1)
  container = @new_containers[0]

  out, status = Open3.capture2e($docker, "inspect", '--format={{.Id}}', name)
  unless status.success?
    log(out)
    raise "Container not found"
  end

  expect(out).to start_with(container)
  @container = container
end

Then(/^the container is based on image ([^\s]+)$/) do |image|
  out, status = Open3.capture2e($docker, "inspect", '--format={{.Config.Image}}', @container)
  unless status.success?
    log(out)
    raise "Could not determine container image"
  end

  expect(out.strip).to eq(image)
end

Then(/^the container has working directory ([^\s]+)$/) do |work_dir|
  out, status = Open3.capture2e($docker, "inspect", '--format={{.Config.WorkingDir}}', @container)
  unless status.success?
    log(out)
    raise "Could not determine container working directory"
  end
  out = '/' if out.strip == ""
  expect(out.strip).to eq(work_dir)

  out, status = Open3.capture2e($docker, "exec", @container, 'pwd')
  unless status.success?
    log(out)
    raise "Could not run command on container"
  end
  expect(out.strip).to eq(work_dir)
end

def list_container_volume_binds(container)
  out, status = Open3.capture2e($docker, "inspect", '--format={{json .HostConfig.Binds}}', container)
  unless status.success?
    log(out)
    raise "Could not determine container mounts"
  end

  out
end

Then(/^the container has a volume mount for ([^\s]+) at ([^\s]+)$/) do |host_dir, container_dir|
  json_list = list_container_volume_binds(@container)

  host_dir = File.absolute_path(host_dir)
  expect(json_list.strip).to match(%r{"#{host_dir}:#{container_dir}"})
end

Then("the container has no volume mounts") do
  json_list = list_container_volume_binds(@container)

  expect(json_list.strip).to eq("null")
end

def list_published_ports(container)
  out, status = Open3.capture2e($docker, "inspect", '--format={{json .NetworkSettings.Ports}}', container)
  unless status.success?
    log(out)
    raise "Could not determine exposed ports"
  end

  JSON.parse(out).map do |k,v|
    # k ~ "80/tcp"
    # v ~ [{"HostPort": "8080"}]
    [
      k.split("/")[0],
      v.map { |m| m["HostPort"] }
    ]
  end.to_h
end

And("the container publishes no ports") do
  expect(list_published_ports(@container)).to be_empty
end

Then(/^the container publishes port ([0-9]+)(?: as ([0-9]+))?$/) do |container_port, host_port|
  ports_map = list_published_ports(@container)

  expect(ports_map).not_to be_empty
  expect(ports_map).to have_key(container_port)

  if host_port.nil?
    expect(ports_map[container_port]).not_to be_empty
  else
    expect(ports_map[container_port]).to include(host_port)
  end
end

Then(/^the container has an environment variable ([A-Z_]+) with value "([^"]*)"$/) do |key, value|
  out, status = Open3.capture2e($docker, "inspect", '--format={{json .Config.Env}}', @container)
  unless status.success?
    log(out)
    raise "Could not determine container environment variables"
  end

  expect(out.strip).to match(/"#{key}=#{value}"/)
end

# TODO: If we ever test long-running commands _without_ killing them, add this to the exit/output checks:
#
# unless @running_command.nil?
#   expect(@running_command[:thread].value).to_not be_nil
#   @run_status = @running_command[:thread].value.exitstatus
#   @run_output = @running_command[:out].join()
#   @run_err_out = @running_command[:err].join()
#   @running_command = nil
#   @command_just_ran = true
# end

Then('the command exits with status {int}') do |status|
  expect(@run_status).to eq(status)
end

Then("(the )command/its (standard )output is {string}") do |output|
  expect(@run_output.strip).to eq(output)
end

Then("(the )command/its (standard )output is") do |output|
  expect(@run_output.strip).to eq(output)
end

Then("(the )command/its (standard )output contains {string}") do |output|
  expect(@run_output).to match(output)
end

Then("(the )command/its error output is {string}") do |output|
  expect(@run_err_out).to eq(output)
end

Then("(the )command/its error output contains {string}") do |output|
  expect(@run_err_out).to match(output)
end

Then("the container is( still) running") do
  out, status = Open3.capture2e($docker, "inspect", '--format={{.State.Running}}', @container)
  unless status.success?
    log(out)
    raise "Could not determine container state"
  end

  expect(out.strip).to eq("true")
end

Then("the container is not running( anymore)") do
  out, status = Open3.capture2e($docker, "inspect", '--format={{.State.Running}}', @container)
  if status.success?
    expect(out.strip).to eq("false")
  end
end

Then("the container is gone/absent") do
  _, status = Open3.capture2e($docker, "inspect", '--format={{.State.Running}}', @container)
  expect(status.success?).to eq(false)
end

Then("the container is( still) there/present") do
  _, status = Open3.capture2e($docker, "inspect", '--format={{.State.Running}}', @container)
  expect(status.success?).to eq(true)
end


And(/^file ([^\s]+) (?:still |now )?contains$/) do |file_name, expected_content|
  expect(File).to exist(file_name)
  actual_content = File.read(file_name)
  expect(actual_content.strip).to eq(expected_content.strip)
end


Then(/^file ([^\s]+) is a commented valid config file$/) do |config_file|
  # We don't _actually_ have an easy way to parse the file by itself, so we
  # force the app to do it and confirm we're _not_ seeing a config error.
  # Instead, we expect it to run all the way up until it discovers that the
  # placeholder is, in fact, not a valid Docker image.

  # First, uncomment all lines in the file
  lines = File.readlines(config_file)
  File.open(config_file, "w") do |f|
    lines.each do |l|
      f.write(l.sub(/^\s*#\s*/, ""))
    end
  end

  @config_file ||= config_file

  # Now, pretend to run a command
  output, status = Open3.capture2e(@env, "#{@host_workdir}/#{$container_do}", "cat", "/neverthere")
  expect(status.success?).to be(false)
  expect(status.exitstatus).to be == 1
  expect(output).to match("Unable to find image") | match("invalid reference format") # as opposed to a config error
end

And(/^no container was started$/) do
  expect(@containers).to be_empty
end

Then(/^a command matching \/(.*)\/ is running in the container$/) do |pattern|
  out, status = Open3.capture2e($docker, "exec", @container, 'ps', '-eo', 'comm=')

  expect(out).to match(/#{Regexp.escape(pattern)}/)
end

Then(/^no command matching \/(.*)\/ is running in the container$/) do |pattern|
  pending # TODO: docker doesn't forward signals without -t --> issue #6
  out, status = Open3.capture2e($docker, "exec", @container, 'ps', '-eo', 'comm=')

  expect(out).to_not match(/#{Regexp.escape(pattern)}/)
end
