=================
Contributing Guide
=================

Overview
========

This guide provides information for developers interested in contributing to DISCON-Wrapper. We welcome contributions in various forms including bug fixes, feature additions, documentation improvements, and testing enhancements.

Getting Started
=============

1. **Fork the Repository**:
   
   Fork the DISCON-Wrapper repository to your GitHub account to create your own copy where you can make changes.

2. **Clone Your Fork**:

   .. code-block:: bash

       git clone https://github.com/your-username/discon-wrapper.git
       cd discon-wrapper

3. **Set Up Remote**:

   Add the original repository as an upstream remote to stay in sync with it:

   .. code-block:: bash

       git remote add upstream https://github.com/deslaughter/discon-wrapper.git

4. **Create a Branch**:

   Create a new branch for your changes:

   .. code-block:: bash

       git checkout -b feature/my-contribution

Types of Contributions
====================

We welcome various types of contributions:

Bug Fixes
--------

If you find a bug:

1. First check if the issue is already reported in the GitHub issue tracker
2. If not, create a new issue describing the bug, steps to reproduce, and your environment
3. If you want to fix it yourself, comment on the issue to avoid duplicate work
4. Create a pull request referencing the issue number

Feature Additions
---------------

When proposing new features:

1. Start by opening an issue describing the feature and its benefits
2. Discuss the design and implementation approach with maintainers
3. Create a pull request with the implementation
4. Include tests and documentation for the new feature

Documentation Improvements
------------------------

Documentation contributions are highly valued:

1. Correct errors or outdated information
2. Add missing details or examples
3. Improve clarity and organization
4. Add diagrams or illustrations where helpful

Code Style and Guidelines
=======================

All contributions should follow these guidelines:

1. **Go Formatting**:
   
   Format your code using ``go fmt`` before submitting:

   .. code-block:: bash

       go fmt ./...

2. **Code Validation**:

   Run ``go vet`` to catch common errors:

   .. code-block:: bash

       go vet ./...

3. **Testing**:

   Ensure all tests pass and add tests for your changes:

   .. code-block:: bash

       go test ./...

4. **Comments**:

   - Use meaningful comments to explain non-obvious code
   - Document exported functions, types, and variables
   - Follow Go comment conventions (starting with the name of the thing being documented)

5. **Error Handling**:

   - Handle all errors appropriately
   - Avoid using ``_`` to ignore errors unless justified
   - Add context to errors when propagating them

6. **Commit Messages**:

   - Write clear, descriptive commit messages
   - Start with a short summary line
   - Include more details in the body if needed
   - Reference issue numbers when applicable

Pull Request Process
==================

1. **Create a Pull Request**:
   
   When your changes are ready, push to your fork and create a pull request to the main repository.

2. **PR Description**:

   Include in your pull request:
   
   - A summary of the changes
   - References to relevant issues
   - Notes on testing you've performed
   - Any specific areas needing reviewer attention
   - Documentation updates if required

3. **Code Review**:

   - Be open to feedback and make requested changes
   - Respond to comments in a timely manner
   - Keep discussions constructive and focused on the code
   - Update your PR with new commits as needed

4. **CI Checks**:

   Ensure all continuous integration checks pass:
   
   - Tests
   - Code formatting
   - Build verification

5. **Merge**:

   After approval, maintainers will merge your PR.

Development Best Practices
========================

1. **Keep Changes Focused**:
   
   Each pull request should address a single concern. Split large changes into multiple PRs.

2. **Backwards Compatibility**:
   
   Be mindful of backward compatibility, especially for public APIs.

3. **Cross-Platform**:
   
   Ensure code works on both Windows and Linux.

4. **Performance Considerations**:
   
   Be aware of performance implications, especially for code that runs in critical paths.

5. **Security**:
   
   Follow security best practices, especially for code handling file operations or external input.

Getting Help
==========

If you need assistance with your contribution:

1. Comment on the relevant issue
2. Reach out to maintainers
3. Ask questions in discussions or community forums

We appreciate your contribution and are happy to help you through the process.