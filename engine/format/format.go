package format

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

/*
    	 	    BINARY FORMAT

   	 4 BYTE   	  4 BYTE
    [KEY_LEN]	[VAL_LEN]	[KEY]	[VAL]
*/

func EncodeBinary(key, val []byte) []byte {
	keyLen := len(key)
	valLen := len(val)

	bytes := make([]byte, 8+keyLen+valLen)

	binary.BigEndian.PutUint32(bytes[0:4], uint32(keyLen))

	binary.BigEndian.PutUint32(bytes[4:8], uint32(valLen))

	copy(bytes[8:8+keyLen], key)

	copy(bytes[8+keyLen:], val)

	return bytes
}

func DecodeBinary(file *os.File, offset int64) (key, val []byte, nextOffset int64, err error) {

	header := make([]byte, 8)

	_, err = file.ReadAt(header, offset)
	if err != nil {
		if err == io.EOF {
			return nil, nil, offset, io.EOF
		}
		return nil, nil, offset, fmt.Errorf("failed to read header: %v", err)
	}

	keyLen := binary.BigEndian.Uint32(header[0:4])
	valLen := binary.BigEndian.Uint32(header[4:8])

	totalLen := int64(keyLen) + int64(valLen)
	data := make([]byte, totalLen)

	_, err = file.ReadAt(data, offset+8)

	if err != nil {
		return nil, nil, offset, fmt.Errorf("failed to read key/value: %v", err)
	}

	key = data[:keyLen]
	val = data[keyLen:]
	nextOffset = offset + 8 + totalLen

	return key, val, nextOffset, nil
}
