# BÁO CÁO THIẾT KẾ KIẾN TRÚC VÀ BẢO MẬT HỆ THỐNG
## DỰ ÁN: HỆ THỐNG THI TRẮC NGHIỆM TRỰC TUYẾN ĐẠI HỌC (UNIVERSITY EXAMINATION SYSTEM)

---

## 1. TỔNG QUAN KIẾN TRÚC HỆ THỐNG (SYSTEM ARCHITECTURE OVERVIEW)

Hệ thống **Online Examination System** được xây dựng theo mô hình **Client-Server kiến trúc phân tầng (Layered Clean Architecture)**, tách biệt độc lập giữa tầng hiển thị giao diện (Frontend SPA) và tầng xử lý nghiệp vụ/lưu trữ dữ liệu (Backend RESTful API & Relational Database).

```
+-------------------------------------------------------------------------+
|                         CLIENT LAYER (BROWSER)                          |
|   React 19 + Vite + TailwindCSS + React Router v6 + Axios Interceptor   |
|   Phòng thi chuẩn hóa 2 cột: Khối câu hỏi + Bảng điều khiển 1..N       |
+-------------------------------------------------------------------------+
                                    |
                            HTTP/REST (JSON)
                    Authorization: Bearer <JWT Token>
                                    v
+-------------------------------------------------------------------------+
|                         BACKEND LAYER (GOLANG)                          |
|  +-------------------------------------------------------------------+  |
|  | Gin Gonic Engine + CORS Middleware (Cổng 8081)                    |  |
|  +-------------------------------------------------------------------+  |
|  | Auth Middleware: JWT Verification & RBAC Guard (Teacher / Student)|  |
|  +-------------------------------------------------------------------+  |
|  | Handlers:                                                         |  |
|  |   - Auth & User Management                                        |  |
|  |   - Teacher Examination & Question Full CRUD + Bulk Import        |  |
|  |   - Teacher Score Overrides & Attempt Reset (DB Transactions)     |  |
|  |   - Teacher Student Account Management (CRUD)                     |  |
|  |   - Student Available Exams Portal & Real-time Exam Room          |  |
|  |   - History & Academic Appeals Workflow                           |  |
|  +-------------------------------------------------------------------+  |
|  | Security Guards: QuestionSafe DTO, Dual-Timer, FOR UPDATE Lock   |  |
|  +-------------------------------------------------------------------+  |
|  | Data Access: GORM ORM (Transactions, Atomic UPSERT, Prepared Stmt)|  |
|  +-------------------------------------------------------------------+  |
+-------------------------------------------------------------------------+
                                    |
                            Database Driver (TCP)
                                    v
+-------------------------------------------------------------------------+
|                        PERSISTENCE LAYER (RDBMS)                        |
|       PostgreSQL 15 (Docker) / SQLite In-Memory (Test Suite)            |
|       8 Normalized Tables + 2 Unique Composite Indexes + Cascades       |
+-------------------------------------------------------------------------+
```

### 1.1. Công nghệ lựa chọn (Tech Stack) & Lý do kỹ thuật
- **Backend - Golang (Gin Framework & GORM)**: 
  - Khả năng xử lý đa luồng đồng thời (concurrency) vượt trội thông qua cơ chế Goroutine, tiêu tốn rất ít bộ nhớ RAM, đáp ứng hoàn hảo kịch bản hàng trăm hoặc hàng nghìn thí sinh cùng nộp bài hoặc auto-save đồng thời.
  - Gin Framework cung cấp router Radix tree tốc độ cao, khả năng benchmark hàng chục nghìn req/s với độ trễ micro-giây.
- **Database - PostgreSQL 15**: 
  - Hệ quản trị cơ sở dữ liệu quan hệ mạnh mẽ, chuẩn ACID, hỗ trợ tính năng khoá dòng bi quan (`SELECT ... FOR UPDATE`), câu lệnh nguyên tử `ON CONFLICT DO UPDATE` (Atomic UPSERT) và Composite Unique Index, loại bỏ hoàn toàn nguy cơ mất toàn vẹn dữ liệu.
- **Frontend - React 19, Vite & TailwindCSS**: 
  - Vite tối ưu tốc độ đóng gói (Hot Module Replacement cực nhanh).
  - Axios Interceptor xử lý đính kèm Bearer JWT tự động vào toàn bộ các request có bảo mật.
  - Phân vùng trạng thái và điều hướng theo vai trò người dùng (Role-Based Routing).
  - Giao diện phòng thi trực tuyến thiết kế tối giản, loại bỏ hoàn toàn biểu tượng cảm xúc (emojis) màu mè, hướng tới trải nghiệm học thuật chuyên nghiệp của môi trường đại học.

---

## 2. THIẾT KẾ CƠ SỞ DỮ LIỆU & SƠ ĐỒ ERD (DATABASE SCHEMA & ERD)

Cơ sở dữ liệu được chuẩn hóa ở dạng chuẩn 3NF (Third Normal Form) gồm đúng **8 bảng cốt lõi**, phản ánh chính xác nghiệp vụ khảo thí, quản trị học viên, phúc khảo khiếu nại, nhật ký can thiệp điểm số và hệ thống thông báo nội bộ.

### 2.1. Sơ đồ thực thể liên kết (ERD Diagram)

