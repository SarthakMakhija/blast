<p align="center">
    <img alt="blast" src="https://github.com/SarthakMakhija/blast/assets/21108320/ffb3336c-688f-4b33-b99f-3a26fb35982d" />
</p>

| Platform       | Build Status                                                                                                                  |
|----------------|:------------------------------------------------------------------------------------------------------------------------------|
| Ubuntu latest  | [![blast_ubuntu_latest](https://github.com/SarthakMakhija/blast/actions/workflows/build_ubuntu_latest.yml/badge.svg)](https://github.com/SarthakMakhija/blast/actions/workflows/build_ubuntu_latest.yml)|
| macOS 12       | [![blast_macos_12](https://github.com/SarthakMakhija/blast/actions/workflows/build_macos_12.yml/badge.svg)](https://github.com/SarthakMakhija/blast/actions/workflows/build_macos_12.yml)|


**blast** is a load generator for TCP servers, especially if such servers maintain persistent connections.

## Content Organization

## Why blast

I am a part of the team that is developing a strongly consistent distributed key/value storage engine with support for rich queries.
The distributed key/value storage engine has TCP servers that implement [Single Socket Channel](https://martinfowler.com/articles/patterns-of-distributed-systems/single-socket-channel.html) and [Request Pipeline
](https://martinfowler.com/articles/patterns-of-distributed-systems/request-pipeline.html). 

We needed a way to send load on our servers and get a report with some details including total connections created, total requests sent, total responses read and time to get those responses back etc.

Our servers accept protobuf encoded messages as byte slices, so the tool should be able to send the load (/byte slice) in a format that the target servers
can decode. Almost all distributed systems accept payloads in a very specific format. For example, [JunoDB](https://github.com/paypal/junodb) sends (and receives) [OperationalMessage](https://github.com/paypal/junodb/blob/ca68aa14734768fd047b66ea0b7e6316b15fef16/pkg/proto/opMsg.go#L33) encoded as byte slice.

All we needed was a tool that can send load (or specific load) on target TCP servers, read responses from those servers and present a decent :) report. This was an opportunity to build **blast**.


The logo is built using [logo.com](logo.com)
