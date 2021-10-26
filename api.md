# api doc


## 0. status code

```go
	codeSuccess               = "80000"
	codeParamParseErr         = "80001"
	codeSymbolErr             = "80002"
	codeStafiAddressErr       = "80003"
	codeBlockHashErr          = "80004"
	codeTxHashErr             = "80005"
	codeSignatureErr          = "80006"
	codePubkeyErr             = "80007"
	codeInternalErr           = "80008"
	codePoolAddressErr        = "80009"
	codeTxDuplicateErr        = "80010"
	codeTokenPriceErr         = "80011"
	codeInAmountFormatErr     = "80012"
	codeMinOutAmountFormatErr = "80013"
	codePriceSlideErr         = "80014"
	codeMinLimitErr           = "80015"
	codeMaxLimitErr           = "80016"
	codeSwapInfoNotExistErr   = "80017"
	codeBundleIdNotExistErr   = "80018"
```



## 1. post swap info

### (1) description

*  post user swap info

### (2) path

* /feeStation/api/v1/station/swapInfo

### (3) request method

* post

### (4) request payload 

* data format: application/json
* data detail:

| field        | type   | notice                                                          |
| :----------- | :----- | :-------------------------------------------------------------- |
| stafiAddress | string | user stafi address, hex string with 0x prefix                   |
| symbol       | string | support: DOT KSM ATOM ETH                                       |
| blockHash    | string | block hash, hex string with 0x prefix                           |
| txHash       | string | tx hash, hex string with 0x prefix                              |
| poolAddress  | string | pool address, get from api                                      |
| signature    | string | signature, hex string with 0x prefix                            |
| pubkey       | string | pubkey, hex string with 0x prefix                               |
| inAmount     | string | in token amount, decimal string, decimals equal to native token |
| minOutAmount | string | min out amount, decimal string, decimals 12                     |

* native token decimals

DOT 10, KSM/FIS 12, ETH 18, ATOM 6


### (5) response
* include status、data、message fields
* status、message must be string format, data must be object

| grade 1 | grade 2 | grade 3 | type   | must exist? | encode type | description |
| :------ | :------ | :------ | :----- | :---------- | :---------- | :---------- |
| status  | N/A     | N/A     | string | Yes         | null        | status code |
| message | N/A     | N/A     | string | Yes         | null        | status info |
| data    | N/A     | N/A     | object | Yes         | null        | data        |
          
          
## 2. get pool info

### (1) description

*  get pool info

### (2) path

* /feeStation/api/v1/station/poolInfo

### (3) request method

* get

### (4) request payload 

* null
 
### (5) response
* include status、data、message fields
* status、message must be string format,data must be object

| grade 1 | grade 2      | grade 3     | type   | must exist? | encode type | description      |
| :------ | :----------- | :---------- | :----- | :---------- | :---------- | :--------------- |
| status  | N/A          | N/A         | string | Yes         | null        | status code      |
| message | N/A          | N/A         | string | Yes         | null        | status info      |
| data    | N/A          | N/A         | object | Yes         | null        | data             |
|         | poolInfoList | N/A         | list   | Yes         | null        | list             |
|         |              | symbol      | string | Yes         | null        | DOT KSM ATOM ETH |
|         |              | poolAddress | string | Yes         | null        | pool address     |
|         |              | swapRate    | string | Yes         | null        | decimals 6       |
|         | swapMaxLimit | N/A         | string | Yes         | null        | decimals 12      |
|         | swapMinLimit | N/A         | string | Yes         | null        | decimals 12      |


## 3. get swap info

### (1) description

*  get swap info

### (2) path

* /feeStation/api/v1/station/swapInfo

### (3) request method

* get

### (4) request param 

* `symbol`: support `DOT KSM ATOM ETH`
* `blockHash`: hex string with 0x prefix
* `txHash`: hex string with 0x prefix
 
### (5) response
* include status、data、message fields
* status、message must be string format,data must be object

| grade 1 | grade 2    | grade 3 | type   | must exist? | encode type | description |
| :------ | :--------- | :------ | :----- | :---------- | :---------- | :---------- |
| status  | N/A        | N/A     | string | Yes         | null        | status code |
| message | N/A        | N/A     | string | Yes         | null        | status info |
| data    | N/A        | N/A     | object | Yes         | null        | data        |
|         | swapStatus | N/A     | number | Yes         | null        | swap status |



* swap status detail

| swap status | description       |
| :---------- | :---------------- |
| 0           | VerifySigs        |
| 1           | VerifyTxOk        |
| 2           | PayOk             |
| 3           | BlockHashFailed   |
| 4           | TxHashFailed      |
| 5           | AmountFailed      |
| 6           | PubkeyFailed      |
| 7           | PoolAddressFailed |
| 8           | MemoFailed        |