```mermaid
erDiagram
    users ||--o{ exams : "teacher creates"
    users ||--o{ exam_attempts : "student takes"
    users ||--o{ appeals : "student submits"
    users ||--o{ appeals : "teacher responds"
    users ||--o{ audit_logs : "actor performs"
    users ||--o{ notifications : "receives"
    exams ||--o{ questions : "contains (CASCADE)"
    exams ||--o{ exam_attempts : "has (CASCADE)"
    exams ||--o{ appeals : "relates to (CASCADE)"
    exam_attempts ||--o{ attempt_answers : "records (CASCADE)"
    exam_attempts ||--o{ appeals : "appeals for"
    questions ||--o{ attempt_answers : "answered for"

    users {
        bigint id PK
        varchar email UK "Normalized, unique"
        varchar password_hash "Bcrypt hash"
        varchar full_name "User full name"
        varchar role "'teacher' or 'student'"
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at "Soft delete"
    }

    exams {
        bigint id PK
        bigint teacher_id FK "References users.id"
        varchar exam_code UK "Unique code (e.g. IT2026)"
        varchar title "Exam title"
        text description "Detailed description / notes"
        int duration "Duration in minutes"
        timestamp start_time "Exam window start"
        timestamp end_time "Exam window end"
        timestamp published_at "Publication timestamp"
        varchar status "'draft' or 'published'"
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at "Soft delete"
    }

    questions {
        bigint id PK
        bigint exam_id FK "References exams.id"
        int sort_order "Question sequence"
        text prompt "Question text"
        text option_a "Option A"
        text option_b "Option B"
        text option_c "Option C"
        text option_d "Option D"
        varchar correct_option "A, B, C, or D"
        decimal points "Default 1.0"
        timestamp created_at
        timestamp updated_at
    }

    exam_attempts {
        bigint id PK
        bigint exam_id FK "References exams.id"
        bigint user_id FK "References users.id"
        varchar status "'in_progress' or 'completed'"
        decimal score "Rounded to 2 decimals"
        int correct_count "Number of correct answers"
        int wrong_count "Number of wrong/blank answers"
        int violation_count "Count of tab switches"
        text teacher_note "Adjustment notes by teacher"
        timestamp started_at "Attempt start timestamp"
        timestamp last_saved_at "Last auto-save timestamp"
        timestamp submitted_at "Submission timestamp"
    }

    attempt_answers {
        bigint id PK
        bigint attempt_id FK "References exam_attempts.id"
        bigint question_id FK "References questions.id"
        varchar chosen_option "Nullable: 'A'|'B'|'C'|'D'"
        boolean is_correct "True if matches correct_option"
        decimal points_awarded "Awarded points for question"
        timestamp updated_at "Save timestamp"
    }

    appeals {
        bigint id PK
        bigint exam_id FK "References exams.id"
        bigint submission_id FK "References exam_attempts.id"
        bigint student_id FK "References users.id"
        bigint teacher_id FK "References users.id"
        varchar title "Appeal subject"
        text content "Appeal explanation"
        varchar status "'PENDING'|'REPLIED'|'REJECTED'|'RESOLVED'"
        text teacher_response "Response from teacher"
        timestamp responded_at "Response timestamp"
        timestamp created_at
        timestamp updated_at
    }

    audit_logs {
        bigint id PK
        bigint actor_id FK "References users.id"
        varchar action "'UPDATE_SCORE' or 'RESET_ATTEMPT'"
        bigint target_id "Target attempt ID"
        text details "Change details or reason"
        timestamp created_at
    }

    notifications {
        bigint id PK
        bigint user_id FK "References users.id"
        varchar title "Notification title"
        text message "Notification body"
        boolean is_read "Read flag"
        timestamp created_at
    }
```

### 2.2. Phân tích chi tiết các bảng dữ liệu

1. **`users`**:
   - Lưu trữ thông tin tài khoản người dùng với phân quyền (`role`: `teacher` | `student`).
   - Cột `email` có ràng buộc `UNIQUE` và được chuẩn hóa chữ thường trước khi lưu.
   - Mật khẩu lưu dưới dạng Bcrypt hash an toàn.
   - Hỗ trợ Soft-delete qua cột `deleted_at`.

2. **`exams`**:
   - Chứa thông tin đề thi do Giảng viên tạo (`teacher_id`).
   - `exam_code`: Mã kỳ thi gồm chữ hoa/số ngẫu nhiên (hoặc mã định danh như `IT2026`), có ràng buộc `UNIQUE`.
   - `description`: Lưu trữ mô tả đề thi, ghi chú nội dung môn thi.
   - `status`: Quản lý trạng thái đề thi (`draft`: đang soạn thảo; `published`: đã phát hành). Đề thi chỉ được phép xuất bản khi có ít nhất 1 câu hỏi.
   - `start_time` và `end_time`: Xác định khung giờ mở cửa kỳ thi trên hệ thống.
   - `published_at`: Lưu mốc thời gian xuất bản chính thức của đề thi.

3. **`questions`**:
   - Lưu nội dung câu hỏi trắc nghiệm 4 đáp án (A, B, C, D) và đáp án đúng (`correct_option`).
   - Khóa ngoại `exam_id` trỏ đến `exams.id`.
   - Cột `points` (mặc định 1.0) cho phép thiết lập trọng số điểm linh hoạt từng câu.

4. **`exam_attempts`**:
   - Đại diện cho một lượt làm bài của một thí sinh cụ thể đối với một đề thi cụ thể.
   - `started_at`: Thời điểm thí sinh bấm nút bắt đầu làm bài.
   - `status`: Chuyển đổi trạng thái từ `in_progress` sang `completed` khi nộp bài hoặc khi bị hệ thống cưỡng chế auto-submit do hết giờ.
   - `score`, `correct_count`, `wrong_count`: Lưu kết quả chấm thi cuối cùng.
   - `violation_count`: Biến đếm số lần thí sinh rời khỏi màn hình làm bài (chuyển tab hoặc thu nhỏ trình duyệt). Khi `violation_count >= 3`, hệ thống tự động khóa và thu bài thi.
   - `teacher_note`: Lưu trữ ghi chú của giảng viên khi điều chỉnh điểm thi hoặc can thiệp bài làm.

5. **`attempt_answers`**:
   - Lưu trữ chi tiết lựa chọn của thí sinh cho từng câu hỏi trong bài thi.
   - `chosen_option`: Lựa chọn của thí sinh (`A`, `B`, `C`, `D` hoặc `NULL` nếu thí sinh chưa chọn).
   - `is_correct` và `points_awarded`: Được tính toán và cập nhật khi bài thi được nộp và chấm điểm.

6. **`appeals`**:
   - Lưu trữ đơn khiếu nại/phúc khảo điểm thi của thí sinh gửi lên giảng viên phụ trách môn thi.
   - Liên kết 4 khóa ngoại: `exam_id`, `submission_id`, `student_id`, `teacher_id`.
   - `status`: Quản lý tiến trình xử lý đơn (`PENDING`: Chờ xử lý, `REPLIED`: Đã phản hồi, `REJECTED`: Đã từ chối, `RESOLVED`: Đã xử lý).

7. **`audit_logs`**:
   - Bảng lưu vết minh bạch các hành vi can thiệp nhạy cảm của Giảng viên.
   - `actor_id`: Khóa ngoại trỏ về `users.id` (Giảng viên thực hiện hành vi).
   - `action`: Mã định danh hành vi (`UPDATE_SCORE` hoặc `RESET_ATTEMPT`).
   - `target_id`: ID của lượt thi (`exam_attempts.id`) bị can thiệp.
   - `details`: Nội dung chi tiết giải trình, ví dụ: "Cập nhật điểm từ 8.00 thành 9.00: Đính chính điểm câu 4".
   - `created_at`: Mốc thời gian ghi nhận hành vi.

