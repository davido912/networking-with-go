## Chapter 4 notes (page 115)

### Connection handling

#### Initialisation
composed of a three step handshake between client and server (SYN/ACKs) which in the end results in a successful
connection establishment.

#### Communication
client communicates receive buffer memory space to server, server then knows its limit

#### Termination
FIN signal that leads to eventual closure (either side)

### Utilities 
The following tools can be used to monitor networking in programming:
* https://www.wireshark.org (book Practical Packet Analysis). can help in inspecting TCP packets


### Good to knows
* -race flag in testing can help detect data races
* ICMP echo request and response can check availability of a service - alternatively a TCP connection can be sufficient but 
is with overhead that should questioned