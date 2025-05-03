==================
OpenFAST Integration
==================

Overview
========

This guide provides detailed information about integrating DISCON-Wrapper with OpenFAST, including advanced configuration options and best practices for optimal performance.

OpenFAST DISCON Interface
========================

OpenFAST interacts with wind turbine controllers through the DISCON interface, which follows the Bladed-style controller API. The DISCON interface consists of:

1. A function call with specific parameters:

   .. code-block:: c

       void DISCON(float *avrSWAP, int *aviFAIL, char *accINFILE, char *avcOUTNAME, char *avcMSG);

2. ServoDyn input parameters that specify:
   - The controller DLL path
   - The input file name
   - The procedure name to call

The discon-client library implements this interface and forwards the calls to the discon-server.

Modifying ServoDyn Input Files
============================

To use DISCON-Wrapper with OpenFAST, modify your ServoDyn input file as follows:

1. Change the ``DLL_FileName`` to point to the discon-client library:

   .. code-block:: text

       "discon-client_amd64.dll"    DLL_FileName - Name/location of the dynamic library

2. Keep the original ``DLL_InFile`` (this file will be automatically transferred to the server if found locally):

   .. code-block:: text

       "DISCON.IN"                  DLL_InFile   - Name of input file sent to the DLL (-)

3. Change the ``DLL_ProcName`` to "DISCON" (this is the fixed procedure name in discon-client):

   .. code-block:: text

       "DISCON"                     DLL_ProcName - Name of procedure in DLL to be called (-)

Environment Variables for OpenFAST
================================

Before running OpenFAST, set the following environment variables:

.. list-table::
   :widths: 25 75
   :header-rows: 1

   * - Environment Variable
     - Description
   * - DISCON_SERVER_ADDR
     - The server address, e.g., ``localhost:8080``
   * - DISCON_LIB_PATH
     - Path to the actual controller library on the server
   * - DISCON_LIB_PROC
     - Name of the procedure in the controller library (e.g., ``DISCON`` or ``CONTROL``)

Additional optional variables:

.. list-table::
   :widths: 25 75
   :header-rows: 1

   * - Environment Variable
     - Description
   * - DISCON_CLIENT_DEBUG
     - Debug level or CSV filename for SWAP arrays
   * - DISCON_ADDITIONAL_FILES
     - Comma-separated list of additional files to transfer

Running Multiple OpenFAST Instances
=================================

DISCON-Wrapper supports running multiple OpenFAST instances simultaneously. Each instance:

1. Creates its own WebSocket connection to the server
2. Gets a unique connection ID
3. Has its own isolated copy of the controller library on the server

This allows for parallel simulations without interference between controllers.

When using multiple instances:

- Each instance can use different environment variables if needed
- With discon-manager, each instance can use a different controller version
- File transfers are handled independently for each instance

Performance Considerations
========================

To achieve optimal performance with DISCON-Wrapper:

1. **Network latency**: Run the client and server on the same network for minimal latency
2. **File transfers**: Keep input files small or use pre-transferred files when possible
3. **Debug level**: Use debug level 0 in production to avoid logging overhead
4. **Connection setup**: Allow time for the initial connection before simulation starts

ROSCO Controller Integration
==========================

The ROSCO controller is commonly used with OpenFAST and works well with DISCON-Wrapper:

1. Build a ROSCO-specific Docker image (see example in ``docker/Dockerfile.rosco``)
2. Configure the controller in the database (for discon-manager)
3. Use the appropriate procedure name (typically ``DISCON``)

Example ServoDyn snippet for ROSCO with DISCON-Wrapper:

.. code-block:: text

    ---------------------- BLADED INTERFACE ---------------------- (Bladed Interface)
    "discon-client_amd64.dll"    DLL_FileName - Name/location of the dynamic library
    "DISCON.IN"                  DLL_InFile   - Name of input file sent to the DLL (-)
    "DISCON"                     DLL_ProcName - Name of procedure in DLL to be called (-)

Example environment variables:

.. code-block:: bash

    export DISCON_SERVER_ADDR=localhost:8080
    export DISCON_LIB_PATH=/app/build/libdiscon.so
    export DISCON_LIB_PROC=DISCON
    export DISCON_ADDITIONAL_FILES=DISCON.IN,DISCON.DBUG

Debugging OpenFAST Integration
============================

When troubleshooting OpenFAST integration:

1. Start with high debug levels on both client and server:

   .. code-block:: bash

       export DISCON_CLIENT_DEBUG=2
       # Start server with --debug=2

2. Check for file transfer issues:

   - Ensure input files exist in the correct locations
   - Check permissions on input files
   - Verify the paths in error messages

3. Examine the SWAP array:

   .. code-block:: bash

       export DISCON_CLIENT_DEBUG=my_simulation

   This creates CSV files with the SWAP array values that you can analyze.

4. Testing sequential calls:

   Run a simple test case first to verify the setup works before running complex simulations.