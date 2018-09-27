#!/usr/bin/env bats

@test "Store file in cache" {
  mkdir tmp && touch tmp/example.file
  run bash -c './cache store --key v4 --path tmp'
  [ "$status" -eq 0 ]
  [[ ${lines[4]} =~ "Starting upload" ]]
}
