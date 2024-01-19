package main

import (
	"encoding/json"
	"net/http"

	"github.com/kukymbr/partupload"
)

func main() {
	handler, err := partupload.NewFileStorageReceiver("./example/server/uploads")
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		if req.Method == http.MethodOptions {
			return
		}

		state, err := handler.Receive(req)
		if err != nil {
			httpErr := partupload.HttpErrorFromAny(err)

			w.WriteHeader(httpErr.GetStatus())
			_, _ = w.Write([]byte(httpErr.Error()))

			return
		}

		if state.Status != partupload.StatusComplete {
			data, _ := json.Marshal(state)

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(data)

			return
		}

		_, _ = w.Write([]byte("Done: " + state.GetTargetFilePath()))
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
