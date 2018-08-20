package requestmanager

import (
	"context"
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	webwire "github.com/qbeon/webwire-go"
	pld "github.com/qbeon/webwire-go/payload"
)

// RequestIdentifier represents the universally unique minified UUIDv4 identifier of a request.
type RequestIdentifier = [8]byte

// reply is used by the request manager to represent the results
// of a request (both failed and succeeded)
type reply struct {
	Reply webwire.Payload
	Error error
}

// Request represents a request created and tracked by the request manager
type Request struct {
	// manager references the RequestManager instance managing this request
	manager *RequestManager

	// identifier represents the unique identifier of this request
	identifier RequestIdentifier

	// timeout represents the configured timeout duration of this request
	timeout time.Duration

	// reply represents a channel for asynchronous reply handling
	reply chan reply
}

// Identifier returns the assigned request identifier
func (req *Request) Identifier() RequestIdentifier {
	return req.identifier
}

// AwaitReply blocks the calling goroutine
// until either the reply is fulfilled or failed, the request timed out
// a user-defined deadline was exceeded or the request was prematurely canceled.
// The timer is started when AwaitReply is called.
func (req *Request) AwaitReply(ctx context.Context) (webwire.Payload, error) {
	// Start timeout timer
	timeoutTimer := time.NewTimer(req.timeout)
	defer timeoutTimer.Stop()

	// Block until either deadline exceeded, canceled,
	// timed out or reply received
	select {
	case <-ctx.Done():
		req.manager.deregister(req.identifier)
		return nil, webwire.TranslateContextError(ctx.Err())
	case <-timeoutTimer.C:
		req.manager.deregister(req.identifier)
		return &webwire.EncodedPayload{},
			webwire.NewTimeoutErr(fmt.Errorf("timed out"))
	case reply := <-req.reply:
		timeoutTimer.Stop()
		if reply.Error != nil {
			return nil, reply.Error
		}

		// Don't return nil even if the reply is empty
		// to prevent invalid memory access attempts
		// caused by forgetting to check for != nil
		if reply.Reply == nil {
			return &webwire.EncodedPayload{}, nil
		}

		return reply.Reply, nil
	}
}

// RequestManager manages and keeps track of outgoing pending requests
type RequestManager struct {
	lastID uint64
	lock   sync.RWMutex

	// pending represents an indexed list of all pending requests
	pending map[RequestIdentifier]*Request
}

// NewRequestManager constructs and returns a new instance of a RequestManager
func NewRequestManager() RequestManager {
	return RequestManager{
		lastID:  0,
		lock:    sync.RWMutex{},
		pending: make(map[RequestIdentifier]*Request),
	}
}

// Create creates and registers a new request.
// Create doesn't start the timeout timer, this is done in the subsequent request.AwaitReply
func (manager *RequestManager) Create(timeout time.Duration) *Request {
	manager.lock.Lock()

	// Generate unique request identifier by incrementing the last assigned id
	manager.lastID++
	var identifier RequestIdentifier
	idBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(idBytes, manager.lastID)
	copy(identifier[:], idBytes[0:8])

	newRequest := &Request{
		manager,
		identifier,
		timeout,
		make(chan reply),
	}

	// Register the newly created request
	manager.pending[identifier] = newRequest

	manager.lock.Unlock()

	return newRequest
}

// deregister deregisters the given clients session from the list of currently pending requests
func (manager *RequestManager) deregister(identifier RequestIdentifier) {
	manager.lock.Lock()
	delete(manager.pending, identifier)
	manager.lock.Unlock()
}

// Fulfill fulfills the request associated with the given request identifier
// with the provided reply payload.
// Returns true if a pending request was fulfilled and deregistered, otherwise returns false
func (manager *RequestManager) Fulfill(
	identifier RequestIdentifier,
	payload pld.Payload,
) bool {
	manager.lock.RLock()
	req, exists := manager.pending[identifier]
	manager.lock.RUnlock()
	if !exists {
		return false
	}

	req.reply <- reply{
		Reply: &webwire.EncodedPayload{
			Payload: payload,
		},
		Error: nil,
	}
	manager.deregister(identifier)
	return true
}

// Fail fails the request associated with the given request identifier with the provided error.
// Returns true if a pending request was failed and deregistered, otherwise returns false
func (manager *RequestManager) Fail(identifier RequestIdentifier, err error) bool {
	manager.lock.RLock()
	req, exists := manager.pending[identifier]
	manager.lock.RUnlock()
	if !exists {
		return false
	}
	req.reply <- reply{
		Reply: nil,
		Error: err,
	}
	manager.deregister(identifier)
	return true
}

// PendingRequests returns the number of currently pending requests
func (manager *RequestManager) PendingRequests() int {
	manager.lock.RLock()
	len := len(manager.pending)
	manager.lock.RUnlock()
	return len
}

// IsPending returns true if the request associated with the given identifier is pending
func (manager *RequestManager) IsPending(identifier RequestIdentifier) bool {
	manager.lock.RLock()
	_, exists := manager.pending[identifier]
	manager.lock.RUnlock()
	return exists
}
