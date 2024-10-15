# DS-Assignment-1

## Part-1: 
(30 points) Consider some client-server architecture as follows. Several clients are registered
to the server. Periodically, each client sends message to the server. Upon receiving a
message, the server flips a coin and decides to either forward the message to all other
registered clients (excluding the original sender of the message) or drops the message
altogether. To solve this question, you will do the following:
1. Simulate the behaviour of both the server and the registered clients via GO routines.
(10 points)
2. Use Lamport’s logical clock to determine a total order of all the messages received at
all the registered clients. Subsequently, present (i.e., print) this order for all registered
clients to know the order in which the messages should be read. (10 points)
3. Use Vector clock to redo the assignment. Implement the detection of causality violation
and print any such detected causality violation. (10 points)
For all the points above, you should try your solution with at least 10 clients.

## Part-2
(50 points) Use Ring Protocol to implement a working version of replica synchronization. You
may assume that each replica maintains some data structure (that may diverge for arbitrary
reasons), which are periodically synchronized with the coordinator. The coordinator initiates
the synchronization process by sending message to all other machines. Upon receiving the
message from the coordinator, each machine updates its local version of the data structure
with the coordinator’s version. The coordinator, being an arbitrary machine in the network, is
subject to fault. Thus, a new coordinator is chosen by the Ring algorithm. You can assume a
fixed timeout to simulate the behaviour of detecting a fault. The objective is to have a
consensus across all machines (simulated by GO routines) in terms of the newly elected
coordinator.
1. Implement the above protocol of joint synchronization and election via GO (20 points)
2. While implementing your solution, try to simulate both the worst-case and the bestcase situation (i.e., all machines starting elections vs. only one machine starting
election). (10 points)
3. Consider the case where a GO routine fails during the election, yet the routine was
alive when the election started. (5 points + 5 points)
a. The newly elected coordinator fails while circulating the newly chosen
coordinator.
b. The failed node is not the newly elected coordinator.
4. An arbitrary node silently leaves the network (the departed node can be the coordinator
or non-coordinator). (10 points)
