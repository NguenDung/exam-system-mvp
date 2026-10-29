import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './contexts/AuthContext';
import ProtectedRoute from './components/ProtectedRoute';
import TeacherLayout from './components/TeacherLayout';
import StudentLayout from './components/StudentLayout';

// Pages
import LoginPage from './pages/LoginPage';
import RegisterPage from './pages/RegisterPage';
import TeacherExamsPage from './pages/teacher/TeacherExamsPage';
import TeacherExamDetailPage from './pages/teacher/TeacherExamDetailPage';
import TeacherResultsPage from './pages/teacher/TeacherResultsPage';
import TeacherStudentsPage from './pages/teacher/TeacherStudentsPage';
import StudentHomePage from './pages/student/StudentHomePage';
import ExamTakingPage from './pages/student/ExamTakingPage';
import ExamResultPage from './pages/student/ExamResultPage';
import StudentHistoryPage from './pages/student/StudentHistoryPage';
import StudentHistoryDetailPage from './pages/student/StudentHistoryDetailPage';

export default function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          {/* Public */}
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />

          {/* Teacher */}
          <Route element={<ProtectedRoute role="teacher" />}>
            <Route element={<TeacherLayout />}>
              <Route path="/teacher" element={<TeacherExamsPage />} />
              <Route path="/teacher/exams/:id" element={<TeacherExamDetailPage />} />
              <Route path="/teacher/exams/:id/results" element={<TeacherResultsPage />} />
              <Route path="/teacher/students" element={<TeacherStudentsPage />} />
            </Route>
          </Route>

          {/* Student */}
          <Route element={<ProtectedRoute role="student" />}>
            {/* ExamTakingPage is full-screen, no layout wrapper */}
            <Route path="/student/exam/:exam_id/take" element={<ExamTakingPage />} />
            <Route element={<StudentLayout />}>
              <Route path="/student" element={<StudentHomePage />} />
              <Route path="/student/history" element={<StudentHistoryPage />} />
              <Route path="/student/history/:submissionId" element={<StudentHistoryDetailPage />} />
              <Route path="/student/attempts/:attempt_id/result" element={<ExamResultPage />} />
            </Route>
          </Route>

          {/* Fallback */}
          <Route path="/" element={<Navigate to="/login" replace />} />
          <Route path="*" element={<Navigate to="/login" replace />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  );
}
