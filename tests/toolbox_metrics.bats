#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

setup() {
  TMPFILE=$(mktemp /tmp/toolbox-XXXX)
  source /tmp/.env-*
  source /opt/change-erlang-version.sh
  source /opt/change-python-version.sh
  source /opt/change-go-version.sh
  source /opt/change-java-version.sh
  source /opt/change-scala-version.sh
  source ~/.phpbrew/bashrc
  . /home/semaphore/.nvm/nvm.sh
  export PATH="$PATH:/home/semaphore/.yarn/bin"
  export KIEX_HOME="$HOME/.kiex"
  source "/home/semaphore/.kiex/scripts/kiex"
  export PATH="/home/semaphore/.rbenv/bin:$PATH"
  export NVM_DIR=/home/semaphore/.nvm
  export PHPBREW_HOME=/home/semaphore/.phpbrew
  eval "$(rbenv init -)"

  source ~/.toolbox/toolbox
  sem-service start postgres
  sem-service start redis
  sem-version go 1.20
  sed -E -i  '/^semservice,service=[a-z]*,state=(success|fail),version=[0-9a-zA-Z.]+,location=(disk|local|remote) duration=[0-9]+$/d' /tmp/toolbox_metrics 

  sed -E -i  '/^semversion,software=[a-z]*,state=(success|fail),version=[0-9a-zA-Z.-]+,osversion=[0-9.]+ duration=[0-9]+$/d' /tmp/toolbox_metrics

  sed -E -i  "/^libcheckout,provider='?(github|bitbucket)'?,reftype='?[^']*'?,status=(success|fail) size=[0-9]+$/d" /tmp/toolbox_metrics

  sed -E -i  '/^usercache,server=[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3},user=[a-z,0-9,-]+,command=(store|restore),corrupt=[0,1] size=[0-9]+,duration=[0-9]+$/d' /tmp/toolbox_metrics
  
}

@test "metrics file should be empty" {
  if [[ $(wc -c /tmp/toolbox_metrics | awk '{print $1}') -eq 0 ]];then
    rm -f /tmp/toolbox_metrics
  fi

  run cat /tmp/toolbox_metrics

  assert_failure
}
