package evidence

import (
	"encoding/hex"
	"math/big"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/ethereum"
)

const anchorSignature = "anchor(bytes32,uint32)"

var anchorSelector = hex.EncodeToString(ethereum.Keccak256([]byte(anchorSignature))[:4])

func RegistryCalldata(root string, recordCount int) string {
	count := hex.EncodeToString(ethereum.Word(big.NewInt(int64(recordCount))))
	return "0x" + anchorSelector + root[2:] + count
}

type anchorTarget struct {
	to       string
	calldata string
}

func (s *Service) target(batch Batch) anchorTarget {
	if s.config.RegistryAddress != "" {
		return anchorTarget{to: s.config.RegistryAddress, calldata: RegistryCalldata(batch.Root, batch.RecordCount)}
	}
	return anchorTarget{calldata: Calldata(batch.Root)}
}

func (s *Service) registryURL() string {
	if s.config.RegistryAddress == "" || s.config.ExplorerURL == "" {
		return ""
	}
	return s.config.ExplorerURL + "/address/" + s.config.RegistryAddress
}
