=============
Configuration
=============

Overview
========

DISCON-Wrapper offers various configuration options across its components. This section details how to configure each part of the system for optimal performance and compatibility.

Configuration Components
=======================

.. toctree::
   :maxdepth: 2

   client_configuration
   server_configuration
   manager_configuration
   controller_database

Quick Reference
==============

Client Configuration (Environment Variables)
-------------------------------------------

.. list-table::
   :widths: 25 75
   :header-rows: 1

   * - Environment Variable
     - Description
   * - DISCON_SERVER_ADDR
     - Server address (e.g., ``localhost:8080`` or ``https://controller.domain.com``)
   * - DISCON_LIB_PATH
     - Path to controller library on server side
   * - DISCON_LIB_PROC
     - Function name to call in controller library
   * - DISCON_CLIENT_DEBUG
     - Debug level (0=disabled, 1=basic info, 2=verbose)
   * - DISCON_ADDITIONAL_FILES
     - Comma-separated list of additional files to transfer

Server Configuration (Command Line Arguments)
-------------------------------------------

.. list-table::
   :widths: 25 75
   :header-rows: 1

   * - Argument
     - Description
   * - --port
     - Port to listen on (default: 8080)
   * - --debug
     - Debug level (0=disabled, 1=basic info, 2=verbose)

Manager Configuration (YAML)
--------------------------

Key configuration sections:

1. **Server settings**: Port, host, debug level
2. **Docker settings**: Network, resource limits, cleanup timeout
3. **Auth settings**: Authentication options for admin interface
4. **Database settings**: Path to controller database

Controller Database (JSON)
-------------------------

The controller database defines:

1. Available controller versions
2. Docker images to use for each controller
3. Controller-specific settings (library path, procedure name)
4. Port mappings for container communication