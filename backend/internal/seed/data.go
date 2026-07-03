// Package seed holds the initial blog content, transcribed from the ui-design
// mockup (ui-design/Devnote Blog.dc.html). It is consumed by the in-memory store
// and by the MySQL seed script so both backends start with identical data.
package seed

import (
	"time"

	"github.com/ndmt1at21/devlog/backend/internal/domain"
)

func d(y, m, day int) time.Time {
	return time.Date(y, time.Month(m), day, 9, 0, 0, 0, time.UTC)
}

// Series returns the seed series definitions.
func Series() []domain.Series {
	return []domain.Series{
		{
			Slug:        "iam",
			Title:       "Xây dựng hệ thống IAM",
			Description: "Loạt bài thực chiến: thiết kế và xây dựng hệ thống Identity & Access Management từ con số 0.",
		},
	}
}

// Articles returns the seed articles in display order (Ord ascending). The first
// entry is the featured article.
func Articles() []domain.Article {
	return []domain.Article{
		{
			Slug: "ai-agents", Ord: 0, Featured: true, Category: "AI", Author: "Minh Phạm",
			ReadTime: "6 phút đọc", PublishedAt: d(2026, 6, 24),
			Title:   "AI Agents là gì? Tương lai của tự động hóa thông minh",
			Excerpt: "Tìm hiểu agent thực sự là gì, khác workflow ra sao, và khi nào bạn thật sự cần đến chúng.",
			Tags:    []string{"AI", "agent", "automation", "LLM"},
			Body: []domain.Block{
				{Type: "p", Text: "AI Agent là một hệ thống có khả năng tự lập kế hoạch, ra quyết định và thực thi hành động để đạt được mục tiêu — thay vì chỉ trả lời một câu hỏi đơn lẻ như chatbot truyền thống."},
				{Type: "p", Text: "Khác với một prompt thông thường, agent có thể gọi tool, truy vấn dữ liệu và lặp lại nhiều bước cho tới khi hoàn thành nhiệm vụ."},
				{Type: "h", Text: "Vòng lặp hoạt động của agent"},
				{Type: "diagram", Steps: []string{"Quan sát", "Lập kế hoạch", "Hành động", "Phản hồi"}, Caption: "Agent lặp lại chu trình này cho tới khi hoàn thành mục tiêu."},
				{Type: "p", Text: "Ở mỗi vòng lặp, agent quan sát môi trường, để LLM quyết định bước tiếp theo, rồi thực thi hành động đó qua tool hoặc API."},
				{Type: "code", Lang: "javascript", Code: `// Vòng lặp cơ bản của một agent
while (!task.done) {
  const obs = env.observe();
  const action = agent.plan(obs);  // LLM quyết định bước tiếp
  env.execute(action);             // gọi tool / API
}`},
				{Type: "h", Text: "Agent khác gì so với workflow?"},
				{Type: "p", Text: "Workflow đi theo các bước được lập trình sẵn. Agent thì tự quyết định bước tiếp theo dựa trên ngữ cảnh, nên linh hoạt hơn nhưng cũng khó kiểm soát hơn."},
				{Type: "quote", Text: "Hãy bắt đầu từ giải pháp đơn giản nhất. Chỉ dùng agent khi bài toán thực sự cần khả năng suy luận nhiều bước."},
				{Type: "p", Text: "Agent phát huy giá trị khi nhiệm vụ mở, nhiều bước và khó dự đoán trước — ví dụ research, debug hay điều phối nhiều API."},
			},
		},
		{
			Slug: "ts-vs-js", Ord: 1, Category: "Frontend", Author: "An Nguyễn",
			ReadTime: "5 phút đọc", PublishedAt: d(2026, 6, 20),
			Title:   "TypeScript vs JavaScript: Khi nào nên dùng cái nào?",
			Excerpt: "Static typing không phải lúc nào cũng tốt hơn. Cùng phân tích khi nào nên dùng cái nào.",
			Tags:    []string{"TypeScript", "JavaScript", "frontend"},
			Body: []domain.Block{
				{Type: "p", Text: "JavaScript là ngôn ngữ động, linh hoạt và chạy ở mọi nơi. TypeScript bổ sung hệ thống kiểu tĩnh (static typing) lên trên JavaScript."},
				{Type: "p", Text: "Điểm mấu chốt: TypeScript được biên dịch (compile) về JavaScript, nên cuối cùng trình duyệt vẫn chạy JS."},
				{Type: "h", Text: "TypeScript bắt lỗi ngay khi viết"},
				{Type: "code", Lang: "typescript", Code: `function sum(a: number, b: number): number {
  return a + b;
}

sum(2, "3"); // Error: Argument of type "string" is
             // not assignable to parameter "number"`},
				{Type: "h", Text: "Khi nào nên chọn TypeScript?"},
				{Type: "p", Text: "Với dự án lớn, nhiều người cùng làm, TypeScript giúp bắt lỗi ngay khi viết code và tự động gợi ý (autocomplete) tốt hơn."},
				{Type: "p", Text: "Với script nhỏ, prototype nhanh hay landing page đơn giản, JavaScript thuần có thể nhanh và gọn hơn."},
			},
		},
		{
			Slug: "react-perf", Ord: 2, Featured: true, Category: "Frontend", Author: "Linh Trần",
			ReadTime: "7 phút đọc", PublishedAt: d(2026, 6, 18),
			Title:   "Tối ưu hiệu năng React với useMemo và useCallback",
			Excerpt: "Memoization đúng cách — và lý do bạn không nên lạm dụng useMemo cho mọi thứ.",
			Tags:    []string{"React", "performance", "hooks"},
			Body: []domain.Block{
				{Type: "p", Text: "React render lại component mỗi khi state hoặc props thay đổi. Phần lớn trường hợp điều này hoàn toàn ổn — đừng tối ưu khi chưa cần."},
				{Type: "p", Text: "useMemo và useCallback giúp ghi nhớ (memoize) giá trị và hàm để tránh tính toán lại không cần thiết."},
				{Type: "h", Text: "Ví dụ với useMemo"},
				{Type: "code", Lang: "jsx", Code: `const value = useMemo(
  () => heavyCompute(items), // chỉ chạy lại khi items đổi
  [items]
);`},
				{Type: "h", Text: "useMemo hay useCallback?"},
				{Type: "p", Text: "useMemo ghi nhớ kết quả của một phép tính tốn kém. useCallback ghi nhớ chính tham chiếu của một hàm."},
				{Type: "p", Text: "Quy tắc đơn giản: chỉ dùng khi bạn đo được vấn đề hiệu năng thực sự bằng React Profiler, đừng dùng theo thói quen."},
			},
		},
		{
			Slug: "docker-101", Ord: 3, Category: "DevOps", Author: "Hùng Lê",
			ReadTime: "8 phút đọc", PublishedAt: d(2026, 6, 15),
			Title:   "Docker cho người mới: Container hoá trong 15 phút",
			Excerpt: "Container hoá ứng dụng đầu tiên của bạn, giải thích từ con số 0.",
			Tags:    []string{"Docker", "DevOps", "container"},
			Body: []domain.Block{
				{Type: "p", Text: `Container đóng gói ứng dụng cùng toàn bộ phụ thuộc của nó vào một đơn vị chạy được ở bất cứ đâu — giải quyết bài toán kinh điển "máy mình chạy được mà".`},
				{Type: "p", Text: "Docker là công cụ phổ biến nhất để tạo và chạy container."},
				{Type: "h", Text: "Dockerfile đầu tiên của bạn"},
				{Type: "code", Lang: "dockerfile", Code: `# Dockerfile tối giản cho Node app
FROM node:20-alpine
WORKDIR /app
COPY . .
RUN npm ci
CMD ["node", "server.js"]`},
				{Type: "h", Text: "Image và Container"},
				{Type: "p", Text: "Image là bản thiết kế tĩnh, được mô tả trong Dockerfile. Container là một thực thể đang chạy của image đó."},
				{Type: "p", Text: "Bạn build image một lần, rồi chạy nhiều container giống hệt nhau từ nó — nhất quán giữa môi trường dev và production."},
			},
		},
		{
			Slug: "microservices", Ord: 4, Category: "Backend", Author: "Trang Đỗ",
			ReadTime: "9 phút đọc", PublishedAt: d(2026, 6, 12),
			Title:   "Cloud Native & Microservices: Kiến trúc cho hệ thống hiện đại",
			Excerpt: "Khi nào microservices đáng giá, và khi nào một monolith tốt lại thắng thế.",
			Tags:    []string{"microservices", "cloud-native", "backend"},
			Body: []domain.Block{
				{Type: "p", Text: "Kiến trúc microservices chia hệ thống lớn thành nhiều dịch vụ nhỏ, độc lập, mỗi dịch vụ đảm nhận một nghiệp vụ riêng."},
				{Type: "p", Text: "Cloud Native là cách xây dựng ứng dụng tận dụng tối đa môi trường cloud: container, orchestration và auto-scaling."},
				{Type: "h", Text: "Luồng một request đi qua hệ thống"},
				{Type: "diagram", Steps: []string{"Client", "API Gateway", "Microservices", "Database"}, Caption: "Request đi qua gateway rồi được định tuyến tới service phù hợp."},
				{Type: "h", Text: "Đánh đổi cần cân nhắc"},
				{Type: "p", Text: "Microservices giúp đội ngũ triển khai độc lập và mở rộng từng phần. Nhưng nó cũng kéo theo độ phức tạp về network, monitoring và dữ liệu phân tán."},
				{Type: "p", Text: "Với hệ thống nhỏ, một monolith được tổ chức tốt thường là lựa chọn khôn ngoan hơn."},
			},
		},
		{
			Slug: "api-security", Ord: 5, Category: "Bảo mật", Author: "Dũng Vũ",
			ReadTime: "6 phút đọc", PublishedAt: d(2026, 6, 8),
			Title:   "Bảo mật API: JWT, OAuth2 và những best practice cần biết",
			Excerpt: "JWT, OAuth2 và checklist bảo mật API mà mọi backend developer nên thuộc lòng.",
			Tags:    []string{"security", "JWT", "OAuth2", "API"},
			Body: []domain.Block{
				{Type: "p", Text: "Bảo mật API bắt đầu từ hai câu hỏi: bạn là ai (authentication) và bạn được phép làm gì (authorization)."},
				{Type: "p", Text: "JWT (JSON Web Token) là cách phổ biến để truyền thông tin xác thực giữa client và server một cách an toàn."},
				{Type: "h", Text: "Cấu trúc một JWT"},
				{Type: "code", Lang: "text", Code: `header . payload . signature

  └─ thuật toán   └─ dữ liệu   └─ chữ ký

// vd: eyJhbGciOiJIUzI1NiJ9.eyJ1aWQiOjF9.4f9a2c...`},
				{Type: "h", Text: "JWT, OAuth2 và best practices"},
				{Type: "p", Text: `OAuth2 là chuẩn cho phép uỷ quyền truy cập mà không cần chia sẻ mật khẩu — nền tảng của tính năng "Đăng nhập với Google".`},
				{Type: "p", Text: "Luôn dùng HTTPS, đặt thời gian hết hạn ngắn cho token, và không bao giờ lưu secret trong mã nguồn phía client."},
			},
		},
		{
			Slug: "iam-1", Ord: 6, Category: "Bảo mật", Author: "Dũng Vũ",
			ReadTime: "7 phút đọc", PublishedAt: d(2026, 6, 26),
			Series: "iam", Part: 1, PartTitle: "Tổng quan IAM",
			Title:   "Xây dựng hệ thống IAM (Phần 1): Tổng quan Identity & Access Management",
			Excerpt: "Khởi đầu loạt bài IAM — hiểu các khái niệm cốt lõi và kiến trúc tổng thể.",
			Tags:    []string{"IAM", "identity", "security"},
			Body: []domain.Block{
				{Type: "p", Text: "IAM (Identity & Access Management) trả lời hai câu hỏi nền tảng của mọi hệ thống: người dùng là ai, và họ được phép làm gì."},
				{Type: "h", Text: "Bốn trụ cột của IAM"},
				{Type: "diagram", Steps: []string{"Identity", "Authentication", "Authorization", "Audit"}, Caption: "Bốn trụ cột tạo nên một hệ thống IAM hoàn chỉnh."},
				{Type: "p", Text: "Trong các phần tiếp theo, chúng ta sẽ lần lượt xây dựng từng phần: xác thực, phân quyền và các lớp bảo mật nâng cao."},
			},
		},
		{
			Slug: "iam-2", Ord: 7, Category: "Bảo mật", Author: "Dũng Vũ",
			ReadTime: "8 phút đọc", PublishedAt: d(2026, 6, 25),
			Series: "iam", Part: 2, PartTitle: "Authentication với JWT",
			Title:   "Xây dựng hệ thống IAM (Phần 2): Authentication với JWT & refresh token",
			Excerpt: "Triển khai luồng đăng nhập an toàn với access token và refresh token.",
			Tags:    []string{"IAM", "JWT", "authentication"},
			Body: []domain.Block{
				{Type: "p", Text: "Authentication xác minh danh tính người dùng. Mẫu phổ biến nhất hiện nay là cặp access token (ngắn hạn) và refresh token (dài hạn)."},
				{Type: "h", Text: "Luồng cấp và làm mới token"},
				{Type: "diagram", Steps: []string{"Đăng nhập", "Access token", "Hết hạn", "Refresh token"}, Caption: "Access token hết hạn nhanh; refresh token dùng để lấy token mới mà không cần đăng nhập lại."},
				{Type: "code", Lang: "javascript", Code: `// Cấp cặp token khi đăng nhập
const accessToken = sign(user, { expiresIn: "15m" });
const refreshToken = sign(user, { expiresIn: "7d" });`},
			},
		},
		{
			Slug: "iam-3", Ord: 8, Category: "Bảo mật", Author: "Dũng Vũ",
			ReadTime: "7 phút đọc", PublishedAt: d(2026, 6, 24),
			Series: "iam", Part: 3, PartTitle: "Phân quyền RBAC & ABAC",
			Title:   "Xây dựng hệ thống IAM (Phần 3): Phân quyền RBAC và ABAC",
			Excerpt: "So sánh hai mô hình phân quyền phổ biến và cách áp dụng vào hệ thống.",
			Tags:    []string{"IAM", "RBAC", "authorization"},
			Body: []domain.Block{
				{Type: "p", Text: "Authorization quyết định người dùng được làm gì. Hai mô hình phổ biến là RBAC (theo vai trò) và ABAC (theo thuộc tính)."},
				{Type: "h", Text: "RBAC: phân quyền theo vai trò"},
				{Type: "code", Lang: "javascript", Code: `const can = (user, action) =>
  user.roles.some(r => permissions[r]?.includes(action));

can(user, "post:delete"); // true / false`},
				{Type: "p", Text: "RBAC đơn giản và dễ kiểm soát. Khi cần điều kiện phức tạp hơn (theo phòng ban, theo giờ), ABAC sẽ linh hoạt hơn."},
			},
		},
		{
			Slug: "iam-4", Ord: 9, Category: "Bảo mật", Author: "Dũng Vũ",
			ReadTime: "6 phút đọc", PublishedAt: d(2026, 6, 23),
			Series: "iam", Part: 4, PartTitle: "SSO & bảo mật nâng cao",
			Title:   "Xây dựng hệ thống IAM (Phần 4): SSO, OAuth2 và bảo mật nâng cao",
			Excerpt: "Hoàn thiện hệ thống với đăng nhập một lần và các lớp bảo mật bổ sung.",
			Tags:    []string{"IAM", "SSO", "OAuth2"},
			Body: []domain.Block{
				{Type: "p", Text: "SSO (Single Sign-On) cho phép người dùng đăng nhập một lần và truy cập nhiều ứng dụng. OAuth2 và OIDC là nền tảng cho điều này."},
				{Type: "h", Text: "Checklist bảo mật nâng cao"},
				{Type: "p", Text: "Bật multi-factor authentication, ghi log mọi sự kiện đăng nhập, và áp dụng nguyên tắc least-privilege cho mọi quyền truy cập."},
				{Type: "quote", Text: "IAM không phải là việc làm một lần — nó là một quy trình cần được rà soát và củng cố liên tục."},
			},
		},
	}
}

// Comments returns seed comments with CreatedAt computed relative to now.
func Comments() []domain.Comment {
	now := time.Now().UTC()
	return []domain.Comment{
		{ArticleSlug: "ai-agents", Name: "Quang", Body: "Bài viết rất dễ hiểu, cảm ơn tác giả! Mình đang thử build agent với LangChain.", CreatedAt: now.Add(-2 * time.Hour)},
		{ArticleSlug: "ai-agents", Name: "Hà", Body: "Phần phân biệt agent và workflow đúng cái mình đang thắc mắc. Hay quá!", CreatedAt: now.Add(-5 * time.Hour)},
	}
}
