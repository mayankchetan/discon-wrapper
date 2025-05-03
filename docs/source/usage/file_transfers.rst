=============
File Transfers
=============

Overview
========

File transfers are a key feature of DISCON-Wrapper, allowing controller input files to remain on the client machine while being automatically transferred to the server when needed. This guide explains how file transfers work and how to optimize them for your workflow.

How File Transfers Work
======================

The file transfer mechanism follows these steps:

1. OpenFAST calls the DISCON function with a path to an input file (accINFILE parameter)
2. The discon-client checks if the file exists locally
3. If the file exists and hasn't been transferred yet:
   a. The file content is read into memory
   b. The content is sent to the server in a special file transfer payload
   c. The server saves the file to a temporary location
   d. The server returns the temporary file path
4. The client replaces the original path with the server-side path in subsequent calls
5. When the connection closes, all temporary files are automatically cleaned up

Primary Input File Transfer
=========================

The primary input file is typically specified in the ServoDyn input file as ``DLL_InFile`` (e.g., ``DISCON.IN``). This file is automatically detected and transferred when:

1. It exists in the client's working directory
2. The controller attempts to access it

The file transfer happens transparently without requiring additional configuration.

Additional Files Transfer
=======================

Some controllers require multiple input files beyond the primary DISCON.IN file. These additional files can be specified using the ``DISCON_ADDITIONAL_FILES`` environment variable:

.. code-block:: bash

    export DISCON_ADDITIONAL_FILES=additional_params.txt,lookup_table.dat,controller_settings.yaml

Files listed in this variable will be:

1. Transferred to the server before the first controller call
2. Available to the controller in the server's temporary directory
3. Cleaned up when the connection closes

File Reference Updating
=====================

When controllers reference other files inside their input files, DISCON-Wrapper can automatically update these references. For example, if DISCON.IN contains:

.. code-block:: text

    ! Gain schedule table
    "gain_schedule.dat"    GainFile   - File containing gain schedule

This reference will be automatically detected and updated to point to the server-side path when:

1. The file exists locally and is listed in ``DISCON_ADDITIONAL_FILES``
2. The file has been transferred to the server

This feature ensures that file references remain valid on the server side.

Best Practices for File Transfers
===============================

For optimal performance:

1. **Keep files small**: Large files increase initialization time
2. **Transfer only what's needed**: Include only necessary files in ``DISCON_ADDITIONAL_FILES``
3. **Use relative paths**: Avoid absolute paths in file references
4. **Keep files in the working directory**: Files are searched relative to the working directory

File Transfer Limitations
=======================

The current file transfer system has some limitations:

1. **One-time transfer**: Files are transferred only once at the beginning of the connection
2. **No write-back**: Changes made to files on the server are not transferred back to the client
3. **Size limits**: Very large files (>100MB) may cause performance issues
4. **Path complexity**: Complex path structures might not resolve correctly between systems

File Transfer Troubleshooting
===========================

Common file transfer issues and solutions:

1. **File not found errors**:
   - Verify the file exists in the client working directory
   - Check that file paths are correct in ``DISCON_ADDITIONAL_FILES``
   - Ensure file permissions allow reading

2. **File reference problems**:
   - Use relative paths in controller input files
   - Add any referenced files to ``DISCON_ADDITIONAL_FILES``
   - Enable debug output to see which paths are being used

3. **Performance issues**:
   - Reduce the number and size of transferred files
   - Pre-compress large data files if possible
   - Use binary formats instead of text for large data sets

Advanced: Manual File Transfer
============================

For special cases where automatic file transfer doesn't meet your needs, you can manually transfer files to the server before running OpenFAST and then reference them directly:

1. Copy files to the server machine
2. Set ``DISCON_LIB_PATH`` to point to the controller on the server
3. Reference server-side paths directly in your input files

This approach bypasses the automatic file transfer system and may be useful for very large files that remain constant across multiple simulations.