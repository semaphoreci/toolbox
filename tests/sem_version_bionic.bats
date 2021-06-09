#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

rm -rf /home/semaphore/.kiex/elixirs
rm -rf /home/semaphore/.kiex/mix/archives
setup() {
  source /tmp/.env
  source /opt/change-erlang-version.sh
  source /opt/change-python-version.sh
  source /opt/change-go-version.sh
  source /opt/change-java-version.sh
  source /opt/change-scala-version.sh
  source /opt/change-firefox-version.sh
  source ~/.phpbrew/bashrc
  . /home/semaphore/.nvm/nvm.sh
  export PATH="$PATH:/home/semaphore/.yarn/bin"
  source "/home/semaphore/.kiex/scripts/kiex"
  export PATH="/home/semaphore/.rbenv/bin:$PATH"
  export NVM_DIR=/home/semaphore/.nvm
  export PHPBREW_HOME=/home/semaphore/.phpbrew
  eval "$(rbenv init -)"

  source ~/.toolbox/toolbox
}
#  Elixir
@test "change elixir to 1.7.4" {
  sem-version elixir 1.7.4
  assert_success
  run elixir --version
  assert_line --partial "Elixir 1.7.4"
  run ls /home/semaphore/.kiex/mix/archives/elixir-1.7.4/
  assert_success
  assert_line --partial "hex"
}
