# Hệ thống Thi trắc nghiệm Trực tuyến Đại học (University Online Examination System)

Hệ thống thi trắc nghiệm trực tuyến phục vụ công tác khảo thí cấp trường/khoa, được xây dựng theo kiến trúc Client-Server phân tầng, hỗ trợ quản lý toàn diện (Full CRUD) đối với Kỳ thi, Câu hỏi, Điểm số/Lượt thi và Tài khoản Thí sinh, tích hợp phòng thi trực tuyến chuẩn hóa với bảng điều khiển câu hỏi (Question Palette) và cơ chế tự động bảo vệ tính trung thực của kỳ thi.

---

## 30/10/2026 - 12:18 AM Upgrade Features

Bản nâng cấp ngày 30/10/2026 hoàn thiện toàn diện các chức năng quản trị và khảo thí thực tế:

### 1. Quản lý Đề thi và Câu hỏi (Exam & Question Full CRUD)
- **Backend**:
  - `PUT /api/teacher/exams/:id`: Chỉnh sửa tiêu đề, mô tả, thời lượng làm bài, khung giờ mở/đóng và cờ xuất bản (`is_published`).
  - `DELETE /api/teacher/exams/:id`: Xóa vĩnh viễn đề thi và giải phóng toàn bộ dữ liệu phụ thuộc (`questions`, `exam_attempts`, `attempt_answers`, `appeals`) trong một Database Transaction duy nhất (Cascading Deletions).
  - `PUT /api/teacher/questions/:id`: Cho phép cập nhật nội dung câu hỏi, 4 phương án A/B/C/D, đáp án đúng và trọng số điểm khi đề thi ở trạng thái bản nháp (`draft`).
  - `POST /api/teacher/exams/:id/questions/bulk`: Tiếp nhận và thêm hàng loạt câu hỏi vào đề thi trong một giao dịch nguyên tử.
- **Frontend**:
  - Tích hợp Modal Chỉnh sửa đề thi và Modal Xác nhận xóa đề thi trên cả `TeacherExamsPage` và `TeacherExamDetailPage`.
  - Hỗ trợ chỉnh sửa trực tiếp từng câu hỏi trong tab "Câu hỏi & Đáp án".
  - Bổ sung Modal "Nhập nhanh câu hỏi hàng loạt" (Bulk Import) có trình phân tích cú pháp tự động cho cả văn bản phân cách khối (`---`) và mảng dữ liệu JSON chuẩn.

### 2. Quản lý Điểm số và Lượt thi (Score & Attempt Full CRUD)
- **Backend**:
  - `PATCH /api/teacher/attempts/:id/score`: Cho phép giảng viên hiệu chỉnh trực tiếp điểm số của thí sinh và ghi nhận lý do/giải trình vào cột `teacher_note`.
  - `DELETE /api/teacher/attempts/:id`: Xóa toàn bộ dữ liệu của một lượt thi (Reset Attempt) trong Database Transaction, giải phóng Unique Composite Index `idx_exam_attempts_exam_user` để sinh viên được phép thi lại từ đầu khi gặp sự cố mạng hoặc lỗi kỹ thuật.
- **Frontend**:
  - Bổ sung hành động "Sửa điểm" và "Cho thi lại" trên từng hàng thí sinh tại tab "Bảng điểm" của `TeacherExamDetailPage` và trang báo cáo độc lập `TeacherResultsPage`.

### 3. Quản lý Tài khoản Sinh viên (Student Account Full CRUD)
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
  - **Phòng thi Chuẩn hóa (`ExamTakingPage.jsx`)**: Tái cấu trúc theo bố cục 2 cột chuyên nghiệp:
    - Cột trái: Hiển thị câu hỏi hiện tại, 4 phương án lựa chọn dạng radio, nút "Đặt cờ xem lại", cặp nút điều hướng "Câu trước" / "Câu tiếp theo".
    - Cột phải: Bảng điều khiển câu hỏi (Question Palette) hiển thị lưới số thứ tự từ 1 đến N với 4 trạng thái màu sắc theo thời gian thực (Xanh lá = Đã trả lời và đã lưu, Vàng = Đặt cờ xem lại, Xám = Chưa trả lời, Viền xanh = Câu hỏi đang xem). Nhấp vào số thứ tự để chuyển nhanh câu hỏi.

