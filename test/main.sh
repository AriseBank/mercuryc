#!/bin/sh -eu
[ -n "${GOPATH:-}" ] && export "PATH=${GOPATH}/bin:${PATH}"

# Don't translate mercury output for parsing in it in tests.
export "LC_ALL=C"

# Force UTC for consistency
export "TZ=UTC"

if [ -n "${APOLLO_VERBOSE:-}" ] || [ -n "${APOLLO_DEBUG:-}" ]; then
  set -x
fi

if [ -n "${APOLLO_VERBOSE:-}" ]; then
  DEBUG="--verbose"
fi

if [ -n "${APOLLO_DEBUG:-}" ]; then
  DEBUG="--debug"
fi

echo "==> Checking for dependencies"
deps="apollo mercury curl jq git xgettext sqlite3 msgmerge msgfmt shuf setfacl uuidgen"
for dep in $deps; do
  which "${dep}" >/dev/null 2>&1 || (echo "Missing dependency: ${dep}" >&2 && exit 1)
done

if [ "${USER:-'root'}" != "root" ]; then
  echo "The testsuite must be run as root." >&2
  exit 1
fi

if [ -n "${APOLLO_LOGS:-}" ] && [ ! -d "${APOLLO_LOGS}" ]; then
  echo "Your APOLLO_LOGS path doesn't exist: ${APOLLO_LOGS}"
  exit 1
fi

# Helper functions
local_tcp_port() {
  while :; do
    port=$(shuf -i 10000-32768 -n 1)
    nc -l 127.0.0.1 "${port}" >/dev/null 2>&1 &
    pid=$!
    kill "${pid}" >/dev/null 2>&1 || continue
    wait "${pid}" || true
    echo "${port}"
    return
  done
}

# return a list of available storage backends
available_storage_backends() {
  # shellcheck disable=2039
  local backend backends

  backends="dir"
  for backend in btrfs lvm zfs; do
    if which $backend >/dev/null 2>&1; then
      backends="$backends $backend"
    fi
  done
  echo "$backends"
}

# whether a storage backend is available
storage_backend_available() {
  # shellcheck disable=2039
  local backends
  backends="$(available_storage_backends)"
  [ "${backends#*$1}" != "$backends" ]
}

# choose a random available backend, excluding APOLLO_BACKEND
random_storage_backend() {
    # shellcheck disable=2046
    shuf -e $(available_storage_backends) | head -n 1
}

# return the storage backend being used by a APOLLO instance
storage_backend() {
    cat "$1/apollo.backend"
}


if [ -z "${APOLLO_BACKEND:-}" ]; then
  APOLLO_BACKEND=dir
fi

echo "==> Available storage backends: $(available_storage_backends | sort)"
if [ "$APOLLO_BACKEND" != "random" ] && ! storage_backend_available "$APOLLO_BACKEND"; then
  echo "Storage backage \"$APOLLO_BACKEND\" is not available"
  exit 1
fi
echo "==> Using storage backend ${APOLLO_BACKEND}"

# import storage backends
for backend in $(available_storage_backends); do
  # shellcheck disable=SC1090
  . "backends/${backend}.sh"
done


