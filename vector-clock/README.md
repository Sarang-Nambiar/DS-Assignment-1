# Vector Clock

## How to Run

Ensure you are using a Windows environment to run the code. After setting up your environment, simply run the already built executable file located in the root directory.

To run the program, execute the following command in PowerShell:

```powershell
./vector-clock
```

This will begin the execution of the program. The program will continue to run until you manually terminate it.

### Sample Output

A sample output of the execution is shown below:

![Screenshot 2024-10-25 113625](https://github.com/user-attachments/assets/ca933e0a-bd94-477f-9450-a3c5857127f4)

---

## How to Interpret the Output

The CLI output is divided into two sections: **Node Info** and the **Actual Message**. The **Node Info** section represents information about the node that printed the output, while the **Actual Message** section describes the event executed by the client or server.

### Node Info

![image](https://github.com/user-attachments/assets/466ea3b7-9c23-41e1-bb69-1fdd39129242)

The **Node Info** contains three parts separated by a dash(-):

1. **Node Type**: Indicates whether the printed output is from the client or server.
2. **Node ID**: The identifier of the node. The ID of the clients go from 0 to n - 1 where n is the number of clients
3. **Vector Clock Value**: The current vector clock value of that node. The vector clock is array of individual clock values for all the nodes in the network and an additional column for the vector clock of the server. The vector clock values are incremented when 

### Actual Message

![Screenshot 2024-10-25 114517](https://github.com/user-attachments/assets/cea183b1-e9c0-4016-84a7-111d5a764b52)

The **Actual Message** part describes the specific event executed by the client or server.

---

### Vector Clock Increment Logic: 

#### For Clients:

The vector clock is incremented whenever the client sends or receives a message.

#### For Server:

The vector clock is incremented for multiple steps here. It is incremented for the server whenever the server receieves a message, forwards a message or whenever a message is dropped.
