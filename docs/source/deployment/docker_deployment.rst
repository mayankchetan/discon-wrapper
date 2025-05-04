=================
Docker Deployment
=================

Overview
========

Docker provides the simplest way to deploy the DISCON-Wrapper system, particularly when using discon-manager. This guide covers how to set up a complete DISCON-Wrapper environment using Docker and docker-compose.

Prerequisites
============

Before proceeding, ensure you have:

- Docker installed (version 20.10.0 or newer recommended)
- docker-compose installed (version 2.0.0 or newer recommended)
- Basic understanding of Docker concepts

Docker Components
===============

The DISCON-Wrapper system uses the following Docker components:

1. **discon-manager container**: Orchestrates controller containers
2. **Controller containers**: Run discon-server instances with specific controller libraries
3. **Docker network**: Connects the manager and controller containers
4. **Docker volumes**: Store persistent data like configuration and metrics

Simplified Deployment
====================

For most users, the simplest deployment method is using docker-compose:

1. Clone the repository or download the docker-compose.yml file
2. Configure your controllers in db/controllers.json
3. Start the system with docker-compose

Setting Up the Environment
========================

Create a new directory for your DISCON-Wrapper deployment:

.. code-block:: bash

    mkdir discon-wrapper
    cd discon-wrapper

Create the following directory structure:

.. code-block:: text

    discon-wrapper/
    ├── config/
    │   └── config.yaml
    ├── db/
    │   └── controllers.json
    ├── metrics/
    └── docker-compose.yml

Configuration Files
=================

1. Create a config.yaml file:

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

    # Authentication configuration (optional)
    auth:
      enabled: true
      username: "admin"
      password: "changeThisPassword"
      session_timeout: 3600

    # Database configuration
    database:
      controllers_file: "db/controllers.json"

    # Metrics configuration
    metrics:
      enabled: true
      interval: 60
      storage_path: "metrics"

2. Create a controllers.json file:

.. code-block:: json

    {
      "controllers": [
        {
          "id": "default",
          "name": "Default Test Controller",
          "version": "1.0.0",
          "image": "discon-server:latest",
          "library_path": "/app/build/test-discon.dll",
          "proc_name": "discon",
          "ports": {
            "internal": 8080,
            "external": 0
          }
        }
      ]
    }

3. Create a docker-compose.yml file:

.. code-block:: yaml

    version: '3'
    
    services:
      discon-manager:
        image: discon-wrapper/discon-manager:latest
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
        networks:
          - discon-network
    
    networks:
      discon-network:
        driver: bridge

Building Docker Images
====================

You can build the Docker images using the provided build script:

.. code-block:: bash

    ./build.sh

Alternatively, you can build them manually:

.. code-block:: bash

    # Build the manager image
    docker build -t discon-wrapper/discon-manager:latest -f docker/Dockerfile.manager .
    
    # Build the server image
    docker build -t discon-server:latest -f docker/Dockerfile.server .
    
    # Build a custom controller image (e.g., ROSCO)
    docker build -t discon-server-rosco:latest -f docker/Dockerfile.rosco .

Starting the System
=================

Once your configuration files are in place and the Docker images are built, start the system:

.. code-block:: bash

    docker-compose up -d

This will start the discon-manager container in detached mode. You can check its status:

.. code-block:: bash

    docker-compose ps

And view its logs:

.. code-block:: bash

    docker-compose logs -f

Accessing the Administration Interface
===================================

Once the system is running, you can access the administration interface at:

http://localhost:8080/admin

Log in with the username and password specified in your config.yaml file.

Client Configuration
==================

Configure your discon-client to connect to the Docker deployment:

.. code-block:: bash

    # For local connections
    export DISCON_SERVER_ADDR=localhost:8080
    
    # For remote connections
    export DISCON_SERVER_ADDR=server-hostname:8080

Custom Controller Images
======================

To create a custom controller image:

1. Create a Dockerfile based on the discon-server image:

.. code-block:: docker

    FROM discon-server:latest
    
    # Install any additional dependencies
    RUN apt-get update && apt-get install -y \
        some-dependency \
        another-dependency \
        && rm -rf /var/lib/apt/lists/*
    
    # Copy your controller library
    COPY my-controller.dll /controller/my-controller.dll

2. Build the image:

.. code-block:: bash

    docker build -t my-custom-controller:1.0 -f Dockerfile.custom .

3. Update your controllers.json file to use the new image.

Stopping the System
=================

To stop the DISCON-Wrapper system:

.. code-block:: bash

    docker-compose down

This will stop the discon-manager container and remove it, but will preserve your configuration files and metrics.

Updating the System
=================

To update the system with new Docker images:

1. Pull or build the updated images
2. Restart the services:

.. code-block:: bash

    docker-compose down
    docker-compose up -d

Production Considerations
=======================

For production deployments:

1. Enable authentication with a strong password
2. Consider running behind a reverse proxy with TLS for secure WebSocket connections
3. Set appropriate resource limits for containers
4. Set up monitoring and alerting for the system

For more detailed production deployment information, see :doc:`production_deployment`.