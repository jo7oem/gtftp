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

func TestUnmarshalPacket(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    *tftpPacket
		wantErr bool
	}{
		{
			name: "RRQ Packet",
			data: []byte{
				0x00, 0x01, // OpCode for RRQ
				't', 'e', 's', 't', 'f', 'i', 'l', 'e', '.', 't', 'x', 't', 0x00, // Filename
				'o', 'c', 't', 'e', 't', 0x00, // Mode
			},
			want: &tftpPacket{
				OpCode:   OpRRQ,
				Filename: "testfile.txt",
				Mode:     ModeOctet,
			},
			wantErr: false,
		},
		{
			name: "WRQ Packet",
			data: []byte{
				0x00, 0x02, // OpCode for WRQ
				't', 'e', 's', 't', 'f', 'i', 'l', 'e', '.', 't', 'x', 't', 0x00, // Filename
				'n', 'e', 't', 'a', 's', 'c', 'i', 'i', 0x00, // Mode
			},
			want: &tftpPacket{
				OpCode:   OpWRQ,
				Filename: "testfile.txt",
				Mode:     ModeNetASCII,
			},
			wantErr: false,
		},
		{
			name: "Broken Packet (too short)",
			data: []byte{
				0x00, 0x02, // OpCode for WRQ
			},
			wantErr: true,
		},
		{
			name: "Broken Packet (RRQ broken filename)",
			data: []byte{
				0x00, 0x01, // OpCode for RRQ
				't', 'e', 's', 't', 'f', 'i', 'l', 'e', '.', 't', 'x', 't', // Filename without null terminator
			},
			wantErr: true,
		},
		{
			name: "Broken Packet (RRQ broken mode)",
			data: []byte{
				0x00, 0x01, // OpCode for RRQ
				't', 'e', 's', 't', 'f', 'i', 'l', 'e', '.', 't', 'x', 't', 0x00, // Filename without null terminator
				'o', 'c', 't', 'e', 't', // Mode without null terminator
			},
			wantErr: true,
		},
		{
			name: "Broken Packet (mode with extra char)",
			data: []byte{
				0x00, 0x01, // OpCode for RRQ
				't', 'e', 's', 't', 'f', 'i', 'l', 'e', '.', 't', 'x', 't', 0x00, // Filename without null terminator
				'o', 'c', 't', 'e', 't', 0x00, 'a', // Mode without null terminator
			},
			wantErr: false,
			want: &tftpPacket{
				OpCode:   OpRRQ,
				Filename: "testfile.txt",
				Mode:     ModeOctet,
			},
		},
		{
			name: "DATA Packet",
			data: []byte{
				0x00, 0x03, // OpCode for DATA
				0xff, 0xff, // Block number
				'H', 'e', 'l', 'l', 'o', ',', ' ', 'T', 'F', 'T', 'P', '!', // Data
			},
			want: &tftpPacket{
				OpCode: OpDATA,
				Block:  0xffff,
				Data:   []byte("Hello, TFTP!"),
			},
			wantErr: false,
		},
		{
			name: "ACK Packet",
			data: []byte{
				0x00, 0x04, // OpCode for ACK
				0x00, 0x01, // Block number
			},
			want: &tftpPacket{
				OpCode: OpACK,
				Block:  1,
			},
		},
		{
			name: "ACK Packet (broken)",
			data: []byte{
				0x00, 0x04, // OpCode for ACK
				0x00, 0x01, 0x00, // Block number
			},
			wantErr: true,
		},
		{
			name: "ERROR Packet",
			data: []byte{
				0x00, 0x05, // OpCode for ERROR
				0x00, 0x00, // Error code (Not Defined)
				'e', 'r', 'r', 'o', 'r', 0x00, // Error message
			},
			want: &tftpPacket{
				OpCode:    OpERROR,
				ErrorCode: ErrorNotDefined,
				ErrMsg:    "error",
			},
			wantErr: false,
		},
		{
			name: "ERROR Packet (broken, no null terminator)",
			data: []byte{
				0x00, 0x05, // OpCode for ERROR
				0x00, 0x00, // Error code (Not Defined)
				'e', 'r', 'r', 'o', 'r', // Error message
			},
			wantErr: true,
		},
		{
			name: "ERROR Packet (extra char)",
			data: []byte{
				0x00, 0x05, // OpCode for ERROR
				0x00, 0x00, // Error code (Not Defined)
				'e', 'r', 'r', 'o', 'r', 0x00, 'a',
			},
			wantErr: false,
			want: &tftpPacket{
				OpCode:    OpERROR,
				ErrorCode: ErrorNotDefined,
				ErrMsg:    "error",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			packet, err := UnmarshallPacket(tc.data)
			if (err != nil) != tc.wantErr {
				t.Errorf("UnmarshallPacket() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if diff := cmp.Diff(tc.want, packet); diff != "" {
				t.Errorf("UnmarshallPacket() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
