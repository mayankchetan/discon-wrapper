==============
discon-client
==============

Overview
========

The discon-client component is a 64-bit shared library that acts as a bridge between OpenFAST (or any other 64-bit simulation software using the DISCON interface) and the DISCON-Wrapper system. It implements the standard DISCON interface that OpenFAST expects, but instead of directly executing controller logic, it forwards the calls to a discon-server instance via WebSocket.

This approach solves the fundamental issue of running 32-bit controller libraries with 64-bit simulation software, as the actual controller execution happens on the server side.

Implementation
=============

The discon-client is written in Go and compiled to a shared library (.dll on Windows, .so on Linux) with C bindings. It exposes the DISCON function that OpenFAST expects to find in controller libraries.

Key Components
-------------

- **DISCON function**: The primary entry point that implements the Bladed API interface
- **WebSocket client**: Handles communication with the discon-server
- **File transfer system**: Automatically transfers input files to the server
- **Logging and debugging**: Configurable logging capabilities

Functionality
============

When OpenFAST loads the discon-client library and calls the DISCON function, the following sequence occurs:

1. The client initializes a WebSocket connection to the server (if not already connected)
2. Any referenced input files are automatically transferred to the server
3. The controller parameters (avrSWAP, aviFAIL, etc.) are packaged into a payload
4. The payload is sent to the server over WebSocket
5. The client waits for the server to respond with results
6. The results are unpacked and returned to OpenFAST

Environment Variables
====================

The discon-client uses the following environment variables for configuration:

.. list-table::
   :widths: 25 75
   :header-rows: 1

   * - Variable
     - Description
   * - DISCON_SERVER_ADDR
     - The address of the discon-server or discon-manager (e.g., ``localhost:8080`` or ``https://controller.domain.com``)
   * - DISCON_LIB_PATH
     - Path to the actual controller library on the server side
   * - DISCON_LIB_PROC
     - The procedure name to call in the controller library (e.g., ``DISCON`` or ``CONTROL``)
   * - DISCON_CLIENT_DEBUG
     - Debug level (0=disabled, 1=basic info, 2=verbose with payloads) or a filename for CSV output
   * - DISCON_ADDITIONAL_FILES
     - Comma-separated list of additional files to transfer to the server

File Transfer
============

One of the key features of discon-client is its ability to automatically transfer input files to the server. This happens when:

1. The client detects that OpenFAST is trying to access an input file (like DISCON.IN)
2. The client checks if the file exists locally
3. If found, it transfers the file to the server before the controller is called
4. The file path in the DISCON call is updated to point to the server-side location

This allows users to keep their input files on the client machine without manually copying them to the server.

Logging and Debugging
====================

The discon-client provides configurable logging through the DISCON_CLIENT_DEBUG environment variable:

- **Level 0**: No debugging (production mode)
- **Level 1**: Basic information like connections and function calls
- **Level 2**: Verbose output including full payload contents

Additionally, when debug mode is enabled, the client can create CSV files with the SWAP array values sent to and received from the server.

Thread Safety
============

The discon-client is designed to be thread-safe, which is important when OpenFAST runs multiple simulations in parallel threads. Each client connection maintains its own WebSocket connection to the server.