===============================
DISCON-Wrapper Documentation
===============================

.. image:: _static/discon_logo.png
   :alt: DISCON-Wrapper Logo
   :align: center
   :width: 400px

Overview
========

**DISCON-Wrapper** is a system that enables 64-bit wind turbine simulation software like OpenFAST to use 32-bit controller libraries. It solves the binary compatibility problem through a client-server architecture that bridges between different architecture binaries.

Key features:

* Connect 64-bit OpenFAST to 32-bit controller libraries
* Containerized controller execution for isolation and security
* Dynamic container management for multiple controller versions
* Automatic file transfer between client and server
* Low-latency WebSocket communication
* Seamless integration with standard DISCON interface

Documentation Sections
=====================

.. toctree::
   :maxdepth: 2
   :caption: Architecture
   
   architecture/index

.. toctree::
   :maxdepth: 2
   :caption: Components
   
   components/index
   
.. toctree::
   :maxdepth: 2
   :caption: Configuration
   
   configuration/index
   
.. toctree::
   :maxdepth: 2
   :caption: Deployment
   
   deployment/index

.. toctree::
   :maxdepth: 2
   :caption: Usage
   
   usage/index
   
.. toctree::
   :maxdepth: 2
   :caption: Development
   
   development/index

Quick Start
==========

To get started with DISCON-Wrapper quickly:

1. **Install the components**:
   
   Follow the :doc:`deployment/installation` guide to download or build the components.

2. **Start the server**:

   .. code-block:: bash

       ./discon-server_386 --port=8080

3. **Configure the client**:

   .. code-block:: bash

       export DISCON_SERVER_ADDR=localhost:8080
       export DISCON_LIB_PATH=/path/to/controller.dll
       export DISCON_LIB_PROC=DISCON

4. **Run OpenFAST**:

   Configure OpenFAST's ServoDyn to use the discon-client library and run your simulation.

For more detailed instructions, see the :doc:`usage/quickstart` guide.

About DISCON-Wrapper
===================

DISCON-Wrapper was developed to solve the common challenge of using 32-bit controller libraries with modern 64-bit simulation software in wind energy research and industry applications.

License
=======

DISCON-Wrapper is licensed under the Apache License 2.0.

Indices and Tables
================

* :ref:`genindex`
* :ref:`search`