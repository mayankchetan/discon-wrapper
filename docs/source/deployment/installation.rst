============
Installation
============

Overview
========

This guide covers how to install the DISCON-Wrapper components from binaries or build them from source. For Docker-based installation, refer to :doc:`docker_deployment`.

Prerequisites
============

Before installing DISCON-Wrapper, ensure you have:

- **For discon-client**: 64-bit operating system (Windows/Linux)
- **For discon-server**: 32-bit operating system (for 32-bit controllers)
- **For discon-manager**: Docker and docker-compose

Installing from Binaries
=======================

Pre-compiled binaries are available in the `Releases <https://github.com/deslaughter/discon-wrapper/releases>`_ section of the GitHub repository.

discon-client
------------

1. Download the appropriate binary for your system:

   - Windows 64-bit: ``discon-client_amd64.dll``
   - Linux 64-bit: ``libdiscon-client_amd64.so``

2. Place the binary in your OpenFAST simulation directory or another location accessible to OpenFAST.

3. Configure your environment variables as described in the :doc:`../configuration/client_configuration` section.

discon-server
------------

1. Download the appropriate binary for your system:

   - Windows 32-bit: ``discon-server_386.exe``
   - Linux 32-bit: ``discon-server_386``

2. Place the binary in the same directory as your controller libraries or another suitable location.

3. Make the binary executable (Linux only):

   .. code-block:: bash

       chmod +x discon-server_386

discon-manager
-------------

1. Download the appropriate binary for your system:

   - Windows 64-bit: ``discon-manager_amd64.exe``
   - Linux 64-bit: ``discon-manager_amd64``

2. Create the following directory structure:

   .. code-block:: text

       discon-manager/
       ├── config/
       │   └── config.yaml
       ├── db/
       │   └── controllers.json
       └── metrics/

3. Copy the sample configuration files from the repository.

4. Make the binary executable (Linux only):

   .. code-block:: bash

       chmod +x discon-manager_amd64

Building from Source
===================

Prerequisites for building:

- Go 1.24 or newer
- GCC compatible compiler (for CGO)
- Docker (for building containerized versions)

Building the discon-client
------------------------

.. code-block:: bash

    # Navigate to the project directory
    cd discon-wrapper

    # Build the client library
    go build -o build/discon-client.dll -buildmode=c-shared ./discon-client

Building the discon-server
------------------------

.. code-block:: bash

    # For 32-bit builds on a 64-bit system (Linux)
    GOARCH=386 go build -o build/discon-server_386 ./discon-server

    # For Windows 32-bit cross-compilation from Linux
    GOOS=windows GOARCH=386 go build -o build/discon-server_386.exe ./discon-server

Building the discon-manager
-------------------------

.. code-block:: bash

    # Standard build
    go build -o build/discon-manager ./discon-manager

Using the Build Script
--------------------

The repository includes a `build.sh` script that automates the build process:

.. code-block:: bash

    # Make the script executable
    chmod +x build.sh

    # Run the build script
    ./build.sh

This script will:

1. Build the discon-client library
2. Build the discon-server executable
3. Build the test-discon controller
4. Build the Docker images

Verifying the Installation
=========================

discon-client
------------

Verify the client is correctly built and exports the DISCON symbol:

.. code-block:: bash

    # On Linux
    nm -D build/libdiscon-client_amd64.so | grep DISCON

    # On Windows (using Dependency Walker or similar tool)
    # Check that DISCON is exported from discon-client_amd64.dll

discon-server
------------

Start the server and check that it runs correctly:

.. code-block:: bash

    # On Linux
    ./build/discon-server_386 --port=8080 --debug=1

    # On Windows
    build\discon-server_386.exe --port=8080 --debug=1

You should see output indicating the server is running and listening on port 8080.

discon-manager
------------

Start the manager and check that it runs correctly:

.. code-block:: bash

    # On Linux
    ./build/discon-manager --config=config/config.yaml

    # On Windows
    build\discon-manager.exe --config=config\config.yaml

You should see output indicating the manager is running and ready to accept connections.

Next Steps
=========

After installation:

1. Configure each component as described in the :doc:`../configuration/index` section
2. Follow the :doc:`../usage/quickstart` guide to get started quickly
3. For containerized deployment, refer to :doc:`docker_deployment`