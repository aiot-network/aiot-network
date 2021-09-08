# Aiot 命令行钱包说明文档

### 编译命令行钱包
```bash
git clone https://github.com/aiot-network/aiotchain.git

cd aiotchain

sh ./build.sh

cd ./bin/linux/wallet # linux 版
# cd ./bin/darwin/wallet # macOS 版
# cd ./bin/windows/wallet # Windows 版
```

### 配置文件说明
``` toml
#Format switch
Format = true

#testnet or mainnet
TestNet = false

#Directory for managing private key json files（default = "keystore/"）
KeystoreDir = ""

#RPC address
RpcIp = "127.0.0.1"

#RPC port
RpcPort = "23562"

#RPC TLS switch
RpcTLS = false

#If RpcTLS is on, you need to configure RpcCert
RpcCert = ""

#RPC username
RpcUser = ""

#RPC password
RpcPass = ""
```

### 命令行钱包使用

- #### 生成地址
``` bash
./wallet Create 111111
```

``` json
{
    "address": "AiS9RfVfjsb5tdE7xJY7m4kezTQbYnzCvfn",
    "crypto": {
        "cipher": "aes-128-cfb",
        "ciphertext": "0fe7cc9407eecc497435d662bacec66ccffb7adcfc064222850ab6011cd1ec7e45117eb9f2f7fea3d056bd48f46f24d7c055c60c63b1567e845659a618f6e8f3864a96647e1fd612713224a24fdc1043",
        "mnemonic_ciphertext": "b4e290390b01e7b116b163d6446aeef30379c87b1a306a441895bea9341838952988a536f27bffb22fc3b30f0dde4dafe6350f100128f05e24e3110527e836ebc2d1db817a20f6d6acfba160f5d2c93d9566a42a5dd0249254a274bae2c72e4201dfe758fce745758ff8a81449fb8f2c2ad144dfdb1a006bdd7dd1fe18afd8cd54a791678acf9f9bd62a661e488da851cc7003ac131db40d3c044c10673102ad85c6351d46258f0a6bfce7af9e9e261cc78fa7b8ebed1c3da1",
        "salt": "e3654619379b4c69fe1dfd231e9261bb232f555481df64193de9"
    },
    "p2pid": "16Uiu2HAkyRCbzMYjEr3L2LXCFnDJPMLkn46AhTaorCTCuRv4fZnJ"
}
```
- #### 通过助记词导入钱包
``` bash
./wallet MnemonicToAccount "inner board party plunge inmate trumpet orphan lunch stomach uphold priority protect struggle peace victory sand across lucky whisper bone champion adapt genius discover" 111111
```

``` json
{
    "address": "AiS9RfVfjsb5tdE7xJY7m4kezTQbYnzCvfn",
    "crypto": {
        "cipher": "aes-128-cfb",
        "ciphertext": "ea9720f2dd85576df1a842fde28fd6a3ab0955532a687bf499ca53595068b58ec5bcf144d76ce5068b03b73b0fd3be80f4d89f781b9d97140d28a7dafe71841b545ea2dfe02d2994c5380ddd2ff577bc",
        "mnemonic_ciphertext": "b8e44e6b0dba8b6ba1eb534d7928b5f57498f8644e382e1aca80a381892a8f450e86227c27bea76f6299d41873e3085ef085e6edfb67b83ca20a52feaa8cec0ee1c4578b38d209d0c1e2dbcd7d6124b5370b7048f4b5eb94b3b4da2c808d3261ce21cc870587716017f27db737774d571eddd25179bb7d2d0e833ebe47999a2c5aee9c071893dbe2795da9889ac0bc81a9a454c0a9034222e7f68107b352675314e47ad17ad7e32aeb4eb71fb650c0ad738eefc4e5267e4818",
        "salt": "5660e6aa55979a8ef6c4940c2911ef1d0e53159855bb29b9b653"
    },
    "p2pid": "16Uiu2HAkyRCbzMYjEr3L2LXCFnDJPMLkn46AhTaorCTCuRv4fZnJ"
}
```

- #### 导出钱包助记词/私钥
``` bash
./wallet DecryptPrivate AiS9RfVfjsb5tdE7xJY7m4kezTQbYnzCvfn 111111
```

``` json
{
    "private": "962bf2b471fe90060ea42fec920b8851d55f516e53b594463488503fe7ff5384",
    "mnemonic": "inner board party plunge inmate trumpet orphan lunch stomach uphold priority protect struggle peace victory sand across lucky whisper bone champion adapt genius discover"
}
```

- #### 创建派生地址（派生地址需要通过主地址来创建）
```bash
# DerivedAddresses {address} {start} {count} {password}; Generate derived addresses
# {address} 派生主地址
# {start}   起始 index
# {count}   生成地址数量
./wallet DerivedAddresses AiS9RfVfjsb5tdE7xJY7m4kezTQbYnzCvfn 0 5 111111
```
```json
[
    {
        "address": "AiLXhnvEPmz9GFwzfBBrg7RiW6EEzncKqUf",
        "index": 0
    },
    {
        "address": "AiV5QAAzcX3T8ZxqDKzpeoqw171sszmgnoF",
        "index": 1
    },
    {
        "address": "AifJJhhDK2JTQmoCLd76MwtoHRHN8b7dVbx",
        "index": 2
    },
    {
        "address": "AiU8KCKAuWyWxKV9KQwhCGat6LKBgLFYvUh",
        "index": 3
    },
    {
        "address": "AiUv3jvqncVZmDzRdQ4Ri3ZseGg1aRazkDq",
        "index": 4
    }
]
```

- #### 查看地址资产
``` bash
./wallet Account AiS9RfVfjsb5tdE7xJY7m4kezTQbYnzCvfn
```
``` json
{
    "address": "AiS9RfVfjsb5tdE7xJY7m4kezTQbYnzCvfn",
    "nonce": 4,
    "tokens": [
        {
            "address": "AIOT",
            "pledge": 0,
            "balance": 0,
            "locked": 0
        }
    ],
    "confirmed": 413459,
    "work": {
        "cycle": 0,
        "workload": 0,
        "end": 0
    }
}
```

- #### 查看地址列表
``` bash
# 只能查看主地址
./wallet ListAccounts
```
```json 
[
    "AiS9RfVfjsb5tdE7xJY7m4kezTQbYnzCvfn",
]
```

- #### 转账
``` bash
# SendTransaction {from} {token} {to:amount|to:amount} {fees} {password} {nonce}; Send a transaction;
# {token} AIOT 为主币 二层代币是需要替换成对应的合约地址
# {fees}  默认输入 0.0001
# {nonce} 可选 可以通过 Account 命令获取
./wallet SendTransaction AiS9RfVfjsb5tdE7xJY7m4kezTQbYnzCvfn AIOT AiLXhnvEPmz9GFwzfBBrg7RiW6EEzncKqUf:1000 0.0001 111111
```

- #### 派生地址转账
``` bash
# SendDerivedTransaction {from} {index} {token} {to:amount|to:amount} {fees} {password} {nonce}; Send a transaction;
# {index} 指定转出的派生地址
# {token} AIOT 为主币 二层代币是需要替换成对应的合约地址
# {fees}  默认输入 0.0001
# {nonce} 可选 可以通过 Account 命令获取
./wallet SendDerivedTransaction AiS9RfVfjsb5tdE7xJY7m4kezTQbYnzCvfn 1 AIOT AiUv3jvqncVZmDzRdQ4Ri3ZseGg1aRazkDq:1000 0.0001 111111
```