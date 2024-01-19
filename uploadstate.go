package partupload

import (
	"net/http"
	"strconv"
	"strings"
)

// UploadState is an upload process state.
type UploadState struct {
	UploadID string

	ChunkNum    uint64
	ChunksCount uint64

	OriginName string
	OriginSize uint64

	Status Status

	filePath string
}

// GetTargetFilePath returns the uploaded target file path.
func (s *UploadState) GetTargetFilePath() string {
	return s.filePath
}

func uploadStateFromReq(req *http.Request) (*UploadState, error) {
	if req.Method != http.MethodPost && req.Method != http.MethodPatch {
		return nil, ErrMethodNotAllowed
	}

	state := &UploadState{
		UploadID:   strings.TrimSpace(req.Header.Get(HeaderUploadID)),
		OriginName: strings.TrimSpace(req.Header.Get(HeaderOriginName)),
		Status:     StatusProgress,
	}

	if state.UploadID == "" {
		return nil, ErrNoUploadID
	}

	chunkN, err := strconv.ParseUint(req.Header.Get(HeaderChunkNum), 10, 64)
	if err != nil {
		return nil, ErrNoChunkNum
	}

	chunksCount, err := strconv.ParseUint(req.Header.Get(HeaderChunksCount), 10, 64)
	if err != nil {
		return nil, ErrNoChunksCount
	}

	state.ChunkNum = chunkN
	state.ChunksCount = chunksCount
	state.OriginSize, _ = strconv.ParseUint(req.Header.Get(HeaderOriginSize), 10, 64)

	return state, nil
}
