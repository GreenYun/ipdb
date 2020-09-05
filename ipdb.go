package ipdb

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"time"
)

// IPDb stores the whole IPDB file
type IPDb struct {
	FileSize  uint64
	Metadata  Metadata
	Data      []byte
	startFrom uint32
}

// Metadata stores the metadata of the IPDB file
type Metadata struct {
	Build     int64          `json:"build"`
	IPVersion uint16         `json:"ip_version"`
	Languages map[string]int `json:"languages"`
	NodeCount uint32         `json:"node_count"`
	TotalSize uint64         `json:"total_size"`
	Fields    []string       `json:"fields"`
}

var (
	errDatabase = errors.New("database file may be corrupted")
	errNoData   = errors.New("no data")
	errFileSize = errors.New("database file size too small")
	errIPv4Only = errors.New("only IPv4 accepted")
	errIPv6Only = errors.New("only IPv6 accepted")
)

// NewIPDb reads the IPDB file by providing filepath, returning an IPDb instance
// or error.
func NewIPDb(filepath string) (*IPDb, error) {
	finfo, err := os.Stat(filepath)
	if err != nil {
		return nil, err
	}

	fsize := uint64(finfo.Size())
	if fsize < 4 {
		return nil, errFileSize
	}

	fbody, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	metaLen := binary.BigEndian.Uint32(fbody[0:4])
	if fsize < uint64(4+metaLen) {
		return nil, errFileSize
	}

	var metadata Metadata
	if err := json.Unmarshal(fbody[4:4+metaLen], &metadata); err != nil {
		return nil, err
	}

	if fsize != uint64(4+metaLen)+metadata.TotalSize {
		return nil, errFileSize
	}

	ipdb := &IPDb{
		FileSize:  fsize,
		Metadata:  metadata,
		Data:      fbody[4+metaLen:],
		startFrom: 0,
	}

	// To simplify the search process, it omits the useless nodes in IPv4 database
	if metadata.IPVersion == 1 {
		offset := uint32(0)
		for i := 0; i < 96 && offset < metadata.NodeCount; i++ {
			if i >= 80 {
				pos := offset*8 + 4
				offset = binary.BigEndian.Uint32(ipdb.Data[pos : pos+4])
			} else {
				pos := offset * 8
				offset = binary.BigEndian.Uint32(ipdb.Data[pos : pos+4])
			}
		}
		ipdb.startFrom = offset
	}

	return ipdb, nil
}

// GetBuildTime converts build number in metadata into time.Time.
func (d *IPDb) GetBuildTime() time.Time {
	return time.Unix(d.Metadata.Build, 0).In(time.UTC)
}

// GetLanguages returns the languages supported by the database.
func (d *IPDb) GetLanguages() []string {
	languages := make([]string, 0, len(d.Metadata.Languages))
	for k := range d.Metadata.Languages {
		languages = append(languages, k)
	}
	return languages
}

// IsIPv4Db checks if the database an IPv4 database.
func (d *IPDb) IsIPv4Db() bool {
	return d.Metadata.IPVersion == uint16(1)
}

// IsIPv6Db checks if the database an IPv6 database.
func (d *IPDb) IsIPv6Db() bool {
	return d.Metadata.IPVersion == uint16(2)
}
