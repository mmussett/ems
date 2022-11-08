# TIBCO EMS Go Client

[TIBCO](https://tibco.com) is not providing a client package for use with [Go](https://go.dev/). But some enterprise use cases require to connect even lighwight systems as clients to a central message bus like [TIBCO EMS](https://www.tibco.com/de/products/tibco-enterprise-message-service).

Intention of this repository is to demonstrate how the client libraries written in C for Linux x86 based systems can be leveraged to create proper, reliable access for applications written on Go.

This repository contains the source code for the TIBCO EMS Go client library. For a commented introduction on how to setup the environment please consult the [Quickstart](Quickstart.md).


## Installation and Build

The client was designed to work with the EMS 8.4 client libraries as shipped with TIBCO EMS v8.4. It was originally created and tested on Mac OS for x86 CPU architecture.
The solution was successfully tested with the lates TIBCO EMS v10.2 release on Linux (Ubuntu 20.04, x86-64) as well.

You will need to modify the cgo CFLAGS and LDFLAGS directives to the correct location of your local EMS Client Libaries within the Go source file *client.go*.

For running the Go unit tests or any Go appliaction, one must ensure the operating system can find the shared libraries. Either the TIBCO EMS libraries are added to the system environment variable LD_LIBRARY_PATH or symbolic links to needed libraries are set:

On Linux using LD_LIBRARY_PATH (no root access required):
```
export LD_LIBRARY_PATH=/opt/tibco/ems/10.2/lib

echo $LD_LIBRARY_PATH
/opt/tibco/ems/10.2/lib
```

On Mac OS with links (requires root access):
```
ln /opt/tibco/ems/ems841/ems/8.4/lib/libtibems64.dylib /usr/local/lib/.
ln /opt/tibco/ems/ems841/ems/8.4/lib/64/libssl.1.0.0.dylib /usr/local/lib/.
ln /opt/tibco/ems/ems841/ems/8.4/lib/64/libcrypto.1.0.0.dylib /usr/local/lib/.
```

## Reporting bugs

Please report bugs by raising issues for this project in github https://github.com/mmussett/ems/issues


## Update History

07-Nov-2022 - Adoption of some unit tests to avoid a race condition

11-Nov-2019 - Breaking change to Send,SendReceive, and Receive functions

Changed Send, SendReceive, and Receive functions to include destinationType on signature.
destinationType can take 'queue' or 'topic' for the destination type now.
Tested Sending and Receiving message on both Queue and Topic types. 
