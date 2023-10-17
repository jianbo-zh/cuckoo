package util

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/mr-tron/base58/base58"
)

var qrcodeTokenPrefix = "cuckoo:"
var myAlphabet = base58.NewAlphabet("ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz123456789")

type TokenType int

const (
	TokenTypePeer  TokenType = 0
	TokenTypeGroup TokenType = 1
)

type QRCodeToken struct {
	Type        TokenType `json:"t"`
	ID          string    `json:"i"`
	Name        string    `json:"n"`
	Avatar      string    `json:"a"`
	DepositAddr string    `json:"d"`
}

func EncodePeerToken(peerID peer.ID, depositAddr peer.ID, name string, avatar string) string {
	token := QRCodeToken{
		Type:        TokenTypePeer,
		ID:          peerID.String(),
		Name:        base64.RawStdEncoding.EncodeToString([]byte(name)),
		Avatar:      avatar,
		DepositAddr: depositAddr.String(),
	}

	bs, _ := json.Marshal(token)
	return qrcodeTokenPrefix + base58.EncodeAlphabet(bs, myAlphabet)
}

func EncodeGroupToken(groupID string, depositAddr peer.ID, name string, avatar string) string {
	token := QRCodeToken{
		Type:        TokenTypeGroup,
		ID:          groupID,
		Name:        base64.RawStdEncoding.EncodeToString([]byte(name)),
		Avatar:      avatar,
		DepositAddr: depositAddr.String(),
	}

	bs, _ := json.Marshal(token)
	return qrcodeTokenPrefix + base58.EncodeAlphabet(bs, myAlphabet)
}

func DecodeQRCodeToken(token string) (*QRCodeToken, error) {
	if !strings.HasPrefix(token, qrcodeTokenPrefix) {
		return nil, fmt.Errorf("token no prefix")
	}

	bs, err := base58.DecodeAlphabet(strings.TrimPrefix(token, qrcodeTokenPrefix), myAlphabet)
	if err != nil {
		return nil, fmt.Errorf("decode token error: %w", err)
	}

	var qrtoken QRCodeToken
	if err := json.Unmarshal(bs, &qrtoken); err != nil {
		return nil, fmt.Errorf("json.Unmarshal error: %w", err)
	}

	name, err := base64.RawStdEncoding.DecodeString(qrtoken.Name)
	if err != nil {
		return nil, fmt.Errorf("base64.DecodeString error: %w", err)
	}

	qrtoken.Name = string(name)

	return &qrtoken, nil
}
