# Hệ thống Thi trắc nghiệm Trực tuyến (Online Exam System)

Đồ án xây dựng hệ thống thi trắc nghiệm trực tuyến phục vụ kiểm tra, đánh giá trong trường đại học. Hệ thống tách biệt giữa Backend (Golang, Gin API) và Frontend (React SPA), hỗ trợ quản lý lớp học phần, ngân hàng câu hỏi, tổ chức thi trắc nghiệm thời gian thực, chống gian lận, chấm điểm tự động, phúc khảo, quản lý hồ sơ và danh bạ người dùng.

---

## Các tính năng chính

### 1. Quản lý lớp học phần và ghi danh (Classrooms)
- **Tạo và quản lý lớp**: Giảng viên tạo lớp học phần, hệ thống cấp phát mã tham gia ngẫu nhiên 6 ký tự viết hoa (loại trừ các ký tự dễ nhầm lẫn như 0, O, 1, I).
- **Ghi danh bằng mã**: Sinh viên chủ động nhập mã lớp học để ghi danh tham gia lớp học phần.
- **Quản lý thành viên**: Giảng viên có thể tra cứu danh sách sinh viên trong lớp, thêm sinh viên theo email hoặc mã sinh viên, và xóa sinh viên khỏi lớp.
- **Phân bổ đề thi theo lớp**: Đề thi có thể được gán riêng cho từng lớp học (chỉ sinh viên đã ghi danh vào lớp mới được làm bài) hoặc công khai cho toàn bộ sinh viên.

### 2. Quản lý hồ sơ cá nhân và danh bạ người dùng (Profile & Directory)
- **Hồ sơ cá nhân**: Người dùng xem và cập nhật thông tin định danh gồm họ tên, khoa/bộ môn, mã số sinh viên/giảng viên, tiểu sử (bio).
- **Tải ảnh đại diện**: Hỗ trợ dán URL ảnh trực tiếp hoặc tải tệp ảnh từ máy tính (ảnh được chuyển đổi sang chuẩn Base64 Data URI lưu trữ an toàn trong cơ sở dữ liệu).
- **Đổi mật khẩu**: Người dùng tự đổi mật khẩu với xác thực tối thiểu 6 ký tự, mã hóa an toàn qua thuật toán Bcrypt.
- **Danh bạ người dùng**: Tra cứu thông tin giảng viên và sinh viên toàn trường, tìm kiếm đa trường (tên, email, mã số, khoa bộ môn), lọc theo vai trò.
- **Phân quyền người dùng**: Giảng viên có quyền chuyển đổi vai trò của tài khoản (Giảng viên <-> Sinh viên), có chặn thao tác tự đổi vai trò của chính mình.

### 3. Quản lý đề thi và câu hỏi (CRUD)
- **Đề thi**: Tạo, sửa thông tin (tiêu đề, mô tả, thời lượng, khung giờ mở/đóng ca thi, lớp học phần phụ trách), xuất bản đề thi và xóa đề thi. Khi xóa đề thi, hệ thống dùng database transaction để dọn sạch các câu hỏi, lượt làm bài và đơn phúc khảo liên quan (cascade delete).
- **Câu hỏi**: Thêm câu hỏi thủ công (nội dung, 4 đáp án A/B/C/D, đáp án đúng, điểm số), sửa câu hỏi, xóa câu hỏi khi đề còn ở trạng thái nháp (`draft`).
- **Nhập câu hỏi hàng loạt (Bulk Import)**: Hỗ trợ copy/paste văn bản phân cách bằng dấu `---` hoặc mảng JSON để nhập nhiều câu hỏi cùng lúc qua API `POST /api/teacher/exams/:id/questions/bulk`.

### 4. Quản lý điểm số và lượt làm bài
- **Sửa điểm**: Giảng viên có thể điều chỉnh điểm của sinh viên kèm lý do vào trường `teacher_note`. Thao tác này được lưu vết tự động vào `audit_logs` và gửi thông báo đến sinh viên.
- **Cho thi lại (Reset Attempt)**: Xóa lượt làm bài trong transaction để giải phóng ràng buộc `UNIQUE (exam_id, user_id)`. Sinh viên được phép vào thi lại từ đầu khi gặp sự cố mạng hoặc lỗi thiết bị. Thao tác này cũng được ghi vết vào audit log.

