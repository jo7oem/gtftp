package tftp

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
)

// TFTP OpCode定義
const (
	OpRRQ   = 1
	OpWRQ   = 2
	OpDATA  = 3
	OpACK   = 4
	OpERROR = 5
)

// TFTP mode定義
const (
	ModeNetASCII = "netascii"
	ModeOctet    = "octet"
	ModeMail     = "mail" //unsupported
)

// TFTPエラーコード定義
const (
	ErrorNotDefined        = 0
	ErrorFileNotFound      = 1
	ErrorAccessViolation   = 2
	ErrorDiskFull          = 3
	ErrorIllegalOperation  = 4
	ErrorUnknownTransferID = 5
	ErrorFileExists        = 6
	ErrorNoSuchUser        = 7
)

var (
	// ErrInvalidOpCode は無効なオペコードに対するエラーです
	ErrInvalidOpCode = errors.New("invalid opcode")
	// ErrInvalidMode は無効なモードに対するエラーです
	ErrInvalidMode = errors.New("invalid mode")
	// ErrInvalidPacket は無効なパケットに対するエラーです
	ErrInvalidPacket = errors.New("invalid packet")
)

// tftpPacket は TFTPパケットの構造体。
// 各フィールドはTFTPのオペコードに応じて使用されます。
// RFC1350
type tftpPacket struct {
	OpCode    uint16
	Filename  string //RRQ,WRQ
	Mode      string //RRQ,WRQ
	Block     uint16 //DATA,ACK
	Data      []byte //DATA
	ErrorCode uint16 //ERROR
	ErrMsg    string //ERROR
}

// NewRRQPacket は新しいRead Request (RRQ) パケットを作成します。
func NewRRQPacket(filename, mode string) (*tftpPacket, error) {
	if mode != ModeNetASCII && mode != ModeOctet {
		return nil, ErrInvalidMode
	}
	return &tftpPacket{
		OpCode:   OpRRQ,
		Filename: filename,
		Mode:     mode,
	}, nil
}

// NewWRQPacket は新しいWrite Request (WRQ) パケットを作成します。
func NewWRQPacket(filename, mode string) (*tftpPacket, error) {
	if mode != ModeNetASCII && mode != ModeOctet {
		return nil, ErrInvalidMode
	}
	return &tftpPacket{
		OpCode:   OpWRQ,
		Filename: filename,
		Mode:     mode,
	}, nil
}

// NewDATAPacket は新しいData (DATA) パケットを作成します。
func NewDATAPacket(block uint16, data []byte) *tftpPacket {
	return &tftpPacket{
		OpCode: OpDATA,
		Block:  block,
		Data:   data,
	}
}

// NewACKPacket は新しいAcknowledgment (ACK) パケットを作成します。
func NewACKPacket(block uint16) *tftpPacket {
	return &tftpPacket{
		OpCode: OpACK,
		Block:  block,
	}
}

// NewERRORPacket は新しいError (ERROR) パケットを作成します。
func NewERRORPacket(errorCode uint16, errMsg string) (*tftpPacket, error) {
	if errorCode < 0 || errorCode > 7 {
		return nil, ErrInvalidOpCode
	}

	return &tftpPacket{
		OpCode:    OpERROR,
		ErrorCode: errorCode,
		ErrMsg:    errMsg,
	}, nil
}

// MarshallPacket は tftpPacket をバイトスライスに変換。
// 変換後のバイトスライスは、TFTPプロトコルに従った形式になります。
func (p *tftpPacket) MarshallPacket() ([]byte, error) {
	// パケットのサイズを計算
	size := 2 // OpCodeのサイズ
	switch p.OpCode {
	case OpRRQ, OpWRQ:
		size += len(p.Filename) + len(p.Mode) + 2 // ファイル名とモードの長さ + 2バイトの終端
	case OpDATA:
		size += len(p.Data) + 2 // データの長さ + ブロック番号
	case OpACK:
		size += 2 // ブロック番号
	case OpERROR:
		size += 2 + len(p.ErrMsg) + 1 // エラーコードとメッセージの長さ + 1バイトの終端
	default:
		return nil, ErrInvalidOpCode
	}

	packet := make([]byte, size)
	offset := 0

	// OpCodeを設定
	packet[offset] = byte(p.OpCode >> 8)
	packet[offset+1] = byte(p.OpCode)
	offset += 2

	switch p.OpCode {
	case OpRRQ, OpWRQ:
		copy(packet[offset:], p.Filename)
		offset += len(p.Filename) + 1 // ファイル名の後にヌル文字
		copy(packet[offset:], p.Mode)
		offset += len(p.Mode) + 1 // モードの後にヌル文字
	case OpDATA:
		packet[offset] = byte(p.Block >> 8)
		packet[offset+1] = byte(p.Block)
		offset += 2
		copy(packet[offset:], p.Data)
	case OpACK:
		packet[offset] = byte(p.Block >> 8)
		packet[offset+1] = byte(p.Block)
	case OpERROR:
		packet[offset] = byte(p.ErrorCode >> 8)
		packet[offset+1] = byte(p.ErrorCode)
		offset += 2
		copy(packet[offset:], p.ErrMsg)
		offset += len(p.ErrMsg) + 1 // エラーメッセージの後にヌル文字
	default:
		return nil, ErrInvalidOpCode
	}

	return packet, nil
}

// UnmarshallPacket はバイトスライスを tftpPacket に変換。
func UnmarshallPacket(data []byte) (*tftpPacket, error) {
	// データの長さが4バイト未満の場合はエラー
	// rfc1350では、4Byteより短い形式はない。
	if len(data) < 4 {
		return nil, fmt.Errorf("%w, packet too short: %d bytes", ErrInvalidPacket, len(data))
	}

	opCode := uint16(data[0])<<8 | uint16(data[1])
	packet := &tftpPacket{OpCode: opCode}

	offset := 2

	switch opCode {
	case OpRRQ, OpWRQ:
		if data[len(data)-1] != 0 {
			return nil, fmt.Errorf("%w, filename must end with null byte", ErrInvalidPacket)
		}

		split := bytes.Split(data[offset:len(data)-1], []byte{0})

		if len(split) != 2 {
			return nil, fmt.Errorf("%w, broken packet", ErrInvalidPacket)
		}

		packet.Filename = string(split[0])
		packet.Mode = string(split[1])
	case OpDATA:
		packet.Block = uint16(data[offset])<<8 | uint16(data[offset+1])
		offset += 2
		packet.Data = data[offset:]
	case OpACK:
		if len(data) != 4 {
			return nil, fmt.Errorf("%w, ack packet must be 4 bytes", ErrInvalidPacket)
		}

		packet.Block = uint16(data[offset])<<8 | uint16(data[offset+1])
	case OpERROR:
		if len(data) < 5 {
			return nil, fmt.Errorf("%w, error packet must be at least 5 bytes", ErrInvalidPacket)
		}
		packet.ErrorCode = uint16(data[offset])<<8 | uint16(data[offset+1])
		offset += 2

		if data[len(data)-1] != 0 {
			return nil, fmt.Errorf("%w, error message must start with null byte", ErrInvalidPacket)
		}

		buf := data[offset : len(data)-1] // 最後のヌル文字を除外
		if slices.Index(buf, 0) > 0 {
			return nil, fmt.Errorf("%w, broken packet", ErrInvalidPacket)
		}
		packet.ErrMsg = string(buf)
	default:
		return nil, ErrInvalidOpCode
	}

	return packet, nil
}
