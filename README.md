# fani

FANi is a [Function Addressable Compute Network](https://youtu.be/NWGC4S-eZW4) implementation. This repository is strongly inspired on [ipfs-compute](https://github.com/adlrocha/ipfs-compute.

## Overview

The goal of this project is to extend the proof of concept linked above to provide an easy to use CLI and framework to deploy functions to the IPFS-FAN network, and an (ideally) WASM-compatible basic node implementation to execute FAN computations

Like ipfs-compute, FANi utilizes an [ipfs-lite](https://github.com/hsanjuan/ipfs-lite peer to interact with the network.

## WASM Function Compatibility

This project is currently using the excellent [sat](https://github.com/suborbital/sat) [suborbital](https://suborbital.dev/) tools to execute wasm modules. Any "runnable" built using the [subo](https://docs.suborbital.dev/subo/) CLI should work out of the box.

While FANi currently deals only with the wasm modules directly. So while you may develop your functions using subo/atmo, the directives or runnable.yaml files will not be used at all.

Integrating the CLI with suborbital directives or perhaps a more generic configuration/declaration could be a possibility to orchestrate more complex computations and provide metadata about how individual functions should be executed.