8. **`notifications`**:
   - Lưu trữ các thông báo nội bộ thời gian thực gửi tới Giảng viên và Sinh viên.
   - `user_id`: Khóa ngoại trỏ về `users.id` (Người nhận thông báo).
   - `title`: Tiêu đề tóm tắt sự kiện nghiệp vụ (ví dụ: "Có khiếu nại mới từ sinh viên", "Điểm bài thi đã được điều chỉnh", "Khiếu nại đã được giải quyết").
   - `message`: Nội dung thông báo chi tiết kèm ngữ cảnh cụ thể (tên bài thi, điểm số cập nhật, phản hồi của giảng viên).
   - `is_read`: Cờ đánh dấu trạng thái đã đọc (`false`: chưa đọc, `true`: đã đọc).
   - `created_at`: Mốc thời gian phát hành thông báo.

---

### 2.3. Phân tích 2 Unique Composite Index đặc tả

Cơ sở dữ liệu được trang bị 2 chỉ mục tổng hợp duy nhất (Unique Composite Indexes):

#### Index 1: `idx_exam_attempts_exam_user`
```sql
CREATE UNIQUE INDEX IF NOT EXISTS idx_exam_attempts_exam_user
ON exam_attempts (exam_id, user_id);
```
- **Mục đích nghiệp vụ**: Ràng buộc nguyên tắc tối cao của thi trắc nghiệm: **Mỗi học sinh chỉ được phép có duy nhất 1 lượt làm bài (`exam_attempt`) cho một đề thi cụ thể**. Ngăn chặn hoàn toàn việc thí sinh gian lận tạo nhiều bài thi để làm lại từ đầu.
- **Cơ chế Reset Attempt**: Khi thí sinh gặp sự cố mạng hoặc lỗi kỹ thuật, giảng viên sử dụng chức năng "Cho thi lại" (`DELETE /api/teacher/attempts/:id`), lượt làm bài cũ bị xóa vĩnh viễn trong Transaction, qua đó giải phóng Unique Index này và cho phép sinh viên bắt đầu lại một lượt thi mới hợp lệ.

#### Index 2: `idx_attempt_answers_attempt_question`
```sql
CREATE UNIQUE INDEX IF NOT EXISTS idx_attempt_answers_attempt_question
ON attempt_answers (attempt_id, question_id);
```
- **Mục đích nghiệp vụ & Kỹ thuật**:
  - Đảm bảo trong một bài thi (`attempt_id`), mỗi câu hỏi (`question_id`) chỉ tồn tại duy nhất một bản ghi câu trả lời.
  - Cho phép thực hiện câu lệnh **Atomic UPSERT** khi auto-save:
    ```sql
    INSERT INTO attempt_answers (attempt_id, question_id, chosen_option, is_correct, points_awarded, updated_at)
    VALUES (?, ?, ?, false, 0, ?)
    ON CONFLICT (attempt_id, question_id)
    DO UPDATE SET chosen_option = EXCLUDED.chosen_option, updated_at = EXCLUDED.updated_at;
    ```
  - Khi thí sinh thay đổi câu trả lời nhiều lần, cơ sở dữ liệu sẽ tự động thực hiện `UPDATE` thay vì `INSERT`, loại bỏ hiện tượng trùng lặp bản ghi mà không cần phải thực hiện 2 câu lệnh `SELECT` và `UPDATE` riêng biệt (giảm tải 50% số lượng query I/O).

---

## 3. SƠ ĐỒ LUỒNG NGHIỆP VỤ (BUSINESS PROCESS FLOWS)

### 3.1. Luồng A: Vòng đời làm bài thi trắc nghiệm (Exam Lifecycle Sequence)
Quy trình từ lúc Thí sinh xem danh sách bài thi đang mở, Bắt đầu thi, Tương tác với Bảng điều khiển câu hỏi (Question Palette), Auto-save liên tục, Nộp bài và Xem kết quả.

```mermaid
sequenceDiagram
    autonumber
    actor Student as Thí sinh (Browser)
    participant FE as Frontend React (Vite)
    participant Auth as Auth Middleware (JWT)
    participant Handler as Student Handler (Go)
    participant DB as PostgreSQL Database

    Note over Student, DB: GIAI ĐOẠN 1: TRUY CẬP TRANG CHỦ & TRA CỨU
    Student->>FE: Truy cập Trang chủ Sinh viên
    FE->>Auth: GET /api/student/exams/available
    Auth->>Handler: Validate JWT Token (Role: student)
    Handler->>DB: Query các kỳ thi status = 'published'
    DB-->>Handler: Trả về danh sách kỳ thi kèm trạng thái attempt_status
    Handler-->>FE: HTTP 200 (exams: [...])
    FE-->>Student: Hiển thị Thẻ bài thi (Vào thi ngay / Tiếp tục / Xem kết quả)

    Note over Student, DB: GIAI ĐOẠN 2: BẮT ĐẦU LÀM BÀI HOẶC RESUME
    Student->>FE: Bấm "Vào thi ngay"
    FE->>Auth: POST /api/student/exams/:id/start
    Auth->>Handler: Forward request
    Handler->>DB: Transaction: Kiểm tra exam_attempts hiện có
    alt Chưa có bài thi
        Handler->>DB: INSERT INTO exam_attempts (status='in_progress', started_at=NOW())
    else Đã có bài thi dở dang (Resume)
        Handler->>DB: SELECT attempt_answers đã lưu
    end
    Handler->>Handler: Tính toán Effective Deadline (2 đồng hồ)
    Handler->>Handler: Chuyển đổi questions -> QuestionSafe DTO (Ẩn correct_option)
    Handler-->>FE: HTTP 200 (attempt, questions: [QuestionSafe], saved_answers, remaining_seconds)
    FE-->>Student: Hiển thị Giao diện Phòng thi 2 cột & Question Palette

    Note over Student, DB: GIAI ĐOẠN 3: TƯƠNG TÁC LÀM BÀI & AUTO-SAVE
    Student->>FE: Chọn đáp án Câu i & Bấm Đặt cờ xem lại
    FE->>FE: Cập nhật Question Palette (Đổi sang Xanh/Vàng tức thì)
    FE->>Auth: POST /api/student/attempts/:id/save (answers: [...])
    Auth->>Handler: Forward request
    Handler->>DB: Transaction (FOR UPDATE): Kiểm tra deadline + 5s
    alt Còn trong thời hạn làm bài
        Handler->>DB: Atomic UPSERT vào attempt_answers
        Handler-->>FE: HTTP 200 (saved: true)
        FE-->>Student: Hiển thị nhãn "Đã lưu tự động"
    else Quá hạn nộp bài (> deadline + 5s)
        Handler->>DB: Cưỡng chế chấm điểm & đổi status='completed'
        Handler-->>FE: HTTP 409 Conflict (auto_submitted: true)
        FE->>FE: Chuyển hướng sang trang Kết quả thi
    end

    Note over Student, DB: GIAI ĐOẠN 4: NỘP BÀI THI & CHẤM ĐIỂM
    Student->>FE: Bấm Nộp bài thi
    FE->>Auth: POST /api/student/attempts/:id/submit
    Auth->>Handler: Forward request
    Handler->>DB: Transaction (FOR UPDATE): Kiểm tra status=='in_progress'
    Handler->>DB: Đối soát attempt_answers với correct_option của đề thi
    Handler->>DB: UPDATE exam_attempts (score, correct_count, wrong_count, status='completed')
    Handler-->>FE: HTTP 200 (attempt)
    FE-->>Student: Hiển thị Bảng điểm, Số câu đúng/sai & Cho phép phúc khảo
```

