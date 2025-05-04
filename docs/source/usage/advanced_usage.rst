=============
Advanced Usage
=============

Overview
========

This guide covers advanced usage scenarios for DISCON-Wrapper, including integration with CI/CD pipelines, cluster computing, benchmarking, and custom extensions to the system.

Batch Processing
==============

Running Multiple Simulations
--------------------------

DISCON-Wrapper supports running multiple simulations in batch mode:

.. code-block:: bash

    #!/bin/bash
    
    # Set common environment variables
    export DISCON_SERVER_ADDR=localhost:8080
    export DISCON_LIB_PATH=/controller/my_controller.dll
    export DISCON_LIB_PROC=DISCON
    
    # Run multiple simulations
    for wind_speed in 8 10 12 14 16; do
        # Create simulation-specific input files
        sed "s/WIND_SPEED_PLACEHOLDER/$wind_speed/" DISCON.IN.template > DISCON.IN
        
        # Run OpenFAST with this configuration
        openfast simulation_${wind_speed}mps.fst
    done

Dynamic Controller Selection
--------------------------

With discon-manager, you can dynamically select different controllers for each simulation:

.. code-block:: bash

    #!/bin/bash
    
    # Common settings
    export DISCON_SERVER_ADDR=localhost:8080
    export DISCON_LIB_PROC=DISCON
    
    # Run with controller version 1.0
    export DISCON_SERVER_ADDR="localhost:8080/ws?version=1.0"
    openfast simulation_v1.fst
    
    # Run with controller version 2.0
    export DISCON_SERVER_ADDR="localhost:8080/ws?version=2.0"
    openfast simulation_v2.fst

Integration with HPC Environments
===============================

MPI Integration
-------------

When running OpenFAST with MPI for parallel simulations, DISCON-Wrapper can be configured to handle multiple connections:

.. code-block:: bash

    #!/bin/bash
    #SBATCH --nodes=4
    #SBATCH --ntasks-per-node=16
    #SBATCH --time=04:00:00
    
    # Load modules
    module load openfast
    module load mpi
    
    # Set environment variables
    export DISCON_SERVER_ADDR=controller-server.example.com:8080
    export DISCON_LIB_PATH=/app/controllers/controller.dll
    export DISCON_LIB_PROC=DISCON
    
    # Run OpenFAST with MPI
    mpirun -np 64 openfast_mpi simulation.fst

Container Orchestration
---------------------

For large-scale simulations, you can integrate discon-manager with container orchestration systems like Kubernetes:

1. Deploy discon-manager as a Kubernetes service
2. Configure horizontal pod autoscaling based on connection load
3. Use persistent volumes for controller libraries and configuration
4. Create a service for client connections

Example Kubernetes configuration snippets:

.. code-block:: yaml

    # discon-manager deployment
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: discon-manager
    spec:
      replicas: 1
      selector:
        matchLabels:
          app: discon-manager
      template:
        metadata:
          labels:
            app: discon-manager
        spec:
          containers:
          - name: discon-manager
            image: discon-wrapper/discon-manager:latest
            ports:
            - containerPort: 8080
            volumeMounts:
            - name: docker-socket
              mountPath: /var/run/docker.sock
            - name: config
              mountPath: /app/config
            - name: db
              mountPath: /app/db
          volumes:
          - name: docker-socket
            hostPath:
              path: /var/run/docker.sock
          - name: config
            configMap:
              name: discon-manager-config
          - name: db
            configMap:
              name: discon-manager-db

Performance Benchmarking
======================

Measuring Overhead
---------------

To measure the performance overhead of DISCON-Wrapper compared to direct controller loading:

1. Run a baseline simulation with the controller loaded directly
2. Run the same simulation with DISCON-Wrapper
3. Compare simulation duration and CPU/memory usage

Sample benchmark script:

.. code-block:: bash

    #!/bin/bash
    
    # Function to measure execution time
    measure_time() {
        echo "Running $1..."
        START_TIME=$(date +%s)
        $2
        END_TIME=$(date +%s)
        ELAPSED=$(( END_TIME - START_TIME ))
        echo "$1 completed in $ELAPSED seconds"
        return $ELAPSED
    }
    
    # Baseline: Direct controller
    sed -i 's/discon-client_amd64.dll/controller.dll/' ServoDyn.dat
    measure_time "Baseline" "openfast simulation.fst"
    BASELINE_TIME=$?
    
    # DISCON-Wrapper
    sed -i 's/controller.dll/discon-client_amd64.dll/' ServoDyn.dat
    export DISCON_SERVER_ADDR=localhost:8080
    export DISCON_LIB_PATH=/controller/controller.dll
    export DISCON_LIB_PROC=DISCON
    measure_time "DISCON-Wrapper" "openfast simulation.fst"
    WRAPPER_TIME=$?
    
    # Calculate overhead
    OVERHEAD=$(( (WRAPPER_TIME - BASELINE_TIME) * 100 / BASELINE_TIME ))
    echo "Performance overhead: $OVERHEAD%"

Optimizing for Speed
------------------

To minimize performance overhead:

1. **Reduce network latency**:
   - Host server and client on the same machine or LAN
   - Use direct IP addressing rather than DNS resolution

2. **Optimize file transfers**:
   - Pre-transfer all files before starting simulations
   - Keep file sizes small

3. **Minimize logging overhead**:
   - Set DISCON_CLIENT_DEBUG=0
   - Run server with --debug=0
   - Disable container logging for production runs

Extending DISCON-Wrapper
======================

Custom Container Images
--------------------

Creating custom container images allows you to package specific controller versions with their dependencies:

.. code-block:: docker

    # Start from the base discon-server image
    FROM discon-server:latest
    
    # Add controller-specific dependencies
    RUN apt-get update && apt-get install -y \
        liblapack3 \
        libblas3 \
        && rm -rf /var/lib/apt/lists/*
    
    # Copy controller libraries
    COPY controllers/ /controller/
    
    # Set environment variables
    ENV DEFAULT_CONTROLLER=/controller/my_controller.dll
    
    # Keep the default command
    CMD ["--port=8080"]

Adding Authentication
------------------

For secure deployments, you can add authentication to WebSocket connections:

1. Configure the discon-manager to use authentication tokens
2. Modify the client to include authentication headers
3. Set up a token management system

Example client-side authentication code:

.. code-block:: go

    // Add authentication header to WebSocket connection
    header := http.Header{}
    header.Add("Authorization", "Bearer " + os.Getenv("DISCON_AUTH_TOKEN"))
    
    // Connect with the header
    conn, _, err := websocket.DefaultDialer.Dial(u.String(), header)

Building Custom Extensions
-----------------------

To extend DISCON-Wrapper with custom functionality:

1. Fork the repository on GitHub
2. Implement your changes following the existing code structure
3. Use the provided build system to create new binaries
4. Consider contributing back your improvements via pull requests

Custom Monitor Applications
------------------------

You can build custom monitoring applications that connect to the discon-manager's API endpoints:

.. code-block:: python

    import requests
    import json
    import time
    
    # Configure the API endpoint
    base_url = "http://localhost:8080"
    
    # Authentication (if enabled)
    session = requests.Session()
    session.post(f"{base_url}/admin/login", data={
        "username": "admin", 
        "password": "password"
    })
    
    # Monitor active containers
    while True:
        response = session.get(f"{base_url}/containers")
        containers = json.loads(response.text)
        
        print(f"Active containers: {len(containers)}")
        for container in containers:
            print(f"- {container['name']}: {container['status']}")
        
        time.sleep(5)