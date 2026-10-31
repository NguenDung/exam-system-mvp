import { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { profileAPI } from '../api/endpoints';

export default function ProfilePage() {
  const { user, updateUser } = useAuth();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  const [formData, setFormData] = useState({
    full_name: '',
    email: '',
    role: '',
    department: '',
    student_code: '',
    bio: '',
    avatar_url: '',
    new_password: '',
    confirm_password: '',
  });

  const [avatarPreview, setAvatarPreview] = useState('');

  useEffect(() => {
    async function loadProfile() {
      try {
        setLoading(true);
        const { data } = await profileAPI.getProfile();
        const u = data.user;
        setFormData({
          full_name: u.full_name || '',
          email: u.email || '',
          role: u.role || '',
          department: u.department || '',
          student_code: u.student_code || '',
          bio: u.bio || '',
          avatar_url: u.avatar_url || '',
          new_password: '',
          confirm_password: '',
        });
        setAvatarPreview(u.avatar_url || '');
      } catch (err) {
        setError(err.response?.data?.error || 'Không thể tải thông tin hồ sơ');
      } finally {
        setLoading(false);
      }
    }
    loadProfile();
  }, []);

  const handleFileChange = (e) => {
    const file = e.target.files?.[0];
    if (!file) return;

    if (!file.type.startsWith('image/')) {
      setError('Vui lòng chọn tệp hình ảnh hợp lệ (PNG, JPG, WebP)');
      return;
    }

    if (file.size > 2 * 1024 * 1024) {
      setError('Dung lượng ảnh tối đa là 2MB');
      return;
    }

    const reader = new FileReader();
    reader.onload = (event) => {
      const base64 = event.target?.result;
      setFormData((prev) => ({ ...prev, avatar_url: base64 }));
      setAvatarPreview(base64);
      setError('');
    };
    reader.readAsDataURL(file);
  };

  const handleAvatarUrlChange = (e) => {
    const url = e.target.value;
    setFormData((prev) => ({ ...prev, avatar_url: url }));
    setAvatarPreview(url);
  };

  const handleRemoveAvatar = () => {
    setFormData((prev) => ({ ...prev, avatar_url: '' }));
    setAvatarPreview('');
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    if (!formData.full_name.trim()) {
      setError('Họ và tên không được để trống');
      return;
    }

    if (formData.new_password) {
      if (formData.new_password.length < 6) {
        setError('Mật khẩu mới phải có tối thiểu 6 ký tự');
        return;
      }
      if (formData.new_password !== formData.confirm_password) {
        setError('Xác nhận mật khẩu mới không khớp');
        return;
      }
    }

    try {
      setSaving(true);
      const payload = {
        full_name: formData.full_name.trim(),
        department: formData.department.trim(),
        student_code: formData.student_code.trim(),
        bio: formData.bio.trim(),
        avatar_url: formData.avatar_url,
      };

      if (formData.new_password) {
        payload.password = formData.new_password;
      }

      const { data } = await profileAPI.updateProfile(payload);
      if (data.user) {
        updateUser({
          ...user,
          ...data.user,
        });
      }
      setFormData((prev) => ({ ...prev, new_password: '', confirm_password: '' }));
      setSuccess('Cập nhật hồ sơ thành công');
    } catch (err) {
      setError(err.response?.data?.error || 'Cập nhật hồ sơ thất bại');
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="py-12 text-center text-slate-500 text-sm">
        Đang tải thông tin hồ sơ...
      </div>
    );
  }

  const initial = (formData.full_name || 'U').charAt(0).toUpperCase();

  return (
    <div className="max-w-3xl mx-auto space-y-6">
      <div className="border-b border-slate-200 dark:border-slate-800 pb-4">
        <h1 className="text-xl font-bold text-slate-900 dark:text-slate-100 tracking-tight">Hồ sơ cá nhân</h1>
        <p className="text-sm text-slate-500 dark:text-slate-400 mt-1">
          Quản lý thông tin định danh, khoa bộ môn và mật khẩu đăng nhập tài khoản.
        </p>
      </div>

      {error && (
        <div className="p-3 bg-red-50 dark:bg-rose-950/40 border border-red-200 dark:border-rose-800 text-red-700 dark:text-rose-300 text-sm rounded-md">
          {error}
        </div>
      )}

      {success && (
        <div className="p-3 bg-emerald-50 dark:bg-emerald-950/40 border border-emerald-200 dark:border-emerald-800 text-emerald-700 dark:text-emerald-300 text-sm rounded-md">
          {success}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-6 bg-white dark:bg-slate-900 p-6 rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm">
        {/* Avatar Section */}
        <div className="flex flex-col sm:flex-row items-start sm:items-center gap-6 pb-6 border-b border-slate-100 dark:border-slate-800">
          <div className="relative">
            {avatarPreview ? (
              <img
                src={avatarPreview}
                alt="Avatar preview"
                className="w-20 h-20 rounded-full object-cover border-2 border-slate-300 dark:border-slate-700 shadow-sm"
              />
            ) : (
              <div className="w-20 h-20 rounded-full bg-slate-800 text-white flex items-center justify-center font-bold text-2xl border-2 border-slate-300 dark:border-slate-700 shadow-sm">
                {initial}
              </div>
            )}
          </div>

          <div className="flex-1 space-y-2">
            <label className="block text-sm font-semibold text-slate-800 dark:text-slate-200">Ảnh đại diện</label>
            <div className="flex flex-wrap items-center gap-3">
              <label className="btn-secondary text-xs cursor-pointer px-3 py-1.5 border border-slate-300 dark:border-slate-700 rounded hover:bg-slate-50 dark:hover:bg-slate-800">
                Tải ảnh từ máy tính
                <input
                  type="file"
                  accept="image/png, image/jpeg, image/webp"
                  onChange={handleFileChange}
                  className="hidden"
                />
              </label>
              {avatarPreview && (
                <button
                  type="button"
                  onClick={handleRemoveAvatar}
                  className="text-xs text-rose-600 dark:text-rose-400 hover:text-rose-800 font-medium"
                >
                  Gỡ ảnh
                </button>
              )}
            </div>
            <div className="mt-2">
              <input
                type="text"
                placeholder="Hoặc dán URL ảnh trực tiếp (https://...)"
                value={formData.avatar_url.startsWith('data:') ? '' : formData.avatar_url}
                onChange={handleAvatarUrlChange}
                className="w-full text-xs px-3 py-1.5 border border-slate-300 dark:border-slate-700 rounded bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 focus:ring-1 focus:ring-blue-500 focus:outline-none"
              />
            </div>
          </div>
        </div>

        {/* Info Grid */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 uppercase mb-1">
              Email tài khoản
            </label>
            <input
              type="email"
              value={formData.email}
              disabled
              className="w-full text-sm px-3 py-2 bg-slate-100 dark:bg-slate-800/80 border border-slate-200 dark:border-slate-700 rounded text-slate-500 dark:text-slate-400 cursor-not-allowed font-mono"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 uppercase mb-1">
              Vai trò hệ thống
            </label>
            <div className="px-3 py-2 bg-slate-100 dark:bg-slate-800/80 border border-slate-200 dark:border-slate-700 rounded text-sm text-slate-700 dark:text-slate-300 font-medium">
              {formData.role === 'teacher' ? 'Giảng viên' : 'Sinh viên'}
            </div>
          </div>

          <div>
            <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 uppercase mb-1">
              Họ và tên *
            </label>
            <input
              type="text"
              required
              value={formData.full_name}
              onChange={(e) => setFormData({ ...formData, full_name: e.target.value })}
              className="w-full text-sm px-3 py-2 border border-slate-300 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 focus:ring-1 focus:ring-blue-500 focus:outline-none"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 uppercase mb-1">
              {formData.role === 'teacher' ? 'Mã giảng viên' : 'Mã sinh viên'}
            </label>
            <input
              type="text"
              placeholder={formData.role === 'teacher' ? 'GV2026...' : 'SV2026...'}
              value={formData.student_code}
              onChange={(e) => setFormData({ ...formData, student_code: e.target.value })}
              className="w-full text-sm px-3 py-2 border border-slate-300 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 focus:ring-1 focus:ring-blue-500 focus:outline-none"
            />
          </div>

          <div className="sm:col-span-2">
            <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 uppercase mb-1">
              Khoa / Bộ môn
            </label>
            <input
              type="text"
              placeholder="Khoa Công nghệ thông tin, Viện An toàn thông tin..."
              value={formData.department}
              onChange={(e) => setFormData({ ...formData, department: e.target.value })}
              className="w-full text-sm px-3 py-2 border border-slate-300 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 focus:ring-1 focus:ring-blue-500 focus:outline-none"
            />
          </div>

          <div className="sm:col-span-2">
            <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 uppercase mb-1">
              Giới thiệu bản thân (Bio)
            </label>
            <textarea
              rows={3}
              placeholder="Thông tin ngắn về hướng nghiên cứu, lớp sinh hoạt hoặc chuyên ngành..."
              value={formData.bio}
              onChange={(e) => setFormData({ ...formData, bio: e.target.value })}
              className="w-full text-sm px-3 py-2 border border-slate-300 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 focus:ring-1 focus:ring-blue-500 focus:outline-none"
            />
          </div>
        </div>

        {/* Password Change */}
        <div className="pt-4 border-t border-slate-100 dark:border-slate-800">
          <h2 className="text-sm font-bold text-slate-900 dark:text-slate-100 mb-3">Đổi mật khẩu (tùy chọn)</h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
                Mật khẩu mới
              </label>
              <input
                type="password"
                placeholder="Để trống nếu không muốn đổi"
                value={formData.new_password}
                onChange={(e) => setFormData({ ...formData, new_password: e.target.value })}
                className="w-full text-sm px-3 py-2 border border-slate-300 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 focus:ring-1 focus:ring-blue-500 focus:outline-none"
              />
            </div>
            <div>
              <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
                Xác nhận mật khẩu mới
              </label>
              <input
                type="password"
                placeholder="Nhập lại mật khẩu mới"
                value={formData.confirm_password}
                onChange={(e) => setFormData({ ...formData, confirm_password: e.target.value })}
                className="w-full text-sm px-3 py-2 border border-slate-300 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 focus:ring-1 focus:ring-blue-500 focus:outline-none"
              />
            </div>
          </div>
        </div>

        {/* Actions */}
        <div className="flex justify-end pt-4 border-t border-slate-100 dark:border-slate-800">
          <button
            type="submit"
            disabled={saving}
            className="btn-primary px-5 py-2 text-sm font-semibold text-white bg-indigo-600 rounded-md hover:bg-indigo-700 disabled:opacity-50"
          >
            {saving ? 'Đang lưu...' : 'Lưu thay đổi'}
          </button>
        </div>
      </form>
    </div>
  );
}
