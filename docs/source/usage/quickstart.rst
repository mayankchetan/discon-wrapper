==========
Quickstart
==========

Overview
========

This guide provides a quick path to getting started with DISCON-Wrapper. It covers the minimal steps to set up the system and run your first OpenFAST simulation with a remote controller.

Prerequisites
============

Before starting, ensure you have:

- OpenFAST installed on your system
- A controller DLL/SO file that you want to use
- The DISCON-Wrapper components (see :doc:`../deployment/installation`)

Step 1: Set Up the Server
========================

The simplest approach is to run the standalone discon-server:

.. code-block:: bash

    # Start the server (Linux)
    ./discon-server_386 --port=8080 --debug=1
    
    # Start the server (Windows)
    discon-server_386.exe --port=8080 --debug=1

You should see output indicating the server is running and listening on port 8080.

Step 2: Configure the Client
==========================

Place the discon-client library in your OpenFAST working directory:

- Windows: ``discon-client_amd64.dll``
- Linux: ``libdiscon-client_amd64.so``

Set the required environment variables:

Linux:

.. code-block:: bash

    export DISCON_SERVER_ADDR=localhost:8080
    export DISCON_LIB_PATH=/path/to/your/controller.dll
    export DISCON_LIB_PROC=DISCON
    export DISCON_CLIENT_DEBUG=1

Windows:

.. code-block:: batch

    set DISCON_SERVER_ADDR=localhost:8080
    set DISCON_LIB_PATH=controller.dll
    set DISCON_LIB_PROC=DISCON
    set DISCON_CLIENT_DEBUG=1

Step 3: Modify OpenFAST Input Files
=================================

Edit your ServoDyn input file to use the discon-client instead of your original controller:

Original:

.. code-block:: text

    "controller.dll"             DLL_FileName - Name/location of the dynamic library
    "DISCON.IN"                  DLL_InFile   - Name of input file sent to the DLL (-)
    "DISCON"                     DLL_ProcName - Name of procedure in DLL to be called (-)

Modified:

.. code-block:: text

    "discon-client_amd64.dll"    DLL_FileName - Name/location of the dynamic library
    "DISCON.IN"                  DLL_InFile   - Name of input file sent to the DLL (-)
    "DISCON"                     DLL_ProcName - Name of procedure in DLL to be called (-)

Step 4: Run OpenFAST
==================

Run your OpenFAST simulation as usual:

.. code-block:: bash

    # Linux
    openfast your_model.fst
    
    # Windows
    openfast.exe your_model.fst

What Happens Behind the Scenes
============================

When OpenFAST runs:

1. It loads the discon-client library
2. discon-client connects to discon-server via WebSocket
3. discon-client transfers any input files to the server
4. discon-server loads the actual controller library
5. When OpenFAST calls the DISCON function:
   a. discon-client forwards the call to discon-server
   b. discon-server executes the controller function
   c. Results are returned to discon-client and then to OpenFAST

Troubleshooting Common Issues
===========================

Connection Refused
-----------------

If you see "connection refused" errors:

- Ensure the server is running
- Check that the port numbers match
- Verify there's no firewall blocking the connection

Controller Library Not Found
--------------------------

If the server reports it can't find the controller library:

- Verify the path in DISCON_LIB_PATH
- Ensure the controller library exists on the server machine
- Check file permissions

Environment Variables Not Set
---------------------------

If the client reports missing environment variables:

- Double-check that all required variables are set
- Ensure they're set in the same terminal/environment where OpenFAST is run

Using Docker Deployment
=====================

For a more robust setup using Docker:

1. Follow the :doc:`../deployment/docker_deployment` guide
2. Configure your client to connect to the discon-manager
3. Run OpenFAST as described above

Next Steps
=========

Once you have the basic setup working:

- Explore the :doc:`../configuration/index` section for advanced configuration options
- Try running with different controllers using the :doc:`../configuration/controller_database`
- Learn about automatic :doc:`file_transfers` for complex controller setups