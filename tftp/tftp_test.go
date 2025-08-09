package tftp

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMarshalPacket(t *testing.T) {
	tests := []struct {
		name    string
		packet  tftpPacket
		want    []byte
		wantErr bool
	}{
		{
			name: "RRQ Packet",
			packet: tftpPacket{
				OpCode:   OpRRQ,
				Filename: "testfile.txt",
				Mode:     ModeOctet,
			},
			want: []byte{
				0x00, 0x01, // OpCode for RRQ
				't', 'e', 's', 't', 'f', 'i', 'l', 'e', '.', 't', 'x', 't', 0x00, // Filename
				'o', 'c', 't', 'e', 't', 0x00, // Mode
			},
			wantErr: false,
		},
		{
			name: "WRQ Packet",
			packet: tftpPacket{
				OpCode:   OpWRQ,
				Filename: "testfile.txt",
				Mode:     ModeNetASCII,
			},
			want: []byte{
				0x00, 0x02, // OpCode for WRQ
				't', 'e', 's', 't', 'f', 'i', 'l', 'e', '.', 't', 'x', 't', 0x00, // Filename
				'n', 'e', 't', 'a', 's', 'c', 'i', 'i', 0x00, // Mode
			},
			wantErr: false,
		},
		{
			name: "DATA Packet",
			packet: tftpPacket{
				OpCode: OpDATA,
				Block:  1,
				Data:   []byte("Hello, TFTP!"),
			},
			want: []byte{
				0x00, 0x03, // OpCode for DATA
				0x00, 0x01, // Block number
				'H', 'e', 'l', 'l', 'o', ',', ' ', 'T', 'F', 'T', 'P', '!', // Data
			},
			wantErr: false,
		},
		{
			name: "ACK Packet",
			packet: tftpPacket{
				OpCode: OpACK,
				Block:  1,
			},
			want: []byte{
				0x00, 0x04, // OpCode for ACK
				0x00, 0x01, // Block number
			},
			wantErr: false,
		},
		{
			name: "ERROR Packet",
			packet: tftpPacket{
				OpCode:    OpERROR,
				ErrorCode: ErrorNotDefined,
				ErrMsg:    "error",
			},
			want: []byte{
				0x00, 0x05, // OpCode for ERROR
				0x00, 0x00, // Error code (Not Defined)
				'e', 'r', 'r', 'o', 'r', 0x00, // Error message
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := tc.packet.MarshallPacket()
			if (err != nil) != tc.wantErr {
				t.Errorf("MarshalPacket() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if diff := cmp.Diff(tc.want, data); diff != "" {
				t.Errorf("MarshalPacket() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
