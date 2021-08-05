// Copyright 2021 stafiprotocol
// SPDX-License-Identifier: LGPL-3.0-only

package distributor

import (
	"bytes"
	dao_user "drop/dao/user"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/crypto/sha3"
)

var (
	uint256Type, _ = abi.NewType("uint256", "", nil)
	addressType, _ = abi.NewType("address", "", nil)
	arguments0     = abi.Arguments{{Type: uint256Type}}
	arguments1     = abi.Arguments{{Type: addressType}}
	arguments2     = abi.Arguments{{Type: uint256Type}}
)

func ToContent(index *big.Int, account string, amount *big.Int) []byte {
	h := sha3.NewLegacyKeccak256()
	bytes0, _ := arguments0.Pack(index)
	bytes1, _ := arguments1.Pack(common.HexToAddress(account))
	bytes1New := bytes1[12:32]
	bytes2, _ := arguments2.Pack(amount)

	bytes0 = append(bytes0, bytes1New...)
	bytes0 = append(bytes0, bytes2...)
	var buf []byte
	h.Write(bytes0)
	buf = h.Sum(buf)
	return buf
}

func GetRootHash(stats []*dao_user.Snapshot) string {
	bts := make(BufferList, len(stats))
	for i, data := range stats {
		// rewardDecimal := decimal.NewFromFloat(data.DropAmount)
		// rewardDecimal = rewardDecimal.Mul(decimal.NewFromInt(1e18))

		// rewardDecimalStr := rewardDecimal.String()
		// if strings.Contains(rewardDecimalStr, ".") {
		// 	strs := strings.Split(rewardDecimalStr, ".")
		// 	rewardDecimalStr = strs[0]
		// }
		rewardDecimalStrBigInt, _ := new(big.Int).SetString(data.DropAmount, 10)

		bts[i] = *bytes.NewBuffer(ToContent(big.NewInt(int64(i)), data.UserAddress, rewardDecimalStrBigInt))
	}
	var mt MerkleTree
	mt.BuildMerkleTree(&bts)
	data, _ := mt.GetHexRoot()
	return data
}

func GetMerkleTree(stats []*dao_user.Snapshot) *MerkleTree {
	bts := make(BufferList, len(stats))
	for i, data := range stats {
		// rewardDecimal := decimal.NewFromFloat(data.Amount)
		// rewardDecimal = rewardDecimal.Mul(decimal.NewFromInt(1e18))

		// rewardDecimalStr := rewardDecimal.String()
		// if strings.Contains(rewardDecimalStr, ".") {
		// 	strs := strings.Split(rewardDecimalStr, ".")
		// 	rewardDecimalStr = strs[0]
		// }
		rewardDecimalStrBigInt, _ := new(big.Int).SetString(data.DropAmount, 10)

		bts[i] = *bytes.NewBuffer(ToContent(big.NewInt(int64(i)), data.UserAddress, rewardDecimalStrBigInt))
	}
	var mt MerkleTree
	mt.BuildMerkleTree(&bts)
	return &mt
}
