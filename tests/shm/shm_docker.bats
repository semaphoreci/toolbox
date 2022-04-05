#!/usr/bin/env bats

load "../support/bats-support/load"
load "../support/bats-assert/load"

@test "shm: write 512MB to shared memory" {
  run ~/toolbox/tests/shm/shm_512mb
  assert_line --partial "Writing Process: Shared Memory Write: Wrote 536870911 bytes"
  assert_line --partial "Writing Process: Complete"
  [ $(cat /tmp/system-metrics2 | awk -F "," '{print $4}' | grep -o '[0-9]\+' | sort -n | tail -n 1) -ge 510 ]
  run sleep 2
}

@test "shm: write 1024MB to shared memory" {
  run ~/toolbox/tests/shm/shm_1024mb
  assert_line --partial "Writing Process: Shared Memory Write: Wrote 1073741823 bytes"
  assert_line --partial "Writing Process: Complete"
  [ $(cat /tmp/system-metrics2 | awk -F "," '{print $4}' | grep -o '[0-9]\+' | sort -n | tail -n 1) -ge 1020 ]
  run sleep 2
}
