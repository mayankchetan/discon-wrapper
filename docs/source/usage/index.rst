=====
Usage
=====

Overview
========

This section provides practical guides for using the DISCON-Wrapper system with OpenFAST and other wind turbine simulation tools. These guides cover common use cases and workflows to help you get the most out of the system.

.. toctree::
   :maxdepth: 2

   quickstart
   openfast_integration
   file_transfers
   troubleshooting_guide
   advanced_usage

Workflow Overview
===============

The typical workflow for using DISCON-Wrapper involves:

1. **Setup**: Deploy and configure the server components (discon-server or discon-manager)
2. **Client configuration**: Configure the discon-client with environment variables
3. **OpenFAST configuration**: Modify ServoDyn input files to use discon-client
4. **Run simulation**: Execute OpenFAST with the configured environment
5. **Monitoring**: Optionally monitor the connections through the admin interface

Each guide in this section covers one or more aspects of this workflow in detail.