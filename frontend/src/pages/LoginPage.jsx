import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { authAPI } from '../api/endpoints';
import { useAuth } from '../contexts/AuthContext';
import Spinner from '../components/Spinner';

const DEMO_ACCOUNTS = [
  { label: 'Giảng viên', email: 'teacher@test.com', password: '123456' },
  { label: 'Sinh viên 1', email: 'student@test.com', password: '123456' },
  { label: 'Sinh viên 2', email: 'student2@test.com', password: '123456' },
  { label: 'Sinh viên 3', email: 'student3@test.com', password: '123456' },
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
    <div className="min-h-screen bg-slate-100 flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-6">
          <h1 className="text-2xl font-bold text-slate-900 tracking-tight uppercase">Hệ Thống Thi Trực Tuyến</h1>
          <p className="text-slate-600 text-sm mt-1">Cổng khảo thí và đánh giá nội bộ</p>
        </div>

        <div className="bg-white rounded-lg shadow-sm border border-slate-200 p-6 sm:p-8">
          <h2 className="text-lg font-semibold text-slate-800 mb-5">Đăng nhập tài khoản</h2>

          {/* Demo quick fill */}
          <div className="mb-6 p-3 bg-slate-50 rounded-md border border-slate-200">
            <p className="text-xs font-semibold text-slate-600 mb-2 uppercase tracking-wide">Tài khoản thử nghiệm:</p>
            <div className="grid grid-cols-2 gap-2">
              {DEMO_ACCOUNTS.map((acc) => (
                <button
                  key={acc.email}
                  type="button"
                  onClick={() => fillDemo(acc)}
                  className="text-xs py-1.5 px-2.5 rounded border border-slate-300 bg-white text-slate-700 hover:bg-slate-100 font-medium text-left truncate transition-colors"
                >
                  {acc.label}
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
              <div className="p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm">
                {error}
              </div>
            )}

            <button type="submit" disabled={loading} className="btn-primary w-full py-2.5 mt-2">
              {loading ? <Spinner size="sm" /> : null}
              {loading ? 'Đang xác thực...' : 'Đăng nhập'}
            </button>
          </form>

          <p className="mt-6 text-center text-sm text-slate-500">
            Thí sinh chưa có tài khoản?{' '}
            <Link to="/register" className="text-indigo-600 font-medium hover:underline">
              Đăng ký tại đây
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}
