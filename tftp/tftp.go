package tftp

import "errors"

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
