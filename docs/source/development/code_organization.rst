=================
Code Organization
=================

Overview
========

The DISCON-Wrapper codebase is organized into several distinct modules and packages. This structure follows Go's standard package layout and is designed to maintain clear separation of concerns between components.

Repository Structure
==================

The repository is organized as follows:

.. code-block:: text

    discon-wrapper/
    ├── build/                   # Build artifacts directory
    ├── discon-client/           # Client component source code
    │   └── client.go            # Main client implementation
    ├── discon-server/           # Server component source code
    │   ├── load_shared_library.c  # C code for loading shared libraries
    │   ├── server_test.go       # Server tests
    │   ├── server.go            # Server core functionality
    │   └── websocket.go         # WebSocket handling
    ├── discon-manager/          # Manager component source code
    │   ├── admin.go             # Admin interface
    │   ├── config.go            # Configuration handling
    │   ├── database.go          # Controller database
    │   ├── docker.go            # Docker integration
    │   ├── main.go              # Entry point
    │   ├── manager.go           # Core manager functionality
    │   ├── config/              # Sample configuration files
    │   ├── db/                  # Sample database files
    │   └── templates/           # HTML templates for admin interface
    ├── docker/                  # Docker-related files
    │   ├── Dockerfile.manager   # Dockerfile for manager
    │   ├── Dockerfile.rosco     # Dockerfile for ROSCO controller
    │   └── Dockerfile.server    # Dockerfile for server
    ├── shared/                  # Shared code used by multiple components
    │   └── utils/               # Utility functions
    │       ├── file.go          # File handling utilities
    │       ├── logging.go       # Logging utilities
    │       └── transfer.go      # File transfer utilities
    ├── test-app/                # Test application for development
    ├── test-discon/             # Test controller implementation
    ├── build.sh                 # Build script
    ├── docker-compose.yml       # Docker Compose configuration
    ├── go.mod                   # Go module definition
    ├── go.sum                   # Go module checksums
    ├── payload.go               # Shared payload structure definition
    └── README.md                # Project documentation

Core Components
=============

Root Module
----------

The root module contains code shared between all components:

- **payload.go**: Defines the Payload structure used for communication between client and server

discon-client
-----------

The client component includes:

- **client.go**: Implements the DISCON function that OpenFAST calls, along with WebSocket communication, file transfer handling, and environment variable processing

discon-server
-----------

The server component includes:

- **server.go**: Contains the main function and server initialization
- **websocket.go**: Handles WebSocket connections and controller function calls
- **load_shared_library.c**: C code for dynamically loading controller libraries

discon-manager
------------

The manager component includes:

- **main.go**: Entry point and server setup
- **manager.go**: Core functionality for connection management and proxying
- **docker.go**: Docker API integration for container management
- **config.go**: Configuration loading and parsing
- **database.go**: Controller database management
- **admin.go**: Web-based admin interface

shared/utils
-----------

This package contains utility functions used by multiple components:

- **file.go**: File handling utilities
- **logging.go**: Logging utilities
- **transfer.go**: File transfer utilities

Module Dependencies
=================

The dependency tree for DISCON-Wrapper components is as follows:

.. code-block:: text

    discon-client
    ├── root module (payload.go)
    └── shared/utils
    
    discon-server
    ├── root module (payload.go)
    └── shared/utils
    
    discon-manager
    ├── root module (payload.go)
    ├── shared/utils
    └── Docker API libraries

Coding Patterns
=============

Throughout the codebase, several key patterns are used:

1. **Init function** pattern for setup and WebSocket connection in client.go
2. **Connection ID** pattern for isolating concurrent connections
3. **Binary marshaling** for efficient network communication
4. **Mutex-protected maps** for thread-safe operation
5. **Context-based handling** for container lifecycle management
6. **Environment variable configuration** for flexible deployment

Key Data Structures
=================

Payload
------

The Payload structure is the core data structure used for communication between client and server:

.. code-block:: go

    type Payload struct {
        Swap          []float32 // Controller SWAP array
        Fail          int32     // Controller fail flag
        InFile        []byte    // Controller input file path
        OutName       []byte    // Controller output name
        Msg           []byte    // Controller message buffer
        FileContent   []byte    // For file transfers: content of file
        ServerFilePath []byte   // For file transfers: server-side path
    }

ClientConnection (manager)
------------------------

The ClientConnection structure in the manager represents a connected client:

.. code-block:: go

    type ClientConnection struct {
        ID             string
        RemoteAddr     string
        ConnectedAt    time.Time
        LastActivityAt time.Time
        WS             *websocket.Conn
        ContainerID    string
        ContainerInfo  *ContainerInfo
        ProxyCloseCh   chan struct{}
        ControllerPath string
        ProcName       string
        // ... other fields ...
    }

ContainerInfo (manager)
---------------------

The ContainerInfo structure in the manager represents a controller container:

.. code-block:: go

    type ContainerInfo struct {
        ID          string
        Name        string
        Image       string
        ContainerIP string
        Host        string
        Port        int
        Status      string
        CreatedAt   time.Time
        // ... other fields ...
    }

Code Style Guidelines
===================

The DISCON-Wrapper codebase follows these style guidelines:

1. **Go standard formatting**: All code should be formatted with `go fmt`
2. **Meaningful variable names**: Variable names should clearly indicate their purpose
3. **Comments**: Functions and complex logic should be commented
4. **Error handling**: Errors should be properly propagated and handled
5. **Logging levels**: Appropriate logging levels should be used (Debug, Verbose, Error)

File Organization Principles
==========================

1. **Package per component**: Each major component has its own package
2. **Shared code in utils**: Common functionality is extracted to the shared/utils package
3. **Root module for shared structures**: Core structures used by multiple components are in the root module
4. **Separation of concerns**: Each file has a clear, focused responsibility