---

### 3.2. Luồng B: Cơ chế 2 đồng hồ kiểm soát thời gian (Exam Window vs Attempt Duration)

Hệ thống triển khai cơ chế **Server-Authoritative Dual-Clock** để đảm bảo tính công bằng tuyệt đối:

```mermaid
flowchart TD
    Start([Thí sinh bắt đầu thi lúc: started_at]) --> CalcDeadlines[Tính toán 2 mốc thời gian độc lập]
    CalcDeadlines --> D1["Mốc 1: started_at + exam.duration<br/>(Thời lượng bài thi)"]
    CalcDeadlines --> D2["Mốc 2: exam.end_time<br/>(Thời điểm đóng đề thi)"]
    
    D1 & D2 --> Compare{Mốc 1 < Mốc 2?}
    Compare -- "ĐÚNG: Vào thi sớm/đủ giờ" --> Effective1["Hạn chót = started_at + duration<br/>(Làm đủ 100% thời lượng bài thi)"]
    Compare -- "SAI: Vào thi muộn" --> Effective2["Hạn chót = exam.end_time<br/>(Bị ép nộp cùng giờ đóng đề)"]
    
    Effective1 & Effective2 --> SyncClient["remaining_seconds = Effective Deadline - NOW()<br/>Đồng bộ xuống Client"]
```

---

## 4. ĐIỂM NHẤN KỸ THUẬT & CÁC BẢN VÁ BẢO MẬT (SECURITY & TECHNICAL HIGHLIGHTS)

### 4.1. Chống lộ đáp án trắc nghiệm (Zero-Leak Data Transfer Object - DTO)
- **Nguy cơ tiềm ẩn**: Nếu serialize trực tiếp GORM Model `Question` trả về cho học sinh, trường `correct_option` sẽ nằm trong payload JSON. Thí sinh chỉ cần bấm F12 mở tab Network là có thể xem được toàn bộ đáp án đề thi trước khi làm bài.
- **Giải pháp kỹ thuật**:
  - Xây dựng DTO struct `QuestionSafe` loại bỏ hoàn toàn trường `CorrectOption`:
    ```go
    type QuestionSafe struct {
        ID        uint    `json:"id"`
        ExamID    uint    `json:"exam_id"`
        SortOrder int     `json:"sort_order"`
        Prompt    string  `json:"prompt"`
        OptionA   string  `json:"option_a"`
        OptionB   string  `json:"option_b"`
        OptionC   string  `json:"option_c"`
        OptionD   string  `json:"option_d"`
        Points    float64 `json:"points"`
    }
    ```
  - Tại API `StudentStartExam`, 100% câu hỏi được chuyển đổi qua hàm `models.ToQuestionSafe(q)` trước khi gửi về client.
  - **Bảo mật kép tại trang Xem kết quả (`StudentGetResult`)**: Kể cả sau khi nộp bài, nếu đề thi vẫn đang mở (`now.Before(exam.EndTime)`), trường `correct_option` của toàn bộ các câu hỏi vẫn bị ẩn (`nil`). Chỉ khi toàn bộ khung giờ thi kết thúc (`now.After(exam.EndTime)`), thí sinh mới được phép xem đáp án đúng.

---

### 4.2. Chống gian lận thời gian (Server-Authoritative Clock & Grace Period)
- **Nguy cơ tiềm ẩn**: Nếu tin cậy đồng hồ phía Client (JavaScript `Date.now()`), thí sinh có thể chỉnh lùi giờ hệ điều hành trên máy tính để kéo dài thời gian làm bài vô tận.
- **Giải pháp kỹ thuật**:
  - **Server-Authoritative**: Mọi phép tính toán thời gian bắt đầu, thời hạn kết thúc và thời gian còn lại đều lấy mốc từ `time.Now()` trên máy chủ Backend.
  - **Vùng đệm mạng (Network Grace Period)**: Khi gửi request lưu đáp án (`/save`) hoặc nộp bài (`/submit`), hệ thống cho phép vùng đệm trễ mạng 5 giây (`const saveGracePeriod = 5 * time.Second`):
    - Nếu request gửi lên trong khoảng `deadline < now <= deadline + 5s`: Hệ thống vẫn chấp nhận lưu đáp án cuối cùng để bù trừ độ trễ mạng (latency/jitter) của thí sinh.
    - Nếu `now > deadline + 5s`: Hệ thống từ chối lưu đáp án mới, tự động kích hoạt hàm chấm bài dựa trên các đáp án đã lưu hợp lệ trước đó và trả về mã lỗi `409 Conflict`.
  - **Cơ chế quét bài thi hết hạn bị động (Lazy Expiration Grading)**: Khi giáo viên mở bảng điểm hoặc khi học sinh tra cứu đề thi, hàm `gradeExpiredAttempts` tự động chấm điểm toàn bộ các bài thi quá hạn chưa nộp.

---

### 4.3. Chống Race Condition & Double Submit (Pessimistic Locking `FOR UPDATE`)
- **Nguy cơ tiềm ẩn**: Thí sinh mở nhiều tab trình duyệt gửi song song request `/submit` hoặc `/save` gây xung đột dữ liệu.
- **Giải pháp kỹ thuật**:
  - Bọc toàn bộ quy trình chấm thi và lưu bài trong một **Database Transaction** nguyên tử (`config.DB.Transaction`).
  - Khóa dòng bi quan bằng GORM clause:
    ```go
    err := config.DB.Transaction(func(tx *gorm.DB) error {
        var attempt models.ExamAttempt
        if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
            Where("id = ? AND user_id = ?", attemptID, userID).
            First(&attempt).Error; err != nil {
            return err
        }
        if attempt.Status != "in_progress" {
            return errAlreadySubmitted
        }
        ...
    ```
  - Khi một request đang giữ khóa `FOR UPDATE`, tất cả các request song song khác gửi đến cùng một `attempt_id` bắt buộc phải xếp hàng chờ. Request đầu tiên hoàn thành sẽ đổi trạng thái `status = 'completed'`, các request sau sẽ bị từ chối ngay lập tức.

