#!/usr/bin/env bats

load "./support/bats-support/load"
load "./support/bats-assert/load"

@test "Disk IO: Create 6 GiB file" {
  run "if [[ -f /tmp/fill1 ]]; then rm /tmp/fill1; fi"
  run "if [[ -f /tmp/fill2 ]]; then rm /tmp/fill2; fi"
  run rm /tmp/disk-metrics
  run sleep 3
  run dd if=/dev/zero of=/tmp/fill1 bs=1k count=6M
  assert_line --partial "6442450944 bytes (6.4 GB, 6.0 GiB) copied"
  run sleep 6
  [ $(cat /tmp/disk-metrics | awk '{sum+=$19} END {print sum}') -ge 5000 ]
}

@test "Disk IO: Copy 6 GiB file" {
  run rm /tmp/disk-metrics
  run sleep 3
  run cp /tmp/fill1 /tmp/fill2
  run sleep 6
  [ $(cat /tmp/disk-metrics | awk '{sum+=$18} END {print sum}') -ge 5000 ]
  [ $(cat /tmp/disk-metrics | awk '{sum+=$19} END {print sum}') -ge 5000 ]
}
