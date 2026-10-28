# Hệ Thống Thi Trắc Nghiệm Trực Tuyến (ExamMVP)

Hệ thống thi trắc nghiệm trực tuyến hoàn chỉnh chuẩn MVP (Minimum Viable Product), xây dựng với kiến trúc Client-Server hiện đại:
- **Backend**: Golang (Gin Gonic, GORM, JWT Auth, Bcrypt, CORS).
- **Database**: PostgreSQL 15 (Docker Compose), AutoMigrate 5 bảng với unique indexes.
- **Frontend**: React 19, Vite, TailwindCSS, React Router DOM, Axios với JWT Interceptor.
- **Test Suite**: Kiểm thử tích hợp tự động qua 6 Use Cases bằng Go `testing` + `net/http/httptest` (**PASS 100%**).

---

## 🚀 Hướng Dẫn Khởi Chạy Toàn Bộ Hệ Thống

Để khởi chạy và trải nghiệm trực tiếp trên trình duyệt, mở **3 cửa sổ Terminal (PowerShell hoặc CMD)**:

### 🔹 Bước 1: Khởi động Database (PostgreSQL)
Mở Terminal 1 tại thư mục gốc của dự án:
```powershell
docker-compose up -d
```
> Database PostgreSQL sẽ chạy ở cổng `5432`, user `postgres`, pass `postgres`, database `exam_mvp`.

---

### 🔹 Bước 2: Khởi động Backend (Golang - Port 8081)
Mở Terminal 2 tại thư mục gốc:
```powershell
cd backend
go run main.go
```
Khi khởi động lần đầu, Backend sẽ tự động:
1. Kết nối PostgreSQL và chạy `AutoMigrate` cho đúng 5 bảng.
2. Tự động kiểm tra và seed dữ liệu mẫu (Tài khoản Teacher, Student và Đề thi `IT2026`).
3. Lắng nghe tại: **`http://localhost:8081`**.

---

### 🔹 Bước 3: Khởi động Frontend (React Vite - Port 5173)
Mở Terminal 3 tại thư mục gốc:
```powershell
cd frontend
npm run dev
```
*(Nếu gặp thông báo về đường dẫn PowerShell chứa ký tự đặc biệt, chạy lệnh trực tiếp:*
`node .\node_modules\vite\bin\vite.js`*)*