---

### 4.4. Toàn vẹn Dữ liệu Xóa tầng (Cascading Deletions in DB Transaction)
- Khi Giảng viên xóa một đề thi (`DELETE /api/teacher/exams/:id`):
  Hệ thống thực thi trong một Database Transaction duy nhất để dọn sạch toàn bộ các thực thể phụ thuộc:
  1. `appeals` liên quan đến đề thi.
  2. `attempt_answers` thuộc tất cả các lượt thi của đề thi.
  3. `exam_attempts` của đề thi.
  4. `questions` của đề thi.
  5. Xóa chính bản ghi `exams`.
- Tương tự, khi xóa tài khoản sinh viên (`DELETE /api/teacher/students/:id`), toàn bộ khiếu nại, câu trả lời và lượt thi của sinh viên đó được xóa tuần tự trong transaction trước khi xóa bản ghi tài khoản.

---

### 4.5. Cơ chế Nhập nhanh Câu hỏi Hàng loạt (Bulk Question Parser)
- Hỗ trợ Giảng viên tạo cùng lúc hàng chục câu hỏi thông qua văn bản phi cấu trúc:
  ```
  Câu: Ngôn ngữ nào chạy trực tiếp trên máy ảo JVM?
  A. Python
  B. Java
  C. C++
  D. Rust
  Đáp án: B
  Điểm: 1.5
  ---
  ```
- Parser phía Frontend tự động chuẩn hóa văn bản, nhận diện Regex cho các lựa chọn A, B, C, D, đáp án đúng và trọng số điểm; hoặc trực tiếp chấp nhận mảng đối tượng JSON.
- Backend tiếp nhận mảng câu hỏi tại endpoint `POST /api/teacher/exams/:id/questions/bulk`, thực hiện xác thực từng câu hỏi và thêm hàng loạt (Bulk Insert) trong một transaction duy nhất.

---

### 4.6. Thiết kế Phòng thi Trực tuyến Chuẩn mực (Question Palette UX)
- Thay vì dàn trải câu hỏi trên một trang dọc đơn điệu dễ gây rối mắt, giao diện phòng thi được chia thành 2 cột:
  - **Cột trái**: Trọng tâm hiển thị câu hỏi hiện tại, các nút chọn đáp án lớn dễ thao tác, nút "Đặt cờ xem lại" giúp thí sinh đánh dấu các câu cần suy nghĩ thêm, và cặp nút "Câu trước" / "Câu tiếp theo".
  - **Cột phải (Bảng điều khiển câu hỏi)**: Lưới ô vuông 1..N được gán màu theo thời gian thực:
    - **Xanh lá**: Đã chọn đáp án và đã đồng bộ lưu lên máy chủ.
    - **Vàng**: Đang được đặt cờ xem lại.
    - **Xám**: Chưa làm.
    - **Viền sáng**: Vị trí câu hỏi đang xem.
  - Bấm vào bất kỳ ô số nào để nhảy trực tiếp tới câu hỏi tương ứng, giúp thí sinh kiểm soát tiến độ làm bài thi một cách chuyên nghiệp.

---

### 4.7. Động cơ Chống gian lận Phòng thi (Anti-Cheat Engine)
1. **Xáo trộn Câu hỏi Tất định (Deterministic Question Shuffling)**:
   - Thay vì trả về câu hỏi theo thứ tự cố định (dễ dẫn đến việc thí sinh nhìn bài nhau), hệ thống tự động hoán vị mảng câu hỏi dựa trên seed ngẫu nhiên là ID của lượt thi:
     ```go
     rnd := rand.New(rand.NewSource(int64(attempt.ID)))
     rnd.Shuffle(len(questions), func(i, j int) {
         questions[i], questions[j] = questions[j], questions[i]
     })
     ```
   - **Tính tất định (Determinism)**: Mỗi sinh viên có một `attempt.ID` riêng biệt nên thứ tự câu hỏi hoàn toàn khác nhau giữa các thí sinh. Tuy nhiên, cùng một thí sinh khi F5 tải lại trang hoặc bị ngắt kết nối và quay lại (Resume Exam), seed `attempt.ID` không đổi nên thứ tự câu hỏi và số thứ tự 1..N giữ nguyên 100%, không gây xáo trộn tâm lý làm bài.
2. **Giám sát Chuyển Tab & Tự động Đình chỉ thi (Tab-Switch Detection)**:
   - Trình duyệt phía Client lắng nghe đồng thời sự kiện `document.visibilitychange` (khi `document.hidden === true`) và `window.onblur`.
   - Áp dụng cơ chế Throttling 2 giây phía Frontend nhằm khử trùng lặp (Deduplication) khi trình duyệt kích hoạt đồng thời cả 2 sự kiện trên một lần chuyển tab.
   - Gọi API `POST /api/student/attempts/:id/violation` có khóa bi quan `FOR UPDATE` để tăng trường `violation_count` trong cơ sở dữ liệu.
   - **Xử lý vi phạm theo cấp độ**:
     - Vi phạm lần 1 và 2: Hiển thị Modal cảnh báo nghiêm khắc trên màn hình và cập nhật huy hiệu cảnh báo đỏ `Vi phạm: X/3` trên thanh điều hướng.
     - Vi phạm lần 3 (Đạt ngưỡng): Hệ thống tự động thu bài, tính điểm tự động bằng hàm `gradeAttemptInTx`, đổi trạng thái sang `completed` và chuyển hướng sinh viên về trang kết quả kèm thông báo đình chỉ thi do vi phạm quy chế.

---

### 4.8. Giới hạn Tần suất Truy cập Trượt (In-Memory Sliding Window Rate Limiting)
- Triển khai middleware `IPRateLimit(limit int, window time.Duration)` tại tầng Router cho các route nhạy cảm có nguy cơ bị brute-force hoặc DoS cao:
  - `POST /api/auth/login`: Chống vét cạn mật khẩu người dùng.
  - `GET /api/student/exams/lookup`: Chống rà quét dò tìm mã đề thi (`exam_code`).
- Cấu trúc dữ liệu sliding window được bảo vệ bằng `sync.Mutex` an toàn luồng, lưu vết lịch sử timestamp các request trong khung thời gian 1 phút với ngưỡng tối đa 15 request/phút/IP.
- Tích hợp goroutine định kỳ dọn dẹp các IP không hoạt động để giải phóng bộ nhớ RAM.
- **Tính năng Bypass trong Kiểm thử**: Middleware tự động kiểm tra `gin.Mode() == gin.TestMode` hoặc biến môi trường `RATE_LIMIT_DISABLED=true` để bỏ qua giới hạn khi chạy bộ test tự động, bảo đảm không làm gián đoạn CI/CD pipeline.

