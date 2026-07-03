// Frontend-owned translations for the backend's stable API error codes
// (mirrors backend/internal/apierr). The server also returns a default `message`
// in the envelope; translateError prefers this dictionary (by the active locale)
// and falls back to the server message.

import { LOCALE_COOKIE, type Locale } from "./i18n/dictionaries";

const vi: Record<number, string> = {
  1000: "Dữ liệu gửi lên không hợp lệ.",
  1001: "Dữ liệu chưa hợp lệ.",
  1002: "Bạn cần đăng nhập.",
  1003: "Bạn không có quyền thực hiện.",
  1004: "Không tìm thấy.",
  1006: "Đã có lỗi xảy ra. Vui lòng thử lại.",
  1007: "Không kết nối được dịch vụ.",
  1008: "Tính năng chưa được cấu hình.",
  2000: "Đăng nhập chưa được cấu hình.",
  2001: "Email hoặc mật khẩu không đúng.",
  2002: "Không kết nối được dịch vụ xác thực.",
  2003: "Không lấy được thông tin người dùng.",
  2004: "Email này đã được đăng ký.",
  2005: "Email chưa hợp lệ.",
  2006: "Mật khẩu cần tối thiểu 6 ký tự.",
  2007: "Không tạo được phiên đăng nhập.",
  3000: "Không tìm thấy bài viết.",
  3001: "Không tải được danh sách bài viết.",
  3002: "Chưa có bài viết nổi bật.",
  3003: "Không tải được bài viết.",
  3004: "Không tải được danh mục.",
  3005: "Bạn không có quyền tạo bài viết.",
  3006: "Không tạo được bài viết.",
  4000: "Không tải được bình luận.",
  4001: "Không gửi được bình luận.",
  5000: "Gói không hợp lệ.",
  5001: "Không tải được thông tin gói.",
  5002: "Không kích hoạt được Pro.",
  6000: "Số tiền không hợp lệ.",
  6001: "Phương thức thanh toán không hợp lệ.",
  6002: "Không khởi tạo được thanh toán thẻ.",
  6003: "Không khởi tạo được thanh toán MoMo.",
  6004: "Không tạo được đơn hàng.",
  6005: "Không tìm thấy đơn hàng.",
  6006: "Không tải được đơn hàng.",
};

const en: Record<number, string> = {
  1000: "Invalid request data.",
  1001: "Invalid data.",
  1002: "Please log in.",
  1003: "You don’t have permission to do this.",
  1004: "Not found.",
  1006: "Something went wrong. Please try again.",
  1007: "Couldn’t reach the service.",
  1008: "This feature isn’t configured.",
  2000: "Login is not configured.",
  2001: "Incorrect email or password.",
  2002: "Couldn’t reach the authentication service.",
  2003: "Couldn’t fetch user info.",
  2004: "This email is already registered.",
  2005: "Invalid email.",
  2006: "Password must be at least 6 characters.",
  2007: "Couldn’t create the session.",
  3000: "Article not found.",
  3001: "Couldn’t load the article list.",
  3002: "No featured article yet.",
  3003: "Couldn’t load the article.",
  3004: "Couldn’t load categories.",
  3005: "You don’t have permission to publish articles.",
  3006: "Couldn’t create the article.",
  4000: "Couldn’t load comments.",
  4001: "Couldn’t post the comment.",
  5000: "Invalid plan.",
  5001: "Couldn’t load subscription info.",
  5002: "Couldn’t activate Pro.",
  6000: "Invalid amount.",
  6001: "Invalid payment method.",
  6002: "Couldn’t start card payment.",
  6003: "Couldn’t start MoMo payment.",
  6004: "Couldn’t create the order.",
  6005: "Order not found.",
  6006: "Couldn’t load the order.",
};

const MESSAGES: Record<Locale, Record<number, string>> = { vi, en };

/** Read the active locale from the cookie (client-side); defaults to vi. */
function clientLocale(): Locale {
  if (typeof document === "undefined") return "vi";
  const m = document.cookie.match(
    new RegExp(`(?:^|; )${LOCALE_COOKIE}=(en|vi)`),
  );
  return m?.[1] === "en" ? "en" : "vi";
}

/**
 * Resolve a user-facing message for an API error: prefer the FE dictionary (by
 * code + locale), then the server-provided message, then a generic fallback.
 */
export function translateError(
  code: number | undefined,
  serverMessage?: string,
  locale?: Locale,
): string {
  const map = MESSAGES[locale ?? clientLocale()];
  if (code != null && map[code]) return map[code];
  return serverMessage || map[1006];
}
