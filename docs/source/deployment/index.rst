===========
Deployment
===========

Overview
========

This section covers different ways to deploy the DISCON-Wrapper system. Depending on your requirements, you can deploy:

1. **Single-machine setup**: Both client and server running on the same machine
2. **Distributed setup**: Client and server on different machines
3. **Containerized setup**: Using Docker and docker-compose for easy deployment
4. **Production setup**: Enterprise deployment with high availability and security

.. toctree::
   :maxdepth: 2

   installation
   docker_deployment
   production_deployment
   security

Component Dependencies
=====================

Each component of the DISCON-Wrapper system has different dependencies:

discon-client
------------

* 64-bit operating system (Windows/Linux)
* Environment variables for configuration
* Access to the controller input files

discon-server
------------

* 32-bit operating system (for 32-bit controllers) or 64-bit (for 64-bit controllers)
* Controller shared libraries (.dll/.so)
* Runtime libraries required by the controllers (e.g., C, C++, Fortran runtimes)

discon-manager
-------------

* Docker and docker-compose
* Access to the Docker socket
* Controller Docker images
* Controller database configuration

Choosing a Deployment Model
==========================

The right deployment model depends on your needs:

* **Development**: Start with the single-machine setup for easy debugging
* **Small team**: Use the Docker deployment for easy setup and maintenance
* **Enterprise**: Choose the production deployment for robustness and security

For most users, the Docker deployment offers the best balance of simplicity and features.