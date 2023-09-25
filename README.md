# gman

`gman` enables you to manage your man pages in a git monorepo, with support for decentralized organizations with multiple repositories, release cycles, and delivery channels. `gman` is a simple binary which reads an existing documentation repository and renders the documentation to the user's terminal. `gman` also has a built-in [docusaurus](https://docusaurus.io/) web server which can be used to view the documentation in a web browser.

## Table of Contents <!-- omit from toc -->

- [gman](#gman)
  - [Dependencies](#dependencies)
    - [macOS](#macos)
    - [Linux](#linux)
      - [Debian/Ubuntu](#debianubuntu)
      - [Arch](#arch)
      - [Alpine](#alpine)
      - [CentOS/RHEL](#centosrhel)
    - [Windows](#windows)
  - [Installation](#installation)
    - [Docker Image](#docker-image)
  - [Building](#building)
  - [Usage](#usage)
  - [Concepts](#concepts)
    - [Namespaces](#namespaces)
    - [gman page](#gman-page)
    - [gman repo](#gman-repo)
      - [gman repo structure](#gman-repo-structure)
    - [Releases](#releases)
    - [~/.gman](#gman-1)
      - [Configuration](#configuration)
  - [Features](#features)
    - [List](#list)
    - [Search](#search)
    - [Releases](#releases-1)
    - [Print Man Dir](#print-man-dir)
    - [tl;dr](#tldr)
    - [Web](#web)
      - [Deployment](#deployment)


## Dependencies

`gman` requires the following dependencies:

- [git](https://git-scm.com/)
- [pandoc](https://pandoc.org/)
- [groff](https://www.gnu.org/software/groff/)
- [less](https://www.gnu.org/software/less/)
- [xdg-open](https://www.freedesktop.org/wiki/Software/xdg-utils/)
    - only used if `-open` flag is set to `true` and `gman` is unable to fetch the content of a URL
- [node](https://nodejs.org/en/) and [npm](https://www.npmjs.com/)
    - only used if `-web` flag is set to `true`

### macOS

On macOS, you can install these dependencies via [homebrew](https://brew.sh/):

```bash
brew install git pandoc groff less
# to use the web server
brew install nodejs npm
```

### Linux

#### Debian/Ubuntu

On Debian/Ubuntu, you can install these dependencies via `apt`:

```bash
sudo apt install git pandoc groff less xdg-utils
# to use the web server
sudo apt install nodejs npm
```

#### Arch

On Arch, you can install these dependencies via `pacman`:

```bash
sudo pacman -S git pandoc groff less xdg-utils
# to use the web server
sudo pacman -S nodejs npm
```

#### Alpine

On Alpine, you can install these dependencies via `apk`:

```bash
sudo apk add git pandoc groff less xdg-utils
# to use the web server
sudo apk add nodejs npm
```

#### CentOS/RHEL

On CentOS/RHEL, you can install these dependencies via `yum`:

```bash
sudo yum install git pandoc groff less xdg-utils
# to use the web server
sudo yum install nodejs npm
```

### Windows

On Windows, you can install these dependencies via [chocolatey](https://chocolatey.org/):

```bash
choco install git pandoc groff less
# to use the web server
choco install nodejs npm
```

## Installation

First, ensure you have the dependencies installed.

Download the latest release from the [releases page](https://github.com/robertlestak/gman/releases), and place it in your `$PATH`.

You can use [install-release](https://github.com/Rishang/install-release) to do this easily:

```bash
# first install install-release
pip install -U install-release

# then install gman
install-release get https://github.com/robertlestak/gman
```

### Docker Image

A docker image is available at [robertlestak/gman](https://hub.docker.com/r/robertlestak/gman).

To use, simply mount your host `~/.gman` directory and `~/.netrc` file into the container:

```bash
docker run -it --rm -v $HOME/.gman:/root/.gman -v $HOME/.netrc:/root/.netrc robertlestak/gman app1
```

You can set an alias to make this easier:

```bash
alias gman='docker run -it --rm -v $HOME/.gman:/root/.gman -v $HOME/.netrc:/root/.netrc robertlestak/gman'
```

Now, from here on out, you can use `gman` as you normally would:

```bash
gman app1
```

Note that if you use the docker image, the `-open` flag will not work, as the docker container does not have access to your host's default browser. 

Also note, there appears to be a memory leak issue with `pandoc` when running in a docker container if running an emulated architecture (i.e., running an `amd` docker image on an `arm` host). If you face this, ensure you have set your docker memory limit to at least 2GB, and have set your `--platform` flag in your `docker run` command to match your host architecture. 

For example, on a Mac with an `arm` processor, you would run:

```bash
docker run -it --rm -v $HOME/.gman:/root/.gman -v $HOME/.netrc:/root/.netrc --platform linux/arm64 robertlestak/gman app1
```

Alternatively, you can set the `-render=false` flag to disable rendering, which will prevent `pandoc` from being used.

## Building

To build `gman` from source, you will need to have [go](https://golang.org/) installed.

```bash
# clone the repo
git clone https://git.shdw.tech/rob/gman
cd gman
# build and install the binary
# note, it will ask for sudo to install the binary to /usr/local/bin
make install
```

## Usage

The default usage of `gman` aims to mirror the default usage of `man` - that is, given a single argument, `gman` will attempt to return a single man page for the argument.

The additional features of `gman` are available via flags.

```bash
Usage of gman:
  -A	all namespaces
  -branch string
    	git branch (default "main")
  -config string
    	local directory (default "~/.gman")
  -dir
    	print man dir instead of showing contents
  -interval string
    	update interval (default "24h")
  -log string
    	log level (default "info")
  -n string
    	namespace (default "default")
  -notify
    	notify on new releases (default true)
  -ns
    	list namespaces
  -o string
    	output format for lists. text, json, yaml (default "text")
  -open
    	open url on get failure
  -pager string
    	pager (default "less")
  -pull
    	update repo now
  -r	show releases
  -render
    	render markdown (default true)
  -repo string
    	git repo
  -search string
    	search
  -t	show tldr
  -version
    	show version
  -web
    	run web server
  -web-addr string
    	web server address (default ":8080")
  -web-dir string
    	web server directory. (default "~/.gman/web")
```

## Concepts

### Namespaces

In larger organizations there may be name collisions or logical boundaries between groups. To address this, all man pages in `gman` are namespaced. When searching for a man page, if no namespace is specified, first the `default` namespace is searched before searching all other namespaces. The first match is used.

### gman page

In conventional `man` pages, the entire documentation is contained in a [single document](https://man.freebsd.org/cgi/man.cgi?query=mdoc&sektion=7).

`gman` extends this to enable both the full usage documenation (`README.md`) as well as an optional abbreviated version (`TLDR.md`).

Additionally, an optional `examples` directory can be included containing usage examples.

If either `README.md` or `TLDR.md` contain just a URL, `gman` will attempt to fetch the content from that URL and use it as the content of the man page. If `gman` is unable to fetch the content, it will return the URL as-is, and the user can attempt to fetch the content manually. If the `-open` flag is set to `true`, `gman` will attempt to open the URL in the user's default browser.

If you have a `~/.netrc` file with credentials for the URL, `gman` will attempt to use those credentials when fetching the content.

Documentation should be written in [Markdown](https://www.markdownguide.org/cheat-sheet/), and will be rendered to the user's terminal using [pandoc](https://pandoc.org/) and [groff](https://www.gnu.org/software/groff/). To disable rendering, use the `-render=false` flag.

### gman repo

`gman` requires a single monorepo to act as the "single source of truth" used to present the current global man page set. This repo is called the `gman repo`. This repo can include submodules as well as links to external data sources, such as other git repos, wiki pages, or web pages.

#### gman repo structure

The `gman repo` must follow a basic structure in order to be used by `gman`.

```
docs/
    <optional>README.md
    {namespace}/
        {app}/
            README.md
            <optional>TLDR.md
            <optional>examples/
<optional>releases/
    {version}/
        README.md
```

Example:

```
docs/
    README.md
    default/
        app1/
            README.md
            TLDR.md
        app2/
            examples/
                hello.sh
            README.md
releases/
    v0.0.1/
        README.md
    v0.0.2/
        README.md
```

Within the `docs` directory, there should be a directory for each namespace. Within each namespace directory, there should be a directory for each app. Within each app directory, there should be a `README.md` file containing the full documentation for the app. Optionally, there can be a `TLDR.md` file containing a short description of the app. Optionally, there can be an `examples` directory containing usage examples for the app.

Optionally, there can be a `README.md` file in the root of the `docs` directory. This is currently not directly read by `gman`, however if it exists, will be displayed when viewing in a web browser.

Submodules are supported, and can be used to include additional documentation from other repos. Submodules are recursively updated on each update of the `gman repo`.

### Releases

Releases are a way to communicate significant changes or new features to users. Releases are optional, and are not required to use `gman`. When used, releases are stored in the `releases` directory of the `gman repo`.

When releases are used, each time `gman` updates the user's local copy of the `gman repo`, it will check the `releases` directory for any new releases. If a new release is found, `gman` will display the release notes to the user. Unlike `gman` pages, releases are not namespaced.

It is recommended to use [semantic versioning](https://semver.org/) for releases. When used, `gman` will attempt to sort releases by version number, and display the most recent release first.

Users who do not wish to see release notifications can disable them via the `-notify=false` flag, or by setting `notify: false` in their `~/.gman/config.yaml` file.

Releases can be viewed via the `-r` flag. To read a specific release, use the `-r` flag with the release version as an argument, eg:

```bash
gman -r v0.0.1
```

You can also search in releases using the `-search` flag, eg:

```bash
gman -r -search foo
```

### ~/.gman

`gman` uses the `~/.gman` directory to store configuration metadata, as well as the local copies of the `gman repo` and any submodules. If you are familiar with the `$GOPATH` concept, `~/.gman` is conceptually similar to `$GOPATH`. Within `~/.gman` there is a `src` directory which contains the local copies of the `gman repo`(s) and any submodules.

`gman` will automatically update the local copies of the `gman repo` and any submodules on a regular interval. This interval can be configured via the `-interval` flag, or by setting `interval: 24h` in the `~/.gman/config.yaml` file.

#### Configuration

`gman` can be configured via an optional `config.yaml` file in the root of the local directory (default: `~/.gman`). This file is a YAML file with the following structure:

```yaml
---
# git pull interval
interval: 2h
# open URLs in browser on GET failure
open: false
# set a default namespace other than "default"
namespace: foobar
# notify on new releases
notify: false
# pager to use
pager: less
# render markdown
render: false
# show tldr
tldr: true
# web mode
web: false
# web address
webAddr: :8080
# web dir
webDir: web
# default repo to use
repo: foo
# configured repos
repos:
  foo:
    url: https://git.shdw.tech/rob/gman-docs-test
    branch: main
  another:
    url: https://git.shdw.tech/rob/gman-docs-test-2
    branch: develop
```

This enables you to set a default repo to use, as well as additional repos which can be referenced by a given short-name, eg:

```bash
# use the default repo
gman app1
# use a repo by a given short-name rather than full URL and branch
gman -repo another app1
# the above is the same as
gman -repo https://git.shdw.tech/rob/gman-docs-test-2 -branch develop app1
```

## Features

In addition to the basic "show manpage for app", `gman` also supports the following features:

### List

The default usage of `gman` is to list all apps in the current namespace (default: `default`). If the `-A` flag is passed, all apps in all namespaces are listed. If the `-n` flag is passed, all apps in the given namespace are listed. To list all namespaces, use the `-ns` flag.

```bash
# list all apps in the default namespace
gman
# list all apps in all namespaces
gman -A
# list apps in the foo namespace
gman -n foo
# list all namespaces
gman -ns
```

### Search

The `-search` flag will search all apps in the `gman repo` for the given search term. If a namespace is specified, only apps in that namespace will be searched. Both exact string and regex searches are supported.

```bash
# search all apps
gman -search foo
# search apps in a given namespace
gman -search foo -n bar
# regex search
gman -search '^foo.*bar$'
```

### Releases

The `-r` flag will list all releases in the `gman repo`. If a search term is specified, only releases matching the search term will be listed. Both exact string and regex searches are supported.

```bash
# list all releases
gman -r
# list releases matching a search term
gman -r -search foo
# regex search
gman -r -search '^foo.*bar$'
# show a specific release
gman -r v0.0.1
```

### Print Man Dir

By default, `gman` will return the rendered documentation for the given application. If the `-dir` flag is passed, instead of printing the documentation, `gman` will print the local directory containing the documentation.

This allows you to pass this into your IDE or another tool to navigate the documentation and examples.

```bash
# print the directory containing the documentation
gman -dir app1
# open the directory containing the documentation in VSCode
code $(gman -dir app1)
```

Of course, since everything is git, you can edit the documentation directly in the `gman repo` and push your changes back to the remote.

### tl;dr

If a `TLDR.md` file is present in the app's directory, `gman` will print the contents of that file instead of the full documentation if the `-t` flag is passed. If a package does not have a `TLDR.md` file, the full documentation will be printed.

```bash
# print the docs
gman app1
# important bits blah blah blah
gman -t app1
# run this with sudo, to the point
```

### Web

If the `-web` flag is passed (or `web: true` is set in the `~/.gman/config.yaml` file), `gman` will start a [docusaurus](https://docusaurus.io/) web server which can be used to view the documentation in a web browser.

```bash
# start the web server
gman -web
```

This will automatically update the web server when the `gman repo` is updated, at the interval specified in the `~/.gman/config.yaml` file, or via the `-interval` flag.

The `deploy` directory contains an example Kubernetes deployment for the web server.

#### Deployment

First, edit the manifests to suit your needs. "Sensible defaults" have been set, but be sure to review and update as needed.

```bash
kubectl apply -f deploy
```
