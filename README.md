# Hệ thống Thi trắc nghiệm Trực tuyến (Exam MVP)

Hệ thống thi trắc nghiệm trực tuyến phục vụ công tác khảo thí nội bộ, được xây dựng theo kiến trúc Client-Server phân tầng.

## 1. Công nghệ sử dụng

- Backend: Golang (Gin Framework, GORM ORM, JWT, Bcrypt).
- Cơ sở dữ liệu: PostgreSQL 15 (Docker), hỗ trợ giao dịch ACID, khóa dòng bi quan và chỉ mục kết hợp duy nhất (Unique Composite Indexes).
- Frontend: React 19, Vite, TailwindCSS, React Router v6, Axios Interceptor tự động đính kèm JWT Token.
- Test Suite: Kiểm thử tích hợp tự động bao phủ 100% qua các Use Case bằng gói testing và net/http/httptest của Go.

## 2. Các phân hệ chức năng

### 2.1. Phân hệ Giảng viên
- Quản lý danh mục đề thi: Tạo mới đề thi với thời lượng và khung giờ mở/đóng đề thi.
- Soạn thảo nội dung: Thêm, sửa, xóa câu hỏi trắc nghiệm 4 lựa chọn, thiết lập đáp án đúng và trọng số điểm cho từng câu (chỉ cho phép thao tác khi đề ở trạng thái bản nháp).
- Xuất bản đề thi: Kiểm tra điều kiện đề thi có tối thiểu 1 câu hỏi trước khi mở cho thí sinh. Khóa chỉnh sửa nội dung câu hỏi sau khi đã xuất bản.
- Thống kê kết quả: Theo dõi danh sách bài thi đã nộp, điểm trung bình, điểm cao nhất, điểm thấp nhất và thời điểm nộp bài.
- Quản lý khiếu nại: Tiếp nhận đơn khiếu nại điểm thi từ sinh viên, xem lại chi tiết bài làm của thí sinh và thực hiện phản hồi, chấp thuận (đã xử lý) hoặc từ chối khiếu nại kèm ý kiến giải trình.

### 2.2. Phân hệ Sinh viên
- Tra cứu đề thi: Tìm kiếm đề thi hợp lệ bằng mã kỳ thi do giảng viên cung cấp.
- Làm bài thi: Thực hiện làm bài với đồng hồ đếm ngược do máy chủ kiểm soát. Cơ chế tự động lưu bài theo chu kỳ 60 giây hoặc ngay khi thí sinh chọn đáp án.
- Tự động thu bài: Khi hết giờ làm bài, hệ thống tự động khóa phiên và chấm điểm dựa trên các câu trả lời đã lưu hợp lệ trước đó.
- Xem kết quả thi: Xem điểm số, số câu đúng, số câu sai. Đáp án chuẩn chỉ hiển thị sau khi toàn bộ khung giờ thi kết thúc.
- Lịch sử bài làm: Tra cứu danh sách tất cả các bài thi đã hoàn thành kèm bộ lọc tìm kiếm và phân trang.
- Gửi khiếu nại điểm thi: Gửi đơn phúc khảo cho bài thi đã hoàn thành. Theo dõi tiến độ và nội dung phản hồi từ giảng viên phụ trách.

## 3. Cơ chế kỹ thuật và bảo mật

- Phòng chống lộ đáp án (Zero-Leak DTO): API bắt đầu làm bài sử dụng DTO QuestionSafe, loại bỏ trường correct_option trước khi tuần tự hóa JSON trả về trình duyệt.
- Kiểm soát thời gian từ phía máy chủ (Server-Authoritative): Toàn bộ hạn nộp bài và thời gian còn lại được tính toán dựa trên đồng hồ hệ thống của backend, có vùng đệm 5 giây bù trừ trễ mạng. Thao tác sửa giờ máy tính trên client không có tác dụng kéo dài thời gian thi.
- Chống Race Condition và Nộp bài trùng lặp: Áp dụng giao dịch cơ sở dữ liệu (DB Transaction) kết hợp khóa dòng bi quan (Pessimistic Locking: SELECT ... FOR UPDATE) tại các thao tác lưu bài và nộp bài.
- Chống lỗi IDOR: Kiểm tra nghiêm ngặt quyền sở hữu của câu hỏi đối với đề thi trong hàm upsertAnswers và quyền sở hữu bài làm khi gửi khiếu nại.
- Chống SQL Injection: 100% truy vấn sử dụng tham số hóa qua GORM và hàm parseUintParam kiểm tra kiểu dữ liệu đầu vào trên URL.

## 4. Hướng dẫn khởi chạy hệ thống

Chuẩn bị môi trường gồm Docker, Golang 1.21+ và Node.js 18+. Mở 3 cửa sổ Terminal để chạy các thành phần:

### Bước 1: Khởi động Cơ sở dữ liệu (PostgreSQL)
Chạy lệnh tại thư mục gốc của dự án:
```bash
docker-compose up -d
```
Cơ sở dữ liệu PostgreSQL sẽ lắng nghe tại cổng 5432 (User: postgres, Password: postgres, Database: exam_mvp).

### Bước 2: Khởi động Backend (Golang)
Chạy lệnh tại thư mục backend:
```bash
cd backend
go run main.go
```
Backend sẽ tự động thực hiện AutoMigrate 6 bảng (users, exams, questions, exam_attempts, attempt_answers, appeals) kèm các chỉ mục duy nhất, nạp dữ liệu mẫu và lắng nghe tại địa chỉ http://localhost:8081.

### Bước 3: Khởi động Frontend (React Vite)
Chạy lệnh tại thư mục frontend:
```bash
cd frontend
npm run dev
```
Truy cập ứng dụng trên trình duyệt tại địa chỉ: http://localhost:5173.

## 5. Tài khoản thử nghiệm hệ thống

Hệ thống đã khởi tạo sẵn các tài khoản mẫu:

| Vai trò | Địa chỉ Email | Mật khẩu | Mô tả quyền hạn |
| :--- | :--- | :--- | :--- |
| Giảng viên | teacher@test.com | 123456 | Tạo và xuất bản đề thi, xem bảng điểm, giải quyết đơn khiếu nại |
| Sinh viên 1 | student@test.com | 123456 | Tra cứu đề thi, làm bài thi, xem lịch sử và gửi khiếu nại |
| Sinh viên 2 | student2@test.com | 123456 | Tài khoản sinh viên thử nghiệm bổ sung |
| Sinh viên 3 | student3@test.com | 123456 | Tài khoản sinh viên thử nghiệm bổ sung |

Mã đề thi mẫu đã tạo sẵn trên hệ thống: IT2026.

## 6. Kiểm thử tự động (Integration Tests)

Chạy toàn bộ bộ test tích hợp bằng lệnh:
```bash
cd backend
go test -v ./...
```
Bộ test sử dụng cơ sở dữ liệu SQLite in-memory độc lập, kiểm tra toàn diện từ khâu xác thực, chu trình đề thi, tính chính xác của thuật toán chấm điểm đến quyền sở hữu lịch sử bài làm và quy trình khiếu nại.