spawn_apollo() {
  set +x
  # APOLLO_DIR is local here because since $(mercury) is actually a function, it
  # overwrites the environment and we would lose APOLLO_DIR's value otherwise.

  # shellcheck disable=2039
  local APOLLO_DIR apollodir apollo_backend

  apollodir=${1}
  shift

  storage=${1}
  shift

  if [ "$APOLLO_BACKEND" = "random" ]; then
    apollo_backend="$(random_storage_backend)"
  else
    apollo_backend="$APOLLO_BACKEND"
  fi

  # Copy pre generated Certs
  cp deps/server.crt "${apollodir}"
  cp deps/server.key "${apollodir}"

  # setup storage
  "$apollo_backend"_setup "${apollodir}"
  echo "$apollo_backend" > "${apollodir}/apollo.backend"

  echo "==> Spawning apollo in ${apollodir}"
  # shellcheck disable=SC2086
  APOLLO_DIR="${apollodir}" apollo --logfile "${apollodir}/apollo.log" ${DEBUG-} "$@" 2>&1 &
  APOLLO_PID=$!
  echo "${APOLLO_PID}" > "${apollodir}/apollo.pid"
  echo "${apollodir}" >> "${TEST_DIR}/daemons"
  echo "==> Spawned APOLLO (PID is ${APOLLO_PID})"

  echo "==> Confirming apollo is responsive"
  APOLLO_DIR="${apollodir}" apollo waitready --timeout=300

  echo "==> Binding to network"
  # shellcheck disable=SC2034
  for i in $(seq 10); do
    addr="127.0.0.1:$(local_tcp_port)"
    APOLLO_DIR="${apollodir}" mercury config set core.https_address "${addr}" || continue
    echo "${addr}" > "${apollodir}/apollo.addr"
    echo "==> Bound to ${addr}"
    break
  done

  echo "==> Setting trust password"
  APOLLO_DIR="${apollodir}" mercury config set core.trust_password foo
  if [ -n "${DEBUG:-}" ]; then
    set -x
  fi

  echo "==> Setting up networking"
  bad=0
  ip link show apollobr0 || bad=1
  if [ "${bad}" -eq 0 ]; then
    APOLLO_DIR="${apollodir}" mercury network attach-profile apollobr0 default eth0
  fi

  if [ "${storage}" = true ]; then
    echo "==> Configuring storage backend"
    "$apollo_backend"_configure "${apollodir}"
  fi
}

respawn_apollo() {
  set +x
  # APOLLO_DIR is local here because since $(mercury) is actually a function, it
  # overwrites the environment and we would lose APOLLO_DIR's value otherwise.

  # shellcheck disable=2039
  local APOLLO_DIR

  apollodir=${1}
  shift

  echo "==> Spawning apollo in ${apollodir}"
  # shellcheck disable=SC2086
  APOLLO_DIR="${apollodir}" apollo --logfile "${apollodir}/apollo.log" ${DEBUG-} "$@" 2>&1 &
  APOLLO_PID=$!
  echo "${APOLLO_PID}" > "${apollodir}/apollo.pid"
  echo "==> Spawned APOLLO (PID is ${APOLLO_PID})"

  echo "==> Confirming apollo is responsive"
  APOLLO_DIR="${apollodir}" apollo waitready --timeout=300
}

mercury() {
  MERCURY_LOCAL=1
  mercury_remote "$@"
  RET=$?
  unset MERCURY_LOCAL
  return ${RET}
}

mercury_remote() {
  set +x
  injected=0
  cmd=$(which mercury)

  # shellcheck disable=SC2048,SC2068
  for arg in $@; do
    if [ "${arg}" = "--" ]; then
      injected=1
      cmd="${cmd} ${DEBUG:-}"
      [ -n "${MERCURY_LOCAL}" ] && cmd="${cmd} --force-local"
      cmd="${cmd} --"
    elif [ "${arg}" = "--force-local" ]; then
      continue
    else
      cmd="${cmd} \"${arg}\""
    fi
  done

  if [ "${injected}" = "0" ]; then
    cmd="${cmd} ${DEBUG-}"
  fi
  if [ -n "${DEBUG:-}" ]; then
    set -x
  fi
  eval "${cmd}"
}

