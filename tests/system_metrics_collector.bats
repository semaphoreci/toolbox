#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

setup() {
  kill -TERM `ps ax|grep system-metrics-collector|awk '{print $1}'` || true
  rm -f /tmp/system-metrics
  sh ~/toolbox/system-metrics-collector &
  PID=$!
  echo '[A-Z][a-z][a-z] [A-Z][a-z][a-z] [0-9]{1,2} [0-9]{1,2}:[0-9]{1,2}:[0-9]{1,2} [A-Z]{3,4} [0-9]{4} \|  cpu: [0-9]+(\.[0-9]{1,2})?%,  mem: [0-9]{1,2}(.[0-9]{1,2})?%,  system_disk: [0-9]{1,2}(.[0-9]{1,2})?%,  docker_disk: [0-9]{1,2}(.[0-9]{1,2})?%,  shared_memory: [0-9]+ M' > /tmp/pattern.txt
  sleep 5
  kill -TERM $PID
  cat /tmp/system-metrics
}
teardown() {
  rm -f /tmp/system-metrics
}

@test "Test if /tmp/system-metrics format is not empty" {

  result="$(wc -l /tmp/system-metrics | awk '{print $1}')"
  [ "$result" -gt 0 ]
}


@test "Test if /tmp/system-metrics format is correct" {

  run egrep -q -f /tmp/pattern.txt /tmp/system-metrics
  assert_success
}

@test "Test if /tmp/system-metrics has strange lines" {
  
  run egrep -q -v -f /tmp/pattern.txt /tmp/system-metrics
  assert_failure
}
