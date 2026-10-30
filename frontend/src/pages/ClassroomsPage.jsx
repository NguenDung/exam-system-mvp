import { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { classroomAPI } from '../api/endpoints';

export default function ClassroomsPage() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const isTeacher = user?.role === 'teacher';

  const [classes, setClasses] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  // Teacher Create Modal
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newClassName, setNewClassName] = useState('');
  const [newClassDesc, setNewClassDesc] = useState('');
  const [creating, setCreating] = useState(false);

  // Teacher Edit Modal
  const [editingClass, setEditingClass] = useState(null);
  const [editName, setEditName] = useState('');
  const [editDesc, setEditDesc] = useState('');
  const [updating, setUpdating] = useState(false);

  // Student Join Modal
  const [showJoinModal, setShowJoinModal] = useState(false);
  const [joinCode, setJoinCode] = useState('');
  const [joining, setJoining] = useState(false);

  // Copy code feedback
  const [copiedCode, setCopiedCode] = useState('');

  const fetchClasses = useCallback(async () => {
    try {
      setLoading(true);
      setError('');
      const { data } = await classroomAPI.list();
      setClasses(data.classes || []);
    } catch (err) {
      setError(err.response?.data?.error || 'Không thể tải danh sách lớp học');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchClasses();
  }, [fetchClasses]);

  const handleCopyCode = (code) => {
    navigator.clipboard.writeText(code);
    setCopiedCode(code);
    setTimeout(() => setCopiedCode(''), 2000);
  };

  const handleCreateClass = async (e) => {
    e.preventDefault();
    if (!newClassName.trim()) {
      setError('Tên lớp học không được để trống');
      return;
    }
    try {
      setCreating(true);
      setError('');
      setSuccess('');
      const { data } = await classroomAPI.create({
        name: newClassName.trim(),
        description: newClassDesc.trim(),
      });
      setShowCreateModal(false);
      setNewClassName('');
      setNewClassDesc('');
      setSuccess(`Đã tạo lớp học "${data.class.name}" với mã tham gia: ${data.class.code}`);
      fetchClasses();
    } catch (err) {
      setError(err.response?.data?.error || 'Không thể tạo lớp học');
    } finally {
      setCreating(false);
    }
  };

  const handleEditClass = async (e) => {
    e.preventDefault();
    if (!editName.trim()) {
      setError('Tên lớp học không được để trống');
      return;
    }
    try {
      setUpdating(true);
      setError('');
      setSuccess('');
      await classroomAPI.update(editingClass.id, {
        name: editName.trim(),
        description: editDesc.trim(),
      });
      setEditingClass(null);
      setSuccess('Cập nhật thông tin lớp học thành công');
      fetchClasses();
    } catch (err) {
      setError(err.response?.data?.error || 'Không thể cập nhật lớp học');
    } finally {
      setUpdating(false);
    }
  };

  const handleDeleteClass = async (cls) => {
    if (!window.confirm(`Bạn có chắc chắn muốn xóa lớp học "${cls.name}"? Toàn bộ danh sách thành viên sẽ bị giải tán.`)) {
      return;
    }
    try {
      setError('');
      setSuccess('');
      await classroomAPI.delete(cls.id);
      setSuccess(`Đã xóa lớp học "${cls.name}"`);
      setClasses((prev) => prev.filter((c) => c.id !== cls.id));
    } catch (err) {
      setError(err.response?.data?.error || 'Không thể xóa lớp học');
    }
  };

  const handleJoinClass = async (e) => {
    e.preventDefault();
    const code = joinCode.trim().toUpperCase();
    if (!code) {
      setError('Vui lòng nhập mã lớp học');
      return;
    }
    try {
      setJoining(true);
      setError('');
      setSuccess('');
      const { data } = await classroomAPI.join(code);
      setShowJoinModal(false);
      setJoinCode('');
      setSuccess(`Đã tham gia lớp học "${data.class.name}" thành công`);
      fetchClasses();
    } catch (err) {
      setError(err.response?.data?.error || 'Không thể tham gia lớp học. Kiểm tra lại mã lớp.');
    } finally {
      setJoining(false);
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 border-b border-slate-200 pb-4">
        <div>
          <h1 className="text-xl font-bold text-slate-900 tracking-tight">
            {isTeacher ? 'Quản lý Lớp học' : 'Lớp học của tôi'}
          </h1>
          <p className="text-sm text-slate-500 mt-1">
            {isTeacher
              ? 'Tổ chức sinh viên theo lớp học phần, cấp phát mã tham gia và gán đề thi theo lớp.'
              : 'Danh sách các lớp học phần bạn đã tham gia để làm bài kiểm tra và theo dõi đề thi.'}
          </p>
        </div>
        <div>
          {isTeacher ? (
            <button
              onClick={() => setShowCreateModal(true)}
              className="btn-primary text-sm font-semibold px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700"
            >
              Tạo lớp học mới
            </button>
          ) : (
            <button
              onClick={() => setShowJoinModal(true)}
              className="btn-primary text-sm font-semibold px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700"
            >
              Tham gia lớp học
            </button>
          )}
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

      {/* Class List */}
      {loading ? (
        <div className="py-12 text-center text-slate-500 text-sm">
          Đang tải danh sách lớp học...
        </div>
      ) : classes.length === 0 ? (
        <div className="bg-white p-8 rounded-lg border border-slate-200 text-center text-slate-500 text-sm">
          {isTeacher
            ? 'Bạn chưa tạo lớp học nào. Hãy nhấn "Tạo lớp học mới" để bắt đầu.'
            : 'Bạn chưa tham gia lớp học nào. Nhập mã lớp do giảng viên cung cấp để tham gia.'}
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
          {classes.map((cls) => (
            <div
              key={cls.id}
              className="bg-white p-5 rounded-lg border border-slate-200 shadow-sm flex flex-col justify-between hover:border-slate-300 transition-all"
            >
              <div>
                <div className="flex items-start justify-between gap-3 mb-2">
                  <h3 className="text-base font-bold text-slate-900 leading-snug line-clamp-1">
                    {cls.name}
                  </h3>
                  <div className="flex items-center gap-1.5 flex-shrink-0">
                    <span className="font-mono text-xs font-bold bg-indigo-50 text-indigo-700 border border-indigo-200 px-2 py-0.5 rounded">
                      {cls.code}
                    </span>
                    <button
                      type="button"
                      onClick={() => handleCopyCode(cls.code)}
                      className="text-xs text-slate-500 hover:text-slate-800 p-1 rounded hover:bg-slate-100"
                      title="Sao chép mã lớp"
                    >
                      {copiedCode === cls.code ? (
                        <span className="text-emerald-600 font-medium text-[11px]">Đã chép</span>
                      ) : (
                        <span className="text-[11px] font-medium underline">Chép</span>
                      )}
                    </button>
                  </div>
                </div>

                {!isTeacher && cls.teacher_name && (
                  <p className="text-xs text-slate-600 mb-2 font-medium">
                    Giảng viên: <span className="text-slate-900">{cls.teacher_name}</span>
                  </p>
                )}

                <p className="text-xs text-slate-500 line-clamp-2 min-h-[32px] mb-4">
                  {cls.description || 'Không có mô tả cho lớp học này.'}
                </p>

                <div className="flex items-center gap-4 text-xs text-slate-600 pt-3 border-t border-slate-100">
                  <div>
                    <span className="font-bold text-slate-900">{cls.member_count}</span> sinh viên
                  </div>
                  <div>
                    <span className="font-bold text-slate-900">{cls.exam_count}</span> đề thi
                  </div>
                  <div className="text-slate-400 text-[11px] ml-auto">
                    {new Date(cls.created_at).toLocaleDateString('vi-VN')}
                  </div>
                </div>
              </div>

              <div className="mt-4 pt-3 border-t border-slate-100 flex items-center justify-between">
                <button
                  type="button"
                  onClick={() => navigate(`/classes/${cls.id}`)}
                  className="btn-primary text-xs font-semibold px-3 py-1.5 bg-indigo-600 text-white rounded hover:bg-indigo-700"
                >
                  Vào lớp học
                </button>

                {isTeacher && (
                  <div className="flex items-center gap-2">
                    <button
                      type="button"
                      onClick={() => {
                        setEditingClass(cls);
                        setEditName(cls.name);
                        setEditDesc(cls.description || '');
                      }}
                      className="text-xs text-slate-600 hover:text-indigo-600 font-medium"
                    >
                      Sửa
                    </button>
                    <span className="text-slate-300">|</span>
                    <button
                      type="button"
                      onClick={() => handleDeleteClass(cls)}
                      className="text-xs text-red-600 hover:text-red-800 font-medium"
                    >
                      Xóa
                    </button>
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Teacher Create Class Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-slate-900/50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full p-6 space-y-4">
            <h2 className="text-base font-bold text-slate-900">Tạo lớp học mới</h2>
            <form onSubmit={handleCreateClass} className="space-y-4">
              <div>
                <label className="block text-xs font-semibold text-slate-700 uppercase mb-1">
                  Tên lớp học *
                </label>
                <input
                  type="text"
                  required
                  placeholder="Ví dụ: Lập trình Go chuyên sâu - K18"
                  value={newClassName}
                  onChange={(e) => setNewClassName(e.target.value)}
                  className="w-full text-sm px-3 py-2 border border-slate-300 rounded focus:ring-1 focus:ring-indigo-500 focus:outline-none"
                />
              </div>
              <div>
                <label className="block text-xs font-semibold text-slate-700 uppercase mb-1">
                  Mô tả lớp học
                </label>
                <textarea
                  rows={3}
                  placeholder="Mô tả nội dung học phần, ca học hoặc yêu cầu..."
                  value={newClassDesc}
                  onChange={(e) => setNewClassDesc(e.target.value)}
                  className="w-full text-sm px-3 py-2 border border-slate-300 rounded focus:ring-1 focus:ring-indigo-500 focus:outline-none"
                />
              </div>
              <div className="flex justify-end gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className="btn-secondary text-xs px-3 py-2 border border-slate-300 rounded text-slate-700 hover:bg-slate-50"
                >
                  Hủy
                </button>
                <button
                  type="submit"
                  disabled={creating}
                  className="btn-primary text-xs px-4 py-2 bg-indigo-600 text-white font-semibold rounded hover:bg-indigo-700 disabled:opacity-50"
                >
                  {creating ? 'Đang tạo...' : 'Tạo lớp'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Teacher Edit Class Modal */}
      {editingClass && (
        <div className="fixed inset-0 bg-slate-900/50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full p-6 space-y-4">
            <h2 className="text-base font-bold text-slate-900">Chỉnh sửa lớp học</h2>
            <form onSubmit={handleEditClass} className="space-y-4">
              <div>
                <label className="block text-xs font-semibold text-slate-700 uppercase mb-1">
                  Tên lớp học *
                </label>
                <input
                  type="text"
                  required
                  value={editName}
                  onChange={(e) => setEditName(e.target.value)}
                  className="w-full text-sm px-3 py-2 border border-slate-300 rounded focus:ring-1 focus:ring-indigo-500 focus:outline-none"
                />
              </div>
              <div>
                <label className="block text-xs font-semibold text-slate-700 uppercase mb-1">
                  Mô tả lớp học
                </label>
                <textarea
                  rows={3}
                  value={editDesc}
                  onChange={(e) => setEditDesc(e.target.value)}
                  className="w-full text-sm px-3 py-2 border border-slate-300 rounded focus:ring-1 focus:ring-indigo-500 focus:outline-none"
                />
              </div>
              <div className="flex justify-end gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setEditingClass(null)}
                  className="btn-secondary text-xs px-3 py-2 border border-slate-300 rounded text-slate-700 hover:bg-slate-50"
                >
                  Hủy
                </button>
                <button
                  type="submit"
                  disabled={updating}
                  className="btn-primary text-xs px-4 py-2 bg-indigo-600 text-white font-semibold rounded hover:bg-indigo-700 disabled:opacity-50"
                >
                  {updating ? 'Đang lưu...' : 'Lưu thay đổi'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Student Join Class Modal */}
      {showJoinModal && (
        <div className="fixed inset-0 bg-slate-900/50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-sm w-full p-6 space-y-4">
            <h2 className="text-base font-bold text-slate-900">Tham gia lớp học</h2>
            <p className="text-xs text-slate-500">
              Nhập mã gồm 6 ký tự được giảng viên phụ trách cung cấp để ghi danh vào lớp.
            </p>
            <form onSubmit={handleJoinClass} className="space-y-4">
              <div>
                <label className="block text-xs font-semibold text-slate-700 uppercase mb-1">
                  Mã lớp học (6 ký tự) *
                </label>
                <input
                  type="text"
                  required
                  maxLength={6}
                  placeholder="Ví dụ: K8M2P4"
                  value={joinCode}
                  onChange={(e) => setJoinCode(e.target.value.toUpperCase())}
                  className="w-full text-center tracking-widest font-mono text-lg uppercase px-3 py-2 border border-slate-300 rounded focus:ring-1 focus:ring-indigo-500 focus:outline-none"
                />
              </div>
              <div className="flex justify-end gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setShowJoinModal(false)}
                  className="btn-secondary text-xs px-3 py-2 border border-slate-300 rounded text-slate-700 hover:bg-slate-50"
                >
                  Hủy
                </button>
                <button
                  type="submit"
                  disabled={joining}
                  className="btn-primary text-xs px-4 py-2 bg-indigo-600 text-white font-semibold rounded hover:bg-indigo-700 disabled:opacity-50"
                >
                  {joining ? 'Đang tham gia...' : 'Tham gia'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