### 5. Quản lý tài khoản sinh viên
- Giảng viên có thể xem danh sách sinh viên, tìm kiếm theo tên hoặc email (`?q=...`), xem số lượt thi và số bài đã nộp.
- Hỗ trợ thêm tài khoản sinh viên mới, cập nhật thông tin/đặt lại mật khẩu và xóa tài khoản (kèm dọn dẹp dữ liệu bài làm trong transaction).

### 6. Phòng thi trực tuyến và giao diện làm bài
- **Trang chủ sinh viên**: Hiển thị các đề thi đang mở kèm trạng thái làm bài tương ứng (`not_attempted`, `in_progress`, `submitted`, `time_up`). Hỗ trợ tìm đề thi nhanh bằng mã đề.
- **Phòng thi 2 cột**:
  - Cột trái: Hiển thị nội dung câu hỏi, 4 phương án lựa chọn, nút đặt cờ xem lại, nút chuyển câu trước/sau.
  - Cột phải: Bảng số câu hỏi (Question Palette) hiển thị trạng thái từng câu (xanh: đã lưu, vàng: đặt cờ, xám: chưa làm, viền đậm: câu đang xem). Bấm vào số câu để chuyển nhanh.
  - Tự động lưu đáp án khi chọn phương án (UPSERT `ON CONFLICT DO UPDATE`).

### 7. Chống gian lận trong phòng thi
- **Trộn câu hỏi theo seed**: Sử dụng `attempt.ID` làm seed ngẫu nhiên để xáo trộn thứ tự câu hỏi cho từng sinh viên. Khi sinh viên F5 hoặc đổi máy, thứ tự câu vẫn giữ nguyên không bị đảo lộn.
- **Phát hiện chuyển tab (Tab-switch Detection)**: Lắng nghe sự kiện `visibilitychange` và `window.onblur`. Mỗi lần sinh viên rời khỏi màn hình thi, hệ thống gọi API `POST /api/student/attempts/:id/violation` để tăng biến đếm `violation_count`. Nếu vi phạm từ 3 lần trở lên, bài thi sẽ bị tự động thu bài và chấm điểm ngay lập tức.
- **Đồng hồ đếm ngược từ server**: Server tính toán số giây còn lại (`remaining_seconds`) dựa trên `started_at` và `duration`, ngăn việc gian lận bằng cách chỉnh đồng hồ máy trạm.
- **Không lộ đáp án đúng (Zero-Leak)**: API phòng thi sử dụng DTO `QuestionSafe` đã lọc bỏ trường `correct_option`. Đáp án đúng chỉ được trả về sau khi ca thi kết thúc (`now.After(exam.EndTime)`).

### 8. Phúc khảo bài thi và thông báo có điều hướng sâu
- **Quy trình phúc khảo 4 trạng thái**: Sinh viên xem lại bài làm sau khi thi và có thể gửi đơn phúc khảo (tối đa 1 đơn mỗi bài thi). Trạng thái xử lý tuân theo máy trạng thái: `PENDING` -> `REPLIED` -> `REJECTED` hoặc `RESOLVED`.
- **Thông báo điều hướng sâu (Deep-link Navigation)**: Tự động gửi thông báo khi có đơn phúc khảo mới, có phản hồi phúc khảo, sửa điểm, cấp quyền thi lại hoặc tham gia lớp học. Mỗi thông báo lưu kèm `target_url` cho phép bấm vào là chuyển hướng ngay đến đúng màn hình chi tiết tương ứng.
- **Quản lý danh sách thông báo**: Hỗ trợ đánh dấu đã đọc từng thông báo hoặc toàn bộ, xóa từng thông báo thủ công hoặc nút xóa nhanh toàn bộ thông báo đã đọc để tránh làm dài danh sách.
- **Polling thời gian thực**: Frontend kiểm tra thông báo nền định kỳ mỗi 5 giây để cập nhật tức thời số lượng thông báo mới.

