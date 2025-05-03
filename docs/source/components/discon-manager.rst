=================
discon-manager
=================

Overview
========

The discon-manager is a container orchestration layer for the DISCON-Wrapper system. It dynamically creates and manages Docker containers running discon-server instances, allowing for flexible deployment of multiple controller versions.

While the discon-client and discon-server can operate in a direct client-server model, the discon-manager adds significant capabilities for production environments:

- Management of multiple controller versions
- Container-based isolation for enhanced security and stability
- Dynamic scaling based on client demand
- Centralized administration and monitoring
- Resource management and constraints

Implementation
=============

The discon-manager is implemented in Go and is designed to run as a Docker container with access to the Docker socket. It uses the Docker API to create and manage containers and provides a WebSocket proxy to route client connections to the appropriate container.

Key Components
-------------

- **WebSocket proxy**: Routes client connections to the appropriate container
- **Docker controller**: Manages container lifecycle (creation, monitoring, cleanup)
- **Controller database**: Tracks available controller versions and their configurations
- **Connection manager**: Maintains client connections and their associated containers
- **Admin interface**: Web-based UI for administration and monitoring
- **Metrics collector**: Gathers performance metrics and health data

Architecture
===========

The discon-manager acts as an intermediary between OpenFAST clients (running discon-client) and discon-server containers:

::

    ┌─────────┐                  ┌────────────────┐                  ┌───────────────────┐
    │         │                  │                │                  │ Container 1       │
    │ OpenFAST│  WebSocket       │  DisconManager │   WebSocket      │ ┌───────────────┐ │
    │ Client  │◄──────────────► │                │◄─────────────────┤►│ DisconServer   │ │
    │         │                  │   (Proxy)      │                  │ │ + Controller 1 │ │
    └─────────┘                  └────────────────┘                  │ └───────────────┘ │
                                        │                           └───────────────────┘
                                        │                                    
                                        │                           ┌───────────────────┐
                                        │                           │ Container 2       │
                                        │                           │ ┌───────────────┐ │
                                        └───────────────────────────┤►│ DisconServer   │ │
                                                                    │ │ + Controller 2 │ │
                                                                    │ └───────────────┘ │
                                                                    └───────────────────┘

When a client connects, the manager:

1. Parses the WebSocket query parameters to identify the requested controller version
2. Checks if a suitable container is already running
3. If not, creates a new container with the specified controller image
4. Establishes a WebSocket connection to the container
5. Proxies communication between the client and container

Container Management
===================

The discon-manager handles the complete lifecycle of controller containers:

- **Creation**: Containers are created on-demand when clients request a particular controller version
- **Monitoring**: Active containers are monitored for health and resource usage
- **Proxy**: WebSocket traffic is proxied between clients and containers
- **Cleanup**: Containers are automatically stopped and removed after a period of inactivity

Configuration
===========

The discon-manager is configured through a YAML file (typically at `config/config.yaml`). Key configuration areas include:

Server Settings
--------------

.. code-block:: yaml

    server:
      port: 8080          # Port to listen on
      host: "0.0.0.0"     # Interface to bind to
      debug_level: 1      # 0=disabled, 1=basic info, 2=verbose

Docker Settings
--------------

.. code-block:: yaml

    docker:
      network_name: "discon-network"      # Docker network name
      container_prefix: "discon-controller-" # Prefix for container names
      memory_limit: "512m"                # Memory limit per container
      cpu_limit: 1.0                      # CPU limit per container
      cleanup_timeout: 30                 # Cleanup after inactivity (seconds)

Controller Database
==================

The discon-manager uses a JSON database (typically at `db/controllers.json`) to track available controller versions:

.. code-block:: json

    {
      "controllers": [
        {
          "id": "default",
          "name": "Default Controller",
          "version": "1.0.0",
          "image": "discon-server:latest",
          "library_path": "/app/build/test-discon.dll",
          "proc_name": "discon",
          "ports": {
            "internal": 8080,
            "external": 0
          }
        },
        {
          "id": "rosco",
          "name": "ROSCO Controller",
          "version": "2.6.0",
          "image": "discon-server-rosco:latest",
          "library_path": "/app/build/libdiscon.so",
          "proc_name": "DISCON",
          "ports": {
            "internal": 8080,
            "external": 0
          }
        }
      ]
    }

This allows administrators to:

1. Register different controller versions
2. Specify different Docker images for each controller
3. Configure controller-specific settings (library path, procedure name)

Administration Interface
======================

The discon-manager provides a web-based administration interface for:

1. Monitoring active connections and containers
2. Viewing system metrics and health status
3. Managing controller versions
4. Testing controller configurations
5. Viewing logs and diagnostics

Endpoints
=========

The discon-manager provides several HTTP endpoints:

.. list-table::
   :widths: 25 75
   :header-rows: 1

   * - Endpoint
     - Description
   * - /ws
     - WebSocket endpoint for client connections
   * - /health
     - Health check endpoint
   * - /metrics
     - Basic metrics endpoint
   * - /containers
     - List of active containers
   * - /controllers
     - List of available controllers
   * - /admin
     - Administration web interface

Client Connection
===============

Clients connect to the discon-manager using a WebSocket URL:

::

    ws://hostname:8080/ws?controller=ID&proc=PROCNAME

With query parameters:

- `controller=ID`: Use a specific controller by ID
- `version=VERSION`: Use a specific controller by version
- `path` (optional): Override controller library path
- `proc` (optional): Override controller function name

For example:

::

    ws://localhost:8080/ws?controller=rosco&proc=DISCON