=======
Payload
=======

Overview
========

The Payload component is the core data structure used for communication between the discon-client and discon-server components. It defines how controller function parameters and file transfers are serialized and transmitted over WebSocket connections.

Purpose and Design
=================

The Payload structure serves several key purposes:

1. Encapsulating all parameters needed for controller function calls
2. Enabling efficient binary serialization for network transmission
3. Supporting file transfers between client and server
4. Maintaining backward and forward compatibility

Implementation
=============

The Payload is defined in the root module of the DISCON-Wrapper project (in ``payload.go``) and is imported by both the client and server components. It is implemented as a Go struct with methods for binary serialization and deserialization.

Structure Definition
------------------

The core Payload structure includes:

.. code-block:: go

    type Payload struct {
        Swap          []float32 // Controller SWAP array
        Fail          int32     // Controller fail flag
        InFile        []byte    // Controller input file path
        OutName       []byte    // Controller output name
        Msg           []byte    // Controller message buffer
        FileContent   []byte    // For file transfers: content of file
        ServerFilePath []byte   // For file transfers: server-side path
    }

This structure maps directly to the parameters of the standard DISCON interface:

- ``avrSWAP`` → ``Swap``
- ``aviFAIL`` → ``Fail``
- ``accINFILE`` → ``InFile``
- ``avcOUTNAME`` → ``OutName``
- ``avcMSG`` → ``Msg``

Additionally, the ``FileContent`` and ``ServerFilePath`` fields enable file transfers between client and server.

Binary Serialization
------------------

The Payload implements binary serialization and deserialization methods:

- ``MarshalBinary()``: Converts the Payload to a binary representation for network transmission
- ``UnmarshalBinary()``: Recreates the Payload from its binary representation

This binary format is carefully designed to be:

1. Compact: Minimizing network bandwidth requirements
2. Efficient: Optimized for fast serialization and deserialization
3. Extensible: Able to accommodate future extensions

File Transfer Support
===================

A Payload can represent either a standard controller function call or a file transfer request:

1. **Controller function call**:
   - Contains controller parameters in the Swap, Fail, InFile, OutName, and Msg fields
   - FileContent is empty

2. **File transfer request**:
   - Contains the file content in the FileContent field
   - Contains the server-side path in the ServerFilePath field

Helper functions in the shared utilities package can detect whether a payload represents a file transfer:

.. code-block:: go

    func IsFileTransfer(payload *Payload) bool {
        return len(payload.FileContent) > 0 && len(payload.ServerFilePath) > 0
    }

Usage in Client-Server Communication
==================================

The communication flow using the Payload structure is:

1. **Client side**:
   - The client creates a Payload with controller parameters or file content
   - Serializes it using MarshalBinary()
   - Sends the binary data over WebSocket

2. **Server side**:
   - Receives binary data over WebSocket
   - Deserializes it using UnmarshalBinary()
   - Processes controller call or file transfer as appropriate
   - Creates a response Payload
   - Serializes and sends it back to the client

Version Compatibility
===================

The Payload structure maintains backward and forward compatibility through careful design:

- Fixed-size fields are placed at the beginning
- Variable-length fields include size prefixes
- New fields can be added to the end of the structure
- Older clients/servers can still work with newer versions by ignoring unknown fields