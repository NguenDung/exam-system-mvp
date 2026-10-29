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

### 5. Củng cố Kiểm thử Tự động (Integration Tests)
- Bổ sung 4 kịch bản kiểm thử tích hợp độc lập trong `backend/main_test.go`:
  - `TestTeacher_ExamCRUD_And_BulkQuestions`: Kiểm tra sửa đề, nhập câu hỏi hàng loạt, sửa câu hỏi và xóa đề thi cascade.
  - `TestTeacher_ScoreUpdate_And_ResetAttempt`: Kiểm tra sửa điểm kèm ghi chú và xóa lượt thi cho phép sinh viên bắt đầu lại.
  - `TestTeacher_StudentManagementCRUD`: Kiểm tra danh sách, tạo mới, sửa thông tin, đặt lại mật khẩu và xóa tài khoản sinh viên.
  - `TestStudent_AvailableExams`: Kiểm tra danh sách bài thi đang mở và trạng thái lượt thi của sinh viên.
- Toàn bộ 17/17 kịch bản kiểm thử đạt kết quả **PASS 100%**.

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
4. **Phân quyền và Chống lỗi IDOR**: Kiểm tra nghiêm ngặt quyền sở hữu đối với từng thao tác sửa/xóa câu hỏi, sửa/xóa đề thi, sửa điểm hoặc xóa lượt thi. Thí sinh chỉ được thao tác trên bài thi của chính mình.
5. **Chống SQL Injection**: 100% câu truy vấn được tham số hóa thông qua GORM ORM kết hợp hàm `parseUintParam` xác thực kiểu dữ liệu đầu vào trên URL parameters.

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

Chạy toàn bộ 17 kịch bản kiểm thử tích hợp tự động bao phủ tất cả các phân hệ:
```bash
cd backend
go test -v ./...
```
Toàn bộ kịch bản kiểm thử chạy trên cơ sở dữ liệu SQLite in-memory giả lập đầy đủ ràng buộc ACID và xử lý kiểu dữ liệu thời gian thực.
