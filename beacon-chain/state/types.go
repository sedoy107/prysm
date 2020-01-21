package state

import (
	"sync"

	pbp2p "github.com/prysmaticlabs/prysm/proto/beacon/p2p/v1"
	"github.com/prysmaticlabs/prysm/shared/bytesutil"
	"github.com/prysmaticlabs/prysm/shared/hashutil"
	"github.com/prysmaticlabs/prysm/shared/stateutil"
)

type BeaconState struct {
	state        *pbp2p.BeaconState
	lock         sync.RWMutex
	merkleLayers [][][]byte
}

func Initialize(st *pbp2p.BeaconState) (*BeaconState, error) {
	fieldRoots, err := stateutil.ComputeFieldRoots(st)
	if err != nil {
		return nil, err
	}
	layers := merkleize(fieldRoots)
	return &BeaconState{
		state:        st,
		merkleLayers: layers,
	}, nil
}

func (b *BeaconState) HashTreeRoot() [32]byte {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return bytesutil.ToBytes32(b.merkleLayers[len(b.merkleLayers)-1][0])
}

func merkleize(leaves [][]byte) [][][]byte {
	for len(leaves) != 32 {
		leaves = append(leaves, make([]byte, 32))
	}
	currentLayer := leaves
	layers := make([][][]byte, 5)
	layers[0] = currentLayer

	// We keep track of the hash layers of a Merkle trie until we reach
	// the top layer of length 1, which contains the single root element.
	//        [Root]      -> Top layer has length 1.
	//    [E]       [F]   -> This layer has length 2.
	// [A]  [B]  [C]  [D] -> The bottom layer has length 4 (needs to be a power of two).
	i := 1
	for len(currentLayer) > 1 && i < len(layers) {
		layer := make([][]byte, 0)
		for i := 0; i < len(currentLayer); i += 2 {
			hashedChunk := hashutil.Hash(append(currentLayer[i], currentLayer[i+1]...))
			layer = append(layer, hashedChunk[:])
		}
		currentLayer = layer
		layers[i] = currentLayer
		i++
	}
	return layers
}