Truy cập giao diện Web tại:
👉 **[http://localhost:5173](http://localhost:5173)**

---

## 🔑 Tài Khoản Demo Sẵn Có (Seed Data)

Trang Đăng nhập (`http://localhost:5173/login`) đã tích hợp sẵn **2 nút bấm tự điền nhanh**:

| Vai trò | Email | Mật khẩu | Chức năng chính |
| :--- | :--- | :--- | :--- |
| **Giáo viên** | `teacher@test.com` | `123456` | Tạo đề thi, soạn câu hỏi trắc nghiệm, xuất bản (publish), xem thống kê kết quả thí sinh |
| **Học sinh** | `student@test.com` | `123456` | Nhập mã `IT2026` tra cứu đề, làm bài có đếm ngược, tự động lưu (auto-save), nộp bài & xem điểm |

*(Bạn cũng có thể vào trang `/register` để tự đăng ký một tài khoản Học sinh mới bất kỳ).*

---

## 🧪 Chạy Kiểm Thử Tự Động (Integration Tests)

Dự án đi kèm bộ test toàn diện trong `backend/main_test.go` bao quát đầy đủ 6 Use Case nghiệp vụ:
- **UC01**: Đăng ký Student mới, kiểm tra trùng lặp email (409), đăng nhập Teacher & Student nhận JWT token.
- **UC02**: Teacher tạo đề thi (`draft`), thêm câu hỏi, xuất bản (`published`). Kiểm tra bảo mật: cấm thêm câu hỏi khi đã xuất bản (400), cấm Student truy cập API Teacher (403).
- **UC03**: Student tra cứu đề thi bằng mã `exam_code`, kiểm tra trạng thái theo thời gian thực.
- **UC04**: Student bắt đầu thi (`/start`), bảo mật không để lộ `correct_option`, auto-save đáp án, tính năng Resume khi tải lại trang khôi phục chính xác câu trả lời và thời gian làm bài còn lại.
- **UC05**: Nộp bài (`/submit`) và tự động chấm điểm chính xác trong 1 Transaction (`score`, `correct_count`, `wrong_count`), ngăn chặn nộp bài 2 lần.
- **UC06**: Student xem kết quả sau nộp (ẩn đáp án đúng khi đề còn mở) và Teacher xem thống kê (tổng thí sinh, điểm TB, cao nhất/thấp nhất).

Để chạy test:
```powershell
cd backend
go test -v ./...
```
**Kết quả kiểm thử thực tế:**
```
PASS: TestMVP_FullFlow (UC01 -> UC06)
PASS: TestHealth
PASS: TestAuth_BadInputs
PASS: TestTeacher_ExamLifecycle
PASS: TestStudent_GradingAccuracy
ok   exam-system-mvp  1.066s (100% PASS)
```

---

## 📁 Cấu Trúc Mã Nguồn

```
exam-system-mvp/
├── docker-compose.yml              # Container PostgreSQL 15
├── README.md
│
├── backend/                        # Toàn bộ mã nguồn Golang
│   ├── config/database.go          # Kết nối GORM + cấu hình DB
│   ├── models/
│   │   ├── models.go               # 5 GORM Models + QuestionSafe DTO
│   │   ├── migrate.go              # AutoMigrate & Unique composite indexes
│   │   └── seed.go                 # Seeder tài khoản & đề mẫu IT2026
│   ├── middleware/auth.go          # JWT Auth & phân quyền Role (teacher/student)
│   ├── handlers/
│   │   ├── auth.go                 # Register, Login
│   │   ├── teacher.go              # 6 APIs quản lý đề thi, câu hỏi, kết quả
│   │   └── student.go              # 7 APIs lookup, start, auto-save, submit, result
│   ├── router/router.go            # Khởi tạo Gin Router & cấu hình CORS (5173, 3000)
│   ├── utils/helpers.go            # JWT sign/verify, sinh mã ngẫu nhiên exam_code
│   ├── main.go                     # Entrypoint server (port 8081)
│   └── main_test.go                # Bộ Integration Test tự động (6 Use Cases)
│
└── frontend/                       # Toàn bộ mã nguồn React Vite
    ├── src/
    │   ├── api/
    │   │   ├── axios.js            # Axios instance + Request Interceptor gắn Bearer JWT
    │   │   └── endpoints.js        # Khai báo các hàm gọi API
    │   ├── contexts/AuthContext.jsx# Quản lý phiên đăng nhập và User state
    │   ├── components/
    │   │   ├── ProtectedRoute.jsx  # Bảo vệ route theo quyền (teacher / student)
    │   │   ├── TeacherLayout.jsx   # Layout Header cho Giáo viên
    │   │   └── StudentLayout.jsx   # Layout Header cho Học sinh
    │   ├── pages/
    │   │   ├── LoginPage.jsx       # Đăng nhập + nút bấm điền nhanh Demo
    │   │   ├── RegisterPage.jsx    # Đăng ký học sinh
    │   │   ├── teacher/
    │   │   │   ├── TeacherExamsPage.jsx      # Danh sách đề + Modal tạo đề mới
    │   │   │   ├── TeacherExamDetailPage.jsx # Soạn thảo câu hỏi + Xuất bản đề
    │   │   │   └── TeacherResultsPage.jsx    # Bảng điểm & 4 thẻ thống kê
    │   │   └── student/
    │   │       ├── StudentHomePage.jsx       # Tra cứu đề thi theo mã + Lịch sử thi
    │   │       ├── ExamTakingPage.jsx        # Countdown Timer + Auto-save + Resume
    │   │       └── ExamResultPage.jsx        # Xem điểm, số câu đúng/sai & chi tiết
    │   ├── App.jsx                 # Routing phân hệ
    │   └── main.jsx
    └── package.json
```

---

## 📋 Danh Sách RESTful APIs

### Public Auth
- `POST /api/auth/register`: Đăng ký học sinh.
- `POST /api/auth/login`: Đăng nhập, trả về JWT Token và User Info.

### Phân Hệ Giáo Viên (`Bearer Token`, role = `teacher`)
- `GET /api/teacher/exams`: Danh sách đề thi do giáo viên tạo.
- `POST /api/teacher/exams`: Tạo đề thi mới (trạng thái `draft`, tự sinh `exam_code` 6 ký tự).
- `GET /api/teacher/exams/:id`: Xem chi tiết đề thi kèm danh sách câu hỏi.
- `POST /api/teacher/exams/:id/questions`: Thêm danh sách câu hỏi vào đề thi (`draft`).
- `DELETE /api/teacher/questions/:question_id`: Xóa câu hỏi khỏi đề thi (`draft`).
- `PATCH /api/teacher/exams/:id/publish`: Xuất bản đề thi (yêu cầu tối thiểu 1 câu hỏi).
- `GET /api/teacher/exams/:id/results`: Thống kê kết quả thi (tổng số HS, điểm TB, max/min, bảng điểm chi tiết).

### Phân Hệ Học Sinh (`Bearer Token`, role = `student`)
- `GET /api/student/exams/lookup?code=XYZ`: Tra cứu đề thi đã xuất bản theo mã.
- `GET /api/student/attempts`: Lịch sử các bài thi học sinh đã tham gia.
- `POST /api/student/exams/:id/start`: Bắt đầu hoặc Resume bài thi (không trả về `correct_option`).
- `POST /api/student/attempts/:id/save`: Auto-save câu trả lời (UPSERT vào database).
- `POST /api/student/attempts/:id/submit`: Nộp bài và chấm điểm tự động trong 1 Transaction.
- `GET /api/student/attempts/:id/result`: Xem kết quả bài thi sau nộp (chỉ hiện đáp án đúng khi đề thi đã kết thúc).
