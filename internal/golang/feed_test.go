package golang

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

const feedJSON = `[
 {"version":"go1.27.4","stable":true,"files":[
  {"filename":"go1.27.4.src.tar.gz","os":"","arch":"","sha256":"aa","kind":"source"},
  {"filename":"go1.27.4.linux-amd64.tar.gz","os":"linux","arch":"amd64","sha256":"a1","kind":"archive"},
  {"filename":"go1.27.4.linux-arm64.tar.gz","os":"linux","arch":"arm64","sha256":"a2","kind":"archive"},
  {"filename":"go1.27.4.windows-amd64.zip","os":"windows","arch":"amd64","sha256":"a3","kind":"archive"}]},
 {"version":"go1.26.11","stable":true,"files":[
  {"filename":"go1.26.11.linux-amd64.tar.gz","os":"linux","arch":"amd64","sha256":"b1","kind":"archive"},
  {"filename":"go1.26.11.linux-arm64.tar.gz","os":"linux","arch":"arm64","sha256":"b2","kind":"archive"}]},
 {"version":"go1.28rc1","stable":false,"files":[]}
]`

func serve(t *testing.T, code int, body string) string {
	t.Helper()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(s.Close)
	return s.URL
}

func TestFetchLatestIgnoresOlderAndUnstable(t *testing.T) { // AC5, AC14
	rel, err := FetchLatest(context.Background(), http.DefaultClient, serve(t, 200, feedJSON))
	if err != nil {
		t.Fatal(err)
	}
	if rel.Version.String() != "1.27.4" || rel.Files["amd64"].SHA256 != "a1" || rel.Files["arm64"].Filename != "go1.27.4.linux-arm64.tar.gz" {
		t.Errorf("unexpected release %+v", rel)
	}
}

func TestFetchLatestRejectsUnstableOnly(t *testing.T) { // AC5
	if _, err := FetchLatest(context.Background(), http.DefaultClient, serve(t, 200, `[{"version":"go1.28rc1","stable":false,"files":[]}]`)); err == nil {
		t.Error("expected error")
	}
}

func TestFetchLatestMissingArch(t *testing.T) { // AC7
	body := `[{"version":"go1.27.4","stable":true,"files":[{"filename":"go1.27.4.linux-amd64.tar.gz","os":"linux","arch":"amd64","sha256":"a1","kind":"archive"}]}]`
	if _, err := FetchLatest(context.Background(), http.DefaultClient, serve(t, 200, body)); err == nil {
		t.Error("expected error for missing arm64")
	}
}

func TestFetchLatestBadFeed(t *testing.T) { // AC8
	for name, url := range map[string]string{"500": serve(t, 500, "oops"), "json": serve(t, 200, "{"), "empty": serve(t, 200, "[]")} {
		if _, err := FetchLatest(context.Background(), http.DefaultClient, url); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}

func TestFetchLatestBadVersionString(t *testing.T) { // AC6
	body := `[{"version":"go1.27.4; rm -rf /","stable":true,"files":[]}]`
	if _, err := FetchLatest(context.Background(), http.DefaultClient, serve(t, 200, body)); err == nil {
		t.Error("expected error")
	}
}

func TestFetchLatestCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := FetchLatest(ctx, http.DefaultClient, serve(t, 200, feedJSON)); err == nil {
		t.Error("expected error")
	}
}
