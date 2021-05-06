package utils

import (
	"fmt"
	"os"
)

func SplitFile(file *os.File, bNumber int) ([]string, []int) {
	var r []string
	var n []int
	fs, _ := file.Stat()
	totalBytes := fs.Size()
	fmt.Printf("totalBytes: %d\n", totalBytes)
	blockBytes := totalBytes / (int64(bNumber)) + 1
	fmt.Printf("blockBytes: %d\n", blockBytes)
	file.Seek(0, 0)
	data := make([]byte, totalBytes+1)
	file.Read(data)
	for i:=0; i < bNumber; i++ {
		var blockData = make([]byte, blockBytes + 1)
		if i != bNumber - 1 {
			blockData = data[int64(i)*blockBytes: int64(i)*blockBytes + blockBytes]
			fmt.Printf("index: %d, length: %d\n", i, len(blockData))
		} else {
			blockData = data[int64(i)*blockBytes:]
			fmt.Printf("index: %d, length: %d\n", i, len(blockData))
		}
		bName := fmt.Sprintf("%s.b-%d", file.Name(), i)
		bfile, _ := os.OpenFile(bName, os.O_CREATE | os.O_WRONLY, 0600)
		bfile.Write(blockData)
		bfile.Close()
		r = append(r, bName)
		n = append(n, i)
	}
	return r, n
}