import api from './axios';

// ── Auth ─────────────────────────────────────────────────────────────────────
export const authAPI = {
  login: (data) => api.post('/auth/login', data),
  register: (data) => api.post('/auth/register', data),
};

// ── Teacher ──────────────────────────────────────────────────────────────────
export const teacherAPI = {
  listExams: () => api.get('/teacher/exams'),
  createExam: (data) => api.post('/teacher/exams', data),
  getExam: (id) => api.get(`/teacher/exams/${id}`),
  addQuestions: (id, data) => api.post(`/teacher/exams/${id}/questions`, data),
  deleteQuestion: (questionId) => api.delete(`/teacher/questions/${questionId}`),
  publishExam: (id) => api.patch(`/teacher/exams/${id}/publish`),
  getResults: (id) => api.get(`/teacher/exams/${id}/results`),
};

// ── Student ──────────────────────────────────────────────────────────────────
export const studentAPI = {
  lookupExam: (code) => api.get(`/student/exams/lookup?code=${code}`),
  listAttempts: () => api.get('/student/attempts'),
  startExam: (examId) => api.post(`/student/exams/${examId}/start`),
  saveAnswers: (attemptId, data) => api.post(`/student/attempts/${attemptId}/save`, data),
  submitExam: (attemptId, data) => api.post(`/student/attempts/${attemptId}/submit`, data),
  getResult: (attemptId) => api.get(`/student/attempts/${attemptId}/result`),
};
