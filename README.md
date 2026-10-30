# Hệ thống Thi trắc nghiệm Trực tuyến Đại học (University Online Examination System)

Hệ thống thi trắc nghiệm trực tuyến phục vụ công tác khảo thí cấp trường và khoa, được xây dựng theo kiến trúc Client-Server phân tầng độc lập. Hệ thống hỗ trợ quản trị toàn diện (Full CRUD) đối với Kỳ thi, Ngân hàng Câu hỏi, Điểm số, Lượt thi và Tài khoản Thí sinh; tích hợp phòng thi trực tuyến chuẩn mực với Bảng điều khiển câu hỏi (Question Palette), cơ chế tự động bảo vệ tính trung thực (Anti-Cheat), hệ thống Thông báo nội bộ đa chiều và Nhật ký kiểm toán bất biến (Audit Logs).

---

## Tính năng Nâng cấp Hệ thống (Bản cập nhật Ngày 30/10/2026)

Bản nâng cấp ngày 30/10/2026 hoàn thiện toàn diện các chức năng quản trị và khảo thí thực tế:

### 1. Quản lý Đề thi và Ngân hàng Câu hỏi (Exam & Question Full CRUD)
- **Backend**:
  - `PUT /api/teacher/exams/:id`: Chỉnh sửa tiêu đề, mô tả, thời lượng làm bài, khung giờ mở/đóng và cờ xuất bản (`is_published`).
  - `DELETE /api/teacher/exams/:id`: Xóa vĩnh viễn đề thi và giải phóng toàn bộ dữ liệu phụ thuộc (`questions`, `exam_attempts`, `attempt_answers`, `appeals`) trong một Database Transaction duy nhất (Cascading Deletions).
  - `PUT /api/teacher/questions/:id`: Cho phép cập nhật nội dung câu hỏi, 4 phương án A/B/C/D, đáp án đúng và trọng số điểm khi đề thi ở trạng thái bản nháp (`draft`).
  - `DELETE /api/teacher/questions/:question_id`: Xóa câu hỏi khỏi đề thi bản nháp kèm xác minh quyền sở hữu.
  - `POST /api/teacher/exams/:id/questions/bulk`: Tiếp nhận và thêm hàng loạt câu hỏi vào đề thi trong một giao dịch nguyên tử.
- **Frontend**:
  - Tích hợp Modal Chỉnh sửa đề thi và Modal Xác nhận xóa đề thi trên cả `TeacherExamsPage` và `TeacherExamDetailPage`.
  - Hỗ trợ chỉnh sửa trực tiếp từng câu hỏi trong tab "Câu hỏi & Đáp án".
  - Bổ sung Modal "Nhập nhanh câu hỏi hàng loạt" (Bulk Import) có trình phân tích cú pháp tự động cho cả văn bản phân cách khối (`---`) và mảng dữ liệu JSON chuẩn.

### 2. Quản lý Điểm số và Lượt thi (Score & Attempt Full CRUD)
- **Backend**:
  - `PATCH /api/teacher/attempts/:id/score`: Cho phép Giảng viên hiệu chỉnh trực tiếp điểm số của thí sinh, ghi chú giải trình vào cột `teacher_note`, tự động lưu vết vào `audit_logs` và sinh thông báo gửi tới sinh viên.
  - `DELETE /api/teacher/attempts/:id`: Xóa toàn bộ dữ liệu của một lượt thi (Reset Attempt) trong Database Transaction, giải phóng Unique Composite Index `idx_exam_attempts_exam_user` để sinh viên được phép thi lại từ đầu khi gặp sự cố kỹ thuật, đồng thời ghi vết vào `audit_logs` và thông báo cho sinh viên.
- **Frontend**:
  - Bổ sung hành động "Sửa điểm" và "Cho thi lại" trên từng hàng thí sinh tại tab "Bảng điểm" của `TeacherExamDetailPage` và trang báo cáo độc lập `TeacherResultsPage`.

