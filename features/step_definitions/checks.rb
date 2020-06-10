# frozen_string_literal: true

require 'open3'

Given(/^(\w+) is installed$/) do |command|
  raise "#{command} not installed" unless system("which #{command}")
end

Then(/^(?:the|a) container (?:has|is started with) name ([^\s]+)$/) do |name|
  expect(@new_containers.count).to eq(1)

  out, status = Open3.capture2e("docker", "inspect", '--format={{.Id}}', name)
  unless status.success?
    log(out)
    raise "Container not found"
  end

  expect(out).to start_with(@new_containers[0])
end

And(/^(?:the|a) container is (?:started )?based on image ([^\s]+)$/) do |image|
  expect(@new_containers.count).to eq(1)
  container = @new_containers[0]

  out, status = Open3.capture2e("docker", "inspect", '--format={{.Config.Image}}', container)
  unless status.success?
    log(out)
    raise "Could not determine container image"
  end

  expect(out.strip).to eq(image)
end

And(/^(?:the|a) container (?:has|is started with) working directory ([^\s]+)$/) do |work_dir|
  expect(@new_containers.count).to eq(1)
  container = @new_containers[0]

  out, status = Open3.capture2e("docker", "inspect", '--format={{.Config.WorkingDir}}', container)
  unless status.success?
    log(out)
    raise "Could not determine container working directory"
  end
  out = '/' if out.strip == ""
  expect(out.strip).to eq(work_dir)

  out, status = Open3.capture2e("docker", "exec", container, 'pwd')
  unless status.success?
    log(out)
    raise "Could not run command on container"
  end
  expect(out.strip).to eq(work_dir)
end

And(/^(?:the|a) container (?:has|is started with) a volume mount for ([^\s]+) at ([^\s]+)$/) do |host_dir, container_dir|
  expect(@new_containers.count).to eq(1)
  container = @new_containers[0]

  out, status = Open3.capture2e("docker", "inspect", '--format={{json .HostConfig.Binds}}', container)
  unless status.success?
    log(out)
    raise "Could not determine container mounts"
  end

  host_dir = File.absolute_path(host_dir)
  expect(out.strip).to match(%r{"#{host_dir}:#{container_dir}"})
end

And(/^(?:the|a) container (?:has|is started with) an environment variable ([A-Z_]+) with value "([^"]*)"$/) do |key, value|
  expect(@new_containers.count).to eq(1)
  container = @new_containers[0]

  out, status = Open3.capture2e("docker", "inspect", '--format={{json .Config.Env}}', container)
  unless status.success?
    log(out)
    raise "Could not determine container environment variables"
  end

  expect(out.strip).to match(/"#{key}=#{value}"/)
end

And(/^the command exits with status (\d+) and output "([^"]*)"$/) do |status, output|
  expect(@run_status).to eq(int(status))
  expect(@run_output).to eq(output)
end
