import { useState, useEffect, useCallback } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { directoryAPI, teacherAPI } from '../api/endpoints';

export default function DirectoryPage() {
  const { user: currentUser } = useAuth();
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  const [searchQuery, setSearchQuery] = useState('');
  const [roleFilter, setRoleFilter] = useState('');
  const [updatingId, setUpdatingId] = useState(null);

  const fetchDirectory = useCallback(async () => {
    try {
      setLoading(true);
      setError('');
      const params = {};
      if (searchQuery.trim()) params.q = searchQuery.trim();
      if (roleFilter) params.role = roleFilter;

      const { data } = await directoryAPI.getDirectory(params);
      setUsers(data.users || []);
    } catch (err) {
      setError(err.response?.data?.error || 'Không thể tải danh bạ người dùng');
    } finally {
      setLoading(false);
    }
  }, [searchQuery, roleFilter]);

  useEffect(() => {
    const timer = setTimeout(() => {
      fetchDirectory();
    }, 250);
    return () => clearTimeout(timer);
  }, [fetchDirectory]);

  const handleRoleChange = async (targetUser, newRole) => {
    if (targetUser.id === currentUser?.id) {
      setError('Bạn không thể tự thay đổi vai trò của chính mình');
      return;
    }
    if (!window.confirm(`Xác nhận chuyển vai trò của ${targetUser.full_name} thành "${newRole === 'teacher' ? 'Giảng viên' : 'Sinh viên'}"?`)) {
      return;
    }

    try {
      setUpdatingId(targetUser.id);
      setError('');
      setSuccess('');
      await teacherAPI.updateUserRole(targetUser.id, newRole);
      setUsers((prev) =>
        prev.map((u) => (u.id === targetUser.id ? { ...u, role: newRole } : u))
      );
      setSuccess(`Đã cập nhật vai trò cho ${targetUser.full_name}`);
    } catch (err) {
      setError(err.response?.data?.error || 'Không thể cập nhật vai trò');
    } finally {
      setUpdatingId(null);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 border-b border-slate-200 pb-4">
        <div>
          <h1 className="text-xl font-bold text-slate-900 tracking-tight">Danh bạ người dùng</h1>
          <p className="text-sm text-slate-500 mt-1">
            Tra cứu thông tin giảng viên, sinh viên và đơn vị đào tạo trong toàn hệ thống.
          </p>
        </div>
      </div>

      {error && (
        <div className="p-3 bg-red-50 border border-red-200 text-red-700 text-sm rounded-md">
          {error}
        </div>
      )}

      {success && (
        <div className="p-3 bg-emerald-50 border border-emerald-200 text-emerald-700 text-sm rounded-md">
          {success}
        </div>
      )}

      {/* Filter and Search Bar */}
      <div className="bg-white p-4 rounded-lg border border-slate-200 shadow-sm flex flex-col sm:flex-row gap-3">
        <div className="flex-1">
          <input
            type="text"
            placeholder="Tìm theo họ tên, email, mã số SV/GV, khoa bộ môn..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full text-sm px-3.5 py-2 border border-slate-300 rounded-md focus:ring-1 focus:ring-indigo-500 focus:outline-none"
          />
        </div>
        <div className="w-full sm:w-48">
          <select
            value={roleFilter}
            onChange={(e) => setRoleFilter(e.target.value)}
            className="w-full text-sm px-3 py-2 border border-slate-300 rounded-md focus:ring-1 focus:ring-indigo-500 focus:outline-none bg-white text-slate-700"
          >
            <option value="">Tất cả vai trò</option>
            <option value="teacher">Giảng viên</option>
            <option value="student">Sinh viên</option>
          </select>
        </div>
      </div>

      {/* User Cards Grid */}
      {loading ? (
        <div className="py-12 text-center text-slate-500 text-sm">
          Đang tải danh sách người dùng...
        </div>
      ) : users.length === 0 ? (
        <div className="bg-white p-8 rounded-lg border border-slate-200 text-center text-slate-500 text-sm">
          Không tìm thấy tài khoản nào khớp với điều kiện tìm kiếm.
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {users.map((item) => {
            const initial = (item.full_name || 'U').charAt(0).toUpperCase();
            const isTeacher = item.role === 'teacher';
            const isSelf = item.id === currentUser?.id;

            return (
              <div
                key={item.id}
                className="bg-white p-5 rounded-lg border border-slate-200 shadow-sm flex flex-col justify-between hover:border-slate-300 transition-colors"
              >
                <div>
                  <div className="flex items-start gap-3.5 mb-3">
                    {item.avatar_url ? (
                      <img
                        src={item.avatar_url}
                        alt={item.full_name}
                        className="w-12 h-12 rounded-full object-cover border border-slate-300 bg-slate-100 flex-shrink-0"
                      />
                    ) : (
                      <div className="w-12 h-12 rounded-full bg-slate-800 text-white flex items-center justify-center font-bold text-lg border border-slate-300 flex-shrink-0">
                        {initial}
                      </div>
                    )}
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-2">
                        <h3 className="text-sm font-bold text-slate-900 truncate">
                          {item.full_name}
                        </h3>
                        {isSelf && (
                          <span className="text-[10px] bg-slate-100 text-slate-600 px-1.5 py-0.5 rounded font-medium">
                            Tôi
                          </span>
                        )}
                      </div>
                      <p className="text-xs text-slate-500 truncate mt-0.5">{item.email}</p>
                      <div className="mt-1.5 flex items-center gap-2">
                        <span
                          className={`text-[11px] font-semibold px-2 py-0.5 rounded ${
                            isTeacher
                              ? 'bg-amber-50 text-amber-800 border border-amber-200'
                              : 'bg-indigo-50 text-indigo-700 border border-indigo-200'
                          }`}
                        >
                          {isTeacher ? 'Giảng viên' : 'Sinh viên'}
                        </span>
                        {item.student_code && (
                          <span className="text-[11px] text-slate-600 font-mono bg-slate-100 px-1.5 py-0.5 rounded">
                            {item.student_code}
                          </span>
                        )}
                      </div>
                    </div>
                  </div>

                  {item.department && (
                    <div className="text-xs text-slate-700 mt-2 bg-slate-50 p-2 rounded border border-slate-100">
                      <span className="font-semibold text-slate-500">Khoa / Đơn vị: </span>
                      {item.department}
                    </div>
                  )}

                  {item.bio && (
                    <p className="text-xs text-slate-600 mt-2 line-clamp-2 leading-relaxed">
                      {item.bio}
                    </p>
                  )}
                </div>

                {/* Role Switcher for Teachers */}
                {currentUser?.role === 'teacher' && !isSelf && (
                  <div className="pt-3 mt-3 border-t border-slate-100 flex items-center justify-between">
                    <span className="text-[11px] text-slate-500">Phân quyền:</span>
                    <button
                      type="button"
                      disabled={updatingId === item.id}
                      onClick={() => handleRoleChange(item, isTeacher ? 'student' : 'teacher')}
                      className="text-xs text-indigo-600 hover:text-indigo-800 font-medium hover:underline disabled:opacity-50"
                    >
                      {updatingId === item.id
                        ? 'Đang đổi...'
                        : isTeacher
                        ? 'Chuyển sang Sinh viên'
                        : 'Nâng lên Giảng viên'}
                    </button>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
