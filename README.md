# fani

FANi is a [Function Addressable Compute Network](https://youtu.be/NWGC4S-eZW4) implementation. This repository is strongly inspired on [ipfs-compute](https://github.com/adlrocha/ipfs-compute).

## Status

Currently just a very rough POC.

## Overview

The goal of this project is to extend the proof of concept linked above to provide an easy to use CLI and framework to deploy functions to the IPFS-FAN network, and an (ideally) WASM-compatible basic node implementation to execute FAN computations

Like ipfs-compute, FANi utilizes an [ipfs-lite](https://github.com/hsanjuan/ipfs-lite) peer to interact with the network.

## WASM Function Compatibility

This project is currently using the excellent [sat](https://github.com/suborbital/sat) and [subo](https://docs.suborbital.dev/subo/usage) tools from [suborbital](https://suborbital.dev/) to execute wasm modules. Any "runnable" built using the [subo](https://docs.suborbital.dev/subo/) CLI should work out of the box.

While FANi currently deals only with the wasm modules directly. So while you may develop your functions using subo/atmo, the directives or runnable.yaml files will not be used at all.

Integrating the CLI with suborbital directives or perhaps a more generic configuration/declaration could be a possibility to orchestrate more complex computations and provide metadata about how individual functions should be executed.

## Usage

Follow the [subo](https://docs.suborbital.dev/subo/usage) documentation to create a basic helloworld rust function wasm module.

```
➜ subo create runnable rs-runnable
➜ subo build rs-runnable
```

Then deploy the compiled wasm function

```
➜ ./fani deploy rs-hello rs-runnable/rs-runnable.wasm
created libp2p host:  QmTw68adhzSKncUW4y79kmvhcjEnezwJBo8xL1boYaGojC
listenting on:  /ip4/192.168.86.54/tcp/37551
added bytecode bafybeicfd7o3glgy6746vkhg54syzwaf5bvdplo5iz63df3w45h4ljunom
added abi bafybeicon5jwwcukiy7pyuqnjltbtrgyswv7iggqrghy6zndtrtmb5ir2m
sitting idle to provide deployed dag
```

Keep fani running, so it's ipfs-lite peer can provide the function content

In a second terminal, call the function via it's ABI CID

```
➜ ./fani call bafybeicon5jwwcukiy7pyuqnjltbtrgyswv7iggqrghy6zndtrtmb5ir2m

created libp2p host:  QmTtxyw8tji2Ncae7tNjPGPChvb7hH8MbGr3WZvLAGpbVM
listenting on:  /ip4/192.168.86.54/tcp/37795
resolving ABI:  bafybeicon5jwwcukiy7pyuqnjltbtrgyswv7iggqrghy6zndtrtmb5ir2m
{ID:rs-hello ByteCode:bafybeicfd7o3glgy6746vkhg54syzwaf5bvdplo5iz63df3w45h4ljunom Args:[]}
getting bytecode
result:
hello
result added to network:  bafybeiauirvsiecy5h6spipofhjsqpz26mxj65yazekpvw4w3askiflr2i
```

You may also pass args after the function CID. They will be added to the
network, and their content retrieved via CID will be passed to the function

```
➜ ./fani call bafybeicon5jwwcukiy7pyuqnjltbtrgyswv7iggqrghy6zndtrtmb5ir2m "world, \nlove fani"

created libp2p host:  QmXsBCLexfzRbRitkmKjWomiAwFhAsmL9axjxQMgX1y8CT
listenting on:  /ip4/192.168.86.54/tcp/36715
resolving ABI:  bafybeicon5jwwcukiy7pyuqnjltbtrgyswv7iggqrghy6zndtrtmb5ir2m
discovered new peer: QmZeEFgBZivfHc4cXig99MgaWWR2MkYHbvAvajvdGoL2ah
{ID:rs-hello ByteCode:bafybeicfd7o3glgy6746vkhg54syzwaf5bvdplo5iz63df3w45h4ljunom Args:[]}
getting bytecode
result:
hello world, \nlove fani
result added to network:  bafybeibwt4ul52ct2azbteljy2c6sp7iidqwiitdxfmacao6dvsgrd6cdy
```

You can then open a third terminal, and retrieve the result from ipfs

```
➜ ipfs cat bafybeibwt4ul52ct2azbteljy2c6sp7iidqwiitdxfmacao6dvsgrd6cdy
hello world, \nlove fani
```