### 5. Chống gian lận Phòng thi (Anti-Cheat)
- **Trộn thứ tự câu hỏi tất định (Deterministic Question Shuffling)**: Khi sinh viên lấy đề thi (`POST /api/student/exams/:id/start`), danh sách câu hỏi được xáo trộn ngẫu nhiên dựa trên seed là `attempt.ID`. Mỗi thí sinh nhận một mã đề hoán vị ngẫu nhiên khác nhau nhưng khi tải lại trang (F5 / Resume), thứ tự các câu hỏi vẫn được giữ nguyên vẹn cố định.
- **Phát hiện chuyển Tab & Đình chỉ thi (Tab-switch Detection)**:
  - Bổ sung trường `violation_count` (int, mặc định 0) vào bảng `exam_attempts`.
  - API `POST /api/student/attempts/:id/violation`: Tăng biến đếm vi phạm với cơ chế khóa bi quan trong giao dịch. Nếu vi phạm đạt ngưỡng (>= 3 lần), hệ thống tự động khóa bài, chấm điểm và chuyển trạng thái sang `completed`.
  - `ExamTakingPage.jsx` lắng nghe sự kiện `visibilitychange` (`document.hidden === true`) và `window.onblur`. Cảnh báo trực quan qua modal và huy hiệu vi phạm trên thanh điều hướng phòng thi.
  - Bổ sung cột "Số lần vi phạm" với nhãn cảnh báo trực quan trên giao diện Giảng viên (`TeacherExamDetailPage` và `TeacherResultsPage`).

### 6. Bảo mật & Nhật ký Can thiệp (Security & Audit Trail)
- **Giới hạn Tần suất Truy cập (In-Memory Sliding Window Rate Limiting)**: Áp dụng giới hạn tối đa 15 request/phút/IP trên các endpoint nhạy cảm (`POST /api/auth/login`, `GET /api/student/exams/lookup`) để phòng chống tấn công brute-force và rà quét dữ liệu. Tự động bypass trong chế độ kiểm thử tự động để bảo đảm tính ổn định của CI/CD.
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
  - Hàm helper `CreateNotification(userID uint, title, message string)` ghi nhận thông báo ngầm, an toàn và không gây gián đoạn luồng nghiệp vụ chính.
  - Tự động kích hoạt thông báo tại 3 thời điểm nghiệp vụ:
    - Khi sinh viên nộp đơn khiếu nại mới -> Thông báo tức thời đến Giảng viên phụ trách đề thi.
    - Khi Giảng viên phản hồi, giải quyết hoặc từ chối khiếu nại -> Thông báo đến Sinh viên khiếu nại.
    - Khi Giảng viên hiệu chỉnh điểm số thủ công hoặc xóa/reset lượt thi -> Thông báo đến Sinh viên để làm lại bài.
  - Bộ 3 API quản lý thông báo (`/api/notifications`):
    - `GET /api/notifications`: Lấy 20 thông báo mới nhất của người dùng hiện tại kèm số lượng chưa đọc (`unread_count`).
    - `PATCH /api/notifications/:id/read`: Đánh dấu đã đọc 1 thông báo cụ thể.
    - `PATCH /api/notifications/read-all`: Đánh dấu đã đọc toàn bộ thông báo.
- **Frontend**:
  - Tích hợp component `NotificationDropdown` trên thanh điều hướng (Header) của cả Giảng viên (`TeacherLayout`) và Sinh viên (`StudentLayout`).
  - Hiển thị huy hiệu số lượng thông báo chưa đọc (`unread_count`), hỗ trợ xem danh sách thông báo, đánh dấu từng thông báo đã đọc khi nhấp chuột và nút "Đánh dấu đã đọc tất cả".
  - Tự động đồng bộ định kỳ mỗi 30 giây.
  - Loại bỏ hoàn toàn các hàm `window.alert` thô sơ trên toàn bộ giao diện; thay thế bằng các thông báo trạng thái dạng biểu ngữ (inline banner), hộp thoại xác nhận và thẻ thông báo nổi (toast) trang nhã.

