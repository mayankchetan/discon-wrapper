==========
Development
==========

Overview
========

This section is intended for developers who want to contribute to DISCON-Wrapper or understand how the code is structured. It covers the codebase organization, design patterns, testing approach, and guidelines for contributing.

.. toctree::
   :maxdepth: 2

   code_organization
   build_system
   contributing_guide
   testing

Development Environment Setup
===========================

To set up a development environment for DISCON-Wrapper:

1. **Prerequisites**:
   - Go 1.24 or newer
   - GCC compatible compiler for CGO
   - Docker and docker-compose for containerized development
   - A Git client

2. **Clone the repository**:

   .. code-block:: bash

       git clone https://github.com/deslaughter/discon-wrapper.git
       cd discon-wrapper

3. **Install Go dependencies**:

   .. code-block:: bash

       go mod download

4. **Build the components**:

   .. code-block:: bash

       ./build.sh

Development Workflow
==================

The recommended workflow for developing DISCON-Wrapper follows these steps:

1. **Create a feature branch**:

   .. code-block:: bash

       git checkout -b feature/my-new-feature

2. **Make changes to the code**

3. **Write tests for your changes**

4. **Build and test locally**:

   .. code-block:: bash

       go build ./...
       go test ./...

5. **Submit a pull request**

Design Philosophy
===============

DISCON-Wrapper follows several key design principles:

1. **Separation of concerns**: Each component has a well-defined responsibility:
   - discon-client handles interfacing with OpenFAST
   - discon-server handles loading and executing controllers
   - discon-manager handles container orchestration

2. **Minimalistic API**: The core API is kept simple and focused on essential functionality

3. **Isolation**: Each client connection is isolated from others, with its own copy of the controller library

4. **Compatibility**: Maintains compatibility with the standard DISCON interface without requiring changes to controllers or OpenFAST

5. **Extensibility**: Designed to allow extensions and customizations without changing the core functionality

Project Roadmap
=============

Future development plans for DISCON-Wrapper include:

1. **Enhanced security**: Adding more robust authentication and encryption options
2. **Performance optimizations**: Reducing communication overhead and latency
3. **Expanded controller support**: Better handling of complex controller requirements
4. **Improved monitoring**: More comprehensive metrics and visualization tools
5. **Extended documentation**: Detailed API documentation and more usage examples