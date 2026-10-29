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
      setTimeout(() => navigate('/login'), 1500);
    } catch (err) {
      setError(err.response?.data?.error || 'Đăng ký tài khoản không thành công');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-slate-100 flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-6">
          <h1 className="text-2xl font-bold text-slate-900 tracking-tight uppercase">Hệ Thống Thi Trực Tuyến</h1>
          <p className="text-slate-600 text-sm mt-1">Đăng ký tài khoản sinh viên</p>
        </div>

        <div className="bg-white rounded-lg shadow-sm border border-slate-200 p-6 sm:p-8">
          <h2 className="text-lg font-semibold text-slate-800 mb-5">Đăng ký tài khoản mới</h2>

          {success ? (
            <div className="p-4 bg-emerald-50 border border-emerald-200 rounded-md text-emerald-800 text-center">
              <p className="font-semibold">Đăng ký thành công</p>
              <p className="text-sm mt-1 text-emerald-700">Đang chuyển hướng sang trang đăng nhập...</p>
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
                <label className="label">Địa chỉ Email</label>
                <input
                  name="email"
                  type="email"
                  value={form.email}
                  onChange={handleChange}
                  className="input"
                  placeholder="name@student.edu.vn"
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
                <div className="p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm">
                  {error}
                </div>
              )}

              <button type="submit" disabled={loading} className="btn-primary w-full py-2.5 mt-2">
                {loading ? <Spinner size="sm" /> : null}
                {loading ? 'Đang tạo tài khoản...' : 'Xác nhận đăng ký'}
              </button>
            </form>
          )}

          <p className="mt-6 text-center text-sm text-slate-500">
            Đã có tài khoản?{' '}
            <Link to="/login" className="text-indigo-600 font-medium hover:underline">
              Đăng nhập ngay
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}
