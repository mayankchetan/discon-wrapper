===========
Build System
===========

Overview
========

DISCON-Wrapper uses a multi-architecture build system to create binaries for different platforms (32-bit and 64-bit) and different operating systems (Windows and Linux). This guide explains how the build system works and how to use or extend it.

Build Script
===========

The primary build mechanism is the ``build.sh`` script in the repository root. This script automates the building of all components and Docker images.

Script Components
---------------

The build script handles several key tasks:

1. **Client library building**: Compiles the 64-bit client shared library
2. **Server binary building**: Compiles the 32-bit server executable
3. **Test controller building**: Compiles the test controller
4. **Docker image building**: Creates Docker images for server and manager
5. **Output organization**: Places build artifacts in the build directory

Using the Build Script
--------------------

Basic usage:

.. code-block:: bash

    # Make it executable
    chmod +x build.sh
    
    # Run the build
    ./build.sh

You can also run specific parts of the build:

.. code-block:: bash

    # Build only client
    ./build.sh client
    
    # Build only server
    ./build.sh server
    
    # Build only Docker images
    ./build.sh docker

Environment Variables
------------------

The build script respects several environment variables:

.. list-table::
   :widths: 25 75
   :header-rows: 1

   * - Environment Variable
     - Description
   * - GOOS
     - Target operating system (e.g., windows, linux)
   * - GOARCH
     - Target architecture (e.g., amd64, 386)
   * - CGO_ENABLED
     - Enable CGO (required for shared libraries), default: 1
   * - CC
     - C compiler to use (especially important for cross-compilation)

Manual Building
=============

You can also build components manually using Go commands:

Building Client Library
---------------------

.. code-block:: bash

    # For Windows 64-bit
    GOOS=windows GOARCH=amd64 go build -o build/discon-client_amd64.dll -buildmode=c-shared ./discon-client
    
    # For Linux 64-bit
    GOOS=linux GOARCH=amd64 go build -o build/libdiscon-client_amd64.so -buildmode=c-shared ./discon-client

Building Server Binary
--------------------

.. code-block:: bash

    # For Windows 32-bit
    GOOS=windows GOARCH=386 go build -o build/discon-server_386.exe ./discon-server
    
    # For Linux 32-bit
    GOOS=linux GOARCH=386 go build -o build/discon-server_386 ./discon-server

Building Manager Binary
---------------------

.. code-block:: bash

    # For Windows 64-bit
    GOOS=windows GOARCH=amd64 go build -o build/discon-manager_amd64.exe ./discon-manager
    
    # For Linux 64-bit
    GOOS=linux GOARCH=amd64 go build -o build/discon-manager_amd64 ./discon-manager

Docker Builds
===========

The repository includes Dockerfiles for building containerized versions of the components:

Docker Images
-----------

.. list-table::
   :widths: 30 70
   :header-rows: 1

   * - Dockerfile
     - Purpose
   * - docker/Dockerfile.server
     - Base server image with minimal dependencies
   * - docker/Dockerfile.rosco
     - Server image with ROSCO controller and dependencies
   * - docker/Dockerfile.manager
     - Manager image

Building Docker Images
-------------------

To build Docker images manually:

.. code-block:: bash

    # Build server image
    docker build -t discon-server:latest -f docker/Dockerfile.server .
    
    # Build ROSCO image
    docker build -t discon-server-rosco:latest -f docker/Dockerfile.rosco .
    
    # Build manager image
    docker build -t discon-wrapper/discon-manager:latest -f docker/Dockerfile.manager .

Multi-Stage Builds
----------------

The Docker builds use multi-stage builds to minimize image size:

1. **Builder stage**: Compiles the Go binaries
2. **Runtime stage**: Contains only the necessary runtime components

Cross-Compilation Setup
=====================

For cross-compiling between different architectures and operating systems:

Windows to Linux
--------------

Use MinGW for cross-compilation:

.. code-block:: bash

    # Install MinGW
    apt-get install gcc-mingw-w64
    
    # Cross-compile for Windows
    CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build ...

Linux to Windows
--------------

Use MinGW in WSL or native Linux:

.. code-block:: bash

    # Install MinGW
    apt-get install gcc-mingw-w64
    
    # Cross-compile for Windows
    CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build ...

32-bit on 64-bit Systems
----------------------

To build 32-bit binaries on a 64-bit Linux system:

.. code-block:: bash

    # Install 32-bit development libraries
    apt-get install gcc-multilib libc6-dev-i386
    
    # Build for 32-bit
    GOARCH=386 go build ...

Build Artifacts
=============

The build process produces the following key artifacts:

.. list-table::
   :widths: 30 70
   :header-rows: 1

   * - Artifact
     - Description
   * - build/discon-client_amd64.dll
     - Windows 64-bit client library
   * - build/libdiscon-client_amd64.so
     - Linux 64-bit client library
   * - build/discon-server_386.exe
     - Windows 32-bit server binary
   * - build/discon-server_386
     - Linux 32-bit server binary
   * - build/discon-manager_amd64.exe
     - Windows 64-bit manager binary
   * - build/discon-manager_amd64
     - Linux 64-bit manager binary

CI/CD Integration
===============

For continuous integration, the build script can be used in CI pipelines:

GitHub Actions Example
-------------------

.. code-block:: yaml

    name: Build DISCON-Wrapper
    
    on:
      push:
        branches: [ main ]
      pull_request:
        branches: [ main ]
    
    jobs:
      build:
        runs-on: ubuntu-latest
        
        steps:
        - uses: actions/checkout@v2
        
        - name: Set up Go
          uses: actions/setup-go@v2
          with:
            go-version: 1.24
            
        - name: Install dependencies
          run: |
            sudo apt-get update
            sudo apt-get install -y gcc-multilib libc6-dev-i386
            
        - name: Build
          run: ./build.sh
            
        - name: Test
          run: go test ./...
            
        - name: Upload artifacts
          uses: actions/upload-artifact@v2
          with:
            name: discon-wrapper-binaries
            path: build/

Troubleshooting
=============

Common build issues and their solutions:

1. **CGO_ENABLED required**:
   
   If you get errors about CGO, ensure CGO_ENABLED=1:
   
   .. code-block:: bash
   
       CGO_ENABLED=1 go build ...

2. **Missing 32-bit libraries**:
   
   On 64-bit Linux systems, install multilib support:
   
   .. code-block:: bash
   
       sudo apt-get install gcc-multilib libc6-dev-i386

3. **Windows DLL export issues**:
   
   Ensure the DISCON function is properly exported:
   
   .. code-block:: bash
   
       nm -D build/libdiscon-client_amd64.so | grep DISCON