gen_cert() {
  # Temporarily move the existing cert to trick MERCURY into generating a
  # second cert.  MERCURY will only generate a cert when adding a remote
  # server with a HTTPS scheme.  The remote server URL just needs to
  # be syntactically correct to get past initial checks; in fact, we
  # don't want it to succeed, that way we don't have to delete it later.
  [ -f "${APOLLO_CONF}/${1}.crt" ] && return
  mv "${APOLLO_CONF}/client.crt" "${APOLLO_CONF}/client.crt.bak"
  mv "${APOLLO_CONF}/client.key" "${APOLLO_CONF}/client.key.bak"
  echo y | mercury_remote remote add "$(uuidgen)" https://0.0.0.0 || true
  mv "${APOLLO_CONF}/client.crt" "${APOLLO_CONF}/${1}.crt"
  mv "${APOLLO_CONF}/client.key" "${APOLLO_CONF}/${1}.key"
  mv "${APOLLO_CONF}/client.crt.bak" "${APOLLO_CONF}/client.crt"
  mv "${APOLLO_CONF}/client.key.bak" "${APOLLO_CONF}/client.key"
}

my_curl() {
  curl -k -s --cert "${APOLLO_CONF}/client.crt" --key "${APOLLO_CONF}/client.key" "$@"
}

wait_for() {
  addr=${1}
  shift
  op=$("$@" | jq -r .operation)
  my_curl "https://${addr}${op}/wait"
}

ensure_has_localhost_remote() {
  addr=${1}
  if ! mercury remote list | grep -q "localhost"; then
    mercury remote add localhost "https://${addr}" --accept-certificate --password foo
  fi
}

ensure_import_testimage() {
  if ! mercury image alias list | grep -q "^| testimage\s*|.*$"; then
    if [ -e "${APOLLO_TEST_IMAGE:-}" ]; then
      mercury image import "${APOLLO_TEST_IMAGE}" --alias testimage
    else
      deps/import-busybox --alias testimage
    fi
  fi
}

check_empty() {
  if [ "$(find "${1}" 2> /dev/null | wc -l)" -gt "1" ]; then
    echo "${1} is not empty, content:"
    find "${1}"
    false
  fi
}

check_empty_table() {
  if [ -n "$(sqlite3 "${1}" "SELECT * FROM ${2};")" ]; then
    echo "DB table ${2} is not empty, content:"
    sqlite3 "${1}" "SELECT * FROM ${2};"
    false
  fi
}