### 3. Quản trị Tài khoản Sinh viên (Student Account Full CRUD)
- **Backend**:
  - `GET /api/teacher/students`: Danh sách tài khoản sinh viên có hỗ trợ tìm kiếm theo họ tên hoặc email (`?q=...`), thống kê tổng số lượt thi và số bài thi đã hoàn thành (`completed_attempts / total_attempts`).
  - `POST /api/teacher/students`: Tạo tài khoản sinh viên mới với mật khẩu ban đầu được mã hóa Bcrypt an toàn.
  - `PUT /api/teacher/students/:id`: Cập nhật họ tên, email hoặc đặt lại mật khẩu mới cho sinh viên.
  - `DELETE /api/teacher/students/:id`: Xóa tài khoản sinh viên và đồng thời xóa sạch toàn bộ lượt thi, bài làm và đơn khiếu nại liên quan trong Database Transaction.
- **Frontend**:
  - Xây dựng trang mới `TeacherStudentsPage.jsx` với giao diện bảng dữ liệu chuẩn mực, tích hợp modal thêm, sửa và xác nhận xóa an toàn.
  - Bổ sung mục "Quản lý sinh viên" vào thanh điều hướng chính của giảng viên (`TeacherLayout.jsx`) và định tuyến tại `App.jsx`.

### 4. Nâng cấp Giao diện Phòng thi và Cổng Khảo thí Sinh viên
- **Backend**:
  - `GET /api/student/exams/available`: API cung cấp danh mục các đề thi đang mở kèm trạng thái làm bài tương ứng của chính sinh viên (`not_attempted`, `in_progress`, `submitted`, `time_up`).
- **Frontend**:
  - **Trang chủ Sinh viên (`StudentHomePage.jsx`)**: Tự động tải và hiển thị danh mục bài thi đang mở dạng thẻ với các thông số: thời lượng, số câu hỏi, tổng điểm, hạn nộp; nút hành động thích ứng theo trạng thái ("Vào thi ngay", "Tiếp tục làm bài", "Xem kết quả"); kèm hộp tra cứu mã đề thủ công.
  - **Phòng thi Chuẩn hóa (`ExamTakingPage.jsx`)**: Bố cục 2 cột chuyên nghiệp:
    - Cột trái: Hiển thị câu hỏi hiện tại, 4 phương án lựa chọn dạng radio, nút "Đặt cờ xem lại", cặp nút điều hướng "Câu trước" / "Câu tiếp theo".
    - Cột phải: Bảng điều khiển câu hỏi (Question Palette) hiển thị lưới số thứ tự từ 1 đến N với 4 trạng thái màu sắc theo thời gian thực (Xanh lá = Đã trả lời và đã lưu, Vàng = Đặt cờ xem lại, Xám = Chưa trả lời, Viền xanh = Câu hỏi đang xem). Nhấp vào số thứ tự để chuyển nhanh câu hỏi.

### 5. Chống gian lận Phòng thi (Anti-Cheat Engine)
- **Trộn thứ tự câu hỏi tất định (Deterministic Question Shuffling)**: Khi sinh viên lấy đề thi (`POST /api/student/exams/:id/start`), danh sách câu hỏi được xáo trộn ngẫu nhiên dựa trên seed là `attempt.ID`. Mỗi thí sinh nhận một mã đề hoán vị ngẫu nhiên khác nhau nhưng khi tải lại trang (F5 / Resume), thứ tự các câu hỏi vẫn được giữ nguyên vẹn cố định.
- **Phát hiện chuyển Tab & Đình chỉ thi (Tab-switch Detection)**:
  - Bổ sung trường `violation_count` (int, mặc định 0) vào bảng `exam_attempts`.
  - API `POST /api/student/attempts/:id/violation`: Tăng biến đếm vi phạm với cơ chế khóa bi quan trong giao dịch. Nếu vi phạm đạt ngưỡng (>= 3 lần), hệ thống tự động khóa bài, chấm điểm và chuyển trạng thái sang `completed`.
  - `ExamTakingPage.jsx` lắng nghe sự kiện `visibilitychange` (`document.hidden === true`) và `window.onblur`. Cảnh báo trực quan qua modal và huy hiệu vi phạm trên thanh điều hướng phòng thi.
  - Bổ sung cột "Số lần vi phạm" với nhãn cảnh báo trực quan trên giao diện Giảng viên.