### 9. Gia cố Bảo mật & Chống Xung đột Cơ sở dữ liệu (Security Hardening & Stability)
- **Chống lỗi Khóa Cơ sở dữ liệu SQLite (Database is Locked)**: Bật chế độ ghi nhật ký `WAL` (Write-Ahead Logging) và thiết lập thời gian chờ bận `busy_timeout=5000` (`PRAGMA journal_mode=WAL;` và `PRAGMA busy_timeout=5000;`) trong cấu hình kết nối `config/database.go` và môi trường kiểm thử.
- **Đồng bộ Thời gian Chuẩn từ Máy chủ (Server-Authoritative Timer)**: API `/start` tính toán trực tiếp `remaining_seconds` trên máy chủ và đồng bộ xuống đồng hồ đếm ngược của phòng thi, ngăn chặn hoàn toàn gian lận bằng cách chỉnh lùi giờ hệ thống trên máy tính thí sinh.
- **Bảo mật Đáp án Tuyệt đối (Zero-Leak Guarantee)**: Toàn bộ API lấy đề thi phòng thi đều sử dụng DTO `QuestionSafe` loại bỏ triệt để trường `correct_option`. Kể cả sau khi nộp bài, đáp án đúng chỉ được hiển thị khi đề thi đã kết thúc thời gian mở đề thi (`now.After(exam.EndTime)`).
- **Bộ Tiêu đề Bảo mật HTTP & Giới hạn Kích thước Body**: Middleware `SecurityMiddleware` tự động gắn các HTTP Security Headers (`X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: strict-origin-when-cross-origin`) và áp dụng giới hạn kích thước phần thân request tối đa 2MB (`http.MaxBytesReader`) trên toàn bộ hệ thống để phòng chống tấn công DoS / Memory Exhaustion.

### 10. Củng cố Kiểm thử Tự động (Integration Tests)
- Bổ sung 9 kịch bản kiểm thử tích hợp độc lập trong `backend/main_test.go`:
  - `TestTeacher_ExamCRUD_And_BulkQuestions`: Kiểm tra sửa đề, nhập câu hỏi hàng loạt, sửa câu hỏi và xóa đề thi cascade.
  - `TestTeacher_ScoreUpdate_And_ResetAttempt`: Kiểm tra sửa điểm kèm ghi chú và xóa lượt thi cho phép sinh viên bắt đầu lại.
  - `TestTeacher_StudentManagementCRUD`: Kiểm tra danh sách, tạo mới, sửa thông tin, đặt lại mật khẩu và xóa tài khoản sinh viên.
  - `TestStudent_AvailableExams`: Kiểm tra danh sách bài thi đang mở và trạng thái lượt thi của sinh viên.
  - `TestAntiCheat_QuestionShufflingAndViolationLimit`: Kiểm tra xáo trộn câu hỏi tất định, đếm vi phạm rời tab và tự động nộp bài khi vi phạm 3 lần.
  - `TestTeacher_AuditLogsAndCSVExport`: Kiểm tra ghi nhận audit log và xuất file CSV có UTF-8 BOM chuẩn.
  - `TestRateLimiter_Enforcement`: Kiểm tra chặn request khi vượt ngưỡng tần suất gọi API.
  - `TestNotifications_FullWorkflow`: Kiểm tra toàn bộ vòng đời thông báo (khiếu nại, phản hồi, sửa điểm, đánh dấu đã đọc đơn lẻ và tất cả).
  - `TestSecurityMiddleware_HeadersAndBodyLimit`: Kiểm tra các HTTP Security Headers và cơ chế từ chối request body vượt quá 2MB.
