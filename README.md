# Hệ Thống Thi Trắc Nghiệm Trực Tuyến (ExamMVP)

Hệ thống thi trắc nghiệm trực tuyến chuẩn MVP (Minimum Viable Product), được thiết kế theo kiến trúc Client-Server hiện đại, bảo mật cao và tối ưu hiệu năng:

- **Backend**: Golang (Gin Framework, GORM ORM, JWT Authentication, Bcrypt, CORS).
- **Database**: PostgreSQL 15 (Docker Compose), AutoMigrate các bảng dữ liệu, gồm lịch sử bài làm/khiếu nại, cùng các Unique Composite Index chống trùng lặp.
- **Frontend**: React 19, Vite, TailwindCSS, React Router DOM, Axios với Request Interceptor tự động đính kèm Bearer JWT.
- **Bảo mật & Nghiệp vụ**: Phân quyền teacher/student; lịch sử bài làm, khiếu nại và phản hồi; chống lộ đáp án trong lúc thi, chống gian lận thời gian (Server-authoritative), chống Race Condition (DB Transaction & Row-level Locking `FOR UPDATE`), chống IDOR chéo đề thi, và chống SQL Injection.
- **Database**: PostgreSQL 15 (Docker Compose), AutoMigrate 5 bảng với các Unique Composite Index chống trùng lặp dữ liệu.
- **Frontend**: React 19, Vite, TailwindCSS, React Router DOM, Axios với Request Interceptor tự động đính kèm Bearer JWT.
- **Bảo mật & Nghiệp vụ**: Chống lộ đáp án trắc nghiệm, chống gian lận thời gian (Server-authoritative), chống Race Condition (DB Transaction & Row-level Locking `FOR UPDATE`), chống IDOR chéo đề thi, và chống SQL Injection.
- **Kiểm thử tự động (Test Suite)**: Bộ Integration Test tự động bao phủ 100% qua toàn bộ Use Cases bằng gói `testing` và `net/http/httptest` của Go.

