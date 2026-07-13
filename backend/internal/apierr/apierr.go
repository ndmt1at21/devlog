// Package apierr defines the API's stable, unique error codes and the typed
// error carried through handlers. Each code is stable across releases so the
// frontend can map it to a localized message; the Message here is the server's
// default (also returned in the envelope as a fallback / for debugging).
package apierr

import "net/http"

// Error is a typed API error: a stable Code, the HTTP Status to respond with,
// and a default human Message
type Error struct {
	Code    int
	Status  int
	Message string
}

func (e *Error) Error() string { return e.Message }

// WithMessage returns a copy of the error with a context-specific message but
// the same Code/Status — useful for field-level validation where the code stays
// generic (ErrValidation) but the message names the field.
func (e *Error) WithMessage(msg string) *Error {
	c := *e
	c.Message = msg
	return &c
}

func def(code, status int, msg string) *Error {
	return &Error{Code: code, Status: status, Message: msg}
}

// CodeOK is the success code used in every non-error envelope.
const CodeOK = 0

// Registry. Codes are grouped by domain and MUST remain stable:
//
//	1xxx generic · 2xxx auth · 3xxx content · 4xxx comments · 5xxx pro · 6xxx coffee · 7xxx reactions
var (
	// --- generic (1xxx) ---
	ErrBadRequest   = def(1000, http.StatusBadRequest, "Dữ liệu gửi lên không hợp lệ.")
	ErrValidation   = def(1001, http.StatusBadRequest, "Dữ liệu chưa hợp lệ.")
	ErrUnauthorized = def(1002, http.StatusUnauthorized, "Bạn cần đăng nhập.")
	ErrForbidden    = def(1003, http.StatusForbidden, "Bạn không có quyền thực hiện.")
	ErrNotFound     = def(1004, http.StatusNotFound, "Không tìm thấy.")
	ErrInternal     = def(1006, http.StatusInternalServerError, "Đã có lỗi xảy ra. Vui lòng thử lại.")
	ErrUpstream     = def(1007, http.StatusBadGateway, "Không kết nối được dịch vụ.")
	ErrUnavailable  = def(1008, http.StatusServiceUnavailable, "Tính năng chưa được cấu hình.")

	// --- auth (2xxx) ---
	ErrAuthNotConfigured  = def(2000, http.StatusServiceUnavailable, "Đăng nhập chưa được cấu hình.")
	ErrInvalidCredentials = def(2001, http.StatusUnauthorized, "Email hoặc mật khẩu không đúng.")
	ErrAuthUpstream       = def(2002, http.StatusBadGateway, "Không kết nối được dịch vụ xác thực.")
	ErrUserInfo           = def(2003, http.StatusBadGateway, "Không lấy được thông tin người dùng.")
	ErrEmailTaken         = def(2004, http.StatusConflict, "Email này đã được đăng ký.")
	ErrInvalidEmail       = def(2005, http.StatusBadRequest, "Email chưa hợp lệ.")
	ErrWeakPassword       = def(2006, http.StatusBadRequest, "Mật khẩu cần tối thiểu 6 ký tự.")
	ErrSessionCreate      = def(2007, http.StatusInternalServerError, "Không tạo được phiên đăng nhập.")

	// --- content (3xxx) ---
	ErrArticleNotFound  = def(3000, http.StatusNotFound, "Không tìm thấy bài viết.")
	ErrArticleList      = def(3001, http.StatusInternalServerError, "Không tải được danh sách bài viết.")
	ErrFeaturedNotFound = def(3002, http.StatusNotFound, "Chưa có bài viết nổi bật.")
	ErrArticleLoad      = def(3003, http.StatusInternalServerError, "Không tải được bài viết.")
	ErrCategoryList     = def(3004, http.StatusInternalServerError, "Không tải được danh mục.")
	ErrArticleForbidden = def(3005, http.StatusForbidden, "Bạn không có quyền tạo bài viết.")
	ErrArticleCreate    = def(3006, http.StatusInternalServerError, "Không tạo được bài viết.")
	// Image uploads (presigned direct-to-bucket).
	ErrUploadNotConfigured = def(3007, http.StatusServiceUnavailable, "Tải ảnh chưa được cấu hình.")
	ErrUploadType          = def(3008, http.StatusBadRequest, "Chỉ hỗ trợ ảnh JPEG, PNG, WebP, GIF hoặc AVIF.")
	ErrUploadTooLarge      = def(3009, http.StatusBadRequest, "Ảnh tối đa 5 MB.")
	ErrUploadCreate        = def(3010, http.StatusInternalServerError, "Không tạo được liên kết tải ảnh.")
	ErrImageHost           = def(3011, http.StatusBadRequest, "Ảnh trong bài phải được tải lên từ trình soạn thảo.")
	// Editing an existing article (author-only).
	ErrArticleUpdate        = def(3012, http.StatusInternalServerError, "Không cập nhật được bài viết.")
	ErrArticleEditForbidden = def(3013, http.StatusForbidden, "Bạn không có quyền sửa bài viết này.")

	// --- comments (4xxx) ---
	ErrCommentList   = def(4000, http.StatusInternalServerError, "Không tải được bình luận.")
	ErrCommentCreate = def(4001, http.StatusInternalServerError, "Không gửi được bình luận.")

	// --- pro / subscription (5xxx) ---
	ErrPlanInvalid      = def(5000, http.StatusBadRequest, "Gói không hợp lệ.")
	ErrSubscriptionLoad = def(5001, http.StatusInternalServerError, "Không tải được thông tin gói.")
	ErrSubscribeFailed  = def(5002, http.StatusInternalServerError, "Không kích hoạt được Pro.")

	// --- coffee (6xxx) ---
	ErrCoffeeAmount        = def(6000, http.StatusBadRequest, "Số tiền không hợp lệ.")
	ErrCoffeeMethod        = def(6001, http.StatusBadRequest, "Phương thức thanh toán không hợp lệ.")
	ErrCoffeeCheckoutCard  = def(6002, http.StatusBadGateway, "Không khởi tạo được thanh toán thẻ.")
	ErrCoffeeCheckoutMomo  = def(6003, http.StatusBadGateway, "Không khởi tạo được thanh toán MoMo.")
	ErrCoffeeOrderCreate   = def(6004, http.StatusInternalServerError, "Không tạo được đơn hàng.")
	ErrCoffeeOrderNotFound = def(6005, http.StatusNotFound, "Không tìm thấy đơn hàng.")
	ErrCoffeeLoad          = def(6006, http.StatusInternalServerError, "Không tải được đơn hàng.")

	// --- reactions: likes & bookmarks (7xxx) ---
	ErrReactionLoad   = def(7000, http.StatusInternalServerError, "Không tải được lượt thích.")
	ErrReactionUpdate = def(7001, http.StatusInternalServerError, "Không cập nhật được tương tác.")
	ErrReactionKind   = def(7002, http.StatusBadRequest, "Loại tương tác không hợp lệ.")
	ErrBookmarkList   = def(7003, http.StatusInternalServerError, "Không tải được bài viết đã lưu.")
)
