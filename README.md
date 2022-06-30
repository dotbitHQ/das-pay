* [Prerequisites](#prerequisites)
* [Install &amp; Run](#install--run)
     * [Source Compile](#source-compile)
     * [Docker](#docker)
* [Usage](#usage)
* [Documents](#documents)

# das-pay

Build & run with [das-register](https://github.com/dotbitHQ/das-register). Support CKB, TRX, BNB, ETH and Matic to pay the registration fee.

## Prerequisites

* Ubuntu 18.04 or newer
* MYSQL >= 8.0
* go version >= 1.15.0
* [CKB Node](https://github.com/nervosnetwork/ckb)
* [ETH Node](https://ethereum.org/en/community/support/#building-support)
* [BSC Node](https://docs.binance.org/smart-chain/developer/fullnode.html)
* [Tron Node](https://developers.tron.network/docs/fullnode)
* [das-database](https://github.com/dotbitHQ/das-database)
* [das-register](https://github.com/dotbitHQ/das-register)

## Install & Run

### Source Compile

```bash
# get the code
git clone https://github.com/dotbitHQ/das-pay.git

# edit config/config.yaml and run das-database and init db of das-register before run das_pay_server

# compile and run
cd das-pay
make pay
./das_pay_server --config=config/config.yaml
```

### Docker
* docker >= 20.10
* docker-compose >= 2.2.2

```bash
sudo curl -L "https://github.com/docker/compose/releases/download/v2.2.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
sudo ln -s /usr/local/bin/docker-compose /usr/bin/docker-compose
docker-compose up -d
```

_if you already have a mysql installed, just run_
```bash
docker run -dv $PWD/config/config.yaml:/app/config/config.yaml --name das-pay-server slagga/das-pay
```


## Usage
Set the gateway address of each chain in `conf/config.yaml` ( `private`, the private key of address, is for refund). The whole workflow as below:


```
                ┌───────┐
                │ start │
                └───┬───┘
                    │
                    │
                    │
        ┌───────────▼───────────┐
┌───────┤       sync block?     │◄─┬───┐
│       │ CKB/ETH/BSC/TRX/MATIC │  │   │
│       └───────────┬───────────┘  │   │
│                   │              │   │
│                   Y              │   │
│                   │              │   │
│         ┌─────────▼──────────┐   │   │
│         │ parse txs in block │   │   │
│         └─────────┬──────────┘   │   │
│                   │              │   │
│                   │              │   │
│                   │              N   │
N                   │              │   │
│         ┌─────────▼──────────┐   │   │
│         │ pay for the order? ├───┘   │
│         └─────────┬──────────┘       │
│                   │                  │
│                   Y                  │
│                   │                  │
│                   │                  │
│       ┌───────────▼─────────────┐    │
│       │ update the order status ├────┘
│       └─────────────────────────┘
│
│
│
│
│               ┌───────┐
└──────────────►│  end  │
                └───────┘

```
## Documents
* [What is DAS](https://github.com/dotbitHQ/das-contracts/blob/master/docs/en/Overview-of-DAS.md)
* [What is a DAS transaction on CKB](https://github.com/dotbitHQ/das-contracts/blob/master/docs/en/Data-Structure-and-Protocol/Transaction-Structure.md)
