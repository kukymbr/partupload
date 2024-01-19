// Package partupload is a tool to receive
// partial uploads over the HTTP(S).
package partupload

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	HeaderUploadID    = "Part-Upload-ID"
	HeaderChunkNum    = "Part-Upload-Chunk-Num"
	HeaderChunksCount = "Part-Upload-Chunks-Count"
	HeaderOriginName  = "Part-Upload-Origin-Name"
	HeaderOriginSize  = "Part-Upload-Origin-Size"
)

var (
	ErrMethodNotAllowed = NewHttpError(http.StatusMethodNotAllowed)
	ErrNoUploadID       = NewHttpError(http.StatusBadRequest, "no upload ID given")
	ErrNoChunkNum       = NewHttpError(http.StatusBadRequest, "no upload chunk number given")
	ErrNoChunksCount    = NewHttpError(http.StatusBadRequest, "no chunks count given")
)

type Receiver interface {
	Receive(req *http.Request) (*UploadState, error)
	Cancel(req *http.Request) (*UploadState, error)
}

func NewFileStorageReceiver(targetDir string) (Receiver, error) {
	targetDir = filepath.Clean(targetDir)

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, fmt.Errorf("target dir %s does not exist and cannot be created: %w", targetDir, err)
	}

	return &fileStorage{dir: targetDir}, nil
}

type fileStorage struct {
	dir string
}

func (f *fileStorage) Receive(req *http.Request) (*UploadState, error) {
	state, err := uploadStateFromReq(req)
	if err != nil {
		return nil, err
	}

	state.filePath = f.getFilePath(state.UploadID)
	chunkPath := f.getPartPath(state.UploadID, state.ChunkNum)

	if err := saveToFile(req.Body, chunkPath); err != nil {
		return nil, NewHttpError(http.StatusInternalServerError, err)
	}

	parts, err := f.getCompletePartsPaths(state.UploadID)
	if err != nil {
		return nil, NewHttpError(http.StatusInternalServerError, err)
	}

	if uint64(len(parts)) == state.ChunksCount {
		state.Status = StatusComplete

		if err := joinParts(state.filePath, parts...); err != nil {
			return nil, NewHttpError(http.StatusInternalServerError, err)
		}

		_ = cleanup(parts...)
	}

	return state, nil
}

func (f *fileStorage) Cancel(req *http.Request) (*UploadState, error) {
	//TODO implement me
	panic("implement me")
}

func (f *fileStorage) getFilePath(uploadID string) string {
	return filepath.Join(f.dir, getSHA1(uploadID))
}

func (f *fileStorage) getPartPath(uploadID string, chunkN uint64) string {
	return fmt.Sprintf("%s.%d", f.getFilePath(uploadID), chunkN)
}

func (f *fileStorage) getCompletePartsPaths(uploadID string) ([]string, error) {
	paths := make([]string, 0)
	uploadID = getSHA1(uploadID)

	err := filepath.WalkDir(f.dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == f.dir {
			return nil
		}

		if entry.IsDir() {
			return filepath.SkipDir
		}

		nameParts := strings.Split(entry.Name(), ".")
		if len(nameParts) != 2 || nameParts[0] != uploadID {
			return nil
		}

		_, err = strconv.ParseUint(nameParts[1], 10, 64)
		if err != nil {
			return nil
		}

		if isFileLocked(path) {
			return nil
		}

		paths = append(paths, path)

		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(paths, func(i, j int) bool {
		n1, _ := strconv.ParseUint(strings.Split(paths[i], ".")[1], 10, 64)
		n2, _ := strconv.ParseUint(strings.Split(paths[j], ".")[1], 10, 64)

		return n1 < n2
	})

	return paths, nil
}