---

### 4.9. Hệ thống Nhật ký Can thiệp Minh bạch (Audit Logging System)
- Khi Giảng viên thực hiện các thao tác nhạy cảm tác động trực tiếp đến kết quả thi của sinh viên:
  1. Hiệu chỉnh điểm thi (`PATCH /api/teacher/attempts/:id/score` -> Action: `UPDATE_SCORE`).
  2. Cho phép thí sinh thi lại / Xóa lượt thi (`DELETE /api/teacher/attempts/:id` -> Action: `RESET_ATTEMPT`).
- Hệ thống tự động khởi tạo bản ghi trong bảng `audit_logs` ngay trong Database Transaction:
  - `actor_id`: ID của Giảng viên thực hiện.
  - `action`: Mã hành động chuẩn hóa.
  - `target_id`: ID của lượt thi (`attempt.ID`).
  - `details`: Lưu vết điểm cũ, điểm mới và lý do giải trình do giảng viên cung cấp.
- Cung cấp API `GET /api/teacher/audit-logs` và Modal "Nhật ký can thiệp" trực quan, giúp hội đồng khảo thí tra cứu lịch sử minh bạch bất cứ lúc nào.

---

### 4.10. Xuất Bảng điểm Chuẩn Excel (Grade CSV Export with UTF-8 BOM)
- Cung cấp endpoint `GET /api/teacher/exams/:id/results/export` phục vụ công tác lưu trữ và báo cáo điểm thi cấp khoa/trường.
- **Xử lý Mã hóa Ký tự Tiếng Việt**: Để tương thích hoàn toàn với Microsoft Excel trên hệ điều hành Windows mà không xảy ra hiện tượng vỡ font (Mojibake), phản hồi CSV được gắn tiền tố Byte Order Mark (UTF-8 BOM: `\xef\xbb\xbf`).
- Định dạng chuẩn 6 cột thông tin:
  ```csv
  Họ tên,Email,Điểm số,Trạng thái,Số lần vi phạm,Thời gian nộp
  ```
- Nút "Xuất bảng điểm (CSV)" được tích hợp trên giao diện Giảng viên, tải xuống tệp tin qua Blob URL tức thì.

---

### 4.11. Hệ thống Thông báo Nội bộ (In-App Notification Engine) & Khử bỏ Window Alert
- **Tự động kích hoạt theo sự kiện nghiệp vụ (Event-Driven Notifications)**: Hệ thống tự động phát hành thông báo đến đúng người nhận tại các mốc thời gian quan trọng:
  1. Khi sinh viên nộp đơn khiếu nại mới -> Giảng viên sở hữu đề thi nhận thông báo: `"Khiếu nại mới từ sinh viên: [Họ tên]"` kèm tiêu đề và mã đề thi.
  2. Khi giảng viên phản hồi, chấp thuận hoặc từ chối khiếu nại -> Sinh viên nộp đơn nhận thông báo cập nhật trạng thái đơn phúc khảo.
  3. Khi giảng viên can thiệp hiệu chỉnh điểm thi hoặc cấp quyền thi lại (Reset Attempt) -> Sinh viên nhận thông báo minh bạch về số điểm mới hoặc trạng thái mở lại đề.
- **Thiết kế An toàn Luồng (Non-blocking / Fail-safe Architecture)**:
  - Hàm helper `CreateNotification(userID uint, title, message string)` được bao bọc khối phục hồi lỗi (`defer func() { if r := recover(); r != nil { ... } }`).
  - Lỗi phát sinh trong quá trình ghi nhận thông báo được ghi log cục bộ mà **tuyệt đối không bao giờ làm fail transaction nghiệp vụ chính** (gửi khiếu nại, sửa điểm, reset bài).
- **Khử bỏ hoàn toàn `window.alert()` trên Giao diện**:
  - Loại bỏ triệt để 100% các hộp thoại `window.alert()` nguyên thủy, thô sơ và gây gián đoạn luồng người dùng.
  - Thay thế bằng hệ thống thông báo chuẩn mực gồm: Banners cảnh báo tích hợp (Inline Alert Banners), Modal xác nhận hành vi nguy hiểm và Toast Notifications tự biến mất sau 4 giây.
  - Dropdown Thông báo trên thanh điều hướng (Navbar) cho cả Giảng viên và Sinh viên:
    - Huy hiệu (Badge) đếm số lượng thông báo chưa đọc.
    - Cơ chế polling tự động ngầm mỗi 30 giây để cập nhật thông báo mới.
    - Đánh dấu đã đọc đơn lẻ khi click vào từng thông báo (`PATCH /api/notifications/:id/read`).
    - Nút tác vụ "Đánh dấu đã đọc tất cả" (`PATCH /api/notifications/read-all`).

---

### 4.12. Gia cố Bảo mật Tiêu đề HTTP (Security Headers) & Giới hạn Body Size 2MB
- **Middleware `SecurityMiddleware`**: Được đăng ký toàn cục tại tầng đầu tiên của Gin Engine, thực thi trước toàn bộ các route và controller:
  - **`X-Content-Type-Options: nosniff`**: Vô hiệu hóa tính năng MIME-sniffing của trình duyệt, ngăn chặn kẻ tấn công ngụy trang file thực thi độc hại dưới dạng file tĩnh hoặc JSON.
  - **`X-Frame-Options: DENY`**: Chống tấn công Clickjacking, nghiêm cấm hoàn toàn việc nhúng hệ thống thi vào bất kỳ `<iframe>`, `<frame>` hay `<object>` nào.
  - **`Referrer-Policy: strict-origin-when-cross-origin`**: Bảo vệ thông tin định danh và token nhạy cảm trong URL, chỉ gửi origin khi chuyển hướng cross-origin qua HTTPS.
- **Giới hạn kích thước Body 2MB (DoS & Resource Exhaustion Defense)**:
  - Sử dụng `http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)` với hạn mức tối đa 2MB (2 * 1024 * 1024 bytes).
  - Ngay lập tức từ chối các request có payload khổng lồ cố tình gây tràn RAM hoặc làm cạn kiệt tài nguyên máy chủ.

---

