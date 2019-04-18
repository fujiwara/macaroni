package macaroni

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var testMetadataPath = "/v3/20ed294b-9e28-499d-b4f6-33848085dd98"

func TestECSMetadata(t *testing.T) {
	defer func() { env = nil }()

	ts := newECSMetadataEndpoint()
	defer ts.Close()

	env = map[string]string{
		"ECS_CONTAINER_METADATA_URI": ts.URL + testMetadataPath,
	}
	meta, err := getECSMetadata()
	if err != nil {
		t.Error(err)
	}
	expected := &ECSMetadata{
		Cluster:       "api",
		TaskARN:       "arn:aws:ecs:ap-northeast-1:999999999999:task/965d53cd-8dd8-483a-b9c6-f0910c3892a4",
		ContainerName: "app",
	}
	if diff := cmp.Diff(meta, expected); diff != "" {
		t.Error(diff)
	}
}

func TestECSMetadataNull(t *testing.T) {
	defer func() { env = nil }()

	env = map[string]string{
		"ECS_CONTAINER_METADATA_URI": "",
	}
	meta, err := getECSMetadata()
	if err != nil {
		t.Error(err)
	}
	if meta != nil {
		t.Errorf("unexpected meta %#v", meta)
	}
}

func newECSMetadataEndpoint() *httptest.Server {
	mux := http.NewServeMux()
	genHandler := func(filename string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			f, err := os.Open(filename)
			if err != nil {
				w.WriteHeader(404)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			io.Copy(w, f)
		}
	}
	mux.HandleFunc(testMetadataPath, genHandler("test/container.json"))
	mux.HandleFunc(testMetadataPath+"/task", genHandler("test/task.json"))

	return httptest.NewServer(mux)
}
