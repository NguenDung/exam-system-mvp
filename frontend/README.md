# Online Exam System — Frontend

Giao diện web Single-Page Application (SPA) cho hệ thống thi trắc nghiệm trực tuyến, xây dựng bằng React 19, Vite, TailwindCSS và React Router.

## Cấu trúc chính

- `src/api/`: cấu hình Axios client (`axios.js`) và tập hợp hàm gọi REST API (`endpoints.js`).
- `src/contexts/`: `AuthContext.jsx` (quản lý JWT và thông tin tài khoản) và `ThemeContext.jsx` (chuyển đổi giao diện Sáng / Tối).
- `src/components/`: layout cho Sinh viên và Giảng viên, thanh thông báo `NotificationDropdown.jsx`, `ProtectedRoute.jsx`, `ThemeToggle.jsx`, `Spinner.jsx`.
- `src/pages/`: các màn hình nghiệp vụ (`LoginPage`, `RegisterPage`, `ProfilePage`, `DirectoryPage`, `ClassroomsPage`, `ClassDetailPage`, `ExamSchedulePage`, cùng hai thư mục `student/` và `teacher/`).
- `nginx.conf` & `Dockerfile`: đóng gói bundle production và reverse proxy `/api/` tới dịch vụ backend trong Docker Compose.

## Biến môi trường

- `VITE_API_URL` (tùy chọn): URL gốc của backend API khi chạy ngoài Docker. Mặc định:
  - Development (`npm run dev`): `http://localhost:8081/api`
  - Production (`npm run build` / Nginx Docker): `/api`

## Lệnh phát triển

```bash
npm install
npm run dev      # Khởi chạy Vite dev server tại http://localhost:5173
npm run build    # Tạo bản build production trong thư mục dist/
npm run lint     # Kiểm tra mã nguồn bằng oxlint
npm run preview  # Xem trước bản build production
```
