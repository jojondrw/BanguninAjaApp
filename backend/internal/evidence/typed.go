package evidence

import (
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/ethereum"
)

const (
	domainName    = "BanguninAja"
	domainVersion = "1"
	primaryType   = "OwnerClaimApproval"
	domainType    = "EIP712Domain(string name,string version,uint256 chainId)"
	approvalType  = "OwnerClaimApproval(string project,string contract,string term,uint256 progress,uint256 amount,string claimedOn,bytes32 evidenceRoot)"
)

type ApprovalMessage struct {
	Project      string `json:"project"`
	Contract     string `json:"contract"`
	Term         string `json:"term"`
	Progress     string `json:"progress"`
	Amount       string `json:"amount"`
	ClaimedOn    string `json:"claimedOn"`
	EvidenceRoot string `json:"evidenceRoot"`
}

type typedField struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type TypedData struct {
	Types       map[string][]typedField `json:"types"`
	PrimaryType string                  `json:"primaryType"`
	Domain      map[string]any          `json:"domain"`
	Message     ApprovalMessage         `json:"message"`
}

func NewTypedData(chainID int64, message ApprovalMessage) TypedData {
	return TypedData{
		Types: map[string][]typedField{
			"EIP712Domain": {{"name", "string"}, {"version", "string"}, {"chainId", "uint256"}},
			primaryType: {
				{"project", "string"}, {"contract", "string"}, {"term", "string"}, {"progress", "uint256"},
				{"amount", "uint256"}, {"claimedOn", "string"}, {"evidenceRoot", "bytes32"},
			},
		},
		PrimaryType: primaryType,
		Domain:      map[string]any{"name": domainName, "version": domainVersion, "chainId": chainID},
		Message:     message,
	}
}

func ApprovalDigest(chainID int64, message ApprovalMessage) []byte {
	keccakText := func(value string) []byte { return ethereum.Keccak256([]byte(value)) }
	number := func(value string) []byte {
		parsed, _ := new(big.Int).SetString(value, 10)
		if parsed == nil {
			parsed = new(big.Int)
		}
		return ethereum.Word(parsed)
	}
	root, _ := hex.DecodeString(strings.TrimPrefix(message.EvidenceRoot, "0x"))
	rootWord := make([]byte, 32)
	copy(rootWord, root)

	domain := ethereum.Keccak256(
		keccakText(domainType), keccakText(domainName), keccakText(domainVersion), ethereum.Word(big.NewInt(chainID)),
	)
	structHash := ethereum.Keccak256(
		keccakText(approvalType), keccakText(message.Project), keccakText(message.Contract), keccakText(message.Term),
		number(message.Progress), number(message.Amount), keccakText(message.ClaimedOn), rootWord,
	)
	return ethereum.Keccak256([]byte{0x19, 0x01}, domain, structHash)
}
