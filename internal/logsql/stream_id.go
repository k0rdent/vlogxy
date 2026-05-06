package logstorage

import (
	"encoding/hex"
	"fmt"
)

// streamID is an internal id of log stream.
//
// Blocks are ordered by streamID inside parts.
type streamID struct {
	// tenantID is a tenant id for the given stream.
	// It is located at the beginning of streamID in order
	// to physically group blocks for the same tenants on the storage.
	tenantID TenantID

	// id is internal id, which uniquely identifies the stream in the tenant by its labels.
	// It is calculated as a hash of canonically sorted stream labels.
	//
	// Streams with identical sets of labels, which belong to distinct tenants, have the same id.
	id u128
}

// marshalString returns _stream_id value for the given sid.
func (sid *streamID) marshalString(dst []byte) []byte {
	dst = sid.tenantID.marshalString(dst)
	dst = sid.id.marshalString(dst)
	return dst
}

func (sid *streamID) tryUnmarshalFromString(s string) bool {
	data, err := hex.DecodeString(s)
	if err != nil {
		return false
	}
	tail, err := sid.unmarshal(data)
	if err != nil || len(tail) > 0 {
		return false
	}
	return true
}

// String returns human-readable representation for sid.
func (sid *streamID) String() string {
	return fmt.Sprintf("(tenant_id={account_id=%d, project_id=%d}, id=%s)", sid.tenantID.AccountID, sid.tenantID.ProjectID, &sid.id)
}

// unmarshal unmarshals sid from src and returns the tail from src.
func (sid *streamID) unmarshal(src []byte) ([]byte, error) {
	srcOrig := src
	tail, err := sid.tenantID.unmarshal(src)
	if err != nil {
		return srcOrig, err
	}
	src = tail
	tail, err = sid.id.unmarshal(src)
	if err != nil {
		return srcOrig, err
	}
	return tail, nil
}
