package handler_test

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/ndmt1at21/devlog/backend/internal/config"
)

// s3TestCfg fully configures uploads against a fake R2 account.
func s3TestCfg() config.Config {
	return config.Config{
		DBDriver:          "memory",
		S3Endpoint:        "https://acc.r2.cloudflarestorage.com",
		S3Bucket:          "devlog-images",
		S3Region:          "auto",
		S3AccessKeyID:     "test-ak",
		S3SecretAccessKey: "test-sk",
		ImageBaseURL:      "https://img.example.com",
	}
}

func TestCreateUploadRequiresLogin(t *testing.T) {
	srv, _ := newAuthedClientCfg(t, true, s3TestCfg())
	status, env := postJSON(t, http.DefaultClient, srv.URL+v1+"/uploads", `{"type":"image/png","size":100}`)
	if status != http.StatusUnauthorized || env.Code != 1002 {
		t.Fatalf("anonymous upload: status=%d code=%d, want 401/1002", status, env.Code)
	}
}

func TestCreateUploadForbiddenWithoutPermission(t *testing.T) {
	srv, client := newAuthedClientCfg(t, false, s3TestCfg())
	status, env := postJSON(t, client, srv.URL+v1+"/uploads", `{"type":"image/png","size":100}`)
	if status != http.StatusForbidden || env.Code != 3005 {
		t.Fatalf("upload without permission: status=%d code=%d, want 403/3005", status, env.Code)
	}
}

func TestCreateUploadNotConfigured(t *testing.T) {
	srv, client := newAuthedClient(t, true) // default cfg: no S3 settings
	status, env := postJSON(t, client, srv.URL+v1+"/uploads", `{"type":"image/png","size":100}`)
	if status != http.StatusServiceUnavailable || env.Code != 3007 {
		t.Fatalf("unconfigured upload: status=%d code=%d, want 503/3007", status, env.Code)
	}
}

func TestCreateUploadValidation(t *testing.T) {
	srv, client := newAuthedClientCfg(t, true, s3TestCfg())
	cases := map[string]struct {
		body string
		code int
	}{
		"svg rejected":  {`{"type":"image/svg+xml","size":100}`, 3008},
		"non-image":     {`{"type":"application/pdf","size":100}`, 3008},
		"zero size":     {`{"type":"image/png","size":0}`, 3009},
		"too large":     {`{"type":"image/png","size":5242881}`, 3009},
		"unknown field": {`{"type":"image/png","size":1,"key":"../x"}`, 1000},
	}
	for name, tc := range cases {
		status, env := postJSON(t, client, srv.URL+v1+"/uploads", tc.body)
		if status < 400 || env.Code != tc.code {
			t.Errorf("%s: status=%d code=%d, want code %d", name, status, env.Code, tc.code)
		}
	}
}

func TestCreateUploadTicket(t *testing.T) {
	srv, client := newAuthedClientCfg(t, true, s3TestCfg())
	status, env := postJSON(t, client, srv.URL+v1+"/uploads", `{"type":"image/webp","size":123456}`)
	if status != http.StatusCreated || env.Code != 0 {
		t.Fatalf("upload ticket: status=%d code=%d message=%q, want 201/0", status, env.Code, env.Message)
	}
	var ticket struct {
		UploadURL string `json:"uploadUrl"`
		PublicURL string `json:"publicUrl"`
	}
	if err := json.Unmarshal(env.Data, &ticket); err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(ticket.PublicURL, "https://img.example.com/img/") || !strings.HasSuffix(ticket.PublicURL, ".webp") {
		t.Errorf("publicUrl = %q", ticket.PublicURL)
	}

	u, err := url.Parse(ticket.UploadURL)
	if err != nil {
		t.Fatalf("uploadUrl does not parse: %v", err)
	}
	if u.Host != "acc.r2.cloudflarestorage.com" || !strings.HasPrefix(u.Path, "/devlog-images/img/") {
		t.Errorf("uploadUrl host/path = %s %s", u.Host, u.Path)
	}
	// The object key of both URLs must agree, and the signature must pin the
	// validated content type and byte size.
	key := strings.TrimPrefix(u.Path, "/devlog-images/")
	if got := strings.TrimPrefix(ticket.PublicURL, "https://img.example.com/"); got != key {
		t.Errorf("public key %q != upload key %q", got, key)
	}
	q := u.Query()
	if q.Get("X-Amz-SignedHeaders") != "content-length;content-type;host" {
		t.Errorf("SignedHeaders = %q", q.Get("X-Amz-SignedHeaders"))
	}
	if len(q.Get("X-Amz-Signature")) != 64 {
		t.Errorf("signature = %q", q.Get("X-Amz-Signature"))
	}
}

// TestCreateArticleImageHostPolicy: with IMAGE_BASE_URL set, article bodies may
// only embed images under that origin.
func TestCreateArticleImageHostPolicy(t *testing.T) {
	srv, client := newAuthedClientCfg(t, true, s3TestCfg())

	foreign := `{"title":"Ảnh ngoài","category":"Backend","format":"blocks",
		"body":[{"type":"img","src":"https://evil.example.net/x.png","alt":"x"}]}`
	status, env := postJSON(t, client, srv.URL+v1+"/articles", foreign)
	if status != http.StatusBadRequest || env.Code != 3011 {
		t.Fatalf("foreign image host: status=%d code=%d, want 400/3011", status, env.Code)
	}

	ok := `{"title":"Ảnh nhà","category":"Backend","format":"markdown",
		"content":"Mở đầu.\n\n![Sơ đồ](https://img.example.com/img/a.png)"}`
	status, env = postJSON(t, client, srv.URL+v1+"/articles", ok)
	if status != http.StatusCreated || env.Code != 0 {
		t.Fatalf("own image host: status=%d code=%d message=%q, want 201/0", status, env.Code, env.Message)
	}
	var created struct {
		Body []struct {
			Type string `json:"type"`
			Src  string `json:"src"`
			Alt  string `json:"alt"`
		} `json:"body"`
	}
	if err := json.Unmarshal(env.Data, &created); err != nil {
		t.Fatal(err)
	}
	if len(created.Body) != 2 || created.Body[1].Type != "img" ||
		created.Body[1].Src != "https://img.example.com/img/a.png" || created.Body[1].Alt != "Sơ đồ" {
		t.Fatalf("created body = %+v", created.Body)
	}
}
