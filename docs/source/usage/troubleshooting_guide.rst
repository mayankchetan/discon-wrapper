===================
Troubleshooting Guide
===================

Overview
========

This guide provides solutions to common issues encountered when using DISCON-Wrapper. It covers problems related to installation, configuration, connection, and runtime behavior.

Connection Issues
===============

Failed to Connect to Server
-------------------------

**Symptoms**: The client reports "connection refused" or "failed to connect to server" errors.

**Solutions**:

1. **Check server status**:
   - Verify the server is running with ``ps aux | grep discon-server``
   - Check server logs for startup errors

2. **Verify connection parameters**:
   - Ensure DISCON_SERVER_ADDR is correctly set
   - Confirm the port number matches the server's configuration

3. **Network issues**:
   - Check firewall settings: ``sudo iptables -L | grep <port>``
   - Verify network connectivity: ``ping <server_host>``

4. **For discon-manager**:
   - Ensure the Docker network exists: ``docker network ls``
   - Check that container creation succeeds: ``docker-compose logs discon-manager``

Connection Times Out
------------------

**Symptoms**: The client connects initially but then disconnects with a timeout error.

**Solutions**:

1. **Check server load**:
   - Monitor server CPU/memory: ``top -b -n 1 | grep discon-server``
   - If overloaded, increase server resources

2. **Network stability**:
   - Check for packet loss: ``ping -c 20 <server_host>``
   - Monitor network latency during simulation

3. **Timeout settings**:
   - For long-running simulations, adjust WebSocket timeouts in code

WebSocket Handshake Failed
------------------------

**Symptoms**: Error message about WebSocket handshake failure.

**Solutions**:

1. **Check protocol**:
   - Ensure the URL uses the correct ws:// or wss:// prefix
   - For secure connections, verify certificates

2. **Proxy issues**:
   - If using a reverse proxy, ensure it supports WebSockets
   - Check proxy timeout settings

Controller Issues
===============

Controller Library Not Found
-------------------------

**Symptoms**: Error message "Error loading shared library" or "Library not found".

**Solutions**:

1. **Verify library path**:
   - Check DISCON_LIB_PATH environment variable
   - Ensure the path is correct from the server's perspective
   - Look for path case sensitivity issues on Linux

2. **Library permissions**:
   - Check file permissions: ``ls -la <library_path>``
   - Ensure the server process can read the library

3. **Dependencies**:
   - Verify library dependencies using ``ldd`` (Linux) or Dependency Walker (Windows)
   - Install any missing runtime dependencies

Controller Function Not Found
--------------------------

**Symptoms**: Error message "Error loading function from shared library".

**Solutions**:

1. **Function name mismatch**:
   - Verify DISCON_LIB_PROC environment variable
   - Check function name case (some linkers are case-sensitive)
   - Use ``nm -D <library>`` on Linux to list exported symbols

2. **Function export issues**:
   - Ensure the function is properly exported from the library
   - Check for name mangling (C++ functions)

Controller Execution Crashes
-------------------------

**Symptoms**: The server crashes or returns an error during controller execution.

**Solutions**:

1. **Memory issues**:
   - Check for buffer overflows in controller code
   - Verify array sizes in the SWAP array
   - Monitor memory usage with ``docker stats`` (for containerized setup)

2. **Runtime errors**:
   - Increase debug level to get more detailed error messages
   - Check for division by zero or other numerical issues

3. **Binary compatibility**:
   - Ensure the controller was compiled for the correct architecture
   - Verify that required runtime libraries are installed

File Transfer Issues
==================

File Not Found
------------

**Symptoms**: "File not found" errors during file transfer.

**Solutions**:

1. **Check file existence**:
   - Verify the file exists at the specified path
   - Check file permissions

2. **Path issues**:
   - Use relative paths rather than absolute paths
   - For additional files, check DISCON_ADDITIONAL_FILES format

3. **Working directory**:
   - Run OpenFAST from the directory containing input files
   - Or specify full paths to input files

File Transfer Fails
----------------

**Symptoms**: File transfer starts but fails to complete.

**Solutions**:

1. **File size issues**:
   - Large files may cause memory problems, try reducing file size
   - Split large data files into smaller files

2. **Transfer timeouts**:
   - Increase timeout settings for large files
   - Check network bandwidth between client and server

3. **Storage issues**:
   - Ensure sufficient disk space on server
   - Check file system permissions

Performance Issues
================

High Latency
----------

**Symptoms**: The simulation runs slower than expected due to communication overhead.

**Solutions**:

1. **Network optimization**:
   - Run client and server on the same network
   - Minimize the number of network hops

2. **Reduce debug output**:
   - Set DISCON_CLIENT_DEBUG=0 for production runs
   - Disable verbose logging on the server

3. **Optimization flags**:
   - Compile the controller with optimization flags
   - Use release builds instead of debug builds

Memory Usage Growth
----------------

**Symptoms**: The server's memory usage increases over time.

**Solutions**:

1. **Check for memory leaks**:
   - Use tools like Valgrind to check for leaks in controller code
   - Monitor memory usage patterns

2. **Container resource limits**:
   - Set appropriate memory limits in docker-compose.yml
   - Monitor container memory usage with ``docker stats``

3. **Garbage collection**:
   - Ensure proper cleanup of temporary files
   - Check for accumulated connections

Docker-specific Issues
====================

Container Creation Fails
---------------------

**Symptoms**: discon-manager fails to create controller containers.

**Solutions**:

1. **Docker access**:
   - Verify Docker socket access: ``ls -la /var/run/docker.sock``
   - Check user permissions for Docker socket

2. **Image availability**:
   - Ensure all required images exist: ``docker images``
   - Pull or build missing images

3. **Resource constraints**:
   - Check system resources (CPU, memory, disk space)
   - Adjust resource limits in docker-compose.yml

Container Networking Issues
------------------------

**Symptoms**: Containers start but can't communicate.

**Solutions**:

1. **Network existence**:
   - Verify network exists: ``docker network ls``
   - Check network configuration

2. **Container IP assignment**:
   - Inspect container network settings: ``docker inspect <container_id>``
   - Check for IP address conflicts

3. **DNS resolution**:
   - Test DNS resolution within containers
   - Check /etc/hosts and /etc/resolv.conf

Collecting Diagnostic Information
===============================

When reporting issues, collect the following diagnostic information:

1. **Version information**:
   - DISCON-Wrapper component versions
   - OpenFAST version
   - Operating system version
   - Docker version (if applicable)

2. **Log files**:
   - Client debug output
   - Server logs
   - Docker container logs: ``docker-compose logs``

3. **Configuration**:
   - Environment variables
   - config.yaml (for discon-manager)
   - controllers.json (for discon-manager)

4. **Error messages**:
   - Copy the full error message text
   - Include context around when the error occurred