- Toàn bộ 22/22 kịch bản kiểm thử đạt kết quả **PASS 100%**.

---

## 1. Công nghệ sử dụng

- **Backend**: Golang (Gin Framework, GORM ORM, JWT, Bcrypt).
- **Cơ sở dữ liệu**: PostgreSQL 15 (Docker container), hỗ trợ giao dịch ACID, khóa bi quan dòng dữ liệu (`SELECT ... FOR UPDATE`), câu lệnh nguyên tử (`ON CONFLICT DO UPDATE`) và chỉ mục kết hợp duy nhất (Composite Unique Indexes).
- **Frontend**: React 19, Vite, TailwindCSS, React Router v6, Axios Interceptor tự động quản lý và đính kèm JWT Bearer Token.
- **Môi trường kiểm thử**: Gói `testing` và `net/http/httptest` tiêu chuẩn của Go, bao phủ 100% các Use Case tích hợp trên cơ sở dữ liệu SQLite in-memory độc lập.

---

## 2. Các phân hệ và tính năng nghiệp vụ

### 2.1. Phân hệ Giảng viên (Teacher & Administration)
- **Quản lý Đề thi (Full CRUD)**:
  - Tạo mới đề thi với tiêu đề, mô tả, thời lượng làm bài (phút), thời điểm mở đề và thời điểm đóng đề.
  - Chỉnh sửa thông tin đề thi, thời lượng, khung giờ thi và trạng thái đóng/mở đề thi.
  - Xóa đề thi kèm cơ chế xóa theo tầng (Cascade) đối với toàn bộ câu hỏi, lượt thi, bài làm và đơn khiếu nại liên quan trong DB Transaction.
  - Xuất bản đề thi chính thức: Kiểm tra tối thiểu 1 câu hỏi trước khi xuất bản.
- **Quản lý Câu hỏi (Full CRUD & Bulk Import)**:
  - Thêm mới từng câu hỏi trắc nghiệm 4 lựa chọn (A, B, C, D), thiết lập đáp án đúng và trọng số điểm.
  - Chỉnh sửa nội dung câu hỏi, các lựa chọn và đáp án đúng khi đề thi ở trạng thái bản nháp.
  - Xóa câu hỏi đơn lẻ kèm kiểm tra quyền sở hữu và trạng thái bản nháp.
  - Nhập nhanh câu hỏi hàng loạt (Bulk Import): Hỗ trợ cú pháp văn bản phân cách khối câu hỏi (`---`) hoặc mảng định dạng JSON chuẩn.
- **Quản lý Điểm số & Lượt thi (Score & Attempt Management)**:
  - Xem bảng điểm tổng hợp, thống kê phân bố điểm (Điểm trung bình, Điểm cao nhất, Điểm thấp nhất, Tỷ lệ hoàn thành).
  - Hiệu chỉnh điểm thi trực tiếp: Cho phép giảng viên sửa điểm bài thi của thí sinh kèm lý do/ghi chú giải trình (`teacher_note`).
  - Cho phép thi lại (Reset Attempt): Xóa lượt thi gặp sự cố kỹ thuật của thí sinh trong DB Transaction, giải phóng Unique Constraint để sinh viên thi lại từ đầu.
- **Quản lý Tài khoản Sinh viên (Student Account CRUD)**:
  - Liệt kê danh sách sinh viên kèm bộ lọc tìm kiếm theo họ tên hoặc email.
  - Thống kê số lượt thi và số bài thi hoàn thành của từng sinh viên.
  - Thêm mới tài khoản sinh viên với mật khẩu khởi tạo (mã hóa Bcrypt).
  - Cập nhật họ tên, email hoặc đặt lại mật khẩu mới cho sinh viên.
  - Xóa tài khoản sinh viên kèm cơ chế dọn dẹp toàn vẹn dữ liệu.
- **Xử lý Phúc khảo & Khiếu nại**:
  - Tiếp nhận đơn khiếu nại điểm thi từ sinh viên.
  - Phản hồi giải trình, chấp thuận cập nhật điểm hoặc từ chối khiếu nại.