kill_apollo() {
  # APOLLO_DIR is local here because since $(mercury) is actually a function, it
  # overwrites the environment and we would lose APOLLO_DIR's value otherwise.

  # shellcheck disable=2039
  local APOLLO_DIR daemon_dir daemon_pid check_leftovers apollo_backend

  daemon_dir=${1}
  APOLLO_DIR=${daemon_dir}
  daemon_pid=$(cat "${daemon_dir}/apollo.pid")
  check_leftovers="false"
  apollo_backend=$(storage_backend "$daemon_dir")
  echo "==> Killing APOLLO at ${daemon_dir}"

  if [ -e "${daemon_dir}/unix.socket" ]; then
    # Delete all containers
    echo "==> Deleting all containers"
    for container in $(mercury list --fast --force-local | tail -n+3 | grep "^| " | cut -d' ' -f2); do
      mercury delete "${container}" --force-local -f || true
    done

    # Delete all images
    echo "==> Deleting all images"
    for image in $(mercury image list --force-local | tail -n+3 | grep "^| " | cut -d'|' -f3 | sed "s/^ //g"); do
      mercury image delete "${image}" --force-local || true
    done

    # Delete all networks
    echo "==> Deleting all networks"
    for network in $(mercury network list --force-local | grep YES | grep "^| " | cut -d' ' -f2); do
      mercury network delete "${network}" --force-local || true
    done

    # Delete all profiles
    echo "==> Deleting all profiles"
    for profile in $(mercury profile list --force-local | tail -n+3 | grep "^| " | cut -d' ' -f2); do
      mercury profile delete "${profile}" --force-local || true
    done

    echo "==> Deleting all storage pools"
    for storage in $(mercury storage list --force-local | tail -n+3 | grep "^| " | cut -d' ' -f2); do
      mercury storage delete "${storage}" --force-local || true
    done

    echo "==> Checking for locked DB tables"
    for table in $(echo .tables | sqlite3 "${daemon_dir}/apollo.db"); do
      echo "SELECT * FROM ${table};" | sqlite3 "${daemon_dir}/apollo.db" >/dev/null
    done

    # Kill the daemon
    apollo shutdown || kill -9 "${daemon_pid}" 2>/dev/null || true

    # Cleanup shmounts (needed due to the forceful kill)
    find "${daemon_dir}" -name shmounts -exec "umount" "-l" "{}" \; >/dev/null 2>&1 || true
    find "${daemon_dir}" -name devapollo -exec "umount" "-l" "{}" \; >/dev/null 2>&1 || true

    check_leftovers="true"
  fi

  if [ -n "${APOLLO_LOGS:-}" ]; then
    echo "==> Copying the logs"
    mkdir -p "${APOLLO_LOGS}/${daemon_pid}"
    cp -R "${daemon_dir}/logs/" "${APOLLO_LOGS}/${daemon_pid}/"
    cp "${daemon_dir}/apollo.log" "${APOLLO_LOGS}/${daemon_pid}/"
  fi

  if [ "${check_leftovers}" = "true" ]; then
    echo "==> Checking for leftover files"
    rm -f "${daemon_dir}/containers/mercury-monitord.log"
    rm -f "${daemon_dir}/security/apparmor/cache/.features"
    check_empty "${daemon_dir}/containers/"
    check_empty "${daemon_dir}/devices/"
    check_empty "${daemon_dir}/images/"
    # FIXME: Once container logging rework is done, uncomment
    # check_empty "${daemon_dir}/logs/"
    check_empty "${daemon_dir}/security/apparmor/cache/"
    check_empty "${daemon_dir}/security/apparmor/profiles/"
    check_empty "${daemon_dir}/security/seccomp/"
    check_empty "${daemon_dir}/shmounts/"
    check_empty "${daemon_dir}/snapshots/"

    echo "==> Checking for leftover DB entries"
    check_empty_table "${daemon_dir}/apollo.db" "containers"
    check_empty_table "${daemon_dir}/apollo.db" "containers_config"
    check_empty_table "${daemon_dir}/apollo.db" "containers_devices"
    check_empty_table "${daemon_dir}/apollo.db" "containers_devices_config"
    check_empty_table "${daemon_dir}/apollo.db" "containers_profiles"
    check_empty_table "${daemon_dir}/apollo.db" "networks"
    check_empty_table "${daemon_dir}/apollo.db" "networks_config"
    check_empty_table "${daemon_dir}/apollo.db" "images"
    check_empty_table "${daemon_dir}/apollo.db" "images_aliases"
    check_empty_table "${daemon_dir}/apollo.db" "images_properties"
    check_empty_table "${daemon_dir}/apollo.db" "images_source"
    check_empty_table "${daemon_dir}/apollo.db" "profiles"
    check_empty_table "${daemon_dir}/apollo.db" "profiles_config"
    check_empty_table "${daemon_dir}/apollo.db" "profiles_devices"
    check_empty_table "${daemon_dir}/apollo.db" "profiles_devices_config"
    check_empty_table "${daemon_dir}/apollo.db" "storage_pools"
    check_empty_table "${daemon_dir}/apollo.db" "storage_pools_config"
    check_empty_table "${daemon_dir}/apollo.db" "storage_volumes"
    check_empty_table "${daemon_dir}/apollo.db" "storage_volumes_config"
  fi

  # teardown storage
  "$apollo_backend"_teardown "${daemon_dir}"

  # Wipe the daemon directory
  wipe "${daemon_dir}"

  # Remove the daemon from the list
  sed "\|^${daemon_dir}|d" -i "${TEST_DIR}/daemons"
}

