======================
Server Configuration
======================

Overview
========

The discon-server component provides a WebSocket server that loads and interfaces with actual controller libraries. It has a simpler configuration model than the client, primarily using command line arguments.

Command Line Arguments
=====================

.. list-table::
   :widths: 25 75
   :header-rows: 1

   * - Argument
     - Description
   * - --port
     - Port number to listen on. Default: ``8080``
   * - --debug
     - Debug level: 0=disabled, 1=basic info, 2=verbose with payloads. Default: ``0``

Example Usage
============

Basic usage:

.. code-block:: bash

    # Start with default settings (port 8080, no debug)
    discon-server

With port and debugging options:

.. code-block:: bash

    # Start on port 8181 with basic debug info
    discon-server --port=8181 --debug=1

    # Start with verbose debugging
    discon-server --port=8080 --debug=2

When running inside a Docker container, the port is often published to the host:

.. code-block:: bash

    docker run -p 8080:8080 discon-server:latest --debug=1

Controller Libraries
===================

The discon-server doesn't require pre-configuration for controller libraries. Instead, it dynamically loads libraries based on client connection parameters:

1. The client specifies the controller library path and function name in its connection request
2. The server creates a temporary copy of the library for isolation
3. The library is loaded when the client connects, and unloaded when the connection closes

This dynamic loading approach allows:

- Multiple different controllers to be used simultaneously
- Multiple clients to use the same controller without interference
- Clean unloading of controllers when no longer needed

File Handling
============

The discon-server creates temporary working directories and files for each client connection. These temporary files include:

1. A copy of the controller library (renamed with a connection-specific suffix)
2. Any input files transferred from the client

These temporary files are automatically cleaned up when the connection closes, ensuring no leftover files accumulate over time.

Security Considerations
=====================

When deploying discon-server, consider the following security aspects:

1. **WebSocket Security**: By default, WebSocket connections are not encrypted. For production use, consider running behind a reverse proxy with TLS.

2. **Library Validation**: The server will attempt to load any library specified by the client. In production environments, consider restricting which libraries can be loaded.

3. **Resource Limits**: Consider running the server with resource limits to prevent a single controller from consuming excessive resources.

4. **File System Access**: The server requires access to the controller libraries and temporary directories. Restrict file system access to only what's necessary.

Containerization
===============

The discon-server is designed to be run in a container, particularly when used with discon-manager. The container should:

1. Include all necessary runtime libraries (C, C++, Fortran runtimes)
2. Mount a volume containing controller libraries or include them in the image
3. Expose the configured WebSocket port

A typical Dockerfile might look like:

.. code-block:: docker

    FROM ubuntu:24.04

    # Install runtime dependencies
    RUN apt-get update && apt-get install -y \
        libc6 \
        libstdc++6 \
        libgcc-s1 \
        libgfortran5 \
        liblapack3 \
        libblas3 \
        && rm -rf /var/lib/apt/lists/*

    # Copy the server binary
    COPY discon-server /usr/local/bin/discon-server

    # Create a directory for controller libraries
    RUN mkdir -p /controller

    # Expose the WebSocket port
    EXPOSE 8080

    # Start the server
    ENTRYPOINT ["/usr/local/bin/discon-server"]
    CMD ["--port=8080"]

Logging
=======

The server outputs logs to stdout/stderr, making it compatible with container logging systems. The verbosity depends on the debug level:

- **Level 0**: Minimal logging (errors only)
- **Level 1**: Basic operational logging (connections, library loading, etc.)
- **Level 2**: Verbose logging including full payload contents

For containerized deployments, these logs are typically captured by the container runtime and can be viewed with commands like ``docker logs``.