# TTK4145 - Elevator project
## Description 
In this repository you can find our implementation of the elevator project. For assigning elevator orders we have a primary, which can be reconfigured if necessary, i.e. the execution of orders is based on a master-slave structure. Furthermore, alive messages will be broadcasted so that the other nodes know which elevators are connected to the network at all times. If an elevator cannot connect to the network, it will work independently of the other elevator, in a "single elevator" mode. 

The repository contains the following packages:
- Project types
- Connection and communication
- Alive observation
- Primary execution
- Primary reconfiguration
- Hall request assigner
- Driver functions
- Process pair

## How to run project code

Run (specify the correct path for the hall request assigner executable!):

`go run main.go --cport 15657 --cport 10001 --hra="/home/student/Project-ressources/cost_fns/hall_request_assigner/hall_request_assigner"`

If multiple instances on same PC (i.e. using Simulators), then
- use flag `--eport` to specify the port an other simulator/elevator-server is running on and
- use flag `--cport` to specify the communication port the system shall use. The delta of the communicationports between several instances on the same PC must be at least 4.
See file `main.go` for all available flags.

### Disclaimer ;)

it creates a process pair, so you need to be fast and close both tabs quick after each other to stop everything.
