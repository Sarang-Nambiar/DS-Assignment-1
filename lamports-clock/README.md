# Lamport's Clock

## How to Run

Ensure you are using a Windows environment to run the code. After setting up your environment, simply run the already built executable file located in the root directory.

To run the program, execute the following command in PowerShell:

```powershell
./lamports-clock
```

This will begin the execution of the program. The program will continue to run until you manually terminate it.

### Sample Output

A sample output of the execution is shown below:

![Screenshot 2024-10-25 111203](https://github.com/user-attachments/assets/7694b289-a62d-4d77-b9ae-4456446c030a)

---

## How to Interpret the Output

The CLI output is divided into two sections: **Node Info** and the **Actual Message**. The **Node Info** section represents information about the node that printed the output, while the **Actual Message** section describes the event executed by the client or server.

### Node Info

![Screenshot 2024-10-25 111050](https://github.com/user-attachments/assets/813fc78a-b22a-48d1-bc13-8516bb4174b8)

The **Node Info** contains three parts separated by a dash(-):

1. **Node Type**: Indicates whether the printed output is from the client or server.
2. **Node ID**: The identifier of the node. The ID of the clients go from 0 to n - 1 where n is the number of clients
3. **Logical Clock Value**: The current logical clock value of that node.

### Actual Message

![Screenshot 2024-10-25 114358](https://github.com/user-attachments/assets/ab081816-1e3d-4871-b62c-9dd319fd7c25)

The **Actual Message** part describes the specific event executed by the client or server.

---

### Logical Clock Increment Logic: 

#### For Clients:

The logical clock is incremented whenever the client sends or receives a message.

#### For Server:

The logical clock is incremented for multiple steps here. It is incremented for the server whenever the server receieves a message, forwards a message or whenever a message is dropped.
