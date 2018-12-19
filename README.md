![Logo](https://user-images.githubusercontent.com/5942370/50045447-d14f3400-0060-11e9-8e98-78cfdcf85a75.png)

# Taask CLI

:wave: Welcome!

Taask Core is an open source system for running arbitrary jobs on any infrastructure. In other words, `cloud native tasks-as-a-service`.

taaskctl is a command-line utility for interacting with a Taask Core installation.

## Project status
:warning: Taask is in *pre-Alpha* and should not be used for critical workloads. When all critical components have been implemented, and the platform's security has been fully reviewed, it will graduate to alpha.

## More info
taaskctl is implemented as a user interface for [client-golang](https://github.com/taask/client-golang).

## Capabilities
taaskctl is currently a test bed for Taask Core

It allows running tasks from a YAML or JSON file that implements the Task spec using `taaskctl create`. That spec is described shown in exampletask.yaml.

It also allows for some basic load testing of a Taask cluster with the `taaskctl chaos` command.

## Plans
Eventually, taaskctl will become the main interface for accessing, administering, and operating a Taask cluster. It will grow to include commands such as:
- `taaskctl install [kubernetes | docker-swarm | etc]` to install taaskctl with one commend
- `taaskctl list` to list tasks in various states, of various kinds.