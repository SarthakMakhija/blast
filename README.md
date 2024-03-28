<p align="center">
   <img alt="blast" src="https://github.com/SarthakMakhija/blast/assets/21108320/0c282eb8-fb21-4294-bccd-12a81426b894" />
</p>

| Platform      | Build Status                                                                                                                                                                                             |
|---------------|:---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Ubuntu latest | [![blast_ubuntu_latest](https://github.com/SarthakMakhija/blast/actions/workflows/build_ubuntu_latest.yml/badge.svg)](https://github.com/SarthakMakhija/blast/actions/workflows/build_ubuntu_latest.yml) |
| macOS 12      | [![blast_macos_12](https://github.com/SarthakMakhija/blast/actions/workflows/build_macos_12.yml/badge.svg)](https://github.com/SarthakMakhija/blast/actions/workflows/build_macos_12.yml)                |


**blast** is a load generator for TCP servers, especially if such servers maintain persistent connections.

## Content Organization

- [Why blast](#why-blast)
- [Features](#features)
- [Installation](#installation)
- [FAQs](#faqs)
- [Screenshots](#screenshots)
- [References](#references)

## Why blast

I am a part of the team that is developing a strongly consistent distributed key/value storage engine with support for rich queries.
The distributed key/value storage engine has TCP servers that implement [Single Socket Channel](https://martinfowler.com/articles/patterns-of-distributed-systems/single-socket-channel.html) and [Request Pipeline
](https://martinfowler.com/articles/patterns-of-distributed-systems/request-pipeline.html). 

We needed a way to send load on our servers and get a report with details including total connections established, total requests sent, total responses read and time to get those responses back etc.

Another detail, our servers accept protobuf encoded messages as byte slices, so the tool should be able to send the load (/byte slice) in a format that the target servers
can decode. Almost all distributed systems accept payloads in a very specific format. For example, [JunoDB](https://github.com/paypal/junodb) sends (and receives) [OperationalMessage](https://github.com/paypal/junodb/blob/ca68aa14734768fd047b66ea0b7e6316b15fef16/pkg/proto/opMsg.go#L33) encoded as byte slice.

All we needed was a tool that can send load (or specific load) on target TCP servers, read responses from those servers and present a decent :) report. This was an opportunity to build **blast**. **blast** is inspired from [hey](https://github.com/rakyll/hey), which is an HTTP load generator in golang.

*With version 0.0.3, blast is a very thin CLI with a dependency on [blast-core](https://github.com/SarthakMakhija/blast-core).*  

## Features

**blast** provides the following features:
1. Support for sending **N requests per second per worker**.
2. Support for **reading N total responses** from the target server.
3. Support for **reading N successful responses** from the target server.
4. Support for **customizing** the **load** **duration**. By default, blast runs for 20 seconds.
5. Support for sending requests to the target server with the specified **concurrency level**.
6. Support for **establishing N connections** to the target server.
7. Support for specifying the **connection timeout**.
8. Support for **printing** the **report**.
9. Support for sending dynamic payloads with **PayloadGenerator**.

## Installation

### MacOS

1. **Download the current release**

`wget -o - https://github.com/SarthakMakhija/blast/releases/download/v0.0.6/blast_Darwin_x86_64.tar.gz`

3. **Unzip the release in a directory**

`mkdir blast && tar xvf blast_Darwin_x86_64.tar.gz -C blast`

### Linux AMD64

1. **Download the current release**

`wget -o - https://github.com/SarthakMakhija/blast/releases/download/v0.0.6/blast_Linux_x86_64.tar.gz`

2. **Unzip the release in a directory**

`mkdir blast && tar xvf blast_Linux_x86_64.tar.gz -C blast`

## Supported flags

| **Flag** | **Description**                                                                                                                                                                                                                                                                                                                           |
|----------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| c        | Number of workers to run concurrently. **Default is 50.**                                                                                                                                                                                                                                                                                 |
| f        | File path containing the payload.                                                                                                                                                                                                                                                                                                         |
| rps      | Rate limit in requests per second per worker. **Default is 50.**                                                                                                                                                                                                                                                                          |
| z        | Duration of the blast to run. **Default is 20 seconds.**                                                                                                                                                                                                                                                                                  |
| t        | Timeout for establishing connection with the target server. **Default is 3 seconds.**                                                                                                                                                                                                                                                     |
| Rr       | Read responses from the target server. **Default is false.**                                                                                                                                                                                                                                                                              |
| Rrs      | Read response size is the size of the responses in bytes returned by the target server.                                                                                                                                                                                                                                                   |
| Rrd      | Read response deadline defines the deadline for the read calls on connection. **Default is no deadline** which means the read calls do not timeout.                                                                                                                                                                                       |
| Rtr      | Read total responses is the total responses to read from the target server. The load generation will stop if either the duration (-z) has exceeded or the total responses have been read. This flag is applied only if "Read responses" (-Rr) is true.                                                                                    |
| Rsr      | Read successful responses  is the successful responses to read from the target server. The load generation will stop if either the duration (-z) has exceeded or the total successful responses have been read. Either of "-Rtr" or "-Rsr" must be specified, if -Rr is set. This flag is applied only if "Read responses" (-Rr) is true. |
| conn     | Number of connections to open with the target server. **Default is 1.**                                                                                                                                                                                                                                                                   |
| cpu      | Number of cpu cores to use. **Default is the number of logical CPUs.**                                                                                                                                                                                                                                                                    |

## FAQs

1. **Can I use blast to only send the load and not worry about getting the responses back?**
   
Yes.

The following command sends 100 requests per second per worker, over 10 TCP connections using 100 concurrent workers.
```sh
./blast -rps 100 -c 100 -conn 10 -f ./payload localhost:8989
```

2. **Are the workers implemented using goroutines?**
   
Yes, workers are implemented as cooperative goroutines. You can refer the code [here](https://github.com/SarthakMakhija/blast-core/blob/main/workers/worker.go).

3. **Can I create more connections than workers?**

No, you can not create more connections that workers. The relationship between the concurrency and the workers is simple: `concurrency % workers must be equal to zero`.
This means, we can have 100 workers with 10 connections, where a group of 10 workers will share one connection.

You can refer the code [here](https://github.com/SarthakMakhija/blast-core/blob/main/workers/worker_group.go#L77).

4. **My server takes a protobuf encoded byte slice. How do I pass the payload to blast?**

**blast** supports reading the payload from a file. The payload that needs to be sent to the target server can be written to a file in a separate process and then the file can be passed
as an option to the **blast**. Let's look at the pseudocode:

```go
    func main() {
        message := &ProtoMessage {....}
        encoded, err := proto.Marshal(message)
        assert.Nil(err)

        file, err := os.Create("payload")
        assert.Nil(err)
        defer func() {
            _ = file.Close()
        }()

        _, err = file.Write(encoded)
	    assert.Nil(t, err)
    }
```

The above code creates a protobuf encoded message and writes it to a file. The file can then be provided using `-f` option to the **blast**.

5. **blast provides a feature to read responses. How is response reading implemented?**

[ResponseReader](https://github.com/SarthakMakhija/blast-core/blob/main/report/response_reader.go) implements one goroutine per `net.Conn` to read responses from connections.
The goroutine keeps on reading from the connection, and tracks successful and failed reads. This design means that there will be 1M response reader goroutines if the user
wants to establish 1M connections and read responses. To handle this, IO multiplexing + pool of ResponseReaders is planned in subsequent release.

6. **What is the significance of Rrs flag in blast?**

To read responses from connections, **blast** needs to know the response payload size. The flag `Rrs` signifies the size of the response payload in bytes (or the size of the
byte slice) that [ResponseReader](https://github.com/SarthakMakhija/blast-core/blob/main/report/response_reader.go) should read in each iteration.

7. **What is the significance of Rrd flag in blast?**

`Rrd` is the read response deadline flag that defines the deadline for the read calls on connections. 
This flag helps in understanding the responsiveness of the target server. Let's consider that we are running **blast** with the following command: 

`./blast -c 100 -conn 100  -f ./payload -Rr -Rrs 19 -Rrd 10ms -Rtr 200000 localhost:8989`.

Here, `Rrd` is 10 milliseconds, this means that the `read` calls in [ResponseReader](https://github.com/SarthakMakhija/blast-core/blob/main/report/response_reader.go) will block for 10ms and 
then timeout if there is no response on the underlying connection.

## References
[hey](https://github.com/rakyll/hey)

*The logo is built using [logo.com](https://logo.com/)*.