### 6. Nhật ký Kiểm toán & Bảo mật (Security & Audit Trail)
- **Giới hạn Tần suất Truy cập (In-Memory Sliding Window Rate Limiting)**: Áp dụng giới hạn tối đa 15 request/phút/IP trên các endpoint nhạy cảm (`POST /api/auth/login`, `GET /api/student/exams/lookup`) để phòng chống tấn công brute-force và rà quét dữ liệu. Tự động bypass trong chế độ kiểm thử tự động.
- **Nhật ký Can thiệp Giảng viên (Audit Log)**:
  - Bảng `audit_logs` ghi nhận các hành vi can thiệp nhạy cảm: Giảng viên sửa điểm bài thi (`UPDATE_SCORE`), Giảng viên cho phép thí sinh thi lại (`RESET_ATTEMPT`). Lưu vết đầy đủ: `actor_id`, `action`, `target_id`, `details` (lý do/giải trình, điểm cũ -> điểm mới), `created_at`.
  - API `GET /api/teacher/audit-logs` và Modal "Nhật ký can thiệp" cho phép Giảng viên tra cứu lịch sử minh bạch trực tiếp trên hệ thống.

### 7. Xuất Bảng điểm Chuẩn Excel (Grade CSV Export with UTF-8 BOM)
- API `GET /api/teacher/exams/:id/results/export`: Xuất dữ liệu bảng điểm dưới định dạng CSV kèm UTF-8 BOM (`\xef\xbb\xbf`) để mở trực tiếp bằng Microsoft Excel mà không bị lỗi hiển thị font tiếng Việt.
- Bảng điểm xuất chuẩn hóa gồm 6 cột: `Họ tên, Email, Điểm số, Trạng thái, Số lần vi phạm, Thời gian nộp`.
- Tích hợp nút "Xuất bảng điểm (CSV)" một chạm trên cả hai trang quản lý điểm của Giảng viên.

### 8. Hệ thống Thông báo Nội bộ (In-App Notifications)
- **Backend**:
  - Model `Notification` (`id`, `user_id`, `title`, `message`, `is_read`, `created_at`) liên kết người dùng và tự động khởi tạo qua `AutoMigrate`.
  - Hàm helper `CreateNotification(userID uint, title, message string)` ghi nhận thông báo ngầm, an toàn và không gây gián đoạn luồng nghiệp vụ chính (`defer recover()`).
  - Tự động kích hoạt thông báo tại 3 thời điểm nghiệp vụ:
    - Sinh viên nộp đơn khiếu nại mới -> Thông báo tức thời đến Giảng viên phụ trách đề thi.
    - Giảng viên phản hồi, giải quyết hoặc từ chối khiếu nại -> Thông báo đến Sinh viên khiếu nại.
    - Giảng viên hiệu chỉnh điểm số thủ công hoặc xóa/reset lượt thi -> Thông báo đến Sinh viên.
  - Bộ 3 API quản lý thông báo (`/api/notifications`):
    - `GET /api/notifications`: Lấy 20 thông báo mới nhất kèm số lượng chưa đọc (`unread_count`).
    - `PATCH /api/notifications/:id/read`: Đánh dấu đã đọc 1 thông báo cụ thể.
    - `PATCH /api/notifications/read-all`: Đánh dấu đã đọc toàn bộ thông báo.
- **Frontend**:
  - Tích hợp component `NotificationDropdown` trên thanh điều hướng (Header) của cả Giảng viên (`TeacherLayout`) và Sinh viên (`StudentLayout`).
  - Hiển thị huy hiệu số lượng thông báo chưa đọc (`unread_count`), hỗ trợ xem danh sách thông báo, đánh dấu từng thông báo đã đọc khi nhấp chuột và nút "Đánh dấu đã đọc tất cả".
  - Tự động đồng bộ định kỳ mỗi 30 giây.
  - Loại bỏ hoàn toàn các hàm `window.alert` thô sơ; thay thế bằng các thông báo trạng thái dạng biểu ngữ (inline banner), hộp thoại xác nhận và thẻ thông báo nổi (toast) trang nhã.

