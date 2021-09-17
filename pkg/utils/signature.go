package utils

import (
	"fmt"

	schnorrkel "github.com/ChainSafe/go-schnorrkel"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
)

func VerifySigsSecp256(sigs, pubkey, message []byte) bool {
	pubKey, err := btcec.ParsePubKey(pubkey, btcec.S256())
	if err != nil {
		return false
	}
	signature, err := btcec.ParseSignature(sigs, btcec.S256())
	if err != nil {
		return false
	}
	// Verify the signature for the message using the public key.
	messageHash := chainhash.DoubleHashB(message)
	return signature.Verify(messageHash, pubKey)
}

func VerifySigsEth(sigs, message []byte, address common.Address) bool {
	useSigs := make([]byte, 65)
	copy(useSigs, sigs)
	if useSigs[64] > 26 {
		useSigs[64] = useSigs[64] - 27
	}
	pubkey, err := ethCrypto.Ecrecover(ethCrypto.Keccak256(message), useSigs)
	if err != nil {
		return false
	}
	recoverAddress := common.BytesToAddress(ethCrypto.Keccak256(pubkey[1:])[12:])
	return recoverAddress == address
}

func VerifySigsEthPersonal(sigs []byte, message string, address common.Address) bool {
	useMessage := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	useSigs := make([]byte, 65)
	copy(useSigs, sigs)
	if useSigs[64] > 26 {
		useSigs[64] = useSigs[64] - 27
	}
	pubkey, err := ethCrypto.Ecrecover(ethCrypto.Keccak256([]byte(useMessage)), useSigs)
	if err != nil {
		return false
	}
	recoverAddress := common.BytesToAddress(ethCrypto.Keccak256(pubkey[1:])[12:])
	return recoverAddress == address
}

var substrateSigningCtx = []byte("substrate")


//only for pokadot web3.js
func VerifiySigsSr25519(sigs, pubkey, message []byte) bool {
	message = append(append([]byte("<Bytes>"), message...), []byte("</Bytes>")...)
	verifyTranscript := schnorrkel.NewSigningContext(substrateSigningCtx, message)
	var usePubkey [32]byte
	copy(usePubkey[:], pubkey)
	p, err := schnorrkel.NewPublicKey(usePubkey)
	if err != nil {
		return false
	}

	sigin := [64]byte{}
	copy(sigin[:], sigs)

	sig := &schnorrkel.Signature{}
	err = sig.Decode(sigin)
	if err != nil {
		return false
	}
	return p.Verify(sig, verifyTranscript)
}
