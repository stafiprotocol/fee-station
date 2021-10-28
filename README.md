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

increase pubkey length
```sql
alter table fee_station_native_chain_txes modify column  sender_pubkey  varchar(560);
alter table swap_infos modify column  pubkey  varchar(560);
alter table fee_station_bundle_addresses modify column  pubkey  varchar(560);

```