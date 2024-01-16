# validator-exporter
Prometheus exporter for cosmos validators metrics.

## Usage
```
Usage of validator-exporter:
  -log-level value
        Set log level
  -p int
        Server port (default 8008)

```
### Supported env vars
```
"GRPC_ADDR" envDefault:"grpc.constantine.archway.tech:443"
"GRPC_TLS_ENABLED" envDefault:"true"
"GRPC_TIMEOUT_SECONDS" envDefault:"5"
"PREFIX" envDefault:"archway"
"CHAIN_NAME" envDefault:"archway"
"CHAIN_ID" envDefault:"constantine-3"
```
### Connecting to archway constantine testnet by default
```
validator-exporter
```
### Connecting to archway mainnet
```
GRPC_ADDR=grpc.mainnet.archway.io:443 validator-exporter
```
### Connecting to localnet
```
GRPC_TLS_ENABLED=false GRPC_ADDR=localhost:9090 validator-exporter --log-level debug
```

## Metrics
```
# HELP cosmos_validator_missed_blocks Returns missed blocks for a validator.
# TYPE cosmos_validator_missed_blocks gauge
cosmos_validator_missed_blocks{bond_status="bonded",chain_id="localnet",chain_name="archway",jailed="false",moniker="validator_1",tombstoned="false",valcons="archwayvalcons1gzfu5aqsqsljmgs5m3eyklwq9ufrpmlfhxhxjd",valoper="archwayvaloper172zqrqtrwfplwhec44050dhuv66ekcmty4hnfv"} 0
cosmos_validator_missed_blocks{bond_status="bonded",chain_id="localnet",chain_name="archway",jailed="false",moniker="validator_2",tombstoned="false",valcons="archwayvalcons1yf253kfa2g3qk2727nltndzll546wrrtdr5d3x",valoper="archwayvaloper1370vgzkv5l3kylcylwekzjcdt2hjk2k8zrht6c"} 0
cosmos_validator_missed_blocks{bond_status="bonded",chain_id="localnet",chain_name="archway",jailed="false",moniker="validator_3",tombstoned="false",valcons="archwayvalcons1ve4g3nth72p6jq92d5cfk9kq0j48jr0m5h7wk5",valoper="archwayvaloper1scxt3mgxmw3z2hpf8k4mlssz5qvljmtaplv6nz"} 2
```
