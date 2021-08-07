package substrate

import (
	"errors"
	scalecodec "github.com/itering/scale.go"
	"github.com/stafiprotocol/go-substrate-rpc-client/types"
)

const (
	ChainTypeStafi    = "stafi"
	ChainTypePolkadot = "polkadot"

	AddressTypeAccountId    = "AccountId"
	AddressTypeMultiAddress = "MultiAddress"
)

var (
	TerminatedError        = errors.New("terminated")
	BondEqualToUnbondError = errors.New("BondEqualToUnbondError")
)

type ChainEvent struct {
	ModuleId string                  `json:"module_id" `
	EventId  string                  `json:"event_id" `
	Params   []scalecodec.EventParam `json:"params"`
}

type Transaction struct {
	ExtrinsicHash  string
	CallModuleName string
	CallName       string
	Address        interface{}
	Params         []ExtrinsicParam
}

type ExtrinsicParam struct {
	Name  string      `json:"name"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}


type Receive struct {
	Recipient []byte
	Value     types.UCompact
}

type AccountInfo struct {
	Nonce     uint32
	Consumers uint32
	Providers uint32
	Data      struct {
		Free       types.U128
		Reserved   types.U128
		MiscFrozen types.U128
		FreeFrozen types.U128
	}
}

const (
	BalancesModuleId        = "Balances"
	TransferKeepAlive       = "transfer_keep_alive"
	Transfer                = "transfer"
	MethodTransferKeepAlive = "Balances.transfer_keep_alive"
	ConstExistentialDeposit = "ExistentialDeposit"

	StakingModuleId           = "Staking"
	StorageActiveEra          = "ActiveEra"
	StorageNominators         = "Nominators"
	StorageErasRewardPoints   = "ErasRewardPoints"
	StorageErasStakersClipped = "ErasStakersClipped"
	StorageEraNominated       = "EraNominated"
	StorageBonded             = "Bonded"
	StorageLedger             = "Ledger"
	MethodPayoutStakers       = "Staking.payout_stakers"
	MethodUnbond              = "Staking.unbond"
	MethodBondExtra           = "Staking.bond_extra"
	MethodWithdrawUnbonded    = "Staking.withdraw_unbonded"
	MethodNominate            = "Staking.nominate"

	MultisigModuleId        = "Multisig"
	NewMultisigEventId      = "NewMultisig"
	MultisigExecutedEventId = "MultisigExecuted"
	StorageMultisigs        = "Multisigs"
	MethodAsMulti           = "Multisig.as_multi"

	SystemModuleId = "System"
	StorageAccount = "Account"

	MethodBatch = "Utility.batch"

	ParamDest     = "dest"
	ParamDestType = "Address"

	ParamValue     = "value"
	ParamValueType = "Compact<Balance>"
)
