# fee-station


## how to use

```sh
make build
# after config conf_station.toml
./build/stationd -C ./conf_station.toml
# after config conf_checker.toml
./build/checkerd -C ./conf_checker.toml
# after config conf_payer.toml
./build/payerd -C ./conf_payer.toml
```