shutdown_apollo() {
  # APOLLO_DIR is local here because since $(mercury) is actually a function, it
  # overwrites the environment and we would lose APOLLO_DIR's value otherwise.

  # shellcheck disable=2039
  local APOLLO_DIR

  daemon_dir=${1}
  APOLLO_DIR=${daemon_dir}
  daemon_pid=$(cat "${daemon_dir}/apollo.pid")
  echo "==> Killing APOLLO at ${daemon_dir}"

  # Kill the daemon
  apollo shutdown || kill -9 "${daemon_pid}" 2>/dev/null || true
}

cleanup() {
  # Allow for failures and stop tracing everything
  set +ex
  DEBUG=

  # Allow for inspection
  if [ -n "${APOLLO_INSPECT:-}" ]; then
    if [ "${TEST_RESULT}" != "success" ]; then
      echo "==> TEST DONE: ${TEST_CURRENT_DESCRIPTION}"
    fi
    echo "==> Test result: ${TEST_RESULT}"

    # shellcheck disable=SC2086
    printf "To poke around, use:\n APOLLO_DIR=%s APOLLO_CONF=%s sudo -E %s/bin/mercury COMMAND\n" "${APOLLO_DIR}" "${APOLLO_CONF}" ${GOPATH:-}
    echo "Tests Completed (${TEST_RESULT}): hit enter to continue"

    # shellcheck disable=SC2034
    read -r nothing
  fi

  echo "==> Cleaning up"

  # Kill all the APOLLO instances
  while read -r daemon_dir; do
    kill_apollo "${daemon_dir}"
  done < "${TEST_DIR}/daemons"

  # Cleanup leftover networks
  # shellcheck disable=SC2009
  ps aux | grep "interface=apollot$$ " | grep -v grep | awk '{print $2}' | while read -r line; do
    kill -9 "${line}"
  done
  if [ -e "/sys/class/net/apollot$$" ]; then
    ip link del apollot$$
  fi

  # Wipe the test environment
  wipe "${TEST_DIR}"

  echo ""
  echo ""
  if [ "${TEST_RESULT}" != "success" ]; then
    echo "==> TEST DONE: ${TEST_CURRENT_DESCRIPTION}"
  fi
  echo "==> Test result: ${TEST_RESULT}"
}

wipe() {
  if which btrfs >/dev/null 2>&1; then
    rm -Rf "${1}" 2>/dev/null || true
    if [ -d "${1}" ]; then
      find "${1}" | tac | xargs btrfs subvolume delete >/dev/null 2>&1 || true
    fi
  fi

  # shellcheck disable=SC2009
  ps aux | grep mercury-monitord | grep "${1}" | awk '{print $2}' | while read -r pid; do
    kill -9 "${pid}" || true
  done

  if [ -f "${TEST_DIR}/loops" ]; then
    while read -r line; do
      losetup -d "${line}" || true
    done < "${TEST_DIR}/loops"
  fi
  if mountpoint -q "${1}"; then
    umount "${1}"
  fi

  rm -Rf "${1}"
}

configure_loop_device() {
  lv_loop_file=$(mktemp -p "${TEST_DIR}" XXXX.img)
  truncate -s 10G "${lv_loop_file}"
  pvloopdev=$(losetup --show -f "${lv_loop_file}")
  if [ ! -e "${pvloopdev}" ]; then
    echo "failed to setup loop"
    false
  fi
  echo "${pvloopdev}" >> "${TEST_DIR}/loops"

  # The following code enables to return a value from a shell function by
  # calling the function as: fun VAR1

  # shellcheck disable=2039
  local  __tmp1="${1}"
  # shellcheck disable=2039
  local  res1="${lv_loop_file}"
  if [ "${__tmp1}" ]; then
      eval "${__tmp1}='${res1}'"
  fi

  # shellcheck disable=2039
  local  __tmp2="${2}"
  # shellcheck disable=2039
  local  res2="${pvloopdev}"
  if [ "${__tmp2}" ]; then
      eval "${__tmp2}='${res2}'"
  fi
}

