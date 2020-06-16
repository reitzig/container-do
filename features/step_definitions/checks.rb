# frozen_string_literal: true

require 'open3'

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

Then(/^the container has a volume mount for ([^\s]+) at ([^\s]+)$/) do |host_dir, container_dir|
  pending # TODO

  out, status = Open3.capture2e($docker, "inspect", '--format={{json .HostConfig.Binds}}', @container)
  unless status.success?
    log(out)
    raise "Could not determine container mounts"
  end

  host_dir = File.absolute_path(host_dir)
  expect(out.strip).to match(%r{"#{host_dir}:#{container_dir}"})
end

Then(/^the container has an environment variable ([A-Z_]+) with value "([^"]*)"$/) do |key, value|
  out, status = Open3.capture2e($docker, "inspect", '--format={{json .Config.Env}}', @container)
  unless status.success?
    log(out)
    raise "Could not determine container environment variables"
  end

  expect(out.strip).to match(/"#{key}=#{value}"/)
end

Then('the command exits with status {int}') do |status|
  expect(@run_status).to eq(status)
end

Then("(the )command/its output is {string}") do |output|
  expect(@run_output).to eq(output)
end

Then("(the )command/its output contains {string}") do |output|
  expect(@run_output).to match(output)
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


And(/^file ([^\s]+) (?:still )?contains$/) do |file_name, expected_content|
  actual_content = File.read(file_name)
  expect(actual_content.strip).to eq(expected_content.strip)
end


Then(/^file ContainerDo\.toml is a valid config file$/) do
  # We don't _actually_ have an easy way to parse the file by itself, so we
  # force the app to do it an confirm we're _not_ seeing a config error.
  # Instead, we expect it to run all the way up until it discovers that the
  # placeholder is, in fact, not a valid Docker image.

  output, status = Open3.capture2e("#{@host_workdir}/#{$container_do}", "cat", "/neverthere")
  expect(status.success?).to be(false)
  expect(status.exitstatus).to be > 1
  expect(output).to match("Unable to find image") | match("docker: invalid reference format")
end

And(/^no container was started$/) do
  expect(@containers).to be_empty
end
