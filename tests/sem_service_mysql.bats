#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

teardown() {
  sem-service stop mysql || true
}

@test "sem-service-check-params mysql passes extra docker flags through" {
  run sem-service-check-params mysql 8.0 --tmpfs /var/lib/mysql:rw,size=2g -e FOO=bar
  assert_success
  assert_output --partial "--tmpfs /var/lib/mysql:rw,size=2g"
  assert_output --partial "-e FOO=bar"
}

@test "sem-service-check-params mysql drops the default bind mount when it is overridden" {
  run sem-service-check-params mysql 8.0 --tmpfs /var/lib/mysql:rw,size=2g
  assert_success
  refute_output --partial "-v /var/tmp/mysql:/var/lib/mysql"
}

@test "sem-service-check-params mysql keeps the default bind mount otherwise" {
  run sem-service-check-params mysql 8.0
  assert_success
  assert_output --partial "-v /var/tmp/mysql:/var/lib/mysql"
}

@test "sem-service-check-params mysql separates mysqld args after '--'" {
  run sem-service-check-params mysql 8.0 --tmpfs /var/lib/mysql:rw,size=2g -- --innodb-doublewrite=0 --skip-log-bin
  assert_success
  mysqld_args="${output#*-- }"
  assert_equal "$(echo "$mysqld_args" | xargs)" "--innodb-doublewrite=0 --skip-log-bin"
}

@test "sem-service-check-params mysql emits an empty mysqld arg list when no '--' is given" {
  run sem-service-check-params mysql 8.0 --username=root
  assert_success
  mysqld_args="${output#*-- }"
  [ -z "$(echo "$mysqld_args" | tr -d '[:space:]')" ]
}

@test "sem-service start mysql accepts a tmpfs datadir and mysqld args end-to-end" {
  sem-service start mysql 8.0 --tmpfs /var/lib/mysql:rw,size=2g -- --innodb-flush-log-at-trx-commit=0 --innodb-doublewrite=0 --skip-log-bin --innodb-flush-method=nosync

  run docker inspect mysql --format '{{json .HostConfig.Tmpfs}}'
  assert_success
  assert_output --partial "/var/lib/mysql"

  run mysql --host=0.0.0.0 -uroot -e 'SHOW VARIABLES LIKE "innodb_doublewrite";'
  assert_success
  assert_output --partial "OFF"

  run mysql --host=0.0.0.0 -uroot -e 'SHOW VARIABLES LIKE "log_bin";'
  assert_success
  assert_output --partial "OFF"
}
