==============
discon-server
==============

Overview
========

The discon-server is a 32-bit executable application that loads and interfaces with actual controller libraries. It serves as the backend of the DISCON-Wrapper system, receiving WebSocket connections from discon-client instances and forwarding controller function calls to the appropriate shared libraries.

The server is designed to run on a 32-bit system or within a 32-bit container, allowing it to load and execute 32-bit controller libraries that cannot be directly used by 64-bit applications like OpenFAST.

Implementation
=============

The discon-server is implemented in Go with C bindings for loading and interfacing with shared libraries. The application provides a WebSocket server that listens for connections and uses CGO to interact with the controller libraries.

Key Components
-------------

- **WebSocket server**: Handles incoming connections from discon-client instances
- **Shared library loader**: Dynamically loads controller libraries and exposes their functions
- **Connection manager**: Tracks active connections and their associated controller libraries
- **File transfer handler**: Manages file transfers from clients
- **Logging system**: Provides configurable logging capabilities

Functionality
============

When the discon-server receives a WebSocket connection, the following sequence occurs:

1. The server parses the query parameters to determine which controller library to load
2. It creates a temporary copy of the controller library for isolation between clients
3. The library is loaded using the dynamic library loading functionality
4. A connection ID is assigned to the client
5. The server enters a loop waiting for messages from the client
6. For each message:
   - The payload is unpacked
   - File transfers are handled if present
   - The appropriate controller function is called with the provided parameters
   - Results are packaged and sent back to the client

When the connection closes, the server unloads the controller library and cleans up any temporary files.

Command Line Arguments
=====================

The discon-server accepts the following command line arguments:

.. list-table::
   :widths: 25 75
   :header-rows: 1

   * - Argument
     - Description
   * - --port
     - Port number to listen on (default: 8080)
   * - --debug
     - Debug level: 0=disabled, 1=basic info, 2=verbose with payloads (default: 0)

Loading Controller Libraries
===========================

The discon-server uses a C-based library loading mechanism to dynamically load controller DLLs or shared objects. This allows it to:

1. Load multiple different controllers simultaneously
2. Isolate controllers from each other through connection-specific IDs
3. Properly unload controllers when connections close
4. Handle different controller function names via the proc parameter

Temporary File Management
========================

For isolation between clients, the discon-server creates temporary copies of controller libraries and input files. These temporary files are:

1. Created when a client connects or transfers a file
2. Used for the duration of the client connection
3. Automatically deleted when the connection closes

This approach ensures that multiple clients can use different versions of controllers or input files without interference.

Error Handling
=============

The discon-server implements several error handling mechanisms:

1. **Library loading failures**: If a controller cannot be loaded, an error is returned to the client
2. **Function call errors**: Controller function errors are captured and returned to the client
3. **Connection timeouts**: Idle connections are automatically closed after a timeout period
4. **File transfer failures**: File transfer errors are reported back to the client

Containerization
===============

While the discon-server can be run as a standalone application, it's often deployed within Docker containers managed by the discon-manager. This allows:

1. Isolation between different controller versions
2. Independent scaling of controller instances
3. Resource limits per controller
4. Clean termination when no longer needed

The discon-server container images typically include the necessary libraries for controller execution, such as C, C++, and Fortran runtime libraries.