### 4.13. Cơ chế Chống xung đột Khóa Cơ sở dữ liệu SQLite (WAL Mode & Busy Timeout)
- **Vấn đề kỹ thuật**: Trong môi trường kiểm thử tích hợp tự động (`go test -v ./...`) và môi trường chạy cục bộ sử dụng SQLite, việc nhiều Goroutine gửi các request ghi song song (`INSERT`, `UPDATE`, `DELETE`) thường xuyên kích hoạt lỗi `database is locked (5)` do cơ chế khóa mặc định đơn luồng của SQLite.
- **Giải pháp xử lý triệt để**:
  - **Kích hoạt chế độ WAL (Write-Ahead Logging)**:
    ```sql
    PRAGMA journal_mode = WAL;
    ```
    Cho phép đọc và ghi diễn ra đồng thời (Concurrent Readers & Single Writer), các truy vấn đọc không bao giờ phong tỏa giao dịch ghi và ngược lại.
  - **Thiết lập Busy Timeout 5000ms**:
    ```sql
    PRAGMA busy_timeout = 5000;
    ```
    Khi gặp lock, driver SQLite sẽ chủ động chờ và thử lại tối đa 5 giây thay vì trả về lỗi ngay lập tức.
  - Hàm `config.EnableSQLitePragmas(db)` được triệu gọi đồng nhất ở cả máy chủ chính lẫn hàm `setupTestServer` trong bộ test, giúp 100% kịch bản kiểm thử chạy mượt mà, loại bỏ hoàn toàn lỗi khóa cơ sở dữ liệu.

---

## 5. DANH MỤC API ENDPOINTS TOÀN DIỆN (API ENDPOINTS SPECIFICATION)

Tất cả các API yêu cầu xác thực đều phải đính kèm Header:
`Authorization: Bearer <jwt_token>`

### 5.1. Authentication APIs (Public)
| Method | Endpoint | Yêu cầu Body | Mã phản hồi | Mô tả |
| :--- | :--- | :--- | :--- | :--- |
| `POST` | `/api/auth/register` | `{full_name, email, password}` | `201 Created` / `400` / `409` | Đăng ký tài khoản sinh viên mới |
| `POST` | `/api/auth/login` | `{email, password}` | `200 OK` / `400` / `401` / `429` | Đăng nhập hệ thống, trả về Token & User Info (Rate limit: 15 req/phút/IP) |

### 5.2. Teacher APIs (Yêu cầu Role = `teacher`)
| Method | Endpoint | Yêu cầu Body / Query Param | Mã phản hồi | Mô tả |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/api/teacher/exams` | Không | `200 OK` | Lấy danh sách đề thi do giảng viên tạo |
| `POST` | `/api/teacher/exams` | `{title, description, duration, start_time, end_time}` | `201 Created` / `400` | Tạo đề thi mới (trạng thái `draft`, sinh `exam_code`) |
| `GET` | `/api/teacher/exams/:id` | Không | `200 OK` / `404` | Xem chi tiết đề thi kèm danh sách câu hỏi |
| `PUT` | `/api/teacher/exams/:id` | `{title, description, duration, start_time, end_time, is_published}` | `200 OK` / `400` / `404` | **Chỉnh sửa thông tin đề thi & đóng/mở đề** |
| `DELETE`| `/api/teacher/exams/:id` | Không | `200 OK` / `404` | **Xóa đề thi và dọn sạch dữ liệu liên quan (Cascade)** |
| `POST` | `/api/teacher/exams/:id/questions` | `{questions: [{prompt, option_a, ...}]}` | `201 Created` / `400` | Thêm danh sách câu hỏi vào đề thi (`draft`) |
| `POST` | `/api/teacher/exams/:id/questions/bulk` | `{questions: [{prompt, option_a, ...}]}` | `201 Created` / `400` | **Nhập hàng loạt câu hỏi từ file/text** |
| `PUT` | `/api/teacher/questions/:id` | `{prompt, option_a, option_b, option_c, option_d, correct_option, points}` | `200 OK` / `400` / `404` | **Sửa câu hỏi và đáp án đúng (khi đề draft)** |
| `DELETE`| `/api/teacher/questions/:question_id` | Không | `200 OK` / `400` / `403` / `404` | Xóa câu hỏi khỏi đề thi (`draft`) |
| `PATCH`| `/api/teacher/exams/:id/publish` | Không | `200 OK` / `400` / `404` | Xuất bản đề thi (yêu cầu tối thiểu 1 câu hỏi) |
| `GET` | `/api/teacher/exams/:id/results` | Không | `200 OK` / `404` | Thống kê kết quả thi: Danh sách điểm, ĐTB, Min, Max |
| `GET` | `/api/teacher/exams/:id/results/export` | Không | `200 OK` / `404` | **Xuất bảng điểm CSV có UTF-8 BOM chuẩn Excel** |
| `PATCH`| `/api/teacher/attempts/:id/score` | `{score, teacher_note}` | `200 OK` / `400` / `404` | **Hiệu chỉnh trực tiếp điểm thi của sinh viên kèm ghi chú & ghi Audit Log** |
| `DELETE`| `/api/teacher/attempts/:id` | Không | `200 OK` / `404` | **Xóa lượt thi (Cho phép sinh viên thi lại từ đầu & ghi Audit Log)** |
| `GET` | `/api/teacher/audit-logs` | Không | `200 OK` | **Lấy danh sách nhật ký can thiệp điểm số và lượt thi** |
| `GET` | `/api/teacher/students` | Query: `?q=search_term` | `200 OK` | **Danh sách sinh viên, số bài thi đã tham gia** |
| `POST` | `/api/teacher/students` | `{full_name, email, password}` | `201 Created` / `400` / `409` | **Tạo tài khoản sinh viên mới** |
| `PUT` | `/api/teacher/students/:id` | `{full_name, email, password?}` | `200 OK` / `400` / `404` / `409` | **Cập nhật họ tên, email hoặc đổi mật khẩu sinh viên** |
| `DELETE`| `/api/teacher/students/:id` | Không | `200 OK` / `404` | **Xóa tài khoản sinh viên kèm dữ liệu làm bài** |
| `GET` | `/api/teacher/exams/:id/appeals` | Query: `page`, `page_size`, `status`, `q` | `200 OK` / `404` | Lấy danh sách đơn khiếu nại của đề thi |
| `GET` | `/api/teacher/appeals/:appealId` | Không | `200 OK` / `404` | Xem chi tiết đơn khiếu nại và bài làm của thí sinh |
| `PATCH`| `/api/teacher/appeals/:appealId/reply` | `{response}` | `200 OK` / `400` / `404` / `409` | Phản hồi ý kiến cho khiếu nại |
| `PATCH`| `/api/teacher/appeals/:appealId/reject` | `{response}` (tùy chọn) | `200 OK` / `404` / `409` | Từ chối đơn khiếu nại |
| `PATCH`| `/api/teacher/appeals/:appealId/resolve` | `{response}` (tùy chọn) | `200 OK` / `404` / `409` | Chấp thuận / giải quyết xong đơn khiếu nại |

### 5.3. Student APIs (Yêu cầu Role = `student`)
| Method | Endpoint | Yêu cầu Body / Query Param | Mã phản hồi | Mô tả |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/api/student/exams/available` | Không | `200 OK` | **Danh sách các bài thi đang mở và trạng thái tương ứng** |
| `GET` | `/api/student/exams/lookup` | Query: `?code=XYZ` | `200 OK` / `400` / `404` / `429` | Tra cứu đề thi đã xuất bản theo mã kỳ thi (Rate limit: 15 req/phút/IP) |
| `GET` | `/api/student/attempts` | Không | `200 OK` | Xem lịch sử các lượt thi gần đây của thí sinh |
| `GET` | `/api/student/history` | Query: `page`, `page_size`, `q`, `appeal_status` | `200 OK` | Lấy lịch sử các bài thi đã hoàn thành (phân trang) |
| `GET` | `/api/student/history/:submissionId` | Không | `200 OK` / `404` | Xem chi tiết bài làm đã hoàn thành và danh sách khiếu nại |
| `POST` | `/api/student/appeals` | `{submissionId, title, content}` | `201 Created` / `400` / `404` / `409` | Gửi đơn khiếu nại điểm thi |
| `POST` | `/api/student/exams/:id/start` | `{confirmed: true}` | `200 OK` / `400` / `403` / `404` | Bắt đầu hoặc Resume bài thi. Trả về `QuestionSafe` (Đã xáo trộn câu hỏi tất định) |
| `POST` | `/api/student/attempts/:id/violation` | `{reason}` (tùy chọn) | `200 OK` / `400` / `404` | **Ghi nhận vi phạm chuyển tab; tự động nộp bài khi đạt 3 lần** |
| `POST` | `/api/student/attempts/:id/save`| `{answers: [{question_id, chosen_option}]}` | `200 OK` / `400` / `409` | Auto-save câu trả lời (Atomic UPSERT) |
| `POST` | `/api/student/attempts/:id/submit`| `{answers: [...]}` (tùy chọn) | `200 OK` / `400` / `404` | Nộp bài thi và chấm điểm tự động trong DB Transaction |
| `GET` | `/api/student/attempts/:id/result`| Không | `200 OK` / `400` / `404` | Xem điểm số và đối soát bài làm sau khi nộp |

