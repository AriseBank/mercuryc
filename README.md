# APOLLO

REST API, command line tool and OpenStack integration plugin for MERCURY.

APOLLO is pronounced lex-dee.

To easily see what APOLLO is about, you can [try it online](https://linuxcontainers.org/apollo/try-it).

## Status

* GoDoc: [![GoDoc](https://godoc.org/github.com/AriseBank/apollo-controller/client?status.svg)](https://godoc.org/github.com/AriseBank/apollo-controller/client)
* Jenkins (Linux): [![Build Status](https://jenkins.linuxcontainers.org/job/apollo-github-commit/badge/icon)](https://jenkins.linuxcontainers.org/job/apollo-github-commit/)
* Travis (macOS): [![Build Status](https://travis-ci.org/mercury/apollo.svg?branch=master)](https://travis-ci.org/mercury/apollo/)
* AppVeyor (Windows): [![Build Status](https://ci.appveyor.com/api/projects/status/rb4141dsi2xm3n0x/branch/master?svg=true)](https://ci.appveyor.com/project/mercury/apollo/)
* Weblate (translations): [![Translation status](https://hosted.weblate.org/widgets/linux-containers/-/svg-badge.svg)](https://hosted.weblate.org/projects/linux-containers/apollo/)


## Getting started with APOLLO

Since APOLLO development is happening at such a rapid pace, we only provide daily
builds right now. They're available via:

    sudo add-apt-repository ppa:ubuntu-mercury/apollo-git-master && sudo apt-get update
    sudo apt-get install apollo

Because group membership is only applied at login, you then either need to
close and re-open your user session or use the "newgrp apollo" command in the
shell you're going to interact with apollo from.

    newgrp apollo

After you've got APOLLO installed and a session with the right permissions, you
can take your [first steps](#first-steps).

#### Getting started with APOLLO on Windows

APOLLO server is not available on Windows, but it is possible to use
[`mercury` client](https://ci.appveyor.com/project/mercury/apollo/branch/master/artifacts)
with
[some limitations](https://github.com/AriseBank/apollo-controller/issues?utf8=%E2%9C%93&q=is%3Aissue%20is%3Aopen%20windows)
to control remote containers.


## Using the REST API
The APOLLO REST API can be used locally via unauthenticated Unix socket or remotely via SSL encapsulated TCP.

#### via Unix socket
```bash
curl --unix-socket /var/lib/apollo/unix.socket \
    -H "Content-Type: application/json" \
    -X POST \
    -d @hello-ubuntu.json \
    apollo/1.0/containers
```

#### via TCP
TCP requires some additional configuration and is not enabled by default.
```bash
mercury config set core.https_address "[::]:8443"
```
```bash
curl -k -L \
    --cert ~/.config/mercury/client.crt \
    --key ~/.config/mercury/client.key \
    -H "Content-Type: application/json" \
    -X POST \
    -d @hello-ubuntu.json \
    "https://127.0.0.1:8443/1.0/containers"
```
#### JSON payload
The `hello-ubuntu.json` file referenced above could contain something like:
```json
{
    "name":"some-ubuntu",
    "ephemeral":true,
    "config":{
        "limits.cpu":"2"
    },
    "source": {
        "type":"image",
        "mode":"pull",
        "protocol":"simplestreams",
        "server":"https://cloud-images.ubuntu.com/releases",
        "alias":"14.04"
    }
}
```

## Building from source

We recommend having the latest versions of libmercury (>= 2.0.0 required) and CRIU
(>= 1.7 recommended) available for APOLLO development. Additionally, APOLLO requires
Golang 1.5 or later to work. All the right versions dependencies are available
via the APOLLO PPA:

    sudo apt-get install software-properties-common
    sudo add-apt-repository ppa:ubuntu-mercury/apollo-git-master
    sudo apt-get update
    sudo apt-get install acl dnsmasq-base git golang libmercury1 mercury-dev make pkg-config rsync squashfs-tools tar xz-utils

There are a few storage backends for APOLLO besides the default "directory"
backend. Installing these tools adds a bit to initramfs and may slow down your
host boot, but are needed if you'd like to use a particular backend:

    sudo apt-get install lvm2 thin-provisioning-tools
    sudo apt-get install btrfs-tools

To run the testsuite, you'll also need:

    sudo apt-get install curl gettext jq sqlite3 uuid-runtime bzr


### Building the tools

APOLLO consists of two binaries, a client called `mercury` and a server called `apollo`.
These live in the source tree in the `mercury/` and `apollo/` dirs, respectively. To
get the code, set up your go environment:

    mkdir -p ~/go
    export GOPATH=~/go

And then download it as usual:

    go get github.com/mercury/apollo
    cd $GOPATH/src/github.com/mercury/apollo
    make

...which will give you two binaries in $GOPATH/bin, `apollo` the daemon binary,
and `mercury` a command line client to that daemon.

### Machine Setup

You'll need sub{u,g}ids for root, so that APOLLO can create the unprivileged
containers:

    echo "root:1000000:65536" | sudo tee -a /etc/subuid /etc/subgid

Now you can run the daemon (the --group sudo bit allows everyone in the sudo
group to talk to APOLLO; you can create your own group if you want):

    sudo -E $GOPATH/bin/apollo --group sudo

## First steps

APOLLO has two parts, the daemon (the `apollo` binary), and the client (the `mercury`
binary). Now that the daemon is all configured and running (either via the
packaging or via the from-source instructions above), you can create a container:

    $GOPATH/bin/mercury launch ubuntu:14.04

Alternatively, you can also use a remote APOLLO host as a source of images.
One comes pre-configured in APOLLO, called "images" (images.linuxcontainers.org)

    $GOPATH/bin/mercury launch images:centos/7/amd64 centos

## Bug reports

Bug reports can be filed at https://github.com/AriseBank/apollo-controller/issues/new

## Contributing

Fixes and new features are greatly appreciated but please read our
[contributing guidelines](CONTRIBUTING.md) first.

Contributions to this project should be sent as pull requests on github.

## Hacking

Sometimes it is useful to view the raw response that APOLLO sends; you can do
this by:

    mercury config set core.trust_password foo
    mercury remote add local 127.0.0.1:8443
    wget --no-check-certificate https://127.0.0.1:8443/1.0 --certificate=$HOME/.config/mercury/client.crt --private-key=$HOME/.config/mercury/client.key -O - -q

## Upgrading

The `apollo` and `mercury` (`apollo-client`) binaries should be upgraded at the same time with:

    apt-get update
    apt-get install apollo apollo-client

## Support and discussions

We use the MERCURY mailing-lists for developer and user discussions, you can
find and subscribe to those at: https://lists.linuxcontainers.org

If you prefer live discussions, some of us also hang out in
[#mercuryontainers](http://webchat.freenode.net/?channels=#mercuryontainers) on irc.freenode.net.


## FAQ

#### How to enable APOLLO server for remote access?

By default APOLLO server is not accessible from the networks as it only listens
on a local unix socket. You can make APOLLO available from the network by specifying
additional addresses to listen to. This is done with the `core.https_address`
config variable.

To see the current server configuration, run:

    mercury config show

To set the address to listen to, find out what addresses are available and use
the `config set` command on the server:

    ip addr
    mercury config set core.https_address 192.168.1.15

#### When I do a `mercury remote add` over https, it asks for a password?

By default, APOLLO has no password for security reasons, so you can't do a remote
add this way. In order to set a password, do:

    mercury config set core.trust_password SECRET

on the host APOLLO is running on. This will set the remote password that you can
then use to do `mercury remote add`.

You can also access the server without setting a password by copying the client
certificate from `.config/mercury/client.crt` to the server and adding it with:

    mercury config trust add client.crt


#### How do I configure APOLLO storage?

APOLLO supports btrfs, directory, lvm and zfs based storage.

First make sure you have the relevant tools for your filesystem of
choice installed on the machine (btrfs-progs, lvm2 or zfsutils-linux).

By default, APOLLO comes with no configured network or storage.
You can get a basic configuration done with:

    apollo init

"apollo init" supports both directory based storage and ZFS.
If you want something else, you'll need to use the "mercury storage" command:

    mercury storage create default BACKEND [OPTIONS...]
    mercury profile device add default root disk path=/ pool=default

BACKEND is one of "btrfs", "dir", "lvm" or "zfs".

Unless specified otherwise, APOLLO will setup loop based storage with a sane default size.

For production environments, you should be using block backed storage
instead both for performance and reliability reasons.

#### How can I live migrate a container using APOLLO?

Live migration requires a tool installed on both hosts called
[CRIU](http://criu.org), which is available in Ubuntu via:

    sudo apt-get install criu

Then, launch your container with the following,

    mercury launch ubuntu $somename
    sleep 5s # let the container get to an interesting state
    mercury move host1:$somename host2:$somename

And with luck you'll have migrated the container :). Migration is still in
experimental stages and may not work for all workloads. Please report bugs on
mercury-devel, and we can escalate to CRIU lists as necessary.

#### Can I bind mount my home directory in a container?

Yes. The easiest way to do that is using a privileged container:

1.a) create a container.

    mercury launch ubuntu privilegedContainerName -c security.privileged=true

1.b) or, if your container already exists.

        mercury config set privilegedContainerName security.privileged true
2) then.

    mercury config device add privilegedContainerName shareName disk source=/home/$USER path=/home/ubuntu

#### How can I run docker inside a APOLLO container?

In order to run Docker inside a APOLLO container the `security.nesting` property of the container should be set to `true`. No other changes should be necessary.

    mercury config set <container> security.nesting true
