package main

import (
	"fmt"
	"gtftp/tftp"
)

func main() {
	// TFTPクライアントの初期化や使用例をここに追加できます。
	// 例えば、NewClient()を呼び出して新しいクライアントを作成するなど。
	client := tftp.NewClient("localhost:69")
	// ここでclientを使用してTFTP操作を行うことができます。

	//data, err := client.Recv("text.txt", "octet")
	data, err := client.Recv("1m.bin", "octet")
	if err != nil {
		panic(err)
	}
	// 受信したデータを処理する
	fmt.Println("Received data length:", len(data))
	//fmt.Println("Received data:", string(data))
}