### 9. Gia cố Bảo mật & Chống Xung đột Cơ sở dữ liệu (Security Hardening & Stability)
- **Chống lỗi Khóa Cơ sở dữ liệu SQLite (Database is Locked)**: Bật chế độ ghi nhật ký `WAL` (Write-Ahead Logging) và thiết lập thời gian chờ bận `busy_timeout=5000` (`PRAGMA journal_mode=WAL;` và `PRAGMA busy_timeout=5000;`) trong cấu hình kết nối `config/database.go` và môi trường kiểm thử.
- **Đồng bộ Thời gian Chuẩn từ Máy chủ (Server-Authoritative Timer)**: API `/start` tính toán trực tiếp `remaining_seconds` trên máy chủ và đồng bộ xuống đồng hồ đếm ngược của phòng thi, ngăn chặn hoàn toàn gian lận bằng cách chỉnh lùi giờ hệ thống trên máy tính thí sinh.
- **Bảo mật Đáp án Tuyệt đối (Zero-Leak Guarantee)**: Toàn bộ API phòng thi sử dụng DTO `QuestionSafe` loại bỏ triệt để trường `correct_option`. Kể cả sau khi nộp bài, đáp án đúng chỉ được hiển thị khi đề thi đã kết thúc thời gian mở đề thi (`now.After(exam.EndTime)`).
- **Bộ Tiêu đề Bảo mật HTTP & Giới hạn Kích thước Body**: Middleware `SecurityMiddleware` tự động gắn các HTTP Security Headers (`X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: strict-origin-when-cross-origin`) và áp dụng giới hạn kích thước phần thân request tối đa 2MB (`http.MaxBytesReader`) trên toàn bộ hệ thống để phòng chống tấn công DoS / Memory Exhaustion.

### 10. Củng cố Kiểm thử Tự động (Integration Tests)
- Toàn bộ **22/22 kịch bản kiểm thử tích hợp độc lập** trong `backend/main_test.go` đạt kết quả **PASS 100%**, bao phủ toàn diện từ phân quyền RBAC, phòng thi, chấm điểm, anti-cheat, thông báo tới bảo mật hệ thống.

---

## Cấu trúc Thư mục Dự án