### 9. Thời khóa biểu & Lịch thi trực quan (/schedule)
- **Phân nhóm thông minh theo mốc thời gian**: Tự động nhóm các kỳ thi thành 3 nhóm trạng thái: "Đang diễn ra", "Sắp diễn ra" và "Đã kết thúc".
- **Đồng hồ đếm ngược trực tiếp**: Hiển thị thời gian còn lại (đối với ca thi đang mở) hoặc thời gian chờ mở ca thi (đối với ca thi sắp diễn ra) cập nhật theo từng giây.
- **Lọc theo phạm vi & Lớp học**: Sinh viên chỉ thấy các ca thi toàn trường hoặc thuộc các lớp học phần mình đã ghi danh. Giảng viên theo dõi toàn bộ các ca thi do mình quản lý kèm số lượt sinh viên đã thi.
- **Điều hướng nhanh**: Cung cấp nút CTA "Vào phòng thi ngay", "Tiếp tục làm bài", "Xem kết quả" (Sinh viên) hoặc "Cấu hình đề", "Xem bảng điểm" (Giảng viên).

### 10. Xuất bảng điểm CSV và bảo mật
- **Xuất CSV UTF-8 BOM**: API `GET /api/teacher/exams/:id/results/export` xuất bảng điểm dạng CSV có gắn UTF-8 BOM (`\xef\xbb\xbf`) để mở trực tiếp bằng Excel không bị lỗi font tiếng Việt.
- **Rate Limiting**: Giới hạn 15 request/phút/IP trên các endpoint nhạy cảm (`/api/auth/login`, `/api/student/exams/lookup`) bằng sliding window in-memory.
- **HTTP Security Headers**: Middleware tự động gắn `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: strict-origin-when-cross-origin` và giới hạn request body tối đa 2MB.
- **SQLite WAL Mode**: Cấu hình `PRAGMA journal_mode=WAL` và `busy_timeout=5000` giúp môi trường test chạy song song mượt mà, không bị lock database.

### 11. Bảng điểm cá nhân & Cảnh báo học vụ (/student/grades)
- **Học bạ tín chỉ chuẩn đại học**: Thống kê tự động CPA (toàn khóa) và GPA từng học kỳ dựa trên kết quả thi thực tế từ database.
- **Biểu đồ trực quan hóa**: Cung cấp chế độ xem biến thiên điểm số theo môn học và biểu đồ phân bổ tín chỉ học lực theo thang chữ (A, B+, B, C+, C, D, F).
- **Hệ thống Cảnh báo học vụ 3 cấp độ**: Tự động rà soát điểm số và số lượng học phần chưa đạt, phát cảnh báo học vụ Mức 1, 2 hoặc 3 theo đúng quy chế đào tạo tín chỉ của Bộ GD&ĐT.
- **Bộ dự báo mục tiêu GPA**: Hỗ trợ sinh viên tính toán số tín chỉ và điểm cần đạt để nâng hạng học lực tốt nghiệp (Khá, Giỏi, Xuất sắc).

### 12. Sổ điểm lớp học tổng hợp (Class Gradebook)
- **Bảng ma trận điểm số chi tiết**: Giảng viên theo dõi toàn bộ thành viên trong lớp kèm điểm số từng bài thi, điểm trung bình lớp và xếp loại học lực.
- **Thống kê tổng quan thời gian thực**: Đo lường điểm trung bình lớp, biên độ điểm ($Max / Min$) và số lượng sinh viên có nguy cơ phải thi lại (Điểm TB $< 4.0$).
- **Bộ lọc đa tiêu chí**: Tra cứu nhanh theo tên, email, mã số sinh viên hoặc lọc theo danh mục học lực (*Xuất sắc*, *Giỏi*, *Khá*, *Trung bình*, *Yếu / Cảnh báo*).
- **Xuất báo cáo lớp học**: Một click xuất toàn bộ sổ điểm ra file `.csv` chuẩn UTF-8 BOM mở bằng Microsoft Excel.

### 13. Giao diện Đa chế độ Sáng / Tối (Dark / Light Mode)
- **Chuyển đổi tức thời (Zero-flicker)**: Hỗ trợ dark mode toàn diện trên tất cả các trang, thanh sidebar, modal và biểu đồ SVG.
- **Lưu trữ trạng thái người dùng**: Tự động lưu lựa chọn theme vào `localStorage` và đồng bộ với cài đặt hệ điều hành.

