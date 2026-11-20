package camera

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"camera_api/pkg/adapter/database/model"
	"camera_api/pkg/api/request"
	"camera_api/pkg/api/response"
	"camera_api/pkg/utils/crypto"

	"github.com/andriantp/camera"
)

// newTestTarget starts a fake ISAPI camera and returns an adapter plus a
// target pointing at it with an encrypted credential.
func newTestTarget(t *testing.T, handler http.HandlerFunc) (Camera, CameraTarget) {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	adapter, err := NewCamera(Setting{Key: key, Timeout: 2 * time.Second, Download: 2 * time.Second})
	if err != nil {
		t.Fatalf("NewCamera: %v", err)
	}

	credential, err := adapter.EncryptCredential(camera.Credential{
		Host:     strings.TrimPrefix(server.URL, "http://"),
		Username: "admin",
		Password: "secret",
	})
	if err != nil {
		t.Fatalf("EncryptCredential: %v", err)
	}

	return adapter, CameraTarget{Credential: *credential, Protocol: model.ProtocolISAPI}
}

func TestGetCapabilityISAPIWithoutChannel(t *testing.T) {
	var (
		mu    sync.Mutex
		paths []string
	)

	adapter, target := newTestTarget(t, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		paths = append(paths, r.URL.Path)
		mu.Unlock()

		http.Error(w, "not found", http.StatusNotFound)
	})

	// target.Channel is nil: this used to panic with a nil pointer dereference.
	_, err := adapter.GetCapability(context.Background(), target)

	var apiErr response.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusBadGateway {
		t.Fatalf("err = %v, want 502 APIError from the failing camera", err)
	}

	mu.Lock()
	defer mu.Unlock()

	for _, p := range paths {
		if strings.Contains(p, "/channels/1/") {
			return
		}
	}

	t.Fatalf("expected a request using the default channel 1, got %v", paths)
}

func searchResultXML(total int, items ...string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<CMSearchResult>
	<searchID>test</searchID>
	<responseStatus>true</responseStatus>
	<responseStatusStrg>MORE</responseStatusStrg>
	<totalMatches>%d</totalMatches>
	<numOfMatches>%d</numOfMatches>
	<matchList>%s</matchList>
</CMSearchResult>`, total, len(items), strings.Join(items, ""))
}

func matchItemXML(start, end string) string {
	return fmt.Sprintf(`<searchMatchItem>
		<trackID>101</trackID>
		<timeSpan><startTime>%s</startTime><endTime>%s</endTime></timeSpan>
		<mediaSegmentDescriptor>
			<contentType>video</contentType>
			<codecType>H.264</codecType>
			<playbackURI>rtsp://camera/Streaming/tracks/101?starttime=x&amp;size=1048576</playbackURI>
		</mediaSegmentDescriptor>
	</searchMatchItem>`, start, end)
}

func findRequest() request.SearchContent {
	return request.SearchContent{
		From:  time.Date(2026, 8, 18, 0, 0, 0, 0, time.UTC),
		To:    time.Date(2026, 8, 18, 2, 0, 0, 0, time.UTC),
		Limit: 100,
	}
}

func TestFindContentStopsOnEmptyPage(t *testing.T) {
	var requests atomic.Int32

	adapter, target := newTestTarget(t, func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		// Buggy firmware: claims 5 matches but returns none.
		fmt.Fprint(w, searchResultXML(5))
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := adapter.FindContent(ctx, target, findRequest(), "20260818T010000Z_20260818T011000Z.mp4")

	var apiErr response.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusNotFound {
		t.Fatalf("err = %v, want 404 Content not found", err)
	}

	if n := requests.Load(); n != 1 {
		t.Fatalf("camera searched %d times, want 1 (must not loop on an empty page)", n)
	}
}

func TestFindContentPaginates(t *testing.T) {
	var requests atomic.Int32

	adapter, target := newTestTarget(t, func(w http.ResponseWriter, r *http.Request) {
		switch requests.Add(1) {
		case 1:
			fmt.Fprint(w, searchResultXML(2, matchItemXML("2026-08-18T00:00:00Z", "2026-08-18T00:10:00Z")))
		default:
			fmt.Fprint(w, searchResultXML(2, matchItemXML("2026-08-18T01:00:00Z", "2026-08-18T01:10:00Z")))
		}
	})

	record, err := adapter.FindContent(context.Background(), target, findRequest(), "20260818T010000Z_20260818T011000Z.mp4")
	if err != nil {
		t.Fatalf("FindContent: %v", err)
	}

	if record.Filename != "20260818T010000Z_20260818T011000Z.mp4" {
		t.Fatalf("filename = %q", record.Filename)
	}

	if n := requests.Load(); n != 2 {
		t.Fatalf("camera searched %d times, want 2", n)
	}
}
