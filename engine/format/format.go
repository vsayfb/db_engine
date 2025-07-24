package format

import "encoding/binary"

/*
    	 	    BINARY FORMAT

   	 4 BYTE   	  4 BYTE
    [KEY_LEN]	[VAL_LEN]	[KEY]	[VAL]
*/

func FormatBinary(key, val []byte) []byte {
	keyLen := len(key)
	valLen := len(val)

	bytes := make([]byte, 8+keyLen+valLen)

	binary.BigEndian.PutUint32(bytes[0:4], uint32(keyLen))

	binary.BigEndian.PutUint32(bytes[4:8], uint32(valLen))

	copy(bytes[8:8+keyLen], key)

	copy(bytes[8+keyLen:], val)

	return bytes
}
