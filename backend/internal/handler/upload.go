package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/ndmt1at21/devlog/backend/internal/apierr"
	"github.com/ndmt1at21/devlog/backend/internal/domain"
	"github.com/ndmt1at21/devlog/backend/internal/platform/id"
	"github.com/ndmt1at21/devlog/backend/internal/upload"
)

// Image bytes never transit the API (JSON bodies are capped at 1 MiB): this
// endpoint only authorizes and returns a presigned direct-to-bucket PUT. The
// declared content type and size are part of the signature, so the client
// can't swap them after validation.
const (
	maxImageBytes = 5 << 20
	presignExpiry = 10 * time.Minute
)

// imageExts maps the allowed upload content types to the stored extension.
// SVG is deliberately excluded: served as a document it can execute script.
var imageExts = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
	"image/avif": ".avif",
}

type uploadInput struct {
	Type string `json:"type"` // MIME content type of the file
	Size int64  `json:"size"` // exact byte size of the file
}

// uploadTicket is the response of POST /uploads: PUT the bytes to UploadURL
// (with matching Content-Type), then embed PublicURL in the article body.
type uploadTicket struct {
	UploadURL string `json:"uploadUrl"`
	PublicURL string `json:"publicUrl"`
}

// createUpload authorizes an image upload for article authoring — same gate as
// publishing (session + IAM "articles:create") — and presigns the object PUT.
func (a *API) createUpload(w http.ResponseWriter, r *http.Request) {
	u, ok := userFrom(r.Context())
	if !ok {
		writeError(w, r, apierr.ErrUnauthorized)
		return
	}
	if a.Auth == nil {
		writeError(w, r, apierr.ErrAuthNotConfigured)
		return
	}
	allowed, err := a.Auth.CheckPermissions(r.Context(), u.Access, []string{articleCreatePermission})
	if err != nil {
		writeError(w, r, apierr.ErrAuthUpstream)
		return
	}
	if !allowed {
		writeError(w, r, apierr.ErrArticleForbidden)
		return
	}
	if !a.Cfg.UploadsEnabled() {
		writeError(w, r, apierr.ErrUploadNotConfigured)
		return
	}

	var in uploadInput
	if !decodeJSON(w, r, &in) {
		return
	}
	ext, ok := imageExts[in.Type]
	if !ok {
		writeError(w, r, apierr.ErrUploadType)
		return
	}
	if in.Size <= 0 || in.Size > maxImageBytes {
		writeError(w, r, apierr.ErrUploadTooLarge)
		return
	}

	// Time-ordered random key; the author never controls the object path.
	key := "img/" + id.NewV7() + ext
	signer := upload.Signer{
		Endpoint:  a.Cfg.S3Endpoint,
		Bucket:    a.Cfg.S3Bucket,
		Region:    a.Cfg.S3Region,
		AccessKey: a.Cfg.S3AccessKeyID,
		SecretKey: a.Cfg.S3SecretAccessKey,
	}
	uploadURL, err := signer.PresignPut(key, in.Type, in.Size, presignExpiry, time.Now())
	if err != nil {
		writeError(w, r, apierr.ErrUploadCreate)
		return
	}
	writeJSON(w, r, http.StatusCreated, uploadTicket{
		UploadURL: uploadURL,
		PublicURL: a.Cfg.ImageBaseURL + "/" + key,
	})
}

// checkImageHosts enforces that every img block points at the configured
// public image origin, so bodies can only embed images that went through the
// upload flow. Without IMAGE_BASE_URL (dev), any valid image URL passes.
func (a *API) checkImageHosts(blocks []domain.Block) error {
	if a.Cfg.ImageBaseURL == "" {
		return nil
	}
	for _, b := range blocks {
		if b.Type == "img" && !strings.HasPrefix(b.Src, a.Cfg.ImageBaseURL+"/") {
			return apierr.ErrImageHost
		}
	}
	return nil
}
