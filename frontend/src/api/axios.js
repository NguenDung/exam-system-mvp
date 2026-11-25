import axios from 'axios';

const preferredLanguages = navigator.languages?.length
  ? navigator.languages.join(',')
  : navigator.language || 'en';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8081/api',
  headers: {
    'Content-Type': 'application/json',
    'Accept-Language': preferredLanguages,
  },
  timeout: 10000,
});

// ── Request interceptor: attach JWT ──────────────────────────────────────────
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// ── Response interceptor: handle 401 globally ────────────────────────────────
api.interceptors.response.use(
  (response) => response,
  (error) => {
    const payload = error.response?.data;
    const traceId = payload?.trace_id || error.response?.headers?.['x-trace-id'];
    if (error.response?.status >= 500 && payload?.error && traceId) {
      const isVietnamese = (navigator.language || '').toLowerCase().startsWith('vi');
      const reference = isVietnamese ? `Mã tra cứu: ${traceId}` : `Reference ID: ${traceId}`;
      payload.error = `${payload.error} (${reference})`;
    }

    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export default api;
