package mytype

import (
	"fmt"
	"strings"

	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	GroupSession   SessionType = "group"
	ContactSession SessionType = "contact"
)

type Session struct {
	ID       SessionID
	Username string
	Content  string
	Unreads  int
}

type SessionType string

type SessionID struct {
	Type  SessionType
	Value []byte
}

func (s *SessionID) String() string {
	if s.Type == ContactSession {
		return fmt.Sprintf("%s_%s", s.Type, peer.ID(s.Value).String())
	}

	return fmt.Sprintf("%s_%s", s.Type, string(s.Value))
}

func GroupSessionID(groupID string) SessionID {
	return SessionID{
		Type:  GroupSession,
		Value: []byte(groupID),
	}
}

func ContactSessionID(contactID peer.ID) SessionID {
	return SessionID{
		Type:  ContactSession,
		Value: []byte(contactID),
	}
}

func DecodeSessionID(sessionID string) (*SessionID, error) {
	arr := strings.Split(sessionID, "_")
	if len(arr) != 2 {
		return nil, fmt.Errorf("sessionID chunks num error")
	}

	var id []byte
	switch arr[0] {
	case string(GroupSession):
		id = []byte(arr[1])
	case string(ContactSession):
		peerID, err := peer.Decode(arr[1])
		if err != nil {
			return nil, fmt.Errorf("peer decode error: %w", err)
		}
		id = []byte(peerID)
	default:
		return nil, fmt.Errorf("unsupport session type")
	}

	return &SessionID{Type: SessionType(arr[0]), Value: id}, nil
}
