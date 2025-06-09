
# Blockchain Topic - Trần Tấn Thành

## Requirements
* [Go v1.24+](https://go.dev/dl/) or newer.
* [Docker](https://docs.docker.com/desktop/setup/install/windows-install/)
* [Protocol buffer compiler](https://github.com/protocolbuffers/protobuf/releases) (Optional: For Regenerate gRPC code)


## Setup
* Install [Go v1.24+](https://go.dev/dl/)

* Install **Go Task** (a very simple library that allows you to write simple “task” scripts in Go and run)
    ```bash
    go install github.com/go-task/task/v3/cmd/task@latest
    ```

* Install necessary libs with **Go Task**
    ```bash
    task install
    ```
    or install without **Go Task**
    ```bash
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    go mod download
    ```

* Install: [Protocol buffer compiler](https://github.com/protocolbuffers/protobuf/releases) ([Reference](https://grpc.io/docs/protoc-installation/))

* Unzip the package and add the `bin` directory to the system `PATH` environment variable.

* Install [Docker](https://docs.docker.com/desktop/setup/install/windows-install/)

## Start Server

* `Start` Docker 

* `Build` docker service
    ```bash
    docker-compose build
    ```

* `Run` docker service
    ```bash
    docker-compose up -d
    ```
    or
* `Build` and `Run` docker service
     ```bash
    docker-compose up -d --build
    ```

## Interact with system - CLI 
### **Client**
* **Create User** ( Create a new user and save information in the `wallet.json` )
    ```bash
    go run ./cmd/cli/main.go create-user --name <username>
    ```

* **Send transaction**  ( Send transaction from client to server )
    ```bash
    go run ./cmd/cli/main.go send-transaction --sender <sender-address> --receiver <recevier-address> --amount <amount>
    ```
    - Optional: --node `<node-target>` ( localhost:`50051`, localhost:`50052` , localhost:`50053` )

* **Get block**
    ```bash
    go run ./cmd/cli/main.go get-block --block-height <block-height>
    ```
    - Optional: --node `<node-target>` ( localhost:`50051`, localhost:`50052` , localhost:`50053` )

* **Get Current Block Height**
    ```bash
    go run ./cmd/cli/main.go get-current-block-height
    ```
    - Optional: --node `<node-target>` ( localhost:`50051`, localhost:`50052` , localhost:`50053` )

* **Monitor All Nodes Status** ( Monitor all running `node statuses` in `real time` within the containers )
    ```bash
    go run .\cmd\cli\main.go monitor-node
    ```

    > **Note**: Default `node-target` is `localhost:50051`

## System Architecture
* **Decisions**
    - **Use `gRPC` to communicate between services**
        + High performance
        + Easy to define message with protobuf

    - **Create a new block every 5 seconds if there is at least one pending transaction**
        + Reduces system resource consumption
        + Minimizes unnecessary memory and storage usage
        + Improves node synchronization time

    - **Use `Base58Check` to encode and decode `PrivateKey` and `PublicKey`**
        + Removes confusing characters (like 0, O, I, l).
        + Easier to read and type.
        + Adds a checksum to detect typing or data errors.

    - **`Sign` Transaction in `Client` with `Private Key` and `Verify` Transaction in `Server` with `Public Key`**
        + `Private key` is `stored only` on the `client side`, keeping it safe even if the server is compromised.
        + Reduces security risk and processing load on the server since it doesn’t manage private keys or handle signing.

    - **Self-implement the basic Merkle tree algorithm**

    - **Implement a mempool to temporarily store pending transactions**

* **Using libraries**
    + [syndtr/goleveldb](github.com/syndtr/goleveldb): Easy interacting with the `LevelDB` database in `golang`
    + [google.golang.org/grpc](google.golang.org/grpc): A high-performance, open-source universal `RPC framework`
    + [google.golang.org/protobuf](google.golang.org/protobuf): Official `Protocol Buffers` for `Go`

* **Folder Struture**
    * `pkg/blockchain`: Contains definitions for Block, Transaction, and logic for generating hashes and the Merkle Tree.
    * `pkg/wallet`: Contains logic for creating and managing ECDSA key pairs, signing, and signature verification.
    * `pkg/p2p`: Handles communication between nodes (via gRPC or HTTP), including transaction broadcasting and block proposal/voting.
    * `pkg/consensus`: Implements the consensus mechanism (e.g., leader election, voting).
    * `pkg/storage`: Interacts with LevelDB.

    * `*pkg/util`: Common helper functions.
    * `*pkg/types`: Shared type definitions.
    * `*pkg/node`: Core node logic.
    * `*pkg/config`: Configuration management.

    * `cmd/cli`: Handles CLI commands.
    * `cmd/node`: Entry point for each validator node.