### 14. Bảo mật giao dịch & Phòng chống rớt mạng
- **Giao dịch nguyên tử khi nộp bài (`GradeAttemptAtomic`)**: Bọc quá trình chấm điểm, lưu đáp án và chuyển trạng thái `completed` trong một GORM Database Transaction duy nhất, loại bỏ triệt để rủi ro lỗi trạng thái dở dang.
- **Bộ đệm Offline LocalStorage**: Tự động sao lưu đáp án đang chọn vào máy trạm, tự phục hồi khi tải lại trang và hỗ trợ nộp bài an toàn khi mạng WiFi bị chập chờn.
- **Khóa hạn nộp bài nghiêm ngặt từ Server**: Chặn đứng các hành vi gian lận can thiệp client-side để gửi đáp án sau khi hết giờ ca thi.
- **Giới hạn thời hạn phúc khảo**: Giới hạn nhận đơn trong vòng 14 ngày kể từ khi kết thúc kỳ thi và ngăn chặn hành vi spam nhiều đơn cùng lúc.

---

## Cấu trúc thư mục dự án

```
exam-system-mvp/
├── ARCHITECTURE_AND_REPORT.md   # Báo cáo kiến trúc và đặc tả kỹ thuật chi tiết
├── README.md                    # Tài liệu hướng dẫn sử dụng và khởi chạy
├── docker-compose.yml           # Cấu hình container 3 tầng (PostgreSQL 15, Go Backend, Nginx Frontend)
├── backend/                     # Backend API (Golang Gin)
│   ├── Dockerfile               # Multi-stage Docker build cho Backend (<20MB)
│   ├── config/
│   │   └── database.go          # Kết nối CSDL PostgreSQL/SQLite, bật WAL Mode
│   ├── dto/                     # Data Transfer Objects & Safe Schemas
│   ├── handlers/
│   │   ├── anti_cheat_audit.go  # Xử lý vi phạm chuyển tab, audit log, xuất CSV
│   │   ├── appeal_handler.go    # Xử lý đơn khiếu nại, phúc khảo điểm
│   │   ├── auth.go              # Đăng ký, đăng nhập, mã hóa bcrypt, JWT
│   │   ├── classroom.go         # API nghiệp vụ lớp học phần và Sổ điểm
│   │   ├── exam_handler.go      # Quản lý đề thi, câu hỏi và xuất bản
│   │   ├── exam_taking_handler.go # Phòng thi trực tuyến, auto-save, nộp bài
│   │   ├── history_appeals.go   # Lịch sử thi, quy trình xử lý đơn phúc khảo
│   │   ├── notifications.go     # Thông báo nội bộ, đánh dấu đã đọc
│   │   ├── schedule.go          # Lịch thi và thời khóa biểu
│   │   ├── teacher_crud.go      # CRUD đề thi, sinh viên, sửa điểm, reset bài
│   │   └── user.go              # Hồ sơ, danh bạ, phân quyền người dùng
│   ├── middleware/
│   │   ├── auth.go              # Xác thực JWT, phân quyền Teacher/Student
│   │   ├── ratelimit.go         # Rate limiter 15 req/phút/IP
│   │   └── security.go          # Security headers, giới hạn body 2MB
│   ├── models/
│   │   ├── migrate.go           # AutoMigrate 10 bảng, tạo composite unique index
│   │   ├── models.go            # GORM Structs (Users, Exams, Classrooms,...)
│   │   └── seed.go              # Dữ liệu mẫu (tài khoản demo, đề thi IT2026)
│   ├── repository/              # Data Access Layer & DB Transactions
│   ├── router/
│   │   └── router.go            # Định tuyến API, cấu hình CORS và middleware
│   ├── service/                 # Tầng logic nghiệp vụ độc lập
│   ├── utils/
│   │   └── helpers.go           # Hàm tiện ích (random code, shuffle slice)
│   ├── main.go                  # Entrypoint server (cổng 8081)
│   └── main_test.go             # 26 kịch bản kiểm thử tích hợp tự động
└── frontend/                    # Frontend SPA (React 19 Vite)
    ├── Dockerfile               # Multi-stage Docker build cho Frontend
    ├── nginx.conf               # Nginx reverse proxy, security headers & Gzip
    ├── src/
    │   ├── api/
    │   │   ├── axios.js         # Axios BaseURL thích ứng Docker/Local & JWT Interceptor
    │   │   └── endpoints.js     # Khai báo các hàm gọi API (Auth, Teacher, Student, Class,...)
    │   ├── components/
    │   │   ├── NotificationDropdown.jsx # Menu thông báo (polling 5s, deep-link navigate)
    │   │   ├── ProtectedRoute.jsx       # Bảo vệ route theo vai trò đăng nhập
    │   │   ├── Spinner.jsx              # Loading spinner
    │   │   ├── StudentLayout.jsx        # Layout Sidebar sinh viên kèm theme toggle
    │   │   ├── TeacherLayout.jsx        # Layout Sidebar giảng viên kèm theme toggle
    │   │   └── ThemeToggle.jsx          # Nút chuyển đổi Dark / Light mode
    │   ├── contexts/
    │   │   ├── AuthContext.jsx  # Quản lý phiên đăng nhập và dữ liệu người dùng
    │   │   └── ThemeContext.jsx # Quản lý trạng thái theme Dark / Light toàn cục
    │   ├── pages/
    │   │   ├── LoginPage.jsx            # Đăng nhập
    │   │   ├── RegisterPage.jsx         # Đăng ký tài khoản sinh viên
    │   │   ├── ProfilePage.jsx          # Hồ sơ cá nhân (Avatar URL/Base64, Bio, Đổi mật khẩu)
    │   │   ├── DirectoryPage.jsx        # Danh bạ người dùng, tra cứu và đổi vai trò
    │   │   ├── ClassroomsPage.jsx       # Danh sách lớp học (Tạo lớp / Tham gia bằng mã)
    │   │   ├── ClassDetailPage.jsx      # Chi tiết lớp (Thành viên, Đề thi & Sổ điểm tổng hợp)
    │   │   ├── ExamSchedulePage.jsx     # Thời khóa biểu và lịch thi
    │   │   ├── student/
    │   │   │   ├── ExamResultPage.jsx          # Kết quả thi và cảnh báo vi phạm
    │   │   │   ├── ExamTakingPage.jsx          # Giao diện phòng thi (Question Palette & Offline Cache)
    │   │   │   ├── StudentGradesPage.jsx       # Bảng điểm cá nhân, biểu đồ GPA & Cảnh báo học vụ
    │   │   │   ├── StudentHistoryDetailPage.jsx# Xem lại bài làm, gửi đơn phúc khảo
    │   │   │   ├── StudentHistoryPage.jsx      # Danh sách lịch sử bài thi
    │   │   │   └── StudentHomePage.jsx         # Đề thi đang mở, vào thi bằng mã đề
    │   │   └── teacher/
    │   │       ├── ExamAppealsTab.jsx          # Xử lý đơn phúc khảo
    │   │       ├── TeacherExamDetailPage.jsx   # Chi tiết đề thi, câu hỏi, bảng điểm
    │   │       ├── TeacherExamsPage.jsx        # Danh sách đề thi và gán lớp học
    │   │       ├── TeacherResultsPage.jsx      # Bảng điểm, sửa điểm, reset bài, xuất CSV
    │   │       └── TeacherStudentsPage.jsx     # Quản lý danh sách sinh viên
    │   ├── App.jsx              # Cấu hình routes và SharedRoleLayout
    │   └── main.jsx             # React entrypoint
    ├── package.json
    └── vite.config.js           # Vite dev server (cổng 5173)
```

