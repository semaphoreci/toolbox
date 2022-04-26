#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

setup() {

  export SEMAPHORE_DOCKERGW_USER='user'
  export SEMAPHORE_DOCKERGW_PASSWORD='password'
  export SEMAPHORE_DOCKERGW_ENDPOINT='127.0.0.1'
  export SEMAPHORE_DOCKERGW_ENDPOINT_PORT='3128'

  mkdir -p /tmp/squid
  tee /tmp/squid/squid.conf > /dev/null <<EOF
auth_param basic program /usr/lib/squid3/basic_ncsa_auth /etc/squid/htaccess
auth_param basic realm proxy
acl SSL_ports port 443
acl CONNECT method CONNECT
http_port 3128
acl authenticated proxy_auth REQUIRED
http_access allow authenticated
http_access deny all
EOF

  htpasswd -bc /tmp/squid/htaccess user password
  docker run --rm -d --name squid -p 3128:3128 -v /tmp/squid:/etc/squid registry.semaphoreci.com/squid:latest
  sleep 10
}

teardown() {

  unset SEMAPHORE_DOCKERGW_USER
  unset SEMAPHORE_DOCKERGW_PASSWORD
  unset SEMAPHORE_DOCKERGW_ENDPOINT
  unset SEMAPHORE_DOCKERGW_ENDPOINT_PORT
  docker stop squid
  sleep 10
}



@test "Account settings are OK" {

  run enetwork start
  assert_success
  assert_output --partial 'Docker daemon reloaded'
}

@test "Stop gateway" {

  run enetwork stop
  run cat /etc/systemd/system/docker.service.d/proxy.conf
  assert_failure
}

@test "Account settings are NOT OK" {
  export SEMAPHORE_DOCKERGW_PASSWORD='password2'
  run enetwork start
  assert_success
  assert_output --partial 'Authentication error'
}
