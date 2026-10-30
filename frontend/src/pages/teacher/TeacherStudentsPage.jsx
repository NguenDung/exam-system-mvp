import { useState, useEffect } from 'react';
import { teacherAPI } from '../../api/endpoints';
import Spinner from '../../components/Spinner';

export default function TeacherStudentsPage() {
  const [students, setStudents] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [searchQuery, setSearchQuery] = useState('');

  // Modals state
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingStudent, setEditingStudent] = useState(null);
  const [deletingStudent, setDeletingStudent] = useState(null);
  const [actionLoading, setActionLoading] = useState(false);
  const [modalError, setModalError] = useState(null);

  // Form fields
  const [formData, setFormData] = useState({
    full_name: '',
    email: '',
    password: '',
  });

  const fetchStudents = async (query = '') => {
    setLoading(true);
    setError(null);
    try {
      const res = await teacherAPI.listStudents(query ? { q: query } : {});
      setStudents(res.data.data || []);
    } catch (err) {
      setError(err.response?.data?.error || 'Không thể tải danh sách sinh viên');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStudents();
  }, []);

  const handleSearch = (e) => {
    e.preventDefault();
    fetchStudents(searchQuery);
  };

  const openCreateModal = () => {
    setFormData({ full_name: '', email: '', password: '' });
    setModalError(null);
    setShowCreateModal(true);
  };

  const handleCreateStudent = async (e) => {
    e.preventDefault();
    setModalError(null);
    setActionLoading(true);
    try {
      await teacherAPI.createStudent(formData);
      setShowCreateModal(false);
      fetchStudents(searchQuery);
    } catch (err) {
      setModalError(err.response?.data?.error || 'Lỗi khi tạo tài khoản sinh viên');
    } finally {
      setActionLoading(false);
    }
  };

  const openEditModal = (student) => {
    setEditingStudent(student);
    setFormData({
      full_name: student.full_name || '',
      email: student.email || '',
      password: '',
    });
    setModalError(null);
  };

  const handleUpdateStudent = async (e) => {
    e.preventDefault();
    if (!editingStudent) return;
    setModalError(null);
    setActionLoading(true);
    try {
      const payload = {
        full_name: formData.full_name,
        email: formData.email,
      };
      if (formData.password.trim()) {
        payload.password = formData.password.trim();
      }
      await teacherAPI.updateStudent(editingStudent.id, payload);
      setEditingStudent(null);
      fetchStudents(searchQuery);
    } catch (err) {
      setModalError(err.response?.data?.error || 'Lỗi khi cập nhật thông tin sinh viên');
    } finally {
      setActionLoading(false);
    }
  };

  const handleDeleteStudent = async () => {
    if (!deletingStudent) return;
    setModalError(null);
    setActionLoading(true);
    try {
      await teacherAPI.deleteStudent(deletingStudent.id);
      setDeletingStudent(null);
      fetchStudents(searchQuery);
    } catch (err) {
      setModalError(err.response?.data?.error || 'Lỗi khi xóa tài khoản sinh viên');
    } finally {
      setActionLoading(false);
    }
  };

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 border-b border-slate-200 pb-5">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-slate-900">Quản lý sinh viên</h1>
        </div>
        <button
          onClick={openCreateModal}
          className="inline-flex items-center justify-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-slate-900 hover:bg-slate-800 transition-colors"
        >
          Thêm sinh viên
        </button>
      </div>

      {/* Search Bar */}
      <form onSubmit={handleSearch} className="flex gap-2 max-w-md">
        <input
          type="text"
          placeholder="Tìm theo họ tên hoặc email..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="flex-1 rounded-md border border-slate-300 px-3 py-2 text-sm shadow-sm focus:border-slate-900 focus:outline-none focus:ring-1 focus:ring-slate-900"
        />
        <button
          type="submit"
          className="px-4 py-2 bg-white border border-slate-300 rounded-md text-sm font-medium text-slate-700 hover:bg-slate-50 transition-colors"
        >
          Tìm kiếm
        </button>
        {searchQuery && (
          <button
            type="button"
            onClick={() => {
              setSearchQuery('');
              fetchStudents('');
            }}
            className="px-3 py-2 text-sm text-slate-500 hover:text-slate-800"
          >
            Đặt lại
          </button>
        )}
      </form>

      {/* Error alert */}
      {error && (
        <div className="rounded-md bg-red-50 p-4 border border-red-200 text-sm text-red-700">
          {error}
        </div>
      )}

      {/* Table list */}
      {loading ? (
        <div className="flex justify-center py-12">
          <Spinner />
        </div>
      ) : students.length === 0 ? (
        <div className="text-center py-12 bg-white rounded-lg border border-slate-200 text-slate-500 text-sm">
          Chưa có sinh viên nào hoặc không tìm thấy kết quả phù hợp.
        </div>
      ) : (
        <div className="bg-white border border-slate-200 rounded-lg shadow-sm overflow-hidden">
          <table className="min-w-full divide-y divide-slate-200 text-sm">
            <thead className="bg-slate-50">
              <tr>
                <th className="px-6 py-3 text-left font-semibold text-slate-700">Mã SV / ID</th>
                <th className="px-6 py-3 text-left font-semibold text-slate-700">Họ và tên</th>
                <th className="px-6 py-3 text-left font-semibold text-slate-700">Email</th>
                <th className="px-6 py-3 text-center font-semibold text-slate-700">Lượt thi</th>
                <th className="px-6 py-3 text-left font-semibold text-slate-700">Ngày tạo</th>
                <th className="px-6 py-3 text-right font-semibold text-slate-700">Thao tác</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-200">
              {students.map((student) => (
                <tr key={student.id} className="hover:bg-slate-50 transition-colors">
                  <td className="px-6 py-4 font-mono text-xs text-slate-500">#{student.id}</td>
                  <td className="px-6 py-4 font-medium text-slate-900">{student.full_name}</td>
                  <td className="px-6 py-4 text-slate-600">{student.email}</td>
                  <td className="px-6 py-4 text-center">
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-slate-100 text-slate-800">
                      {student.completed_attempts || 0} / {student.total_attempts || 0} hoàn thành
                    </span>
                  </td>
                  <td className="px-6 py-4 text-slate-500 text-xs">
                    {student.created_at ? new Date(student.created_at).toLocaleDateString('vi-VN') : '-'}
                  </td>
                  <td className="px-6 py-4 text-right space-x-2">
                    <button
                      onClick={() => openEditModal(student)}
                      className="px-2.5 py-1 text-xs font-medium text-slate-700 bg-slate-100 hover:bg-slate-200 rounded transition-colors"
                    >
                      Sửa
                    </button>
                    <button
                      onClick={() => {
                        setDeletingStudent(student);
                        setModalError(null);
                      }}
                      className="px-2.5 py-1 text-xs font-medium text-red-700 bg-red-50 hover:bg-red-100 rounded transition-colors"
                    >
                      Xóa
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Modal: Create Student */}
      {showCreateModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full p-6 space-y-4">
            <h3 className="text-lg font-bold text-slate-900">Thêm tài khoản sinh viên</h3>
            {modalError && (
              <div className="rounded bg-red-50 p-3 text-xs text-red-700 border border-red-200">
                {modalError}
              </div>
            )}
            <form onSubmit={handleCreateStudent} className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-slate-700 mb-1">Họ và tên</label>
                <input
                  type="text"
                  required
                  value={formData.full_name}
                  onChange={(e) => setFormData({ ...formData, full_name: e.target.value })}
                  placeholder="Ví dụ: Nguyễn Văn A"
                  className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:border-slate-900 focus:outline-none focus:ring-1 focus:ring-slate-900"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-slate-700 mb-1">Email</label>
                <input
                  type="email"
                  required
                  value={formData.email}
                  onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                  placeholder="student@university.edu.vn"
                  className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:border-slate-900 focus:outline-none focus:ring-1 focus:ring-slate-900"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-slate-700 mb-1">Mật khẩu ban đầu</label>
                <input
                  type="password"
                  required
                  minLength={6}
                  value={formData.password}
                  onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                  placeholder="Tối thiểu 6 ký tự"
                  className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:border-slate-900 focus:outline-none focus:ring-1 focus:ring-slate-900"
                />
              </div>
              <div className="flex justify-end gap-2 pt-2 border-t border-slate-100">
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  disabled={actionLoading}
                  className="px-4 py-2 border border-slate-300 rounded-md text-sm text-slate-700 hover:bg-slate-50 transition-colors"
                >
                  Hủy
                </button>
                <button
                  type="submit"
                  disabled={actionLoading}
                  className="px-4 py-2 bg-slate-900 text-white rounded-md text-sm font-medium hover:bg-slate-800 transition-colors disabled:opacity-50"
                >
                  {actionLoading ? 'Đang tạo...' : 'Tạo tài khoản'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Modal: Edit Student */}
      {editingStudent && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full p-6 space-y-4">
            <h3 className="text-lg font-bold text-slate-900">Sửa thông tin sinh viên #{editingStudent.id}</h3>
            {modalError && (
              <div className="rounded bg-red-50 p-3 text-xs text-red-700 border border-red-200">
                {modalError}
              </div>
            )}
            <form onSubmit={handleUpdateStudent} className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-slate-700 mb-1">Họ và tên</label>
                <input
                  type="text"
                  required
                  value={formData.full_name}
                  onChange={(e) => setFormData({ ...formData, full_name: e.target.value })}
                  className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:border-slate-900 focus:outline-none focus:ring-1 focus:ring-slate-900"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-slate-700 mb-1">Email</label>
                <input
                  type="email"
                  required
                  value={formData.email}
                  onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                  className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:border-slate-900 focus:outline-none focus:ring-1 focus:ring-slate-900"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-slate-700 mb-1">
                  Đổi mật khẩu mới <span className="text-slate-400 font-normal">(để trống nếu không đổi)</span>
                </label>
                <input
                  type="password"
                  minLength={6}
                  value={formData.password}
                  onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                  placeholder="Để trống nếu giữ nguyên mật khẩu cũ"
                  className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:border-slate-900 focus:outline-none focus:ring-1 focus:ring-slate-900"
                />
              </div>
              <div className="flex justify-end gap-2 pt-2 border-t border-slate-100">
                <button
                  type="button"
                  onClick={() => setEditingStudent(null)}
                  disabled={actionLoading}
                  className="px-4 py-2 border border-slate-300 rounded-md text-sm text-slate-700 hover:bg-slate-50 transition-colors"
                >
                  Hủy
                </button>
                <button
                  type="submit"
                  disabled={actionLoading}
                  className="px-4 py-2 bg-slate-900 text-white rounded-md text-sm font-medium hover:bg-slate-800 transition-colors disabled:opacity-50"
                >
                  {actionLoading ? 'Đang lưu...' : 'Lưu thay đổi'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Modal: Confirm Delete */}
      {deletingStudent && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full p-6 space-y-4">
            <h3 className="text-lg font-bold text-red-600">Xác nhận xóa tài khoản sinh viên</h3>
            <p className="text-sm text-slate-600">
              Bạn có chắc chắn muốn xóa sinh viên{' '}
              <span className="font-semibold text-slate-900">{deletingStudent.full_name}</span> ({deletingStudent.email})?
            </p>
            <p className="text-xs text-red-600 bg-red-50 p-2.5 rounded border border-red-200">
              Lưu ý: Toàn bộ lượt thi, câu trả lời và khiếu nại của sinh viên này sẽ bị xóa vĩnh viễn theo cơ chế toàn vẹn dữ liệu.
            </p>
            {modalError && (
              <div className="rounded bg-red-50 p-3 text-xs text-red-700 border border-red-200">
                {modalError}
              </div>
            )}
            <div className="flex justify-end gap-2 pt-2 border-t border-slate-100">
              <button
                type="button"
                onClick={() => setDeletingStudent(null)}
                disabled={actionLoading}
                className="px-4 py-2 border border-slate-300 rounded-md text-sm text-slate-700 hover:bg-slate-50 transition-colors"
              >
                Hủy bỏ
              </button>
              <button
                type="button"
                onClick={handleDeleteStudent}
                disabled={actionLoading}
                className="px-4 py-2 bg-red-600 text-white rounded-md text-sm font-medium hover:bg-red-700 transition-colors disabled:opacity-50"
              >
                {actionLoading ? 'Đang xóa...' : 'Xóa tài khoản'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
