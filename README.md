* [Prerequisites](#prerequisites)
* [Install &amp; run](#install--run)
* [Usage](#usage)
* [Documents](#documents)

# das-pay

Build & run with [das-register](https://github.com/DeAccountSystems/das-register). Support CKB, TRX, BNB, ETH and Matic to pay the registration fee.

## Prerequisites

* Ubuntu 18.04 or newer
* MYSQL >= 8.0
* go version >= 1.15.0
* [CKB Node](https://github.com/nervosnetwork/ckb)
* [ETH Node](https://ethereum.org/en/community/support/#building-support)
* [BSC Node](https://docs.binance.org/smart-chain/developer/fullnode.html)
* [Tron Node](https://developers.tron.network/docs/fullnode)
* [das-database](https://github.com/DeAccountSystems/das-database)
* [das-register](https://github.com/DeAccountSystems/das-register)

## Install & run

```bash
# get the code
git clone https://github.com/DeAccountSystems/das-pay.git

# edit config/config.yaml and run das-database and init db of das-register before run das_pay_server

# compile and run
cd das-pay
make pay
./das_pay_server --config=config/config.yaml
```

## Docker Install & Run
```bash
# if you already have a mysql database installed, just run
docker run -dv $PWD/config/config.yaml:/app/config/config.yaml --name bit-pay-server slagga/bit-pay

# if not, you need docker-compose to automate the installation
docker-compose up -d
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
* [What is DAS](https://github.com/DeAccountSystems/das-contracts/blob/master/docs/en/Overview-of-DAS.md)
* [What is a DAS transaction on CKB](https://github.com/DeAccountSystems/das-contracts/blob/master/docs/en/Data-Structure-and-Protocol/Transaction-Structure.md)
