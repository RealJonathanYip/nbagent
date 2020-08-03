package protocol

import (
	"bytes"
	"encoding/binary"
	"git.yayafish.com/nbagent/log"
	"github.com/golang/protobuf/proto"
	"io"
)

type BaseProtocol struct {
	Uri       string
	MsgType   uint8
	RequestID uint32
	Data      []byte
}

const (
	MSG_TYPE_REQUEST    = 1
	MSG_TYPE_RESPONE    = 2
	MSG_TYPE_CAST       = 3
	PROTO_SIZE_BASE     = 5
)

func ReadString(ptrIOReader io.Reader) (string, bool) {
	var nLength uint16
	if anyErr := binary.Read(ptrIOReader, binary.LittleEndian, &nLength); anyErr != nil {
		log.Warningf("read string fail:%v", anyErr)
		return "", false
	}

	ptrBuffer := bytes.NewBuffer([]byte{})
	objLimitReader := io.LimitReader(ptrIOReader, int64(nLength))
	if _, anyErr := ptrBuffer.ReadFrom(objLimitReader); anyErr == nil {
		return string(ptrBuffer.Bytes()), true
	} else {
		log.Warningf("read string fail:%v", anyErr)
		return "", false
	}
}

func WriteString(ptrIOWriter io.Writer, szStr string) bool {
	var nLength uint16
	nLength = uint16(len(szStr))
	if anyErr := binary.Write(ptrIOWriter, binary.LittleEndian, nLength); anyErr != nil {
		log.Warningf("write string fail:%v", anyErr)
		return false
	} else if anyErr := binary.Write(ptrIOWriter, binary.LittleEndian, []byte(szStr)); anyErr != nil {
		log.Warningf("write string fail:%v", anyErr)
		return false
	} else {
		return true
	}
}

// nLength:uint32 + szUri:uint16_string + nMsgType:uint8 + szRequestID:uint16_string + byte:Body
func NewProtocol(szUri string, nMsgType uint8, szRequestID string, objMsg proto.Message) ([]byte, bool) {
	byteBuffer := bytes.NewBuffer([]byte{})
	byteMsg, _ := proto.Marshal(objMsg)
	nSize := PROTO_SIZE_BASE + len(szUri) + len(szRequestID) + len(byteMsg)

	if anyErr := binary.Write(byteBuffer, binary.LittleEndian, uint32(nSize)); anyErr != nil {
		log.Warningf("write package size fail:%v", anyErr)
		return nil, false
	} else if !WriteString(byteBuffer, szUri) {
		return nil, false
	} else if anyErr := binary.Write(byteBuffer, binary.LittleEndian, nMsgType); anyErr != nil {
		log.Warningf("write package nMsgType fail:%v", anyErr)
		return nil, false
	} else if !WriteString(byteBuffer, szRequestID) {
		log.Warningf("write package nRequestID fail:%v", anyErr)
		return nil, false
	} else if anyErr := binary.Write(byteBuffer, binary.LittleEndian, byteMsg); anyErr != nil {
		log.Warningf("write data fail:%v", anyErr)
		return nil, false
	} else {
		return byteBuffer.Bytes(), true
	}
}
