import api from './axios';

export const authAPI = {
  login: (data) => api.post('/auth/login', data),
  register: (data) => api.post('/auth/register', data),
};

export const profileAPI = {
  getProfile: () => api.get('/profile'),
  updateProfile: (data) => api.put('/profile', data),
};

export const directoryAPI = {
  getDirectory: (params) => api.get('/users/directory', { params }),
};

export const classroomAPI = {
  list: () => api.get('/classes'),
  get: (id) => api.get(`/classes/${id}`),
  create: (data) => api.post('/teacher/classes', data),
  update: (id, data) => api.put(`/teacher/classes/${id}`, data),
  delete: (id) => api.delete(`/teacher/classes/${id}`),
  addMember: (id, data) => api.post(`/teacher/classes/${id}/members`, data),
  removeMember: (id, studentId) => api.delete(`/teacher/classes/${id}/members/${studentId}`),
  getGradebook: (id) => api.get(`/teacher/classes/${id}/gradebook`),
  join: (code) => api.post('/student/classes/join', { code }),
};

export const teacherAPI = {
  listExams: () => api.get('/teacher/exams'),
  createExam: (data) => api.post('/teacher/exams', data),
  getExam: (id) => api.get(`/teacher/exams/${id}`),
  updateExam: (id, data) => api.put(`/teacher/exams/${id}`, data),
  deleteExam: (id) => api.delete(`/teacher/exams/${id}`),
  addQuestions: (id, data) => api.post(`/teacher/exams/${id}/questions`, data),
  bulkAddQuestions: (id, data) => api.post(`/teacher/exams/${id}/questions/bulk`, data),
  updateQuestion: (questionId, data) => api.put(`/teacher/questions/${questionId}`, data),
  deleteQuestion: (questionId) => api.delete(`/teacher/questions/${questionId}`),
  publishExam: (id) => api.patch(`/teacher/exams/${id}/publish`),
  getResults: (id) => api.get(`/teacher/exams/${id}/results`),
  updateAttemptScore: (attemptId, data) => api.patch(`/teacher/attempts/${attemptId}/score`, data),
  resetAttempt: (attemptId) => api.delete(`/teacher/attempts/${attemptId}`),
  listAppeals: (examId, params) => api.get(`/teacher/exams/${examId}/appeals`, { params }),
  getAppeal: (appealId) => api.get(`/teacher/appeals/${appealId}`),
  replyAppeal: (appealId, data) => api.patch(`/teacher/appeals/${appealId}/reply`, data),
  rejectAppeal: (appealId) => api.patch(`/teacher/appeals/${appealId}/reject`),
  resolveAppeal: (appealId) => api.patch(`/teacher/appeals/${appealId}/resolve`),
  listStudents: (params) => api.get('/teacher/students', { params }),
  createStudent: (data) => api.post('/teacher/students', data),
  updateStudent: (studentId, data) => api.put(`/teacher/students/${studentId}`, data),
  deleteStudent: (studentId) => api.delete(`/teacher/students/${studentId}`),
  updateUserRole: (id, role) => api.patch(`/teacher/users/${id}/role`, { role }),
  getAuditLogs: () => api.get('/teacher/audit-logs'),
  exportResultsCSV: (examId) => api.get(`/teacher/exams/${examId}/results/export`, { responseType: 'blob' }),
};

export const studentAPI = {
  getAvailableExams: () => api.get('/student/exams/available'),
  lookupExam: (code) => api.get(`/student/exams/lookup?code=${code}`),
  listAttempts: () => api.get('/student/attempts'),
  getHistory: (params) => api.get('/student/history', { params }),
  getHistoryDetail: (submissionId) => api.get(`/student/history/${submissionId}`),
  createAppeal: (data) => api.post('/student/appeals', data),
  startExam: (examId, data = {}) => api.post(`/student/exams/${examId}/start`, data),
  saveAnswers: (attemptId, data) => api.post(`/student/attempts/${attemptId}/save`, data),
  recordViolation: (attemptId, data = {}) => api.post(`/student/attempts/${attemptId}/violation`, data),
  submitExam: (attemptId, data) => api.post(`/student/attempts/${attemptId}/submit`, data),
  getResult: (attemptId) => api.get(`/student/attempts/${attemptId}/result`),
};

export const notificationAPI = {
  getAll: () => api.get('/notifications'),
  markAsRead: (id) => api.patch(`/notifications/${id}/read`),
  markAllAsRead: () => api.patch('/notifications/read-all'),
  delete: (id) => api.delete(`/notifications/${id}`),
  clearRead: () => api.delete('/notifications/clear-read'),
};

export const scheduleAPI = {
  getSchedule: () => api.get('/exams/schedule'),
};
