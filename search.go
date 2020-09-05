package ipdb

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"unsafe"
)

// Search the database for the providing ip, returning the offset in the database
// referring to the data, or error.
func (d *IPDb) Search(ip net.IP) (uint32, error) {
	nodeCount := d.Metadata.NodeCount

	var iplen int
	if d.IsIPv4Db() {
		ip = ip.To4()
		if ip == nil {
			return 0, errIPv4Only
		}
		iplen = 32
	} else {
		if !d.IsIPv6Db() {
			return 0, errDatabase
		}

		ip = ip.To16()
		if ip == nil {
			return 0, errIPv6Only
		}
		iplen = 128
	}

	offset := d.startFrom
	for i := 0; i < iplen; i++ {
		if offset >= nodeCount {
			break
		}

		// To get exactly the ith bit
		bit := 0xFF & ip[i>>3] >> uint(7-(i%8)) & 0x01
		pos := offset*8 + uint32(bit*4)
		offset = binary.BigEndian.Uint32(d.Data[pos : pos+4])
	}

	if offset < nodeCount {
		return 0, errNoData
	}

	return offset, nil
}

// GetRaw returns the raw data by giving the offset, or error.
func (d *IPDb) GetRaw(offset uint32) ([]byte, error) {
	if offset < d.Metadata.NodeCount {
		return nil, errNoData
	}

	dbDataLen := uint32(len(d.Data))

	// pos := d.Metadata.NodeCount*8 + (offset - d.Metadata.NodeCount)
	pos := offset + d.Metadata.NodeCount*7
	if pos >= dbDataLen {
		return nil, errDatabase
	}

	len := binary.BigEndian.Uint16(d.Data[pos : pos+2])
	if pos+2+uint32(len) > dbDataLen {
		return nil, errDatabase
	}

	return d.Data[pos+2 : pos+2+uint32(len)], nil
}

// GetAllFields returns the data mapping by giving the offset, or error.
// BE AWARE: The return value stores the data in the form
// ``map[field]map[lang]data''.
func (d *IPDb) GetAllFields(offset uint32) (map[string]map[string]string, error) {
	raw, err := d.GetRaw(offset)
	if err != nil {
		return nil, err
	}

	rawPtr := (*string)(unsafe.Pointer(&raw))
	rawStrings := strings.Split(*rawPtr, "\t")
	res := make(map[string]map[string]string)

	for lang, pos := range d.Metadata.Languages {
		if pos+len(d.Metadata.Fields) > len(rawStrings) {
			return nil, errDatabase
		}
		for i, field := range d.Metadata.Fields {
			if res[field] == nil {
				res[field] = make(map[string]string)
			}
			res[field][lang] = rawStrings[pos+i]
		}
	}

	return res, nil
}

// GetAllFieldsLocale returns the data mapping by giving the offset, in the
// providing locale ``lang'', or error.
func (d *IPDb) GetAllFieldsLocale(offset uint32, lang string) (map[string]string, error) {
	raw, err := d.GetRaw(offset)
	if err != nil {
		return nil, err
	}

	rawPtr := (*string)(unsafe.Pointer(&raw))
	rawStrings := strings.Split(*rawPtr, "\t")
	res := make(map[string]string)
	pos, ok := d.Metadata.Languages[lang]
	if !ok {
		return nil, fmt.Errorf("language %s not supported", lang)
	}

	if pos+len(d.Metadata.Fields) > len(rawStrings) {
		return nil, errDatabase
	}

	for i, field := range d.Metadata.Fields {
		res[field] = rawStrings[pos+i]
	}

	return res, nil
}

// GetValue returns the values in all locale, by giving offset and field, or error.
func (d *IPDb) GetValue(offset uint32, field string) (map[string]string, error) {
	v, err := d.GetAllFields(offset)
	if err != nil {
		return nil, err
	}

	res, ok := v[field]
	if ok {
		return res, nil
	}
	return nil, fmt.Errorf("field %s not found", field)
}

// GetValueLocale returns the values in the providing locale ``lang'', by giving
// offset and field, or error.
func (d *IPDb) GetValueLocale(offset uint32, field, lang string) (string, error) {
	v, err := d.GetAllFieldsLocale(offset, lang)
	if err != nil {
		return "", err
	}

	res, ok := v[field]
	if ok {
		return res, nil
	}
	return "", fmt.Errorf("field %s not found", field)
}
