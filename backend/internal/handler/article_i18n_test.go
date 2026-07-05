package handler_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

// bilingualArticle has Vietnamese as the primary language and one English
// translation supplied at creation time.
const bilingualArticle = `{
	"lang": "vi",
	"title": "Xin chào",
	"category": "Backend",
	"tags": ["go"],
	"format": "markdown",
	"content": "# Mở đầu\n\nĐây là bản tiếng Việt.",
	"translations": [
		{
			"lang": "en",
			"title": "Hello",
			"format": "markdown",
			"content": "# Intro\n\nThis is the English version."
		}
	]
}`

// i18nDetail is the subset of the detail response the bilingual tests assert on.
type i18nDetail struct {
	Slug           string   `json:"slug"`
	Lang           string   `json:"lang"`
	Title          string   `json:"title"`
	AvailableLangs []string `json:"availableLangs"`
	Body           []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"body"`
	Translations map[string]struct {
		Title   string `json:"title"`
		Excerpt string `json:"excerpt"`
		Body    []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"body"`
	} `json:"translations"`
}

func TestCreateBilingualArticle(t *testing.T) {
	srv, client := newAuthedClient(t, true)

	status, env := postJSON(t, client, srv.URL+v1+"/articles", bilingualArticle)
	if status != http.StatusCreated || env.Code != 0 {
		t.Fatalf("create: status=%d code=%d msg=%q, want 201/0", status, env.Code, env.Message)
	}
	var created i18nDetail
	if err := json.Unmarshal(env.Data, &created); err != nil {
		t.Fatal(err)
	}
	if created.Lang != "vi" {
		t.Errorf("primary lang = %q, want vi", created.Lang)
	}
	if len(created.AvailableLangs) != 2 || created.AvailableLangs[0] != "vi" || created.AvailableLangs[1] != "en" {
		t.Errorf("availableLangs = %v, want [vi en]", created.AvailableLangs)
	}
	en, ok := created.Translations["en"]
	if !ok {
		t.Fatalf("missing en translation in %v", created.Translations)
	}
	if en.Title != "Hello" {
		t.Errorf("en title = %q, want Hello", en.Title)
	}
	if en.Excerpt == "" {
		t.Error("en excerpt should be derived from the english body")
	}

	// Publicly readable, both languages present.
	var got i18nDetail
	if s, _ := getClientEnv(t, client, srv.URL+v1+"/articles/"+created.Slug, &got); s != 200 {
		t.Fatalf("get status = %d", s)
	}
	if got.Title != "Xin chào" {
		t.Errorf("primary title = %q, want Xin chào", got.Title)
	}
	if got.Translations["en"].Title != "Hello" {
		t.Errorf("en title after get = %q", got.Translations["en"].Title)
	}
}

// TestTranslateLater creates a Vietnamese-only article, then adds the English
// translation via an edit — the "publish now, translate later" flow.
func TestTranslateLater(t *testing.T) {
	srv, client := newAuthedClient(t, true)

	create := `{"lang":"vi","title":"Một ngôn ngữ","category":"Backend","tags":["go"],"format":"markdown","content":"# Chào\n\nChỉ có tiếng Việt."}`
	status, env := postJSON(t, client, srv.URL+v1+"/articles", create)
	if status != http.StatusCreated {
		t.Fatalf("create status = %d msg=%q", status, env.Message)
	}
	var created i18nDetail
	if err := json.Unmarshal(env.Data, &created); err != nil {
		t.Fatal(err)
	}
	if len(created.AvailableLangs) != 1 || created.Translations != nil {
		t.Fatalf("new article should be single-language, got langs=%v translations=%v", created.AvailableLangs, created.Translations)
	}

	// Edit to add the English translation.
	edit := `{"lang":"vi","title":"Một ngôn ngữ","category":"Backend","tags":["go"],"format":"markdown","content":"# Chào\n\nChỉ có tiếng Việt.","translations":[{"lang":"en","title":"One language","format":"markdown","content":"# Hi\n\nNow in English too."}]}`
	status, env = putJSON(t, client, srv.URL+v1+"/articles/"+created.Slug, edit)
	if status != http.StatusOK {
		t.Fatalf("edit status = %d code=%d msg=%q", status, env.Code, env.Message)
	}

	var got i18nDetail
	if s, _ := getClientEnv(t, client, srv.URL+v1+"/articles/"+created.Slug, &got); s != 200 {
		t.Fatalf("get status = %d", s)
	}
	if len(got.AvailableLangs) != 2 {
		t.Errorf("after translate-later availableLangs = %v, want 2", got.AvailableLangs)
	}
	if got.Translations["en"].Title != "One language" {
		t.Errorf("added en title = %q", got.Translations["en"].Title)
	}
}

func TestBilingualValidation(t *testing.T) {
	srv, client := newAuthedClient(t, true)

	cases := map[string]string{
		"duplicate primary lang": `{"lang":"vi","title":"T","category":"Backend","format":"markdown","content":"x","translations":[{"lang":"vi","title":"T2","format":"markdown","content":"y"}]}`,
		"invalid translation lang": `{"lang":"vi","title":"T","category":"Backend","format":"markdown","content":"x","translations":[{"lang":"fr","title":"T2","format":"markdown","content":"y"}]}`,
		"translation missing title": `{"lang":"vi","title":"T","category":"Backend","format":"markdown","content":"x","translations":[{"lang":"en","title":"","format":"markdown","content":"y"}]}`,
		"invalid primary lang":       `{"lang":"de","title":"T","category":"Backend","format":"markdown","content":"x"}`,
	}
	for name, body := range cases {
		status, env := postJSON(t, client, srv.URL+v1+"/articles", body)
		if status != http.StatusBadRequest {
			t.Errorf("%s: status=%d code=%d, want 400", name, status, env.Code)
		}
	}
}
