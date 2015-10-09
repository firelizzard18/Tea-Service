# Tea-Service

## A simple aid for service management

`teasvc` runs in two modes - service mode and connect mode. Service mode involves launching a service as a child process and then exporting an object on D-Bus. Connect mode involves using that exported object to connect back to the service-mode teasvc process. Connect mode primarily involves one of four activities: 1) connecting to the output of the child process; 2) connecting to the input and output of the child process; 3) sending a command to the child process; 4) or listing all launched service-mode teasvc processes.

When the service-mode process is launched, it can connect to the system bus or session bus (the default). If it connects to the session bus, only command-mode processes running as the same user can communicate with it. If it connects to the system bus, any process can talk to it.

### `teasvc` is in development and is not ready for use