### 2.2. Phân hệ Sinh viên (Student Portal & Exam Room)
- **Trang chủ Khảo thí (Student Portal)**:
  - Hiển thị danh mục các bài thi đang mở với trạng thái tương ứng (`Chưa thi`, `Đang làm dở`, `Đã nộp bài`).
  - Nút hành động nhanh: "Vào thi ngay", "Tiếp tục làm bài", hoặc "Xem kết quả".
  - Hộp tra cứu nhanh đề thi theo mã phòng thi (Exam Code).
  - Bảng tổng hợp lịch sử các lượt thi gần đây.
- **Phòng thi Chuẩn hóa (Realistic Exam Room Layout)**:
  - Bố cục 2 cột chuyên nghiệp:
    - Cột trái (Khu vực câu hỏi): Hiển thị câu hỏi hiện tại, trọng số điểm, 4 lựa chọn A/B/C/D, nút "Đặt cờ xem lại", nút điều hướng "Câu trước" / "Câu tiếp theo".
    - Cột phải (Bảng điều khiển câu hỏi - Question Palette): Lưới trực quan các câu hỏi từ 1 đến N với 4 trạng thái màu sắc chuẩn mực:
      - Xanh lá: Đã trả lời và đã lưu lên máy chủ.
      - Vàng: Đã đặt cờ xem lại.
      - Xám: Chưa chọn đáp án.
      - Viền xanh: Câu hỏi đang được chọn hiển thị.
    - Click trực tiếp vào số thứ tự trên bảng điều khiển để chuyển nhanh đến câu hỏi tương ứng.
- **Đồng hồ đếm ngược kiểm soát bởi máy chủ (Server-Authoritative)**:
  - Đồng hồ đếm ngược theo thời gian thực.
  - Hộp thoại cảnh báo khi thời gian làm bài còn dưới 1 phút.
  - Tự động khóa đề và nộp bài khi đồng hồ về `00:00`.
- **Cơ chế Tự động Lưu bài (Auto-save)**:
  - Tự động lưu đáp án ngay khi thí sinh nhấp chọn và định kỳ mỗi 60 giây.
  - Hỗ trợ khôi phục trạng thái làm bài (Resume Exam) khi thí sinh vô tình đóng trình duyệt hoặc gặp sự cố đường truyền mạng.
- **Xem Kết quả & Phúc khảo**:
  - Xem điểm số, số câu đúng/sai, chi tiết bài làm.
  - Gửi đơn khiếu nại kèm căn cứ giải trình gửi tới giảng viên phụ trách.

---

## 3. Kiến trúc Kỹ thuật & Bảo mật

1. **Phòng chống Lộ đáp án (Zero-Leak DTO)**: Khi sinh viên bắt đầu làm bài, API trả về DTO `QuestionSafe` đã loại bỏ hoàn toàn trường `correct_option`. Trình duyệt của thí sinh không thể trích xuất đáp án bằng DevTools hay Network Inspector.
2. **Kiểm soát Thời gian Độc quyền phía Máy chủ (Server-Authoritative)**: Mọi mốc thời gian hết hạn nộp bài được tính toán dựa trên đồng hồ của máy chủ backend với vùng đệm 5 giây bù trừ trễ mạng. Thay đổi giờ đồng hồ máy trạm không có tác dụng can thiệp thời gian làm bài.
3. **Chống Race Condition và Nộp bài Trùng lặp**: Thao tác lưu bài và nộp bài được bao bọc trong DB Transaction kết hợp khóa dòng bi quan `SELECT ... FOR UPDATE` trên bảng `exam_attempts` và câu lệnh `ON CONFLICT DO UPDATE` trên bảng `attempt_answers`.
4. **Chống gian lận Phòng thi (Anti-Cheat Engine)**:
   - Xáo trộn danh sách câu hỏi tất định theo seed `attempt.ID` (mỗi thí sinh nhận một thứ tự câu hỏi ngẫu nhiên khác nhau, không nhảy loạn khi tải lại trang).
   - Giám sát chuyển tab qua sự kiện trình duyệt `visibilitychange` và `window.onblur`. Tự động đình chỉ thi và thu bài khi vi phạm 3 lần.
