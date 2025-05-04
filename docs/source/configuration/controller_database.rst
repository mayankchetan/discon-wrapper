========================
Controller Database
========================

Overview
========

The controller database is a JSON file that defines the available controller versions and their configurations. It is used by the discon-manager to determine which Docker image to use for a particular controller version and how to configure the container.

File Structure
=============

The controller database file (typically ``db/controllers.json``) contains a JSON object with a ``controllers`` array:

.. code-block:: json

    {
      "controllers": [
        {
          "id": "controller-id",
          "name": "Controller Name",
          "version": "1.0.0",
          "image": "controller-image:tag",
          "library_path": "/path/to/controller.dll",
          "proc_name": "DISCON",
          "ports": {
            "internal": 8080,
            "external": 0
          }
        },
        // Additional controllers...
      ]
    }

Controller Fields
===============

Each controller entry in the database supports the following fields:

.. list-table::
   :widths: 25 75
   :header-rows: 1

   * - Field
     - Description
   * - id
     - **Required**. Unique identifier for the controller. Used in client requests. **Important**: Do not use periods (``.``) in controller IDs as they can cause issues when the ID is used in URLs for the ``DISCON_LIB_PATH`` environment variable. Use hyphens (``-``) or underscores (``_``) instead.
   * - name
     - **Required**. Human-readable name for the controller.
   * - version
     - **Required**. Version string for the controller. Used for version-based selection.
   * - image
     - **Required**. Docker image to use for the controller container.
   * - library_path
     - **Required**. Path to the controller library inside the container.
   * - proc_name
     - **Required**. Name of the procedure to call in the controller library.
   * - ports.internal
     - **Required**. Port that the discon-server listens on inside the container.
   * - ports.external
     - Optional. Port to expose on the host. Use 0 for dynamic port assignment.

Example Database
==============

Here's an example database with multiple controller versions:

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
        },
        {
          "id": "rosco",
          "name": "ROSCO Controller",
          "version": "2.7.0",
          "image": "discon-server-rosco:2.7",
          "library_path": "/app/build/libdiscon.so",
          "proc_name": "DISCON",
          "ports": {
            "internal": 8080,
            "external": 0
          }
        }
      ]
    }

Version Selection
===============

The discon-manager supports selecting controllers by either:

1. **Controller ID**: Using the ``controller`` query parameter in the WebSocket URL
2. **Controller Version**: Using the ``version`` query parameter in the WebSocket URL

When multiple controllers have the same ID but different versions, the manager will:

- Use the exact version if specified in the ``version`` parameter
- Use the latest version if only the ``controller`` parameter is specified

Default Controller
================

The first controller in the database is considered the default controller. When a client connects without specifying a controller ID or version, the manager will use this controller.

Managing Controllers
==================

To add, update, or remove controllers:

1. Edit the ``controllers.json`` file
2. Restart the discon-manager or use the admin interface to reload the database

Docker Images
============

Each controller entry specifies a Docker image to use. These images should:

1. Be based on the discon-server image
2. Include the specific controller library and any dependencies
3. Expose the internal port specified in the controller configuration

Custom Controller Images
======================

You can create custom controller images by extending the base discon-server image:

.. code-block:: docker

    # Dockerfile for custom controller
    FROM discon-server:latest

    # Install any additional dependencies
    RUN apt-get update && apt-get install -y \
        some-dependency \
        another-dependency \
        && rm -rf /var/lib/apt/lists/*

    # Copy your controller library
    COPY my-controller.dll /controller/my-controller.dll

    # Default command remains the same
    CMD ["--port=8080"]

Build and tag the image to match the image name in your controller database:

.. code-block:: bash

    docker build -t my-custom-controller:1.0 -f Dockerfile.custom .

Then add an entry to your controllers.json file:

.. code-block:: json

    {
      "id": "custom",
      "name": "My Custom Controller",
      "version": "1.0",
      "image": "my-custom-controller:1.0",
      "library_path": "/controller/my-controller.dll",
      "proc_name": "DISCON",
      "ports": {
        "internal": 8080,
        "external": 0
      }
    }