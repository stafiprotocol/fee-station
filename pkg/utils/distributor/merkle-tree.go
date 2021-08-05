// Copyright 2021 stafiprotocol
// SPDX-License-Identifier: LGPL-3.0-only

package distributor

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sort"

	"golang.org/x/crypto/sha3"
)

//叶子节点为layers[0] root为layers[len(layers)-1]
type MerkleTree struct {
	layers        [][]*Node
	leafNodeIndex map[string]int //叶子节点索引
}

type Node struct {
	Hash   bytes.Buffer
	Parent *Node
	left   *Node
	right  *Node
}

func (n *Node) String() string {
	return hex.EncodeToString(n.Hash.Bytes())
}

type BufferList []bytes.Buffer

func (Bs BufferList) Len() int           { return len(Bs) }
func (Bs BufferList) Less(i, j int) bool { return bytes.Compare(Bs[i].Bytes(), Bs[j].Bytes()) < 0 }
func (Bs BufferList) Swap(i, j int)      { Bs[i], Bs[j] = Bs[j], Bs[i] }

func (m *MerkleTree) GetLayers() [][]*Node {
	return m.layers
}

func (m *MerkleTree) BuildMerkleTree(contents *BufferList) {
	//从小到达排序
	sort.Sort(contents)
	m.layers = make([][]*Node, int64(math.Ceil(float64(contents.Len())/2)+1))
	//先将叶子节点放到layer[0]
	m.buildLeafNodes(*contents)

	realHeight := 0
	for i := 0; i < len(m.layers)-1; i++ {
		layer := make([]*Node, int64(math.Ceil(float64(len(m.layers[i]))/2)))
		for j := 0; j < len(m.layers[i]); j = j + 2 {
			if j+1 < len(m.layers[i]) {
				cHash := ConbinedHash(m.layers[i][j].Hash.Bytes(), m.layers[i][j+1].Hash.Bytes())
				node := Node{
					Hash:   cHash,
					Parent: nil,
					left:   m.layers[i][j],
					right:  m.layers[i][j+1],
				}
				layer[j/2] = &node
				m.layers[i][j].Parent = &node
				m.layers[i][j+1].Parent = &node
			} else {
				layer[j/2] = m.layers[i][j]
			}
		}
		m.layers[i+1] = layer
		if len(layer) == 1 {
			realHeight = i + 1
			break
		}
	}
	m.layers = m.layers[0 : realHeight+1]
}

func (m *MerkleTree) GetRootHash() (hash bytes.Buffer, err error) {
	if (len(m.layers[len(m.layers)-1])) != 1 {
		err = errors.New("invalidate tree")
	}
	hash = m.layers[len(m.layers)-1][0].Hash
	return
}

func (m *MerkleTree) GetHexRoot() (hexHash string, err error) {
	if (len(m.layers[len(m.layers)-1])) != 1 {
		err = errors.New("invalidate tree")
	}
	hexHash = hex.EncodeToString(m.layers[len(m.layers)-1][0].Hash.Bytes())
	return
}

func (m *MerkleTree) buildLeafNodes(bs BufferList) {
	m.leafNodeIndex = make(map[string]int)
	m.layers[0] = make([]*Node, bs.Len())
	for i, data := range bs {
		node := Node{
			data,
			nil,
			nil,
			nil,
		}
		m.layers[0][i] = &node
		m.leafNodeIndex[hex.EncodeToString(data.Bytes())] = i
	}
}

func (m *MerkleTree) GetProof(leafNodeBuffer bytes.Buffer) ([]bytes.Buffer, error) {
	proof := make([]bytes.Buffer, 0)
	if index, ok := m.leafNodeIndex[hex.EncodeToString(leafNodeBuffer.Bytes())]; ok {

		for i := 0; i < len(m.layers)-1; i++ {
			node, err := m.getPairElement(index, i)
			if err != nil {
				index = index / 2
				continue
			}
			proof = append(proof, node.Hash)
			index = index / 2
		}

		return proof, nil

	} else {
		return nil, errors.New("leafnode not exist")
	}
}

func VerifyProof(leafNode bytes.Buffer, proof []bytes.Buffer, root bytes.Buffer) bool {
	result := leafNode
	for _, p := range proof {
		result = ConbinedHash(result.Bytes(), p.Bytes())
	}
	return bytes.EqualFold(result.Bytes(), root.Bytes())
}

func (m *MerkleTree) getPairElement(index, layer int) (*Node, error) {
	willUseIndex := 0
	if index%2 == 0 {
		willUseIndex = index + 1
	} else {
		willUseIndex = index - 1
	}
	if willUseIndex <= len(m.layers[layer])-1 {
		return m.layers[layer][willUseIndex], nil
	} else {
		return nil, fmt.Errorf("no pair index %d ,layer %d", index, layer)
	}
}

func ConbinedHash(b0, b1 []byte) bytes.Buffer {
	i := bytes.Compare(b0, b1)
	var bts []byte
	if i == -1 { //a<b
		bts = append(bts, b0...)
		bts = append(bts, b1...)
	} else {
		bts = append(bts, b1...)
		bts = append(bts, b0...)
	}
	h := sha3.NewLegacyKeccak256()
	h.Write(bts)
	return *bytes.NewBuffer(h.Sum(nil))
}