---

## Công nghệ sử dụng

- **Backend**: Golang 1.21+ (Gin Web Framework, GORM ORM, golang-jwt/jwt/v5, bcrypt).
- **Cơ sở dữ liệu**: PostgreSQL 15 (Docker), hỗ trợ transaction ACID, khóa dòng `SELECT ... FOR UPDATE`, lệnh nguyên tử `ON CONFLICT DO UPDATE` và cascade foreign key.
- **Frontend**: React 19, Vite, TailwindCSS, React Router v6, Axios.
- **Kiểm thử**: Gói `testing` và `net/http/httptest` của Go, chạy 25 kịch bản kiểm thử tích hợp trên SQLite in-memory (WAL mode).

---

## Hướng dẫn khởi chạy

### Cách 1: Khởi chạy trọn gói bằng Docker Compose (Khuyên dùng cho Báo cáo & Nghiệm thu)

Toàn bộ hệ thống (PostgreSQL, Golang Backend, Nginx + React Frontend) đã được container hóa hoàn chỉnh. Bạn chỉ cần gõ đúng 1 lệnh duy nhất tại thư mục gốc:

```bash
docker compose up --build -d
```

Sau khi các container khởi động hoàn tất:
- **Giao diện người dùng (Frontend)**: Truy cập tại **`http://localhost:3000`**
- **API Backend**: Truy cập tại **`http://localhost:8081`** (Kiểm tra trạng thái tại `http://localhost:8081/health`)
- **Cơ sở dữ liệu PostgreSQL**: Cổng `5432` (`User: postgres`, `Password: postgres`, `Database: exam_mvp`)

