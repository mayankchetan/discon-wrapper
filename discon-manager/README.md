# DisconManager

DisconManager is a server application that manages Docker containers for OpenFAST controllers. It dynamically creates containers when clients connect, based on the controller version they request, and forwards the client connections to the appropriate container.

## Features

- On-demand creation of Docker containers for different controller versions
- WebSocket proxying to route client connections to appropriate containers
- Automatic container cleanup when connections are closed
- Resource limits for containers (memory, CPU)
- Simple JSON database for controller image mapping
- Health monitoring and basic metrics
- Compatible with existing discon-wrapper client code

## Architecture

DisconManager operates as a proxy between OpenFAST clients and controller-specific Docker containers:

```
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
```

## Configuration

Configuration is managed through a YAML file in `config/config.yaml`:

```yaml
# Server configuration
server:
  port: 8080          # Port to listen on
  host: "0.0.0.0"     # Interface to bind to
  debug_level: 1      # 0=disabled, 1=basic info, 2=verbose

# Docker configuration
docker:
  network_name: "discon-network"      # Docker network name
  container_prefix: "discon-controller-" # Prefix for container names
  memory_limit: "512m"                # Memory limit per container
  cpu_limit: 1.0                      # CPU limit per container
  cleanup_timeout: 30                 # Cleanup after inactivity (seconds)
```

## Controller Database

Controllers are defined in `db/controllers.json`:

```json
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
```

## Building and Running

### Prerequisites

- Docker with Docker Compose
- Access to Docker socket for Docker-out-of-Docker operations

### Building

Run the build script to build all required images:

```bash
./build.sh
```

### Running

Start the DisconManager service:

```bash
docker-compose up -d
```

### Management Endpoints

- `/health` - Health check
- `/metrics` - Basic metrics
- `/containers` - List running containers
- `/controllers` - List available controllers

## Client Connection

### Connecting to DisconManager

The client should connect to:

```
ws://hostname:8080/ws
```

With one of the following query parameter combinations:

- `controller=ID` - Use a specific controller by ID
- `version=VERSION` - Use a specific controller by version
- Or no parameters to use the default controller

Additional optional parameters:
- `path` - Override controller library path
- `proc` - Override controller function name

### Example

```
ws://localhost:8080/ws?controller=rosco&proc=DISCON
```

## Troubleshooting

### Common Issues

1. **Container not starting properly**
   - Check Docker daemon is running
   - Verify Docker network exists: `docker network ls`
   - Check container logs: `docker logs <container-id>`

2. **Connection errors**
   - Verify the DisconManager is running: `docker-compose ps`
   - Check DisconManager logs: `docker-compose logs discon-manager`

3. **Controller not found**
   - Verify the controller ID/version in controllers.json
   - Check that the corresponding Docker image exists: `docker images`

## Resources

- Memory usage per container: 512 MB
- CPU usage per container: 1 core
- Container cleanup: 30 seconds after client disconnection