```
exam-system-mvp/
├── ARCHITECTURE_AND_REPORT.md   # Báo cáo Kiến trúc & Đặc tả Nghiệp vụ Chuyên sâu
├── README.md                    # Tài liệu Hướng dẫn Khởi chạy & Danh mục Tính năng
├── docker-compose.yml           # Khởi chạy dịch vụ PostgreSQL 15 Container
├── backend/                     # Mã nguồn Dịch vụ Backend (Golang & Gin)
│   ├── config/
│   │   └── database.go          # Kết nối CSDL PostgreSQL/SQLite, bật SQLite WAL Mode
│   ├── handlers/
│   │   ├── anti_cheat_audit.go  # Xử lý vi phạm chuyển Tab, Audit Logs, Xuất CSV UTF-8
│   │   ├── auth.go              # Đăng ký, Đăng nhập, Băm Bcrypt, Sinh JWT
│   │   ├── history_appeals.go   # Lịch sử bài làm, Quy trình Phúc khảo 4 trạng thái
│   │   ├── notifications.go     # Thông báo nội bộ đa chiều, Polling, Đánh dấu đã đọc
│   │   ├── student.go           # Phòng thi, Xáo trộn đề tất định, Auto-save (UPSERT), Auto-submit
│   │   ├── teacher.go           # Quản lý đề thi, Xuất bản đề, Thống kê kết quả
│   │   └── teacher_crud.go      # CRUD Đề thi, Nhập câu hỏi hàng loạt, CRUD Sinh viên, Sửa điểm, Reset bài
│   ├── middleware/
│   │   ├── auth.go              # Xác thực JWT Bearer Token, Phân quyền RBAC (Teacher/Student)
│   │   ├── ratelimit.go         # Giới hạn tần suất Sliding Window (15 req/phút/IP)
│   │   └── security.go          # Gắn HTTP Security Headers, Giới hạn Body size 2MB
│   ├── models/
│   │   ├── migrate.go           # AutoMigrate 8 bảng, Khởi tạo 2 Unique Composite Indexes
│   │   ├── models.go            # Khai báo GORM Structs (User, Exam, Question, Attempt, Answer, Appeal, AuditLog, Notification)
│   │   └── seed.go              # Nạp dữ liệu mẫu ban đầu (Giảng viên, Sinh viên, Đề thi IT2026)
│   ├── router/
│   │   └── router.go            # Đăng ký Gin Engine, CORS, Middleware và toàn bộ API Endpoints
│   ├── utils/
│   │   └── helpers.go           # Các hàm tiện ích hỗ trợ xử lý dữ liệu
│   ├── main.go                  # Điểm khởi chạy máy chủ Backend (Cổng 8081)
│   └── main_test.go             # Bộ kiểm thử tích hợp tự động (22 Test Cases)
└── frontend/                    # Ứng dụng Giao diện Người dùng (React 19 & Vite)
    ├── public/                  # Tài nguyên tĩnh (Favicon, Logo)
    ├── src/
    │   ├── api/
    │   │   ├── axios.js         # Cấu hình Axios BaseURL (Cổng 8081) & Interceptor đính kèm JWT
    │   │   └── endpoints.js     # Danh mục hàm gọi API theo từng phân hệ
    │   ├── components/
    │   │   ├── NotificationDropdown.jsx # Dropdown thông báo thời gian thực kèm huy hiệu chưa đọc
    │   │   ├── ProtectedRoute.jsx       # Bộ bảo vệ tuyến đường theo trạng thái đăng nhập và vai trò
    │   │   ├── Spinner.jsx              # Vòng quay hiệu ứng tải trang tối giản
    │   │   ├── StudentLayout.jsx        # Khung giao diện dùng chung cho Phân hệ Sinh viên
    │   │   └── TeacherLayout.jsx        # Khung giao diện dùng chung cho Phân hệ Giảng viên
    │   ├── contexts/
    │   │   └── AuthContext.jsx  # Quản lý trạng thái xác thực toàn cục
    │   ├── pages/
    │   │   ├── LoginPage.jsx    # Trang đăng nhập hệ thống
    │   │   ├── RegisterPage.jsx # Trang đăng ký tài khoản sinh viên
    │   │   ├── student/
    │   │   │   ├── ExamResultPage.jsx          # Trang hiển thị kết quả thi và cảnh báo vi phạm
    │   │   │   ├── ExamTakingPage.jsx          # Phòng thi trực tuyến 2 cột (Question Palette, Auto-save)
    │   │   │   ├── StudentHistoryDetailPage.jsx# Xem lại chi tiết bài làm và gửi khiếu nại
    │   │   │   ├── StudentHistoryPage.jsx      # Danh sách lịch sử các bài thi đã làm
    │   │   │   └── StudentHomePage.jsx         # Cổng khảo thí sinh viên (Danh mục bài thi đang mở)
    │   │   └── teacher/
    │   │       ├── ExamAppealsTab.jsx          # Tab quản lý và xử lý đơn khiếu nại của đề thi
    │   │       ├── TeacherExamDetailPage.jsx   # Chi tiết đề thi, quản lý câu hỏi, bảng điểm, khiếu nại
    │   │       ├── TeacherExamsPage.jsx        # Danh mục đề thi của giảng viên
    │   │       ├── TeacherResultsPage.jsx      # Báo cáo điểm số độc lập, sửa điểm, reset bài, xuất CSV
    │   │       └── TeacherStudentsPage.jsx     # Bảng quản lý tài khoản sinh viên toàn trường
    │   ├── App.jsx              # Định tuyến đường dẫn ứng dụng
    │   ├── main.jsx             # Điểm gắn kết DOM của React
    │   └── index.css            # Cấu hình kiểu dáng cơ bản TailwindCSS
    ├── package.json             # Danh mục thư viện phụ thuộc Frontend
    └── vite.config.js           # Cấu hình Vite Dev Server (Cổng 5173)
```

