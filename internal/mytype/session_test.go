package mytype

import (
	"reflect"
	"testing"

	"github.com/libp2p/go-libp2p/core/peer"
)

func TestDecodeSessionID(t *testing.T) {
	type args struct {
		sessionID string
	}
	tests := []struct {
		name    string
		args    args
		want    *SessionID
		wantErr bool
	}{
		{
			args: args{
				sessionID: "contact_12D3KooWBuFtFZpg2qjYV3Ptb17SS11oRDh7ZpwT7mYoFuxzSFb7",
			},
			want: &SessionID{
				Type: ContactSession,
				Value: func() []byte {
					peerID, _ := peer.Decode("12D3KooWBuFtFZpg2qjYV3Ptb17SS11oRDh7ZpwT7mYoFuxzSFb7")
					return []byte(peerID)
				}(),
			},
			wantErr: false,
		},
		{
			args: args{
				sessionID: "group_debf6dd0-095a-4b20-b31b-2eb0e65db425",
			},
			want: &SessionID{
				Type: GroupSession,
				Value: func() []byte {
					return []byte("debf6dd0-095a-4b20-b31b-2eb0e65db425")
				}(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeSessionID(tt.args.sessionID)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeSessionID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecodeSessionID() = %v, want %v", got, tt.want)
			}
		})
	}
}
