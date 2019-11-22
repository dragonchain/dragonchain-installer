# Dragonchain Installer

This is a cross-platform (windows/mac/linux) tool for aiding in installing and setting up an unmanaged Dragonchain and all its associated components.

This tool utilizes minikube to create a virtual machine to run kubernetes.

**Note:** If you are running this installer in a virtual machine, it may not work because of nesting VMs.

## Installing

In order to run this tool, simply download and run the appropriate executable provided [in the latest release](https://github.com/dragonchain/dragonchain-installer/releases/latest)

If you are on linux or macos, you can alternatively run this command in a terminal as a shortcut:

```sh
curl -Lf https://raw.githubusercontent.com/dragonchain/dragonchain-installer/master/scripts/get_installer.bash -o installer.bash && bash installer.bash
```

## Configuring

Currently, all the configuration options are asked when running the installer.

For Dragon Net support, use the [Dragonchain Console](https://console.dragonchain.com/) to create an unmanaged chain, which will contain the tokens you need to configure with dragon net.

We expect to expand these configuration options in the future.

## User-Feedback

User feedback is encouraged, and can either be provided using [github issues](https://github.com/dragonchain/dragonchain-installer/issues) in this repository, or [joining the dragonchain developer slack](https://forms.gle/ec7sACnfnpLCv6tXA) and talking there.

## Building From Source

In order to build this project from source, ensure that you have golang 1.13 or later installed, then simply run `go get -u github.com/dragonchain/dragonchain-installer/cmd/dc-installer`.

This will download and build the executable `dc-installer` into your `$GOPATH/bin`
