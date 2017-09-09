# Tea-Service

**NOTE: There are no guarantees on the reliability of `teasvc`. See issues.

## A simple aid for service management

`teasvc` has two basic roles. One is to run servces, the other is to connect to a service. Service mode involves launching a service as a child process and then exporting an object on D-Bus. Connect mode involves using that exported object to connect back to the service-mode teasvc process. Connect mode primarily involves one of four activities: 1) connecting to the output of the child process; 2) connecting to the input and output of the child process; 3) sending a command to the child process; 4) or listing all launched service-mode teasvc processes.

When the service-mode process is launched, it can connect to the system bus or session bus (the default). If it connects to the session bus, only command-mode processes running as the same user can communicate with it. If it connects to the system bus, any process can talk to it.

## Contributing

`go fmt` before committing.

### Vendoring

Tea-Service uses `github.com/rancher/trash` for managing vendor packages.

## Todo

- [x] Launch a servce
- [x] Connect to stdout/err of a service
- [x] Send a command (stdin) to a service
- [ ] Open an interactive session with a service
  - [ ] Handle multiple interactive sessions (only when line-buffered)
- [ ] Launch a service as a daemon
  - [ ] Manage pid file

## Issues

`teasvc` leaks. When clients connect, the resulting objects and file descriptors (in the server) are never cleaned up.

There may be other significant issues.