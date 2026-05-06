package logstorage

import (
	"fmt"

	"github.com/VictoriaMetrics/VictoriaMetrics/lib/encoding"
)

// TenantID is an id of a tenant for log streams.
//
// Each log stream is associated with a single TenantID.
type TenantID struct {
	// AccountID is the id of the account for the log stream.
	AccountID uint32 `json:"account_id"`

	// ProjectID is the id of the project for the log stream.
	ProjectID uint32 `json:"project_id"`
}

func (tid *TenantID) marshalString(dst []byte) []byte {
	n := uint64(tid.AccountID)<<32 | uint64(tid.ProjectID)
	dst = marshalUint64Hex(dst, n)
	return dst
}

// unmarshal unmarshals tid from src and returns the remaining tail.
func (tid *TenantID) unmarshal(src []byte) ([]byte, error) {
	if len(src) < 8 {
		return src, fmt.Errorf("cannot unmarshal tenantID from %d bytes; need at least 8 bytes", len(src))
	}
	tid.AccountID = encoding.UnmarshalUint32(src[:4])
	tid.ProjectID = encoding.UnmarshalUint32(src[4:])
	return src[8:], nil
}