deconfigure_loop_device() {
  lv_loop_file="${1}"
  loopdev="${2}"

  SUCCESS=0
  # shellcheck disable=SC2034
  for i in $(seq 10); do
    if losetup -d "${loopdev}"; then
      SUCCESS=1
      break
    fi

    sleep 0.5
  done

  if [ "${SUCCESS}" = "0" ]; then
    echo "Failed to tear down loop device"
    false
  fi

  rm -f "${lv_loop_file}"
  sed -i "\|^${loopdev}|d" "${TEST_DIR}/loops"
}

# Must be set before cleanup()
TEST_CURRENT=setup
TEST_RESULT=failure

trap cleanup EXIT HUP INT TERM

# Import all the testsuites
for suite in suites/*.sh; do
  # shellcheck disable=SC1090
 . "${suite}"
done

# Setup test directory
TEST_DIR=$(mktemp -d -p "$(pwd)" tmp.XXX)
chmod +x "${TEST_DIR}"

if [ -n "${APOLLO_TMPFS:-}" ]; then
  mount -t tmpfs tmpfs "${TEST_DIR}" -o mode=0751
fi

APOLLO_CONF=$(mktemp -d -p "${TEST_DIR}" XXX)
export APOLLO_CONF

APOLLO_DIR=$(mktemp -d -p "${TEST_DIR}" XXX)
export APOLLO_DIR
chmod +x "${APOLLO_DIR}"
spawn_apollo "${APOLLO_DIR}" true
APOLLO_ADDR=$(cat "${APOLLO_DIR}/apollo.addr")
export APOLLO_ADDR

run_test() {
  TEST_CURRENT=${1}
  TEST_CURRENT_DESCRIPTION=${2:-${1}}

  echo "==> TEST BEGIN: ${TEST_CURRENT_DESCRIPTION}"
  START_TIME=$(date +%s)
  ${TEST_CURRENT}
  END_TIME=$(date +%s)

  echo "==> TEST DONE: ${TEST_CURRENT_DESCRIPTION} ($((END_TIME-START_TIME))s)"
}

# allow for running a specific set of tests
if [ "$#" -gt 0 ]; then
  run_test "test_${1}"
  TEST_RESULT=success
  exit
fi

run_test test_check_deps "checking dependencies"
run_test test_static_analysis "static analysis"
run_test test_database_update "database schema updates"
run_test test_remote_url "remote url handling"
run_test test_remote_admin "remote administration"
run_test test_remote_usage "remote usage"
run_test test_basic_usage "basic usage"
run_test test_security "security features"
run_test test_image_expiry "image expiry"
run_test test_image_list_all_aliases "image list all aliases"
run_test test_image_auto_update "image auto-update"
run_test test_image_prefer_cached "image prefer cached"
run_test test_concurrent_exec "concurrent exec"
run_test test_concurrent "concurrent startup"
run_test test_snapshots "container snapshots"
run_test test_snap_restore "snapshot restores"
run_test test_config_profiles "profiles and configuration"
run_test test_config_edit "container configuration edit"
run_test test_config_edit_container_snapshot_pool_config "container and snapshot volume configuration edit"
run_test test_server_config "server configuration"
run_test test_filemanip "file manipulations"
run_test test_network "network management"
run_test test_idmap "id mapping"
run_test test_template "file templating"
run_test test_pki "PKI mode"
run_test test_devapollo "/dev/apollo"
run_test test_fuidshift "fuidshift"
run_test test_migration "migration"
run_test test_fdleak "fd leak"
run_test test_cpu_profiling "CPU profiling"
run_test test_mem_profiling "memory profiling"
run_test test_storage "storage"
run_test test_init_auto "apollo init auto"
run_test test_init_interactive "apollo init interactive"
run_test test_init_preseed "apollo init preseed"
run_test test_storage_profiles "storage profiles"
run_test test_container_import "container import"
run_test test_storage_volume_attach "attaching storage volumes"

TEST_RESULT=success
