# aiot-network

### How to build

####  Prerequisites

- Update Go to version at least 1.15  (required >= **1.15**)

Check your golang version

```bash
~ go version
go version go1.15 darwin/amd64
```

```bash
cd aiot-network/cmd/chain
go build

cd aiot-network/cmd/wallet
go build
```

#### How to use


##### Copy configuration file for reconfiguration

```bash
 cp config.toml.example config.toml
```

##### Modify configuration file

* set RpcPass
* set RpcUser
* set ExternalIp

##### Start the futuremine

```bash

./chain --config config.toml
```

##### Copy wallet configuration file for reconfiguration

```
cd cmd/wallet
cp wallet.toml.example wallet.toml
```

##### Modify wallet configuration file

* set RpcIp
* set RpcUser
* set RpcPass
* If the node has the RpcTLS switch turned on, you need to configure the node's server.pem path to RpcCert and set RpcTLS in wallet.config to true

##### Use wallet

```bash
./wallet --help
```
##### Create an account or set password at create

```bash
./wallet Create 

./wallet Create 123456
```
##### Send transaction

SendTransaction {from} {to} {contract} {amount} {fee} [password] [nonce]

```bash
./wallet SendTransaction xCHiGPLCzgnrdTqjKABXZteAGVJu3jXLjnQ xCE9boXz2TxSE9srVPDdfszyiXtfT3vduc8 AIOT 10 0.1

./wallet SendTransaction xCHiGPLCzgnrdTqjKABXZteAGVJu3jXLjnQ xCE9boXz2TxSE9srVPDdfszyiXtfT3vduc8 AIOT 10 0.1 123456
```

##### Get account balance
Account {address}
```bash
./wallet Account xCHiGPLCzgnrdTqjKABXZteAGVJu3jXLjnQ
```
