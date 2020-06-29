# frozen_string_literal: true

require 'open3'

When(/^container-do is called with `([^`]+)`$/) do |command|
  containers_before, _ = Open3.capture2e($docker, "ps", "-aq")
  command = command.split(/\s+/)
  @run_output, @run_err_out, status = Open3.capture3(@env, "#{@host_workdir}/#{$container_do}", *command)
  @run_status = status.exitstatus
  containers_after, _ = Open3.capture2e($docker, "ps", "-aq")

  @new_containers = containers_after.strip.split("\n") - containers_before.strip.split("\n")
  @containers = (@containers || []).concat(@new_containers)
  @command_just_ran = true
end

When(/^container-do is called with long-running `([^`]+)`$/) do |command|
  containers_before, _ = Open3.capture2e($docker, "ps", "-aq")
  command = command.split(/\s+/)

  @running_command = {
    thread: nil,
    out: [],
    err: []
  }

  Thread.new do
    Open3.popen3(@env, "#{@host_workdir}/#{$container_do}", *command) do |stdin, stdout, stderr, thread|
      @running_command[:thread] = thread

      # Consume output streams in separate threads
      # see: http://stackoverflow.com/a/1162850/83386
      { :out => stdout, :err => stderr }.each do |key, stream|
        Thread.new do
          until (line = stream.gets).nil? do
            @running_command[key].push line
          end
        end
      end

      thread.join # don't exit until the external process is done
    end
  end

  sleep(1) # starting isn't instant
  containers_after, _ = Open3.capture2e($docker, "ps", "-aq")
  @new_containers = containers_after.strip.split("\n") - containers_before.strip.split("\n")
  @containers = (@containers || []).concat(@new_containers)
end

When("(we )wait for( another) {float}s") do |interval|
  sleep(interval.to_f)
end

When(/^(?:we )?send (SIG[A-Z]+) to container-do$/) do |signal|
  begin
    pid = @running_command[:thread].pid
    log("Sending #{signal} to #{pid}")
    Process.kill(signal, pid) unless pid < 1
  rescue => e
    log(e)
    @run_error = e
  end

  @run_status = @running_command[:thread].value.exitstatus
  @run_output = @running_command[:out].join()
  @run_err_out = @running_command[:err].join()

  @command_just_ran = true
end
