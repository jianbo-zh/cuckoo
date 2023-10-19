package sessionds

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	ipfsds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/jianbo-zh/dchat/internal/datastore"
	pb "github.com/jianbo-zh/dchat/protobuf/pb/sessionpb"
	"google.golang.org/protobuf/proto"
)

var _ SessionIface = (*SessionDataStore)(nil)

var sessionDsKey = &datastore.SessionDsKey{}

func SessionWrap(d ipfsds.Batching) *SessionDataStore {
	return &SessionDataStore{Batching: d}
}

type SessionDataStore struct {
	ipfsds.Batching

	unreadsMutex sync.Mutex
}

func (s *SessionDataStore) SetSessionID(ctx context.Context, sessionID string) error {
	if err := s.Put(ctx, sessionDsKey.ListKey(sessionID), []byte(strconv.FormatInt(time.Now().Unix(), 10))); err != nil {
		return fmt.Errorf("ds put key error: %w", err)
	}

	return nil
}

func (s *SessionDataStore) ClearSession(ctx context.Context, sessionID string) error {
	batch, err := s.Batch(ctx)
	if err != nil {
		return fmt.Errorf("ds batch error: %w", err)
	}

	if err := batch.Delete(ctx, sessionDsKey.LastMsgKey(sessionID)); err != nil {
		return fmt.Errorf("ds delete session last msg key error: %w", err)
	}

	if err := batch.Put(ctx, sessionDsKey.UnreadsKey(sessionID), []byte(strconv.FormatInt(0, 10))); err != nil {
		return fmt.Errorf("ds delete session unreads key error: %w", err)
	}

	return batch.Commit(ctx)
}

func (s *SessionDataStore) DeleteSession(ctx context.Context, sessionID string) error {
	batch, err := s.Batch(ctx)
	if err != nil {
		return fmt.Errorf("ds batch error: %w", err)
	}

	if err := batch.Delete(ctx, sessionDsKey.LastMsgKey(sessionID)); err != nil {
		return fmt.Errorf("ds delete session last msg key error: %w", err)
	}

	if err := batch.Delete(ctx, sessionDsKey.UnreadsKey(sessionID)); err != nil {
		return fmt.Errorf("ds delete session unreads key error: %w", err)
	}

	if err := batch.Delete(ctx, sessionDsKey.ListKey(sessionID)); err != nil {
		return fmt.Errorf("ds delete session key error: %w", err)
	}

	return batch.Commit(ctx)
}

func (s *SessionDataStore) GetSessionIDs(ctx context.Context) ([]string, error) {
	results, err := s.Query(ctx, query.Query{
		Prefix: sessionDsKey.ListPrefix(),
		Orders: []query.Order{query.OrderByValueDescending{}},
	})
	if err != nil {
		return nil, fmt.Errorf("ds query error: %w", err)
	}

	var sessionIDs []string
	for result := range results.Next() {
		if result.Error != nil {
			return nil, result.Error
		}
		fmt.Println("session query: ", result.Key)
		sessionIDs = append(sessionIDs, strings.TrimPrefix(result.Key, sessionDsKey.ListPrefix()))
	}

	return sessionIDs, nil
}

func (s *SessionDataStore) GetLastMessage(ctx context.Context, sessionID string) (*pb.SessionLastMessage, error) {

	val, err := s.Get(ctx, sessionDsKey.LastMsgKey(sessionID))
	if err != nil {
		return nil, fmt.Errorf("ds get error: %w", err)
	}

	var lastMsg pb.SessionLastMessage
	if err = proto.Unmarshal(val, &lastMsg); err != nil {
		return nil, fmt.Errorf("proto unmarshal error: %w", err)
	}

	return &lastMsg, nil
}

func (s *SessionDataStore) UpdateSessionTime(ctx context.Context, sessionID string) error {
	if err := s.Put(ctx, sessionDsKey.ListKey(sessionID), []byte(strconv.FormatInt(time.Now().Unix(), 10))); err != nil {
		return fmt.Errorf("ds put error: %w", err)
	}
	return nil
}

func (s *SessionDataStore) SetLastMessage(ctx context.Context, sessionID string, lastMsg *pb.SessionLastMessage) error {
	bs, err := proto.Marshal(lastMsg)
	if err != nil {
		return fmt.Errorf("proto marshal error: %w", err)
	}
	if err := s.Put(ctx, sessionDsKey.LastMsgKey(sessionID), bs); err != nil {
		return fmt.Errorf("ds put error: %w", err)
	}
	return nil
}

func (s *SessionDataStore) GetUnreads(ctx context.Context, sessionID string) (int, error) {
	var unreads int64

	s.unreadsMutex.Lock()
	val, err := s.Get(ctx, sessionDsKey.UnreadsKey(sessionID))
	if err != nil {
		if errors.Is(err, ipfsds.ErrNotFound) {
			unreads = 0

		} else {
			s.unreadsMutex.Unlock()
			return 0, fmt.Errorf("ds get error: %w", err)
		}

	} else {
		unreads, err = strconv.ParseInt(string(val), 10, 64)
		if err != nil {
			s.unreadsMutex.Unlock()
			return 0, fmt.Errorf("strconv.ParseInt error: %w", err)
		}
	}

	s.unreadsMutex.Unlock()
	return int(unreads), nil
}

func (s *SessionDataStore) IncrUnreads(ctx context.Context, sessionID string) error {
	var err error
	var unreads int64

	s.unreadsMutex.Lock()
	val, err := s.Get(ctx, sessionDsKey.UnreadsKey(sessionID))
	if err != nil {
		if errors.Is(err, ipfsds.ErrNotFound) {
			unreads = 0

		} else {
			s.unreadsMutex.Unlock()
			return fmt.Errorf("ds get error: %w", err)
		}

	} else {
		unreads, err = strconv.ParseInt(string(val), 10, 64)
		if err != nil {
			s.unreadsMutex.Unlock()
			return fmt.Errorf("strconv.ParseInt error: %w", err)
		}
	}

	unreads++

	err = s.Put(ctx, sessionDsKey.UnreadsKey(sessionID), []byte(strconv.FormatInt(unreads, 10)))
	if err != nil {
		s.unreadsMutex.Unlock()
		return fmt.Errorf("ds put error: %w", err)
	}
	s.unreadsMutex.Unlock()

	return nil
}

func (s *SessionDataStore) ResetUnreads(ctx context.Context, sessionID string) error {
	s.unreadsMutex.Lock()
	err := s.Put(ctx, sessionDsKey.UnreadsKey(sessionID), []byte("0"))
	if err != nil {
		s.unreadsMutex.Unlock()
		return fmt.Errorf("ds put error: %w", err)
	}
	s.unreadsMutex.Unlock()

	return nil
}
