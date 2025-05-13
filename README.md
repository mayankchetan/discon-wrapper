# DISCON-Wrapper

DISCON-Wrapper is a comprehensive solution for wind turbine controller integration, providing seamless interoperability between different architectures and platforms. This project creates a networked bridge between OpenFAST simulations and controller libraries, solving compatibility issues and enabling advanced container-based orchestration.

## Key Features

- **Cross-Architecture Compatibility**: Run 32-bit controllers with 64-bit OpenFAST simulations
- **Cross-Platform Support**: Bridge Windows controllers to Linux environments and vice versa
- **File Transfer System**: Automatic transfer of input/output files between client and server
- **Container Orchestration**: Dynamic container management for multiple controller versions
- **Admin Interface**: Web-based management of controllers and containers
- **Comprehensive Documentation**: Complete guides for setup, configuration, and troubleshooting

## Components

DISCON-Wrapper consists of three main components: 

1. **discon-client**: A shared library that replaces the controller in OpenFAST
2. **discon-server**: A server application that loads and executes the actual controller
3. **discon-manager**: A container orchestration layer for managing multiple controllers

## When to Use DISCON-Wrapper

This project is especially valuable if:

- You have wind turbine controllers implementing the Bladed API
- Your controllers are 32-bit but your simulation environment is 64-bit
- You need to run controllers across different operating systems
- You want to centrally manage multiple controller versions
- You don't have access to controller source code for recompilation

## Basic Setup

### Server Setup

1. Download `discon-server` from [Releases](https://github.com/deslaughter/discon-wrapper/releases)
2. Place it in the same directory as your controller library
3. Start the server:
   ```bash
   # Linux
   ./discon-server_386 --port=8080 --debug=1
   
   # Windows
   discon-server_386.exe --port=8080 --debug=1
   ```

Debug levels:
- `0`: No debug output (default)
- `1`: Basic debug information
- `2`: Verbose debug output including full payloads

### Client Setup

1. Download `discon-client` from [Releases](https://github.com/deslaughter/discon-wrapper/releases)
2. Update your ServoDyn input file:

```
"discon-client_amd64.dll"    DLL_FileName - Name/location of the dynamic library
"DISCON.IN"                  DLL_InFile   - Name of input file sent to the DLL (-)
"DISCON"                     DLL_ProcName - Name of procedure in DLL to be called (-)
```

3. Set environment variables and run OpenFAST:

```bash
# Linux
export DISCON_SERVER_ADDR=localhost:8080
export DISCON_LIB_PATH=path/to/controller.so
export DISCON_LIB_PROC=DISCON
export DISCON_CLIENT_DEBUG=1
openfast my_turbine.fst

# Windows
set DISCON_SERVER_ADDR=localhost:8080
set DISCON_LIB_PATH=controller.dll
set DISCON_LIB_PROC=CONTROL
set DISCON_CLIENT_DEBUG=1
openfast.exe my_turbine.fst
```

## Advanced Usage with discon-manager

For production environments, the `discon-manager` component provides container orchestration:

1. Start discon-manager:
   ```bash
   docker-compose up -d
   ```

2. Configure your client to connect through the manager:
   ```bash
   export DISCON_SERVER_ADDR=localhost:8080
   export DISCON_CONTROLLER=rosco  # Controller ID in the database
   export DISCON_VERSION=2.6.0     # Optional: specific version
   openfast my_turbine.fst
   ```

The manager automatically:
- Creates containers for requested controllers
- Routes client connections to appropriate containers
- Monitors and cleans up inactive containers
- Provides a web admin interface at `http://localhost:8080/admin`

## File Transfers

DISCON-Wrapper automatically handles file transfers between client and server:

1. When a controller requests an input file, the client checks if it exists locally
2. If found, the file is transferred to the server automatically
3. The file path is updated for the controller on the server side

Benefits:
- Input files stay on the client machine
- Files are transferred only once per simulation
- Automatic cleanup when the connection closes

## Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| DISCON_SERVER_ADDR | Server address with port | localhost:8080 |
| DISCON_LIB_PATH | Path to controller library | controller.dll |
| DISCON_LIB_PROC | Controller procedure name | DISCON |
| DISCON_CLIENT_DEBUG | Debug level (0-2) | 1 |
| DISCON_CONTROLLER | Controller ID for use with manager | rosco |
| DISCON_VERSION | Controller version | 2.6.0 |
| DISCON_ADDITIONAL_FILES | Comma-separated additional files to transfer | file1.txt,file2.dat |

## Container Management

The discon-manager implements modern container lifecycle management:

- **Dynamic Creation**: Containers created on-demand when clients request controllers
- **Graceful Shutdown**: Proper termination with configurable timeouts
- **Resource Limits**: Memory and CPU constraints to prevent resource exhaustion
- **Automatic Cleanup**: Removal of inactive containers to conserve resources

## Documentation

Comprehensive documentation is available in the `docs/` directory:

- **Installation Guide**: Detailed setup instructions
- **Configuration**: Client, server, and manager configuration options
- **Usage Guides**: OpenFAST integration and advanced usage scenarios
- **Architecture**: System design and component interactions
- **Development**: Contributing guidelines and code organization
- **Troubleshooting**: Common issues and solutions

## Status

While DISCON-Wrapper is now used in production environments, it's continuously evolving. We welcome contributions and feedback from the community.

## License

DISCON-Wrapper is licensed under the Apache License 2.0.