> 📖 **Tài liệu kiến trúc chuyên sâu**: Xem chi tiết sơ đồ cơ sở dữ liệu ERD, sơ đồ luồng nghiệp vụ Sequence Diagram, báo cáo phân tích bảo mật và tài liệu API tại file [ARCHITECTURE_AND_REPORT.md](file:///d:/Dung/USTH/CMC/CCYBER/RND/exam-system-mvp/ARCHITECTURE_AND_REPORT.md).

---

## 🚀 Hướng Dẫn Khởi Chạy Hệ Thống

Để khởi chạy toàn bộ hệ thống trên môi trường cục bộ, mở **3 cửa sổ Terminal (PowerShell hoặc CMD)**:

### 🔹 Bước 1: Khởi động Database (PostgreSQL qua Docker)
Mở Terminal 1 tại thư mục gốc dự án:
```powershell
docker-compose up -d
```
*Database PostgreSQL sẽ khởi chạy tại cổng `5432` (User: `postgres`, Password: `postgres`, Database: `exam_mvp`).*

---

### 🔹 Bước 2: Khởi động Backend (Golang - Port 8081)
Mở Terminal 2 tại thư mục gốc:
```powershell
cd backend
go run main.go
```
Khi khởi động, Backend sẽ tự động:
1. Kết nối PostgreSQL và thực hiện `AutoMigrate` cho các bảng dữ liệu, gồm `appeals`, cùng các Unique Index.
1. Kết nối PostgreSQL và thực hiện `AutoMigrate` cho 5 bảng dữ liệu kèm các Unique Index.
2. Kiểm tra và nạp sẵn dữ liệu mẫu (Seed Data) gồm tài khoản Giáo viên, Học sinh và Đề thi demo `IT2026`.
3. Lắng nghe HTTP request tại địa chỉ: **`http://localhost:8081`**.

---

### 🔹 Bước 3: Khởi động Frontend (React Vite - Port 5173)
Mở Terminal 3 tại thư mục gốc:
```powershell
cd frontend
npm run dev
```
*(Trường hợp PowerShell báo lỗi đường dẫn có ký tự đặc biệt, chạy trực tiếp: `node .\node_modules\vite\bin\vite.js`)*

Truy cập ứng dụng trên trình duyệt tại:
👉 **[http://localhost:5173](http://localhost:5173)**

---

## 🔑 Tài Khoản Test Mẫu (Demo Accounts)

Màn hình Đăng nhập (`http://localhost:5173/login`) đã tích hợp sẵn **2 nút bấm tự động điền nhanh** thông tin đăng nhập:

| Vai trò | Email đăng nhập | Mật khẩu | Quyền hạn & Chức năng |
| :--- | :--- | :--- | :--- |
| **Giáo viên** *(Teacher)* | `teacher@test.com` | `123456` | Tạo đề thi mới, soạn câu hỏi trắc nghiệm, xuất bản đề thi (`publish`), theo dõi bảng điểm và 4 chỉ số thống kê kết quả |
| **Student 1** | `student@test.com` | `123456` | Tài khoản học sinh demo |
| **Student 2** | `student2@test.com` | `123456` | Tài khoản học sinh demo |
| **Student 3** | `student3@test.com` | `123456` | Tài khoản học sinh demo |

Ba tài khoản student có cùng quyền: tra cứu kỳ thi `IT2026`, làm bài, xem lịch sử/chi tiết và gửi khiếu nại. Các tài khoản demo được bổ sung an toàn khi backend khởi động, kể cả database đã có dữ liệu. Người dùng cũng có thể tự tạo tài khoản tại `/register`.
| **Học sinh** *(Student)* | `student@test.com` | `123456` | Tra cứu kỳ thi bằng mã `IT2026`, làm bài thi có đồng hồ đếm ngược, tự động lưu bài (auto-save), nộp bài và xem điểm số chi tiết |

*(Học sinh cũng có thể tự tạo tài khoản mới tại trang Đăng ký `/register`).*

---

## 🧪 Chạy Kiểm Thử Tự Động (Integration Tests)

Dự án tích hợp bộ test tự động tại `backend/main_test.go` (sử dụng in-memory SQLite độc lập, không ảnh hưởng database thật):

```powershell
cd backend
go test -v ./...
```

**Kết quả kiểm thử:**
- `TestMVP_FullFlow`: Chạy xuyên suốt UC01 -> UC06 (Đăng ký, Đăng nhập, Tạo đề, Xuất bản, Tra cứu, Bắt đầu thi, Auto-save, Resume khôi phục phiên, Nộp bài, Tự động chấm điểm, Xem bảng thống kê).
- `TestHealth`: Kiểm tra endpoint `/health`.
- `TestAuth_BadInputs`: Kiểm tra validate email/password và chống trùng lặp tài khoản.
- `TestTeacher_ExamLifecycle`: Kiểm tra vòng đời đề thi và ngăn chặn sửa câu hỏi khi đề đã xuất bản.
- `TestStudent_GradingAccuracy`: Kiểm tra tính toán điểm số chính xác và chấm câu chưa làm thành câu sai.
- `TestStudentHistoryAndAppealsOwnershipAndWorkflow`: Kiểm tra lịch sử/chi tiết bài làm, quyền sở hữu, validation khiếu nại, lọc trạng thái và luồng phản hồi/từ chối/xử lý.
- `TestSeedDataEnsuresDemoUsersOnExistingDatabase`: Kiểm tra bổ sung tài khoản demo vào database đã có dữ liệu và seed lặp không tạo bản ghi trùng.
- SQLite integration tests cần môi trường có CGO và C compiler; `go test -run '^$' ./...` có thể dùng để kiểm tra compile khi môi trường không bật CGO.
- **Tất cả các test case đều PASS 100%.**
