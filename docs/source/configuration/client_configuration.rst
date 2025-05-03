=====================
Client Configuration
=====================

Overview
========

The discon-client component is configured primarily through environment variables. These variables control how the client connects to the server, which controller library to use, and debugging options.

Environment Variables
====================

.. list-table::
   :widths: 25 75
   :header-rows: 1

   * - Environment Variable
     - Description
   * - DISCON_SERVER_ADDR
     - **Required**. The address of the discon-server or discon-manager. Can be specified in several formats:
       
       - ``hostname:port`` (e.g., ``localhost:8080``) - Standard WebSocket connection
       - ``domain.name`` (e.g., ``controller.company.com``) - For use with reverse proxies
       - ``http://domain.name`` - Explicit HTTP protocol, uses WebSocket (ws://)
       - ``https://domain.name`` - Secure HTTPS protocol, uses secure WebSocket (wss://)
   * - DISCON_LIB_PATH
     - **Required**. Path to the controller library on the server side. This should be the path relative to the discon-server executable or absolute path in the container.
   * - DISCON_LIB_PROC
     - **Required**. The procedure name to call in the controller library (e.g., ``DISCON`` or ``CONTROL``).
   * - DISCON_CLIENT_DEBUG
     - Optional. Controls debugging output. Can be:
       
       - A number (0=disabled, 1=basic info, 2=verbose with payloads)
       - A filename for CSV output of SWAP arrays
       
       Default: ``0`` (disabled)
   * - DISCON_ADDITIONAL_FILES
     - Optional. Comma-separated list of additional files to transfer to the server before simulation starts. Useful for supplementary input files required by the controller.

OpenFAST Configuration
=====================

To use the discon-client with OpenFAST, you need to modify the ServoDyn input file to point to the discon-client library instead of the actual controller:

Original:

.. code-block:: text

    "controller.dll"             DLL_FileName - Name/location of the dynamic library
    "DISCON.IN"                  DLL_InFile   - Name of input file sent to the DLL (-)
    "CONTROL"                    DLL_ProcName - Name of procedure in DLL to be called (-)

Modified:

.. code-block:: text

    "discon-client_amd64.dll"    DLL_FileName - Name/location of the dynamic library
    "DISCON.IN"                  DLL_InFile   - Name of input file sent to the DLL (-)
    "DISCON"                     DLL_ProcName - Name of procedure in DLL to be called (-)

The ``DLL_InFile`` doesn't need to be changed. Change ``DLL_ProcName`` to ``DISCON`` as that is the name of the procedure in the discon-client library.

Running OpenFAST with discon-client
==================================

Once you've configured the ServoDyn input file, you can run OpenFAST with the discon-client by setting the environment variables and then executing OpenFAST:

.. code-block:: bash

    # Set environment variables
    export DISCON_SERVER_ADDR=localhost:8080
    export DISCON_LIB_PATH=/controller/discon.dll
    export DISCON_LIB_PROC=DISCON
    export DISCON_CLIENT_DEBUG=1
    
    # Run OpenFAST
    openfast my_turbine.fst

Windows Command Prompt:

.. code-block:: batch

    set DISCON_SERVER_ADDR=localhost:8080
    set DISCON_LIB_PATH=controller.dll
    set DISCON_LIB_PROC=CONTROL
    set DISCON_CLIENT_DEBUG=1
    openfast.exe my_turbine.fst

File Transfer Configuration
=========================

The discon-client automatically transfers input files referenced by the controller to the server. This happens when:

1. The file path exists locally
2. The file hasn't already been transferred

For additional files that the controller might need but aren't directly referenced in the DISCON call, use the ``DISCON_ADDITIONAL_FILES`` environment variable:

.. code-block:: bash

    export DISCON_ADDITIONAL_FILES=controller_params.txt,aerodyn.dat,external_gains.csv

Windows Command Prompt:

.. code-block:: batch

    set DISCON_ADDITIONAL_FILES=controller_params.txt,aerodyn.dat,external_gains.csv

These files will be transferred to the server before the simulation starts.

Debugging and Logging
====================

When ``DISCON_CLIENT_DEBUG`` is set to a non-zero value, the discon-client will output debug information:

- **Level 1**: Basic information about connections, function calls, and file transfers
- **Level 2**: Verbose output including full payload contents

To save SWAP array values to CSV files for analysis, set ``DISCON_CLIENT_DEBUG`` to a filename:

.. code-block:: bash

    export DISCON_CLIENT_DEBUG=my_simulation

This will create two files:
- ``my_simulation_sent.csv``: Values sent to the server
- ``my_simulation_recv.csv``: Values received from the server

Connection Security
=================

For secure WebSocket connections (wss://), use an HTTPS URL:

.. code-block:: bash

    export DISCON_SERVER_ADDR=https://controller.example.com

The client will automatically use secure WebSocket (wss://) when an HTTPS URL is provided.