---

## Công nghệ Sử dụng

- **Backend**: Golang 1.21+ (Gin Framework, GORM ORM, JWT, Bcrypt).
- **Cơ sở dữ liệu**: PostgreSQL 15 (Docker container), hỗ trợ giao dịch ACID, khóa bi quan dòng dữ liệu (`SELECT ... FOR UPDATE`), câu lệnh nguyên tử (`ON CONFLICT DO UPDATE`) và chỉ mục kết hợp duy nhất (Composite Unique Indexes).
- **Frontend**: React 19, Vite, TailwindCSS, React Router v6, Axios Interceptor tự động quản lý và đính kèm JWT Bearer Token.
- **Môi trường kiểm thử**: Gói `testing` và `net/http/httptest` tiêu chuẩn của Go, bao phủ 100% các Use Case tích hợp trên cơ sở dữ liệu SQLite in-memory độc lập (chế độ WAL Mode).

---

## Hướng dẫn Khởi chạy Hệ thống

Yêu cầu môi trường máy trạm: Docker Desktop, Golang 1.21+ và Node.js 18+.

### Bước 1: Khởi động Cơ sở dữ liệu (PostgreSQL)
Mở cửa sổ Terminal tại thư mục gốc của dự án:
```bash
docker-compose up -d
```
Cơ sở dữ liệu PostgreSQL sẽ chạy tại cổng `5432` (`User: postgres`, `Password: postgres`, `Database: exam_mvp`).

### Bước 2: Khởi động Dịch vụ Backend (Golang)
Mở một cửa sổ Terminal mới:
```bash
cd backend
go run main.go
```
Backend sẽ tự động thực hiện AutoMigrate 8 bảng và chỉ mục tổng hợp, nạp sẵn tài khoản kiểm thử và lắng nghe tại cổng **`8081`** (`http://localhost:8081`).

### Bước 3: Khởi động Ứng dụng Frontend (React Vite)
Mở một cửa sổ Terminal mới:
```bash
cd frontend
npm install
npm run dev
```
Giao diện người dùng sẽ chạy tại địa chỉ: **`http://localhost:5173`**.

---

## Tài khoản Thử nghiệm Mẫu

Hệ thống được khởi tạo sẵn các tài khoản demo phục vụ kiểm thử nghiệm thu:

| Vai trò | Địa chỉ Email | Mật khẩu | Mô tả quyền hạn |
| :--- | :--- | :--- | :--- |
| **Giảng viên** | `teacher@test.com` | `123456` | Quản lý đề thi, câu hỏi, bảng điểm, học viên, khiếu nại, audit logs |
| **Sinh viên 1** | `student@test.com` | `123456` | Làm bài thi, Question Palette, tự động lưu, lịch sử và khiếu nại |
| **Sinh viên 2** | `student2@test.com` | `123456` | Tài khoản sinh viên kiểm thử bổ sung |
| **Sinh viên 3** | `student3@test.com` | `123456` | Tài khoản sinh viên kiểm thử bổ sung |

Mã đề thi mẫu sẵn có: **`IT2026`** (Kỳ thi Kiểm tra Kiến thức Lập trình & Cơ sở dữ liệu).

---

## Chạy Kiểm thử Tự động (Integration Tests)

Chạy toàn bộ 22 kịch bản kiểm thử tích hợp tự động bao phủ tất cả các phân hệ:
```bash
cd backend
go test -v ./...
```
Toàn bộ kịch bản kiểm thử chạy trên cơ sở dữ liệu SQLite in-memory giả lập đầy đủ ràng buộc ACID, chế độ WAL mode chống lock và xử lý kiểu dữ liệu thời gian thực.
