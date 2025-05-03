==========
Components
==========

DISCON-Wrapper consists of three main components that work together to bridge 64-bit simulation software with 32-bit controller libraries:

1. discon-client
----------------

The discon-client is a 64-bit shared library (.dll on Windows, .so on Linux) that serves as the interface between OpenFAST and the DISCON-Wrapper system. It acts as a replacement for the original controller library.

Key features:
 
* Implements the standard DISCON interface expected by OpenFAST
* Forwards controller calls over WebSocket to discon-server
* Handles automatic file transfers for input files
* Provides configurable logging and debugging capabilities

For detailed information, see :doc:`discon-client`.

2. discon-server
----------------

The discon-server is a 32-bit binary that loads the actual controller library and executes controller function calls. It communicates with the client via WebSocket.

Key features:

* Loads 32-bit controller libraries
* Handles WebSocket communication with the client
* Creates temporary copies of controller libraries for isolation
* Processes file transfer requests from the client
* Provides configurable logging and debugging

For detailed information, see :doc:`discon-server`.

3. discon-manager
----------------

The discon-manager is an optional but recommended component that provides container orchestration for multiple controller versions. It dynamically creates Docker containers running discon-server instances based on client requests.

Key features:

* Dynamic container creation and management
* WebSocket proxying between clients and containers
* Controller version management through a simple database
* Admin interface for monitoring and management
* Resource limiting for containers (memory, CPU)
* Automatic cleanup of inactive containers

For detailed information, see :doc:`discon-manager`.

.. toctree::
   :maxdepth: 2

   discon-client
   discon-server
   discon-manager
   payload