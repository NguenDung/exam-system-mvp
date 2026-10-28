import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { authAPI } from '../api/endpoints';
import Spinner from '../components/Spinner';

export default function RegisterPage() {
  const [form, setForm] = useState({ full_name: '', email: '', password: '', confirm: '' });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);
  const navigate = useNavigate();

  const handleChange = (e) => {
    setForm((f) => ({ ...f, [e.target.name]: e.target.value }));
    setError('');
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (form.password !== form.confirm) {
      setError('Mật khẩu xác nhận không khớp');
      return;
    }
    setLoading(true);
    try {
      await authAPI.register({ full_name: form.full_name, email: form.email, password: form.password });
      setSuccess(true);
      setTimeout(() => navigate('/login'), 2000);
    } catch (err) {
      setError(err.response?.data?.error || 'Đăng ký thất bại');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-indigo-50 via-white to-emerald-50 flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <div className="text-5xl mb-3">📝</div>
          <h1 className="text-3xl font-bold text-gray-900">ExamMVP</h1>
          <p className="text-gray-500 mt-1">Đăng ký tài khoản học sinh</p>
        </div>

        <div className="card shadow-xl">
          <h2 className="text-xl font-semibold text-gray-800 mb-6">Tạo tài khoản mới</h2>

          {success ? (
            <div className="p-4 bg-emerald-50 border border-emerald-200 rounded-xl text-emerald-700 text-center">
              <div className="text-3xl mb-2">✅</div>
              <p className="font-medium">Đăng ký thành công!</p>
              <p className="text-sm mt-1">Đang chuyển hướng đến trang đăng nhập...</p>
            </div>
          ) : (
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="label">Họ và tên</label>
                <input
                  name="full_name"
                  type="text"
                  value={form.full_name}
                  onChange={handleChange}
                  className="input"
                  placeholder="Nguyễn Văn A"
                  required
                />
              </div>
              <div>
                <label className="label">Email</label>
                <input
                  name="email"
                  type="email"
                  value={form.email}
                  onChange={handleChange}
                  className="input"
                  placeholder="your@email.com"
                  required
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
                  placeholder="Tối thiểu 6 ký tự"
                  required
                  minLength={6}
                />
              </div>
              <div>
                <label className="label">Xác nhận mật khẩu</label>
                <input
                  name="confirm"
                  type="password"
                  value={form.confirm}
                  onChange={handleChange}
                  className="input"
                  placeholder="Nhập lại mật khẩu"
                  required
                />
              </div>

              {error && (
                <div className="p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm flex items-center gap-2">
                  <span>⚠️</span> {error}
                </div>
              )}

              <button type="submit" disabled={loading} className="btn-primary w-full btn-lg">
                {loading ? <Spinner size="sm" /> : null}
                {loading ? 'Đang đăng ký...' : 'Đăng ký'}
              </button>
            </form>
          )}

          <p className="mt-5 text-center text-sm text-gray-500">
            Đã có tài khoản?{' '}
            <Link to="/login" className="text-indigo-600 font-medium hover:underline">
              Đăng nhập
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}
