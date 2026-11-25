import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { authAPI } from '../api/endpoints';
import { useAuth } from '../contexts/AuthContext';
import Spinner from '../components/Spinner';
import ThemeToggle from '../components/ThemeToggle';

const DEMO_ACCOUNTS = [
  { label: 'Giảng viên', email: 'teacher@test.com', password: '123456', role: 'Teacher' },
  { label: 'Sinh viên 1', email: 'student@test.com', password: '123456', role: 'Student' },
  { label: 'Sinh viên 2', email: 'student2@test.com', password: '123456', role: 'Student' },
  { label: 'Sinh viên 3', email: 'student3@test.com', password: '123456', role: 'Student' },
];

export default function LoginPage() {
  const [form, setForm] = useState({ email: '', password: '' });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const { login } = useAuth();
  const navigate = useNavigate();

  const handleChange = (e) => {
    setForm((f) => ({ ...f, [e.target.name]: e.target.value }));
    setError('');
  };

  const fillDemo = (acc) => {
    setForm({ email: acc.email, password: acc.password });
    setError('');
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      const { data } = await authAPI.login(form);
      login(data.token, data.user);
      navigate(data.user.role === 'teacher' ? '/teacher' : '/student', { replace: true });
    } catch (err) {
      setError(err.response?.data?.error || 'Đăng nhập không thành công. Vui lòng kiểm tra lại thông tin.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-slate-50 dark:bg-slate-950 flex flex-col justify-center py-12 px-4 sm:px-6 lg:px-8 relative transition-colors">
      {/* Top right theme toggle */}
      <div className="absolute top-4 right-4">
        <ThemeToggle />
      </div>

      <div className="sm:mx-auto sm:w-full sm:max-w-md">
        <div className="flex justify-center mb-3">
          <div className="w-10 h-10 rounded-xl bg-blue-600 flex items-center justify-center text-white font-bold text-sm tracking-wider shadow-sm">
            EX
          </div>
        </div>
        <h1 className="text-center text-xl font-bold tracking-tight text-slate-900 dark:text-slate-100">
          Cổng Thi Trực Tuyến
        </h1>
        <p className="mt-1 text-center text-xs text-slate-500 dark:text-slate-400">
          Hệ thống khảo thí và đánh giá kết quả học tập
        </p>
      </div>

      <div className="mt-6 sm:mx-auto sm:w-full sm:max-w-md">
        <div className="bg-white dark:bg-slate-900 py-8 px-6 sm:px-8 shadow-sm border border-slate-200/90 dark:border-slate-800 rounded-xl">
          {/* Demo quick fill */}
          <div className="mb-6 p-3 bg-slate-50/80 dark:bg-slate-800/60 rounded-lg border border-slate-200/70 dark:border-slate-700">
            <div className="text-[11px] font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider mb-2">
              Tài khoản thử nghiệm:
            </div>
            <div className="grid grid-cols-2 gap-2">
              {DEMO_ACCOUNTS.map((acc) => (
                <button
                  key={acc.email}
                  type="button"
                  onClick={() => fillDemo(acc)}
                  className="p-2 rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 hover:border-blue-400 dark:hover:border-blue-500 hover:bg-blue-50/40 dark:hover:bg-blue-950/40 text-left transition-all"
                >
                  <div className="text-xs font-semibold text-slate-800 dark:text-slate-200 leading-tight">{acc.label}</div>
                  <div className="text-[10px] text-slate-400 dark:text-slate-500 font-mono mt-0.5 truncate">{acc.email}</div>
                </button>
              ))}
            </div>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="label">Địa chỉ Email</label>
              <input
                name="email"
                type="email"
                value={form.email}
                onChange={handleChange}
                className="input"
                placeholder="name@institution.edu.vn"
                required
                autoComplete="email"
              />
            </div>
            <div>
              <label className="label">Mật khẩu</label>
              <input
                name="password"
                type="password"
                value={form.password}
                onChange={handleChange}
                className="input"
                placeholder="••••••"
                required
                autoComplete="current-password"
              />
            </div>

            {error && (
              <div className="p-3 bg-rose-50 dark:bg-rose-950/40 border border-rose-200 dark:border-rose-800 rounded-lg text-rose-700 dark:text-rose-300 text-xs">
                {error}
              </div>
            )}

            <button type="submit" disabled={loading} className="btn-primary w-full py-2.5 mt-2">
              {loading ? <Spinner size="sm" /> : null}
              {loading ? 'Đang xác thực...' : 'Đăng nhập'}
            </button>
          </form>

          <div className="mt-6 pt-5 border-t border-slate-100 dark:border-slate-800 text-center">
            <span className="text-xs text-slate-500 dark:text-slate-400">Chưa có tài khoản khảo thí? </span>
            <Link to="/register" className="text-xs font-semibold text-blue-600 dark:text-blue-400 hover:text-blue-700 hover:underline">
              Đăng ký sinh viên mới
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
