===============
Architecture
===============

System Overview
==============

DISCON-Wrapper follows a client-server architecture designed to bridge compatibility gaps between 64-bit simulation software and 32-bit controller libraries. The system consists of three primary components working together:

.. image:: ../_static/architecture_diagram.png
   :alt: DISCON-Wrapper Architecture
   :align: center
   :width: 600px

Components Interaction
=====================

The following diagram shows how the components interact during a wind turbine simulation:

::

    ┌─────────┐                  ┌────────────────┐                  ┌───────────────────┐
    │         │                  │                │                  │ Container 1       │
    │ OpenFAST│  WebSocket       │  DisconManager │   WebSocket      │ ┌───────────────┐ │
    │ Client  │◄──────────────► │                │◄─────────────────┤►│ DisconServer   │ │
    │         │                  │   (Proxy)      │                  │ │ + Controller 1 │ │
    └─────────┘                  └────────────────┘                  │ └───────────────┘ │
                                        │                           └───────────────────┘
                                        │                                    ▲
                                        │                                    │
                                        │                           ┌───────────────────┐
                                        │                           │ Container 2       │
                                        │                           │ ┌───────────────┐ │
                                        └───────────────────────────┤►│ DisconServer   │ │
                                                                    │ │ + Controller 2 │ │
                                                                    │ └───────────────┘ │
                                                                    └───────────────────┘

Data Flow
=========

The key data flow in the system is:

1. **OpenFAST to discon-client**: OpenFAST loads the discon-client shared library and calls the DISCON function passing controller parameters
2. **discon-client to discon-manager**: The client forwards the data via WebSocket to the manager
3. **discon-manager to container**: The manager proxies the request to the appropriate controller container
4. **container to controller**: The discon-server inside the container forwards the call to the actual controller library
5. **Return path**: The results flow back through the same channels in reverse

Key Design Principles
====================

1. **Compatibility**: Bridge between different architecture binaries (32-bit and 64-bit)
2. **Containerization**: Isolate controller environments for better stability and security
3. **Dynamic scaling**: Create containers on-demand based on client requirements
4. **Low latency**: Minimize overhead in communication to ensure real-time performance
5. **Fault tolerance**: Handle connection drops and errors gracefully

.. toctree::
   :maxdepth: 2

   technical_details
   communication_protocol
   containerization
   security_model