5. **Giới hạn Tần suất (Sliding Window Rate Limiter)**: Ngăn chặn tấn công Brute-force mật khẩu và rà quét mã đề thi với ngưỡng 15 request/phút/IP.
6. **Nhật ký Can thiệp Điểm số & Lượt thi (Audit Trail)**: Ghi vết đầy đủ mọi thao tác sửa điểm hoặc reset bài thi của Giảng viên vào bảng `audit_logs`.
7. **Xuất Bảng điểm Chuẩn Excel**: Hỗ trợ xuất dữ liệu bảng điểm dạng CSV có UTF-8 BOM (`\xef\xbb\xbf`) mở trực tiếp không lỗi font tiếng Việt.
8. **Phân quyền và Chống lỗi IDOR**: Kiểm tra nghiêm ngặt quyền sở hữu đối với từng thao tác sửa/xóa câu hỏi, sửa/xóa đề thi, sửa điểm hoặc xóa lượt thi. Thí sinh chỉ được thao tác trên bài thi của chính mình.
9. **Chống SQL Injection**: 100% câu truy vấn được tham số hóa thông qua GORM ORM kết hợp hàm `parseUintParam` xác thực kiểu dữ liệu đầu vào trên URL parameters.

---

## 4. Hướng dẫn Khởi chạy Hệ thống

Yêu cầu môi trường: Docker Desktop, Golang 1.21+ và Node.js 18+.

### Bước 1: Khởi động Cơ sở dữ liệu (PostgreSQL)
Mở Terminal tại thư mục gốc của dự án:
```bash
docker-compose up -d
```
Cơ sở dữ liệu PostgreSQL sẽ chạy tại cổng `5432` (`User: postgres`, `Password: postgres`, `Database: exam_mvp`).

### Bước 2: Khởi động Backend (Golang)
Mở một cửa sổ Terminal mới:
```bash
cd backend
go run main.go
```
Backend sẽ tự động thực hiện AutoMigrate các bảng và chỉ mục, nạp sẵn tài khoản kiểm thử và lắng nghe tại cổng **`8081`** (`http://localhost:8081`).

### Bước 3: Khởi động Frontend (React Vite)
Mở một cửa sổ Terminal mới:
```bash
cd frontend
npm run dev
```
Giao diện người dùng sẽ chạy tại địa chỉ: **`http://localhost:5173`**.

---

## 5. Tài khoản Thử nghiệm Mẫu

Hệ thống được khởi tạo sẵn các tài khoản demo:

| Vai trò | Địa chỉ Email | Mật khẩu | Mô tả quyền hạn |
| :--- | :--- | :--- | :--- |
| **Giảng viên** | `teacher@test.com` | `123456` | Quản lý đề thi, câu hỏi, bảng điểm, học viên và khiếu nại |
| **Sinh viên 1** | `student@test.com` | `123456` | Làm bài thi, xem bảng câu hỏi, lịch sử và khiếu nại |
| **Sinh viên 2** | `student2@test.com` | `123456` | Tài khoản sinh viên kiểm thử bổ sung |
| **Sinh viên 3** | `student3@test.com` | `123456` | Tài khoản sinh viên kiểm thử bổ sung |

Mã đề thi mẫu sẵn có: **`IT2026`** (Kỳ thi Kiểm tra Kiến thức Lập trình & Cơ sở dữ liệu).

---

## 6. Chạy Kiểm thử Tự động (Integration Tests)

Chạy toàn bộ 22 kịch bản kiểm thử tích hợp tự động bao phủ tất cả các phân hệ:
```bash
cd backend
go test -v ./...
```
Toàn bộ kịch bản kiểm thử chạy trên cơ sở dữ liệu SQLite in-memory giả lập đầy đủ ràng buộc ACID và xử lý kiểu dữ liệu thời gian thực.
