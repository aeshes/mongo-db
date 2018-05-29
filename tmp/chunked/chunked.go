package main

import "net/http"
import "bytes"
import "strconv"
import "fmt"

import "os"
import "math"

type Dummy struct {
	prevPart int
	sent     int
	fullSize int64
}

func check(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func sendChunk(chunk []byte, size int, endpoint string, d Dummy) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(chunk))
	check(err)

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Content-Length", strconv.Itoa(size))
	req.Header.Set("Content-Range", makeContentRange(d))
	req.Header.Set("fileTempId", "5b0abdf267b3db286b88be22")
	res, err := client.Do(req)
	check(err)

	fmt.Println(res)
}

func makeContentRange(d Dummy) string {
	header := "bytes "
	header += strconv.Itoa(d.prevPart)
	header += ("-" + strconv.Itoa(d.sent-1))
	header += ("/" + strconv.FormatInt(d.fullSize, 10))

	return header
}

func sendFile(f string, endpoint string) {
	file, err := os.Open(f)
	check(err)

	defer file.Close()

	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()
	const fileChunk = 0.25 * (1 << 20)
	totalPartsNum := uint64(math.Ceil(float64(fileSize) / float64(fileChunk)))

	fmt.Printf("Splitting to %d pieces.\n", totalPartsNum)

	var sent int
	prevPart := 0
	for i := uint64(0); i < totalPartsNum; i++ {
		partSize := int(math.Min(fileChunk, float64(fileSize-int64(i*fileChunk))))
		sent = sent + partSize
		partBuffer := make([]byte, partSize)

		fmt.Printf("Part size is %d \n", partSize)

		file.Read(partBuffer)
		d := Dummy{prevPart, sent, fileSize}
		sendChunk(partBuffer, partSize, endpoint, d)
		prevPart = sent
	}
}

func main() {
	endpoint := "http://localhost:3000/commonfs/createWithBuilder/appendBytes"
	file := "/home/user/go/src/chunked/opengl.pdf"
	sendFile(file, endpoint)
}
