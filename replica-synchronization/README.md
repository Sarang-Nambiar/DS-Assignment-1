# Replica-Synchronization

## How to run

Ensure you are using a Windows environment to run the code. After setting up your environment, simply run the already built executable file located in the root directory.

To run the program, execute the following command in PowerShell:

```powershell
./replica-synchronization
```

This will begin the execution of the program. The program will continue to run until you manually terminate it.

When you first run the executable, it will start a terminal with the server running as the coordinator node, since no coordinators exist in the network initially. Subsequent executions of the program in separate terminals will create new client nodes and connect them to the existing coordinator.

It is recommended to run the program in 3-4 separate terminals to see the synchronization of the nodes in the network.

Note: The node join feature only works for when the coordinator is of Id 0. If any new coordinator has been elected, the node join feature will not work as intended.

### Sample Output

<!-- Sample output here -->

---

## How to Interpret the Output

Each terminal belongs to a node in the network and will only display the events that occur for that particular node, whether it be a coordinator or a client. The naming convention follows the format that was previously established in Lamport's Clock and the Vector Clock program, i.e. `[<Node Type> - <Node ID>] <Event>`

The replica synchronization is programmed to occur every 5 seconds. Within this 5 seconds, the nodes will periodically modify their local replica copy to simulate a more real world scenario.

### Sample Output

<!-- Sample output here -->

## Election implementation logic

Initially, when the first node joins the network, it is automatically assigned to be the coordinator and subsequent nodes are registered to this coordinator node. The coordinator then starts replica synchronization procedure. However, if the coordinator node fails, a timeout condition is triggered amongst the client nodes and an election process is initiated to elect a new coordinator.

<!-- Insert image of the timeout trigger here -->

The election process is implemented based on ring election protocol.There are two phases to the election:
1. **Discovery Phase:** The client node that triggered the election will start to create a new ring with the client nodes that are still alive. Simutaneously, it will compare the client ID of the nodes in the ring to determine the highest client ID. The client node with the highest client ID will be elected as the new coordinator.
2. **Announcement Phase:**The newly elected coordinator will then circulate the new coordinator ID and the ring structure to all the client nodes in the network. 

## 2. How to simulate worst case and best case scenarios for election

### (a) Worst case scenario:

To simulate the worst case scenario where all the clients simulate the election process simulataneosly, Uncomment line 313 in 'node/client.go' file. This will make sure that the election process is triggered by all client nodes simulataneosly.

Expected output: The client nodes should still be able to elect a new coordinator and continue the replica synchronization process. There are fail safes in place to avoid multiple coordinators of the same client ID from being elected as shown below:

<!-- insert picture of the fail safe -->

### (b) Best case scenario:

To simulate the best case scenario where only one client node triggers the election process, we just need to run the unmodified program as the code is already designed to mimic this scenario.

## 3. How to simulate GO routine fails during election 

### (a) If the newly elected coordinator fails while circulating the newly chosen coordinator

To simulate the scenario where the newly elected coordinator fails while circulating the new coordinator ID and ring structure, uncomment lines 282 or 290 in 'node/client.go' file. This will simulate the failure of the newly elected coordinator in two different cases:
- If the newly elected coordinator fails before circulating the new coordinator ID
- If the newly elected coordinator fails after circulating the new coordinator ID

Expected output: Regardless of when the newly elected coordinator fails, the client nodes after a while will detect the failure of the coordinator due to the timeout condition and will trigger the election process to elect a new coordinator.

### (b) If the failed node is not the newly elected coordinator

To simulate the scenario where a node fails during the election process but it is not the newly elected coordinator, uncomment line 130 in 'node/client.go' file. This will simulate the failure of a node during the election process.

Expected output: The client nodes will detect the failure of the node and just skip the failed node and move onto its successor and so on. However, during this stage, the client nodes will still be updated with the ring structure containing the dead node and the new coordinator ID. The new structure will only be circulated once the coordinator fails and a new discovery phase is initiated.

<!-- insert picture of the skipping successor -->

## 4. How to simulate arbitrary node silently leaving the network(coordinator or non coordinator)

To simulate the scenario where a node silently leaves the network, terminate one of the powershell windows where the node/coordinator is running.

Expected output: If the node that leaves suddenly is a coordinator, then an election is begun once the timeout event is triggered. Or else, if the node is a client, then the client nodes will not be able to detect the failure of the node and the replica synchronization process will continue as normal. It will only be detected in the next election process.