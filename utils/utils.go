package utils

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"github.com/boltdb/bolt"
	"log"
	"os"
	"strconv"
)
// IntToHex converts an int64 to a byte array
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}
// ReverseBytes reverses a byte array
func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}
func ErrorLog(err error){
	if err != nil {
		log.Panic(err)
	}
}
func GobEncode(data interface{}) []byte {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	ErrorLog(err)
	return buff.Bytes()
}
func FileExists(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}
func FindDB(file string) *bolt.DB {
	db, err := bolt.Open(file,0600,nil)
	ErrorLog(err)
	return db
}
func TransformFileIndex(index int) string {
	var bt bytes.Buffer
	dbFileIndex := strconv.Itoa(index)
	for i := 0; i < 6 - len(dbFileIndex); i++ {
		bt.WriteString("0")
	}
	bt.WriteString(dbFileIndex)
	return bt.String()
}