Để dừng hệ thống:
```bash
docker compose down
```

---

### Cách 2: Khởi chạy thủ công từng phần (Dành cho Lập trình viên)

Yêu cầu môi trường: Docker Desktop (để chạy PostgreSQL), Golang 1.21+ và Node.js 18+.

#### Bước 1: Khởi động cơ sở dữ liệu (PostgreSQL)
Chạy lệnh sau tại thư mục gốc của dự án:
```bash
docker compose up postgres -d
```
PostgreSQL sẽ chạy tại cổng `5432` (`User: postgres`, `Password: postgres`, `Database: exam_mvp`).

#### Bước 2: Khởi động Backend (Golang)
Mở terminal tại thư mục `backend`:
```bash
cd backend
go run main.go
```
Backend tự động chạy AutoMigrate 10 bảng và tạo dữ liệu mẫu, lắng nghe tại cổng **`8081`** (`http://localhost:8081`).

#### Bước 3: Khởi động Frontend (React)
Mở terminal tại thư mục `frontend`:
```bash
cd frontend
npm install
npm run dev
```
Truy cập giao diện tại: **`http://localhost:5173`**.

---

## Tài khoản thử nghiệm mẫu

Hệ thống có sẵn các tài khoản demo:

| Vai trò | Email | Mật khẩu | Mô tả quyền hạn |
| :--- | :--- | :--- | :--- |
| **Giảng viên** | `teacher@test.com` | `123456` | Quản lý đề thi, câu hỏi, lớp học, điểm số, sinh viên, phúc khảo, danh bạ |
| **Sinh viên 1** | `student@test.com` | `123456` | Làm bài thi, Question Palette, tự động lưu, tham gia lớp học, xem lịch sử và phúc khảo |
| **Sinh viên 2** | `student2@test.com` | `123456` | Tài khoản sinh viên thử nghiệm |
| **Sinh viên 3** | `student3@test.com` | `123456` | Tài khoản sinh viên thử nghiệm |

Mã đề thi mẫu có sẵn: **`IT2026`** (Kiểm tra Kiến thức Lập trình & Cơ sở dữ liệu).

---

## Chạy kiểm thử tự động

Chạy 26 kịch bản kiểm thử tích hợp toàn diện trong backend:
```bash
cd backend
go test -v ./...
```
Tất cả các test case đều chạy trên SQLite in-memory với các bảng và ràng buộc tương tự môi trường thật.

---

## Tài liệu API & Kiểm thử với Postman / Swagger

Hệ thống được trang bị đầy đủ tài liệu API chuẩn công nghiệp nằm trong thư mục `docs/`:

1. **Postman Collection (`docs/exam_system_postman_collection.json`)**:
   - Chứa toàn bộ các REST API được phân nhóm theo nghiệp vụ (Auth, Exam Management, Exam Taking Flow, History & Appeals, Profile, Notifications).
   - Tích hợp sẵn biến môi trường `baseUrl` (`http://localhost:8081`).
   - Tự động trích xuất và lưu Bearer JWT token sau khi gọi API đăng nhập, không cần copy/paste thủ công.
   - **Cách sử dụng**: Mở phần mềm Postman -> Chọn nút **Import** -> Kéo thả tệp `docs/exam_system_postman_collection.json`.

2. **OpenAPI / Swagger 3.0 Specification (`docs/openapi.yaml`)**:
   - Đặc tả cấu trúc Request/Response, DTOs, mã trạng thái HTTP theo chuẩn OpenAPI 3.0.
   - Có thể mở trực tiếp trên [Swagger Editor](https://editor.swagger.io) hoặc import vào các công cụ API testing khác.
