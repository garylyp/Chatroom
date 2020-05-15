# SimpleChat

A simple chatroom built using Go programming language that allows multiple users to communicate with each other via a server.


## Introduction

SimpleChat is a chatroom hosted on a webserver that provides realtime communication between members in the room. Through the use of websockets, multiple clients can conenct to single server that hosts the chatroom. Clients can then send messages into the chatroom via a terminal and they will be broadcasted instantly to all other clients by the server.

The purpose of this project is to gain a better understanding of development in Go as well as to have a refresher on the software engineering workflow (branching, committing, merging). This project was specifically chosen because I wish to understand deeper on setting up http connections between server and client, which is a fundamental component of many webapp and backend server applications.


## Installation and Usage

1. Start off by cloning the project. Make sure you are in the devbox environment. 
 
    HTTP
    ```
    git clone https://code.byted.org/yan.lim/simple-chat.git
    ```

    SSH
    ```
    git clone git@code.byted.org:yan.lim/simple-chat.git
    ```

2. First, we initialize the server. cd into the simple-chat directory. Then run the following command.
    ```
    go run ./cmd/server/main.go
    ```
    Your should see the following which indicates your server is running and listening for incoming connections.
    ```
    Server running on port: 9000
    ```

3. Next we run multiple clients to simulate the chatroom. Open two other separate terminals and cd into the simple-chat directory for each of them. Then run the following for each terminal
    ```
    go run ./cmd/client/main.go
    ```
    This will establish a websocket connection with the server. A random name starting with `u****` will be assigned to this chatroom user.

4. Enter some inputs into one of the client terminals and observe that the messages are received on other client terminals.

5. A new user can also view the last N messages upon entering the chatroom. Try this out by running another client on a new terminal (as in step 3). N is set to 5 by default, allowing a new user to view the last 5 messages in chronological order.


## Code Structure
The application adopts the following structure. 

The `cmd` folder contains the main entry point to the server and client application. 

The `internal` folder holds other code or data to be used to by main applications. Since this is a small project, it only contains the temporary storage file for storing messages from previous or current sessions of the chatroom.


```
├── cmd
│   ├── server
│   │   └── main.go
│   └── client
│       └── main.go
└── internal
    └── storage
        └── data.json
```

A json file is used for storage in this application. With every new message sent by any client, the json file will be rewritten. Although not very efficient for larger size chatrooms, this design is simple and allows data of messages to be easily accessed when a new user joins the chatroom.

## Tools

* [Golang Websocket Library](golang.org/x/net/websocket)

## Tests
To be updated
```

```


## Takeways
The main takeaway I had was understanding the project structure within a Go project, as well as understanding some of the key features of the Go language, such as concurrency, channels and goroutines. However, this project may be a little too small to help me appreciate how the different packages can interact with each other.



## References
* [Project Layout in Go](https://github.com/golang-standards/project-layout)
* [Creating Chat Application Using Websocket](https://medium.com/@johnshenk77/create-a-simple-chat-application-in-go-using-websocket-d2cb387db836)
* [Building websockets in Go](https://yalantis.com/blog/how-to-build-websockets-in-go/)


