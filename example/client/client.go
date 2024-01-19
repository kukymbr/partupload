package main

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"sync"
)

func main() {
	f, err := os.Open("./example/client/cat.png")
	if err != nil {
		panic(err)
	}

	stat, err := f.Stat()
	if err != nil {
		panic(err)
	}

	chunkSize := 1000000
	chunksCount := int(math.Ceil(float64(stat.Size()) / float64(chunkSize)))
	uploadID := fmt.Sprintf("upload_%d", rand.Intn(12))

	fmt.Printf("File name: %s\n", stat.Name())
	fmt.Printf("File size: %d\n", stat.Size())
	fmt.Printf("Upload ID: %s\n", uploadID)
	fmt.Printf("Chunk size: %d\n", chunkSize)
	fmt.Printf("Chunks count: %d\n", chunksCount)

	wg := &sync.WaitGroup{}

	for chunkN := 0; chunkN < chunksCount; chunkN++ {
		data := make([]byte, 0, chunkSize)
		buf := make([]byte, chunkSize)

		if _, err = f.Seek(int64(chunkN*chunkSize), io.SeekStart); err != nil {
			panic(err)
		}

		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}

		data = buf[:n]

		wg.Add(1)

		go func(chunkN int) {
			defer wg.Done()

			req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/", bytes.NewReader(data))
			if err != nil {
				panic(err)
			}

			req.Header.Set("Part-Upload-ID", uploadID)
			req.Header.Set("Part-Upload-Chunk-Num", fmt.Sprint(chunkN))
			req.Header.Set("Part-Upload-Chunks-Count", fmt.Sprint(chunksCount))
			req.Header.Set("Part-Upload-Origin-Name", stat.Name())
			req.Header.Set("Part-Upload-Origin-Size", fmt.Sprint(stat.Size()))

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				panic(err)
			}

			respBody, _ := io.ReadAll(resp.Body)

			fmt.Printf("Response code: %d\n", resp.StatusCode)
			fmt.Printf("Response body: %s\n", respBody)
		}(chunkN)
	}

	wg.Wait()
}
