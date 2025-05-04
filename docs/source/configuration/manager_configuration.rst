======================
Manager Configuration
======================

Overview
========

The discon-manager component provides a container orchestration layer for the DISCON-Wrapper system. It has the most complex configuration of the three components, using a YAML configuration file for its settings.

Configuration File
=================

The discon-manager is configured through a YAML file, typically located at ``config/config.yaml``. The configuration file is divided into several sections:

Server Settings
--------------

The ``server`` section configures the HTTP/WebSocket server:

.. code-block:: yaml

    server:
      port: 8080          # Port to listen on
      host: "0.0.0.0"     # Interface to bind to (0.0.0.0 = all interfaces)
      debug_level: 1      # Debug level: 0=disabled, 1=basic info, 2=verbose

Docker Settings
-------------

The ``docker`` section configures how containers are managed:

.. code-block:: yaml

    docker:
      network_name: "discon-network"        # Docker network name
      container_prefix: "discon-controller-" # Prefix for container names
      memory_limit: "512m"                  # Memory limit per container
      cpu_limit: 1.0                        # CPU limit per container (cores)
      cleanup_timeout: 30                   # Cleanup after inactivity (seconds)

Authentication Settings
---------------------

The ``auth`` section configures authentication for the admin interface:

.. code-block:: yaml

    auth:
      enabled: true       # Enable/disable authentication
      username: "admin"   # Admin username
      password: "password" # Admin password (use a secure password in production)
      session_timeout: 3600 # Session timeout in seconds (1 hour)

Database Settings
---------------

The ``database`` section configures the controller database:

.. code-block:: yaml

    database:
      controllers_file: "db/controllers.json" # Path to controller database file

Metrics Settings
--------------

The ``metrics`` section configures metrics collection:

.. code-block:: yaml

    metrics:
      enabled: true       # Enable/disable metrics collection
      interval: 60        # Collection interval in seconds
      storage_path: "metrics" # Path to metrics storage directory

Example Configuration File
=========================

Here's a complete example configuration file:

.. code-block:: yaml

    # Server configuration
    server:
      port: 8080
      host: "0.0.0.0"
      debug_level: 1

    # Docker configuration
    docker:
      network_name: "discon-network"
      container_prefix: "discon-controller-"
      memory_limit: "512m"
      cpu_limit: 1.0
      cleanup_timeout: 30

    # Authentication configuration
    auth:
      enabled: true
      username: "admin"
      password: "securepassword123"
      session_timeout: 3600

    # Database configuration
    database:
      controllers_file: "db/controllers.json"

    # Metrics configuration
    metrics:
      enabled: true
      interval: 60
      storage_path: "metrics"

Command Line Arguments
=====================

The discon-manager accepts the following command line arguments:

.. list-table::
   :widths: 25 75
   :header-rows: 1

   * - Argument
     - Description
   * - --config
     - Path to configuration file. Default: ``config/config.yaml``

Docker Integration
=================

The discon-manager needs access to the Docker socket to create and manage containers. When running the manager itself in a container, you must:

1. Mount the Docker socket into the container
2. Ensure the container has permission to access the socket

Example docker-compose.yml:

.. code-block:: yaml

    version: '3'
    
    services:
      discon-manager:
        image: discon-manager:latest
        ports:
          - "8080:8080"
        volumes:
          - /var/run/docker.sock:/var/run/docker.sock
          - ./config:/app/config
          - ./db:/app/db
          - ./metrics:/app/metrics
        environment:
          - DOCKER_HOST=unix:///var/run/docker.sock
        restart: unless-stopped

Environment Variables
====================

The discon-manager also supports configuration through environment variables, which override settings in the configuration file:

.. list-table::
   :widths: 25 75
   :header-rows: 1

   * - Environment Variable
     - Description
   * - DISCON_MANAGER_PORT
     - HTTP/WebSocket server port
   * - DISCON_MANAGER_HOST
     - HTTP/WebSocket server host
   * - DISCON_MANAGER_DEBUG_LEVEL
     - Debug level (0-2)
   * - DISCON_MANAGER_AUTH_ENABLED
     - Enable/disable authentication (true/false)
   * - DISCON_MANAGER_AUTH_USERNAME
     - Admin username
   * - DISCON_MANAGER_AUTH_PASSWORD
     - Admin password
   * - DOCKER_HOST
     - Docker socket path (e.g., unix:///var/run/docker.sock)

Security Considerations
=====================

When deploying discon-manager, consider the following security aspects:

1. **Admin Authentication**: Always enable authentication and use a strong password in production.
2. **Docker Socket**: Access to the Docker socket provides significant privileges. Use appropriate user/group permissions.
3. **Network Security**: Consider running the manager behind a reverse proxy with TLS for secure WebSocket (wss://) connections.
4. **Container Isolation**: Configure appropriate resource limits and network isolation for spawned containers.