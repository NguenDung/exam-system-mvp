import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { teacherAPI } from '../../api/endpoints';
import Spinner from '../../components/Spinner';

function StatusBadge({ status }) {
  return (
    <span
      className={`px-2.5 py-0.5 rounded text-xs font-semibold ${
        status === 'published' ? 'bg-emerald-100 text-emerald-800' : 'bg-amber-100 text-amber-800'
      }`}
    >
      {status === 'published' ? 'Đã xuất bản' : 'Bản nháp'}
    </span>
  );
}

function CreateExamModal({ onClose, onCreated }) {
  const [form, setForm] = useState({
    title: '',
    description: '',
    duration: 30,
    start_time: '',
    end_time: '',
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleChange = (e) => setForm((f) => ({ ...f, [e.target.name]: e.target.value }));

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      const payload = {
        ...form,
        duration: parseInt(form.duration, 10),
        start_time: new Date(form.start_time).toISOString(),
        end_time: new Date(form.end_time).toISOString(),
      };
      const { data } = await teacherAPI.createExam(payload);
      onCreated(data.exam);
    } catch (err) {
      setError(err.response?.data?.error || 'Tạo đề thi không thành công');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg shadow-xl w-full max-w-lg border border-slate-200">
        <div className="flex items-center justify-between p-5 border-b border-slate-200">
          <h3 className="text-base font-bold text-slate-900">Tạo đề thi mới</h3>
          <button onClick={onClose} className="text-slate-400 hover:text-slate-600 text-xl leading-none">&times;</button>
        </div>
        <form onSubmit={handleSubmit} className="p-5 space-y-4">
          <div>
            <label className="label">Tên đề thi / môn thi *</label>
            <input
              name="title"
              value={form.title}
              onChange={handleChange}
              className="input"
              placeholder="Ví dụ: Kiểm tra giữa kỳ Cơ sở dữ liệu"
              required
            />
          </div>
          <div>
            <label className="label">Mô tả kỳ thi</label>
            <textarea
              name="description"
              rows={2}
              value={form.description}
              onChange={handleChange}
              className="input resize-none"
              placeholder="Ví dụ: Đề thi 40 câu trắc nghiệm kiến thức SQL, chuẩn hóa dữ liệu..."
            />
          </div>
          <div>
            <label className="label">Thời lượng làm bài (phút) *</label>
            <input
              name="duration"
              type="number"
              min="1"
              max="300"
              value={form.duration}
              onChange={handleChange}
              className="input"
              required
            />
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div>
              <label className="label">Thời gian bắt đầu *</label>
              <input
                name="start_time"
                type="datetime-local"
                value={form.start_time}
                onChange={handleChange}
                className="input"
                required
              />
            </div>
            <div>
              <label className="label">Thời gian kết thúc *</label>
              <input
                name="end_time"
                type="datetime-local"
                value={form.end_time}
                onChange={handleChange}
                className="input"
                required
              />
            </div>
          </div>
          {error && <p className="text-red-600 text-sm">{error}</p>}
          <div className="flex justify-end gap-3 pt-2">
            <button type="button" onClick={onClose} className="btn-secondary">Hủy bỏ</button>
            <button type="submit" disabled={loading} className="btn-primary">
              {loading ? <Spinner size="sm" /> : null} Tạo đề thi
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

function EditExamModal({ exam, onClose, onUpdated }) {
  const [form, setForm] = useState({
    title: exam.title || '',
    description: exam.description || '',
    duration: exam.duration || 30,
    start_time: exam.start_time ? new Date(exam.start_time).toISOString().slice(0, 16) : '',
    end_time: exam.end_time ? new Date(exam.end_time).toISOString().slice(0, 16) : '',
    is_published: exam.status === 'published',
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleChange = (e) => {
    const { name, value, type, checked } = e.target;
    setForm((f) => ({ ...f, [name]: type === 'checkbox' ? checked : value }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      const payload = {
        title: form.title,
        description: form.description,
        duration: parseInt(form.duration, 10),
        start_time: form.start_time ? new Date(form.start_time).toISOString() : exam.start_time,
        end_time: form.end_time ? new Date(form.end_time).toISOString() : exam.end_time,
        is_published: form.is_published,
      };
      const { data } = await teacherAPI.updateExam(exam.id, payload);
      onUpdated(data.exam);
    } catch (err) {
      setError(err.response?.data?.error || 'Cập nhật đề thi không thành công');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg shadow-xl w-full max-w-lg border border-slate-200">
        <div className="flex items-center justify-between p-5 border-b border-slate-200">
          <h3 className="text-base font-bold text-slate-900">Chỉnh sửa đề thi #{exam.id}</h3>
          <button onClick={onClose} className="text-slate-400 hover:text-slate-600 text-xl leading-none">&times;</button>
        </div>
        <form onSubmit={handleSubmit} className="p-5 space-y-4">
          <div>
            <label className="label">Tên đề thi / môn thi *</label>
            <input
              name="title"
              value={form.title}
              onChange={handleChange}
              className="input"
              required
            />
          </div>
          <div>
            <label className="label">Mô tả kỳ thi</label>
            <textarea
              name="description"
              rows={2}
              value={form.description}
              onChange={handleChange}
              className="input resize-none"
            />
          </div>
          <div>
            <label className="label">Thời lượng làm bài (phút) *</label>
            <input
              name="duration"
              type="number"
              min="1"
              max="300"
              value={form.duration}
              onChange={handleChange}
              className="input"
              required
            />
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div>
              <label className="label">Thời gian bắt đầu *</label>
              <input
                name="start_time"
                type="datetime-local"
                value={form.start_time}
                onChange={handleChange}
                className="input"
                required
              />
            </div>
            <div>
              <label className="label">Thời gian kết thúc *</label>
              <input
                name="end_time"
                type="datetime-local"
                value={form.end_time}
                onChange={handleChange}
                className="input"
                required
              />
            </div>
          </div>
          <div className="flex items-center gap-2 pt-1">
            <input
              id="is_published"
              name="is_published"
              type="checkbox"
              checked={form.is_published}
              onChange={handleChange}
              className="h-4 w-4 rounded border-slate-300 text-slate-900 focus:ring-slate-900"
            />
            <label htmlFor="is_published" className="text-sm font-medium text-slate-700">
              Xuất bản đề thi (Cho phép sinh viên bắt đầu thi)
            </label>
          </div>
          {error && <p className="text-red-600 text-sm">{error}</p>}
          <div className="flex justify-end gap-3 pt-2">
            <button type="button" onClick={onClose} className="btn-secondary">Hủy bỏ</button>
            <button type="submit" disabled={loading} className="btn-primary">
              {loading ? <Spinner size="sm" /> : null} Lưu thay đổi
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

export default function TeacherExamsPage() {
  const [exams, setExams] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingExam, setEditingExam] = useState(null);
  const [deletingExam, setDeletingExam] = useState(null);
  const [deleteLoading, setDeleteLoading] = useState(false);
  const [deleteError, setDeleteError] = useState('');
  const navigate = useNavigate();

  const fetchExams = async () => {
    try {
      const { data } = await teacherAPI.listExams();
      setExams(data.exams || []);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchExams();
  }, []);

  const handleCreated = (exam) => {
    setShowCreateModal(false);
    navigate(`/teacher/exams/${exam.id}`);
  };

  const handleUpdated = (updatedExam) => {
    setEditingExam(null);
    setExams((prev) => prev.map((e) => (e.id === updatedExam.id ? updatedExam : e)));
  };

  const handleDelete = async () => {
    if (!deletingExam) return;
    setDeleteLoading(true);
    setDeleteError('');
    try {
      await teacherAPI.deleteExam(deletingExam.id);
      setDeletingExam(null);
      setExams((prev) => prev.filter((e) => e.id !== deletingExam.id));
    } catch (err) {
      setDeleteError(err.response?.data?.error || 'Lỗi khi xóa bài thi');
    } finally {
      setDeleteLoading(false);
    }
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-xl font-bold text-slate-900">Danh mục đề thi</h1>
          <p className="text-slate-500 text-sm mt-0.5">Quản trị danh sách đề thi trắc nghiệm đã tạo</p>
        </div>
        <button onClick={() => setShowCreateModal(true)} className="btn-primary">
          Tạo đề thi mới
        </button>
      </div>

      {loading ? (
        <div className="flex justify-center py-20"><Spinner size="lg" /></div>
      ) : exams.length === 0 ? (
        <div className="card text-center py-16 border border-slate-200">
          <p className="text-slate-600 font-medium">Chưa có đề thi nào trong danh sách</p>
          <p className="text-slate-400 text-sm mt-1 mb-5">Bấm nút bên dưới để khởi tạo đề thi đầu tiên</p>
          <button onClick={() => setShowCreateModal(true)} className="btn-primary mx-auto">Tạo đề thi</button>
        </div>
      ) : (
        <div className="grid gap-4">
          {exams.map((exam) => (
            <div
              key={exam.id}
              className="bg-white rounded-lg border border-slate-200 p-5 hover:border-slate-300 transition-colors"
            >
              <div className="flex items-start justify-between gap-4 flex-wrap">
                <div
                  className="flex-1 min-w-0 cursor-pointer"
                  onClick={() => navigate(`/teacher/exams/${exam.id}`)}
                >
                  <div className="flex items-center gap-3 flex-wrap">
                    <h3 className="font-semibold text-slate-900 text-base hover:text-indigo-600 transition-colors truncate">
                      {exam.title}
                    </h3>
                    <StatusBadge status={exam.status} />
                  </div>
                  {exam.description && (
                    <p className="text-slate-600 text-sm mt-1 line-clamp-2">{exam.description}</p>
                  )}
                  <div className="flex items-center gap-5 mt-2 text-xs text-slate-500 flex-wrap">
                    <span>
                      Mã kỳ thi: <code className="bg-slate-100 px-2 py-0.5 rounded font-mono font-bold text-slate-800">{exam.exam_code}</code>
                    </span>
                    <span>Thời lượng: <strong>{exam.duration} phút</strong></span>
                    <span>Bắt đầu: {new Date(exam.start_time).toLocaleString('vi-VN')}</span>
                    <span>Kết thúc: {new Date(exam.end_time).toLocaleString('vi-VN')}</span>
                  </div>
                </div>
                <div className="flex items-center gap-2 flex-shrink-0">
                  <button
                    onClick={() => navigate(`/teacher/exams/${exam.id}`)}
                    className="btn-secondary btn-sm"
                  >
                    Chi tiết & Soạn đề
                  </button>
                  <button
                    onClick={() => navigate(`/teacher/exams/${exam.id}/results`)}
                    className="btn-secondary btn-sm"
                  >
                    Bảng điểm
                  </button>
                  <button
                    onClick={() => setEditingExam(exam)}
                    className="btn-secondary btn-sm text-slate-700 hover:bg-slate-100"
                  >
                    Sửa
                  </button>
                  <button
                    onClick={() => {
                      setDeletingExam(exam);
                      setDeleteError('');
                    }}
                    className="btn-secondary btn-sm text-red-600 hover:bg-red-50 hover:border-red-200"
                  >
                    Xóa
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {showCreateModal && (
        <CreateExamModal onClose={() => setShowCreateModal(false)} onCreated={handleCreated} />
      )}

      {editingExam && (
        <EditExamModal
          exam={editingExam}
          onClose={() => setEditingExam(null)}
          onUpdated={handleUpdated}
        />
      )}

      {deletingExam && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full p-6 space-y-4">
            <h3 className="text-lg font-bold text-red-600">Xác nhận xóa đề thi</h3>
            <p className="text-sm text-slate-600">
              Bạn có chắc chắn muốn xóa vĩnh viễn đề thi{' '}
              <span className="font-semibold text-slate-900">{deletingExam.title}</span> (Mã: {deletingExam.exam_code})?
            </p>
            <p className="text-xs text-red-600 bg-red-50 p-2.5 rounded border border-red-200">
              Cảnh báo: Hành động này sẽ xóa toàn bộ câu hỏi, lịch sử làm bài, điểm số và khiếu nại của sinh viên thuộc đề thi này trong DB.
            </p>
            {deleteError && (
              <div className="rounded bg-red-50 p-3 text-xs text-red-700 border border-red-200">
                {deleteError}
              </div>
            )}
            <div className="flex justify-end gap-2 pt-2 border-t border-slate-100">
              <button
                type="button"
                onClick={() => setDeletingExam(null)}
                disabled={deleteLoading}
                className="px-4 py-2 border border-slate-300 rounded-md text-sm text-slate-700 hover:bg-slate-50 transition-colors"
              >
                Hủy bỏ
              </button>
              <button
                type="button"
                onClick={handleDelete}
                disabled={deleteLoading}
                className="px-4 py-2 bg-red-600 text-white rounded-md text-sm font-medium hover:bg-red-700 transition-colors disabled:opacity-50"
              >
                {deleteLoading ? 'Đang xóa...' : 'Xóa đề thi'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