### 5.4. Notification APIs (Yêu cầu Xác thực JWT)
| Method | Endpoint | Yêu cầu Body / Query Param | Mã phản hồi | Mô tả |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/api/notifications` | Không | `200 OK` | **Lấy danh sách 20 thông báo mới nhất kèm số lượng thông báo chưa đọc (`unread_count`)** |
| `PATCH`| `/api/notifications/:id/read` | Không | `200 OK` / `404` | **Đánh dấu một thông báo cụ thể là đã đọc** |
| `PATCH`| `/api/notifications/read-all` | Không | `200 OK` | **Đánh dấu tất cả thông báo của người dùng là đã đọc** |

### 5.5. System Health API (Public)
| Method | Endpoint | Yêu cầu Body | Mã phản hồi | Mô tả |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/health` | Không | `200 OK` | Health-check dịch vụ Backend |

---

## 6. KẾT LUẬN VÀ ĐÁNH GIÁ CHẤT LƯỢNG

1. **Độ phủ Kiểm thử (Test Coverage)**: Hệ thống đạt tỉ lệ **100% PASS** trên toàn bộ **22 kịch bản Integration Test** trong `backend/main_test.go`, bao gồm:
   - Toàn bộ chu trình từ đăng ký, đăng nhập, bảo vệ token JWT, phân quyền RBAC.
   - Luồng thi hoàn chỉnh: Bắt đầu, Auto-save khôi phục đáp án, Auto-submit khi hết giờ, Chấm điểm tự động.
   - Các cơ chế bảo mật nâng cao: Zero-Leak DTO, Chống lộ đáp án trước khi kết thúc đề, Khóa bi quan `FOR UPDATE`, Vùng đệm 5 giây trễ mạng, Chống IDOR.
   - Bộ CRUD hoàn chỉnh: Đề thi, Câu hỏi & Nhập hàng loạt, Sửa điểm, Cho thi lại, Quản lý tài khoản sinh viên và Danh mục bài thi đang mở.
   - Lớp Chống gian lận và Giám sát: Xáo trộn câu hỏi tất định theo seed `attempt.ID`, phát hiện chuyển tab và tự động đình chỉ thi khi vi phạm 3 lần.
   - Cơ chế Bảo mật và Lưu vết: In-Memory Sliding Window Rate Limiting (15 req/phút/IP), bảng Audit Log lưu vết chi tiết mọi can thiệp của Giảng viên.
   - Tiện ích Khảo thí: Xuất bảng điểm CSV có tiền tố UTF-8 BOM chuẩn cho Microsoft Excel.
   - Hệ thống Thông báo Nội bộ (`TestNotifications_FullWorkflow`): Kiểm thử tự động kích hoạt thông báo đa chiều (Học sinh khiếu nại -> Báo Giảng viên; Giảng viên phản hồi -> Báo Học sinh; Giảng viên can thiệp điểm số -> Báo Học sinh), kiểm tra tính năng đánh dấu đã đọc đơn lẻ và đánh dấu đã đọc toàn bộ.
   - Gia cố Bảo mật & Ổn định (`TestSecurityMiddleware_HeadersAndBodyLimit`): Kiểm tra tự động đính kèm Security Headers (`X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: strict-origin-when-cross-origin`) và cơ chế chặn request body vượt quá hạn mức 2MB.
2. **Trải nghiệm Người dùng (UX)**:
   - Loại bỏ hoàn toàn 100% các biểu tượng cảm xúc phi học thuật và hộp thoại `window.alert()` thô sơ.
   - Phòng thi trực tuyến 2 cột với Bảng điều khiển câu hỏi tương tác trực quan (Question Palette 4 trạng thái) đáp ứng các tiêu chuẩn khảo thí đại học thực tế.
   - Trung tâm thông báo trực quan (Notification Dropdown) tích hợp huy hiệu số lượng chưa đọc và tự động polling thời gian thực.
   - Cảnh báo vi phạm quy chế thi trực quan và tức thời.
3. **Tính toàn vẹn Dữ liệu (ACID Integrity) & Chống xung đột**:
   - Các hành vi xóa (đề thi, tài khoản sinh viên, reset bài thi) và chấm thi/sửa điểm đều được thực hiện trọn vẹn trong Database Transaction, đảm bảo hệ thống không có dữ liệu mồ côi hay trạng thái bất đồng nhất.
   - Chế độ SQLite WAL và Busy Timeout 5000ms triệt tiêu hoàn toàn lỗi khóa cơ sở dữ liệu `database is locked (5)` trong môi trường kiểm thử và phát triển.
