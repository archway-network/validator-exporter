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
cosmos_validator_missed_blocks{moniker="validator_1",valcons="archwayvalcons18le5pevj6sdynyksn77n9z9g8394l3xqk04s3z",valoper="archwayvaloper172zqrqtrwfplwhec44050dhuv66ekcmty4hnfv"} 0
cosmos_validator_missed_blocks{moniker="validator_2",valcons="archwayvalcons1z4q9zpe8l8puwv8aq4dqadkz4zm244pnu72qcd",valoper="archwayvaloper1370vgzkv5l3kylcylwekzjcdt2hjk2k8zrht6c"} 0
cosmos_validator_missed_blocks{moniker="validator_3",valcons="archwayvalcons1ep8hnygqw8gvsdfvyanhcfsmvlrvae4s9hljta",valoper="archwayvaloper1scxt3mgxmw3z2hpf8k4mlssz5qvljmtaplv6nz"} 2
```
