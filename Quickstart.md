# How to interact with a TIBCO EMS message broker from GO?

The initial request to use Go programs as proper TIBCO EMS clients to send or receive messages from an enterprise message bus came from a TIBCO enterprise customer. TIBCO as the vendor of EMS does not provide a Go langugae package that could be leveraged.
Google has developed Go to be the "better C++". Hence, it has a close relation to C and offers a relatively simple way to call C/C++ code from Go. Therefore, the idea was to leverage this Go feature to integrate the exsiting C libraries of EMS.

## First steps with GO

To better understand the integration of C-code with Go appliactions its worth to have a look at a more simple example first.

### Installation on Linux

A nice quickstsart for anybody on Linux, more precice Ubuntu 20.04., is the guide [How To Install Go on Ubuntu 20.04 step by step instructions](https://linuxconfig.org/how-to-install-go-on-ubuntu-20-04-focal-fossa-linux). It explaines all steps. On my system `go version go1.13.8 linux/amd64` is available now.

In addition I was using [Visual Studio Code](https://code.visualstudio.com/). After cloning the repository VS-Code was trying to install some GO support. It failed and asked to install some additions 'go get -v golang.org/x/tools/gopls'. Now that the dev environment was prepared I was ready to "Go". :-)

### Calling C Code from GO

Now, the tutorial [Calling C code from go](https://karthikkaranth.me/blog/calling-c-code-from-go/) was what I needed to understand.

In case you place the Go source not at the default location - as I did - you need to compile with different commands:
```
cd ~/Documents/samples/go/src/go2c
gcc -c greeter.c
go build
./go2c

Greetings, ReKie from 1999! We come in peace :)
```

More details on cgo can be found on the [cgo package docs page](https://pkg.go.dev/cmd/cgo).


### Understanding the Result

Building a running sample is not enough as we also need to package and ship an application afterwards. As I had no deep Go skills, I wanted to understand the resulting exeutable and what other resources neeed to be packaged for delivery, e.g. as Docker container (Dockerfile).

To understand the executable dependencies I used some of the options explained within the article [10 ways to analyze binary files on Linux](https://opensource.com/article/20/4/linux-binary-analysis). Of course, that's something should be done for each new application.

```
file go2c 
go2
c: ELF 64-bit LSB executable, x86-64, version 1 (SYSV), dynamically linked, interpreter /lib64/ld-linux-x86-64.so.2, BuildID[sha1]=6b08b5f608c08dc7c6f921734659e2882418a697, for GNU/Linux 3.2.0, not stripped

ldd go2c 
	linux-vdso.so.1 (0x00007fff64c65000)
	libpthread.so.0 => /lib/x86_64-linux-gnu/libpthread.so.0 (0x00007f0cccc8f000)
	libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6 (0x00007f0ccca9d000)
	/lib64/ld-linux-x86-64.so.2 (0x00007f0ccccc6000)
```


## Client Access on EMS Server from a Go Application

Now lets move on to EMS and recreate a sample that makes use of the TIBCO EMS C-client libraries in context of a GO application.
My starting point was the example on Github [TIBCO EMS Go client](https://github.com/mmussett/ems). Although the repo was a bit short on documentation, by following the article above the approach and solutions become quite understandable.

For testing the solution at least an TIBCO EMS server is needed as well as the appropriate C client librarries for EMS.
At the time weritinf the newest EMS version is [TIBCO EMS v10.2](https://edelivery.tibco.com/storefront/eval/tibco-enterprise-message-service-server/prod10929.html).

The EMS wire protocol has been quite stable. Therefore, EMS clients and EMS server version can vary. It should not have an impact. But as always, one should give it a try before use.


### Download and Install TIBCO EMS Server

TIBCO software products are available from [TIBCO eDelivery](https://edelivery.tibco.com/). An active TIBCO account is required to access the software distribution portal!
* TIBCO Enterprise Message Service Client 10.2.0 for Linux (x86-64)
* TIBCO Enterprise Message Service Server 19.2.0 (x86-64)

Unpacking the EMS server (without TIBCO installer):
```
cd /opt/tibco/ems/installer
tar xvzf TIB_ems_10.2.0_linux_x86_64-server.tar.gz
mv ./opt/tibco/ems/installer/opt/tibo/ems/* /opt/tibco/ems
```

For the Go integration we need the EMS C header files as well as the shared libraries. Boths are available already after unpacking the TIBCO EMS server on my test system.

*Attention:* References to "dylib" files are for Mac OS. For Linux on x86-64 we need shared libraries as "lib*.so" files!
```
find /opt/tibco/ems/10* -name libtib*.so
find /opt/tibco/ems/10* -name *.h
```

### Starting and Configuring a local TIBCO EMS Server

Starting the EMS server with its basic configuration (no security enabled; anythig default):
```
cd /opt/tibco/ems/10.2
export EMS_HOME=${PWD}
cp ./samples/config/tibemsd.conf timemsd.conf
mkdir datastore
./bin/tibemsd -config timemsd.conf
```

Setup EMS server to allow dynamic queues and topics created by clients on demand instead having them defined by an EMS administrartor upfront:
```
cd /opt/tibco/ems/10.2/bin
./tibemsadmin
TIBCO Enterprise Message Service Administration Tool.
Copyright 1997-2022 by TIBCO Software Inc.
All rights reserved.

Version 10.2.0 V5 2022-09-30

Type 'help' for commands help, 'exit' to exit:
> connect
Login name (admin): 
Password: 
Connected to: tcp://localhost:7222

tcp://localhost:7222> show server status
 Server:                   EMS-SERVER (version: 10.2.0 V5)
 Hostname:                 tibco-test-vm
 Process Id:               18267
 State:                    active
 Runtime Module Path:      /opt/tibco/ems/10.2/bin/lib:/opt/tibco/ems/10.2/lib:/opt/tibco/ems/10.2/ftl/lib
 Topics:                   1 (0 dynamic, 0 temporary)
 Queues:                   6 (0 dynamic, 1 temporary)
 Client Connections:       0
 Admin Connections:        1
 Sessions:                 1
 Producers:                1
 Consumers:                1
 Durables:                 0
 Pending Messages:         0
 Pending Message Size:     0.0 Kb
 Inbound Messages:         2
 Inbound Message Size:     0.3 Kb
 Outbound Messages:        1
 Outbound Message Size:    2.5 Kb
 Message Memory Usage:     13.2 Kb out of 512MB
 Message Memory Pooled:    53.0 Kb
 Synchronous Storage:      2.0 Kb
 Asynchronous Storage:     3.0 Kb
 Fsync for Sync Storage:   disabled
 Inbound Message Rate:     0 msgs/sec,  0.0 Kb per second
 Outbound Message Rate:    0 msgs/sec,  0.0 Kb per second
 Storage Read Rate:        0 reads/sec,  0.0 Kb per second
 Storage Write Rate:       0 writes/sec, 0.0 Kb per second
 Uptime:                   2 minutes

tcp://localhost:7222> create queue >
tcp://localhost:7222> create topic >

tcp://localhost:7222> show queues
                                                              All Msgs            Persistent Msgs  
  Queue Name                        SNFGXIBCT  Pre  Rcvrs     Msgs    Size        Msgs    Size   
  >                                 ---------    5*     0        0     0.0 Kb        0     0.0 Kb
  $sys.admin                        +--------    5*     0        0     0.0 Kb        0     0.0 Kb
  $sys.lookup                       ---------    5*     0        0     0.0 Kb        0     0.0 Kb
  $sys.redelivery.delay             +--------    5*     0        0     0.0 Kb        0     0.0 Kb
  $sys.undelivered                  +--------    5*     0        0     0.0 Kb        0     0.0 Kb
* $TMP$.EMS-SERVER.475B6368DB7C3.1  ---------    5      1        0     0.0 Kb        0     0.0 Kb
tcp://localhost:7222> show topics
                                                               All Msgs            Persistent Msgs 
  Topic Name                        SNFGEIBCTM  Subs  Durs     Msgs    Size        Msgs    Size   
  >                                 ----------     0     0        0     0.0 Kb        0     0.0 Kb
```

### Configuring EMS client for use with GO

The EMS server came with all files needed for a C based client without the need to install any extra packagaes. The installation is sufficient for local testing. If an executable needs to be compiled and linked on a system without the EMS server installation, one must download *TIB_ems_10.2.0_linux_x86_64.zip* and distribute *TIB_ems_10.2.0_linux_x86_64-c_dev_kit.tar.gz* and *TIB_ems_10.2.0_linux_x86_64-server.tar.gz* from the archive.

As the EMS server was not prepared by an installer (just unpacked from tar) we need to configure the library path to allow the OS to locate and use the needed shared libraries. A good explanation what is needed is explained in the article [Understanding Shared Libraries in Linux](https://www.tecmint.com/understanding-shared-libraries-in-linux), chapter *Locating Shared Libraries in Linux*.

```
cat /etc/ld.so.conf
ls -al /etc/ld.so.conf.d/*.conf

cat /etc/ld.so.conf.d/tibco-ems.conf
# TIBCO EMS libraries
/opt/tibco/ems/10.2/lib
```

*Attention:* The above method of pointing to the needed shared libraries did not work (needs a reboot?). So I followed the method of providing the library path as described in article [How to Check if a Shared Library - How Shared Libraries Are Located](https://www.baeldung.com/linux/check-shared-library-installed#how-shared-libraries-are-located).
```
export LD_LIBRARY_PATH=/opt/tibco/ems/10.2/lib
echo $LD_LIBRARY_PATH
/opt/tibco/ems/10.2/lib
```

### GO Sample Code

Now that we have a local EMS server running, we can start developing our Go application. I have started with the sample repo: `git clone https://github.com/mmussett/ems.git`.

As stated in the README.md some changes are required to meet a local test environmemt. The changes are related to zthe location of the EMS C header files and the location of the shared libraries. Mind the different extension on Linux and Mac OS!

File *client.go* needs to be adopted:
```
...
#cgo CFLAGS: -g -Wall -I/opt/tibco/ems/10.2/include/tibems
#cgo LDFLAGS: -L/opt/tibco/ems/10.2/lib -ltibems
...
```

Now the GO code can be compiled and built.
`go build`

### Test the EMS C-Client called from GO

The Go sample code comes with some unit tests (client_test.go). Those tests are checks for the main EMS client features. If those are successful we have proven that we can interact with our local EMS server from a Go program.

Set the location of the local EMS shared libraries:
```
export LD_LIBRARY_PATH=/opt/tibco/ems/10.2/lib
```

Run individual tests from the set of unit tests *client_test.go*.
```
go test -run 'TestClient_Send'
go test -run 'TestClient_Receive'
```

Run the full test set *client_test.go*.
```
go test -run ''
function: TestNewClient
function: TestClient_Send
Message with text 'hello, world' sent.
function: TestClient_Receive
Received JMS Text Message: hello, world
hello, world
PASS
ok  	_/home/tibco/Documents/go-test/go/ems	0.026s
```

*Hint:* For a quick test it is necessary to skip the test *TestClient_SendReceive* (acknowledgement expected).


## Conclusion

Although TIBCO is not offering an official client package for **GO**lang to interact with **TIBO EMS** the function can be added by leveraging the provided *TIBCO EMS C libraries* and integrate them into a GO program. Sample integration code is provided as Github repository. The GO integration was originally created and tested for *TIBCO EMS v8.4*. Nevertheless, we have sucessfully tested with the latest release *TIBCO EMS v10.2* as well.

## Known Limitations

The example code is illustrating how to connect to an EMS server. In an enterprise context TLS might be needed to ensure a secure communication between a remote client an a central EMS server. That extra bit was not part of the demonstration provided here. The EMS client is able to impelement it. The Go sample code might require extra handling for TLS as well.

The current unit test code is expecting a local EMS server with no security enabled. Those unit tests should be hanced to check for proper client authentication as well.
