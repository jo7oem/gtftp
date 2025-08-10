package tftp

import (
	"fmt"
	"net"
	"time"
)

type Client struct {
	// target はTFTPサーバーのアドレスを表します。
	target string
}

// NewClient は新しいTFTPクライアントを作成します。
func NewClient(target string) *Client {
	return &Client{
		target: target,
	}
}

func (c *Client) Recv(filename string, mode string) ([]byte, error) {
	uAddr, err := net.ResolveUDPAddr("udp", c.target)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	rr, err := NewRRQPacket(filename, mode)
	if err != nil {
		return nil, fmt.Errorf("failed to create RRQ packet: %w", err)
	}
	// RRQ送信
	if err := c.sendPacket(conn, uAddr, rr); err != nil {
		return nil, fmt.Errorf("failed to send RRQ packet: %w", err)
	}

	var result []byte
	blockNum := uint16(1)
	serverAddr := uAddr

	for {
		buf := make([]byte, 516)
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			return nil, fmt.Errorf("failed to read from UDP: %w", err)
		}
		serverAddr = addr // サーバーの転送ポートに更新

		pkt, err := UnmarshallPacket(buf[:n])
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal packet: %w", err)
		}

		switch pkt.OpCode {
		case OpDATA:
			if pkt.Block != blockNum {
				// ブロック番号不一致
				continue
			}
			result = append(result, pkt.Data...)
			ack := NewACKPacket(blockNum)
			ackData, _ := ack.MarshallPacket()
			_, err = conn.WriteToUDP(ackData, serverAddr)
			if err != nil {
				return nil, fmt.Errorf("failed to send ACK: %w", err)
			}
			if len(pkt.Data) < 512 {
				// 最終ブロック
				return result, nil
			}
			blockNum++
		case OpERROR:
			return nil, fmt.Errorf("TFTP error: %d,%s", pkt.ErrorCode, pkt.ErrMsg)
		default:
			return nil, fmt.Errorf("unexpected packet opcode: %d", pkt.OpCode)
		}
	}
}
func (c *Client) sendPacket(conn *net.UDPConn, addr *net.UDPAddr, packt *tftpPacket) error {
	// パケットを送信
	data, err := packt.MarshallPacket()
	if err != nil {
		return fmt.Errorf("failed to marshal packet: %w", err)
	}

	_, err = conn.WriteToUDP(data, addr)

	return err
}
