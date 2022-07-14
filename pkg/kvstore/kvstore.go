// Package kvstore provides a thread-safe key value store.
package kvstore

import "errors"

// Entry is a key value store entry field.
type Entry struct {
	Value string
	Owner string
}

// EntryInfo provides details on a single store key.
type EntryInfo struct {
	Key   string `json:"key"`
	Owner string `json:"owner"`
}

// KVStore is a thread-safe key value store.
type KVStore struct {
	data           map[string]*Entry
	requestChannel chan *request
}

type operation int

const (
	readOperation    operation = iota
	writeOperation   operation = iota
	deleteOperation  operation = iota
	listOperation    operation = iota
	listAllOperation operation = iota
	closeOperation   operation = iota
)

var (
	errUpdateSameUser = errors.New("cannot update entry owned by someone else")
	errDeleteSameUser = errors.New("cannot delete entry owned by someone else")
)

type request struct {
	op     operation
	params interface{}
}

type readRequest struct {
	key             string
	responseChannel chan *readResponse
}

type readResponse struct {
	value string
	ok    bool
}

type writeRequest struct {
	key             string
	value           string
	username        string
	responseChannel chan *writeResponse
}

type writeResponse struct {
	err error
}

type deleteRequest struct {
	key             string
	username        string
	responseChannel chan *deleteResponse
}

type deleteResponse struct {
	deleted bool
	err     error
}

type listRequest struct {
	key             string
	responseChannel chan *listResponse
}

type listResponse struct {
	entryInfo *EntryInfo
}

type listAllRequest struct {
	responseChannel chan *listAllResponse
}

type listAllResponse struct {
	entries []*EntryInfo
}

// NewKVStore returns a new key value store instance.
func NewKVStore() *KVStore {
	store := &KVStore{
		make(map[string]*Entry),
		make(chan *request),
	}

	// start the internal go routine
	handleStoreOperations(store)

	return store
}

// Close shuts down the key value store cleanly.
func Close(s *KVStore) {
	s.requestChannel <- &request{closeOperation, nil}
}

// Read returns the value of the specified key, and a flag
// indicating if the key was present.
//
// Any user can read a key's value.
func Read(s *KVStore, key string) (string, bool) {
	responseChannel := make(chan *readResponse)
	s.requestChannel <- &request{readOperation, &readRequest{key, responseChannel}}

	response := <-responseChannel

	return response.value, response.ok
}

// Write sets or updates the key value.
//
// Only the owning user can update an existing entry.
func Write(s *KVStore, key string, value string, username string) error {
	responseChannel := make(chan *writeResponse)
	s.requestChannel <- &request{writeOperation, &writeRequest{key, value, username, responseChannel}}

	response := <-responseChannel

	return response.err
}

// Delete removes a key, and returns a flag
// indicating if the key was deleted.
//
// Only the owning user can delete a key.
func Delete(s *KVStore, key string, username string) (bool, error) {
	responseChannel := make(chan *deleteResponse)
	s.requestChannel <- &request{deleteOperation, &deleteRequest{key, username, responseChannel}}

	response := <-responseChannel

	return response.deleted, response.err
}

// List returns the value and owner of the specified key, or nil if not present.
//
// Any user can list a key's value.
func List(s *KVStore, key string) *EntryInfo {
	responseChannel := make(chan *listResponse)
	s.requestChannel <- &request{listOperation, &listRequest{key, responseChannel}}

	response := <-responseChannel

	return response.entryInfo
}

// ListAll returns the value and owner of all keys.
//
// Any user can list all keys.
func ListAll(s *KVStore) []*EntryInfo {
	responseChannel := make(chan *listAllResponse)
	s.requestChannel <- &request{listAllOperation, &listAllRequest{responseChannel}}

	response := <-responseChannel

	return response.entries
}

// handleStoreOperations provides thread-safety for the key value store, by performing operations
// on the store in a single go routine in serial, with input provided through messages on a channel.
func handleStoreOperations(s *KVStore) {
	go func() {
		for {
			request := <-s.requestChannel
			switch request.op {
			case readOperation:
				params, ok := request.params.(*readRequest)
				if ok {
					if entry, ok := s.data[params.key]; ok {
						// key is present
						params.responseChannel <- &readResponse{entry.Value, true}
					} else {
						// key not present
						params.responseChannel <- &readResponse{"", false}
					}
				}

			case writeOperation:
				params, ok := request.params.(*writeRequest)
				if ok {
					if entry, ok := s.data[params.key]; ok {
						if entry.Owner == params.username {
							// owner updating key
							s.data[params.key].Value = params.value
							params.responseChannel <- &writeResponse{nil}
						} else {
							// someone else updating key
							params.responseChannel <- &writeResponse{errUpdateSameUser}
						}
					} else {
						// new key
						s.data[params.key] = &Entry{params.value, params.username}
						params.responseChannel <- &writeResponse{nil}
					}
				}

			case deleteOperation:
				params, ok := request.params.(*deleteRequest)
				if ok {
					if entry, ok := s.data[params.key]; ok {
						if entry.Owner == params.username {
							// owner deleting key
							delete(s.data, params.key)
							params.responseChannel <- &deleteResponse{true, nil}
						} else {
							// someone else deleting key
							params.responseChannel <- &deleteResponse{false, errDeleteSameUser}
						}
					} else {
						// key not present
						params.responseChannel <- &deleteResponse{false, nil}
					}
				}

			case listOperation:
				params, ok := request.params.(*listRequest)
				if ok {
					if entry, ok := s.data[params.key]; ok {
						// key is present
						params.responseChannel <- &listResponse{&EntryInfo{params.key, entry.Owner}}
					} else {
						// key not present
						params.responseChannel <- &listResponse{nil}
					}
				}

			case listAllOperation:
				params, ok := request.params.(*listAllRequest)
				if ok {
					// export all entries (if any) into a slice to return
					entries := make([]*EntryInfo, 0, len(s.data))
					for key, entry := range s.data {
						entries = append(entries, &EntryInfo{key, entry.Owner})
					}
					params.responseChannel <- &listAllResponse{entries}
				}

			case closeOperation:
				return
			}
		}
	}()
}
