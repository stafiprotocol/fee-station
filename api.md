# api doc

## 1.post swap info

### (1) description

*  post user swap info

### (2) path

* /feeStation/api/v1/station/swapInfo

### (3) request method

* post

### (4) request payload 

* data format: application/json
* data detail:

| field        | type   | notice                                        |
| :----------- | :----- | :-------------------------------------------- |
| stafiAddress | string | user stafi address, hex string with 0x prefix |
| symbol       | string | support: DOT KSM ATOM ETH                     |
| blockHash    | string | block hash, hex string with 0x prefix         |
| txHash       | string | tx hash, hex string with 0x prefix            |
| poolAddress  | string | pool address, get from api                    |
| signature    | string | signature, hex string with 0x prefix          |
| pubkey       | string | pubkey, hex string with 0x prefix             |
| inAmount     | string | in token amount, decimal string               |
| minOutAmount | string | min out amount, decimal string                |



### (5) response
* include status、data、message fields
* status、message must be string format,data must be object

| grade 1 | grade 2 | grade 3 | type   | must exist? | encode type | description |
| :------ | :------ | :------ | :----- | :---------- | :---------- | :---------- |
| status  | N/A     | N/A     | string | Yes         | null        | status code |
| message | N/A     | N/A     | string | Yes         | null        | status info |
| data    | N/A     | N/A     | object | Yes         | null        | data        |
          
          
