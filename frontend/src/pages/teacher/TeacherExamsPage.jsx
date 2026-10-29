import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { teacherAPI } from '../../api/endpoints';
import Spinner from '../../components/Spinner';

function StatusBadge({ status }) {
  return (
    <span className={`px-2.5 py-0.5 rounded text-xs font-semibold ${
      status === 'published' ? 'bg-emerald-100 text-emerald-800' : 'bg-amber-100 text-amber-800'
    }`}>
      {status === 'published' ? 'Đã xuất bản' : 'Bản nháp'}
    </span>
  );
}

function CreateExamModal({ onClose, onCreated }) {
  const [form, setForm] = useState({
    title: '',
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
        duration: parseInt(form.duration),
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
            <label className="label">Tên đề thi / môn thi</label>
            <input name="title" value={form.title} onChange={handleChange} className="input" placeholder="Ví dụ: Kiểm tra giữa kỳ Lập trình Web" required />
          </div>
          <div>
            <label className="label">Thời lượng làm bài (phút)</label>
            <input name="duration" type="number" min="1" max="300" value={form.duration} onChange={handleChange} className="input" required />
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div>
              <label className="label">Thời gian bắt đầu</label>
              <input name="start_time" type="datetime-local" value={form.start_time} onChange={handleChange} className="input" required />
            </div>
            <div>
              <label className="label">Thời gian kết thúc</label>
              <input name="end_time" type="datetime-local" value={form.end_time} onChange={handleChange} className="input" required />
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

export default function TeacherExamsPage() {
  const [exams, setExams] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
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
    setShowModal(false);
    navigate(`/teacher/exams/${exam.id}`);
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-xl font-bold text-slate-900">Danh mục đề thi</h1>
          <p className="text-slate-500 text-sm mt-0.5">Quản trị danh sách đề thi trắc nghiệm đã tạo</p>
        </div>
        <button onClick={() => setShowModal(true)} className="btn-primary">
          Tạo đề thi mới
        </button>
      </div>

      {loading ? (
        <div className="flex justify-center py-20"><Spinner size="lg" /></div>
      ) : exams.length === 0 ? (
        <div className="card text-center py-16 border border-slate-200">
          <p className="text-slate-600 font-medium">Chưa có đề thi nào trong danh sách</p>
          <p className="text-slate-400 text-sm mt-1 mb-5">Bấm nút bên dưới để khởi tạo đề thi đầu tiên</p>
          <button onClick={() => setShowModal(true)} className="btn-primary mx-auto">Tạo đề thi</button>
        </div>
      ) : (
        <div className="grid gap-4">
          {exams.map((exam) => (
            <div
              key={exam.id}
              className="bg-white rounded-lg border border-slate-200 p-5 hover:border-slate-300 transition-colors cursor-pointer"
              onClick={() => navigate(`/teacher/exams/${exam.id}`)}
            >
              <div className="flex items-center justify-between gap-4 flex-wrap">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-3 flex-wrap">
                    <h3 className="font-semibold text-slate-900 text-base hover:text-indigo-600 transition-colors truncate">
                      {exam.title}
                    </h3>
                    <StatusBadge status={exam.status} />
                  </div>
                  <div className="flex items-center gap-5 mt-2 text-xs text-slate-500 flex-wrap">
                    <span>
                      Mã kỳ thi: <code className="bg-slate-100 px-2 py-0.5 rounded font-mono font-bold text-slate-800">{exam.exam_code}</code>
                    </span>
                    <span>Thời lượng: <strong>{exam.duration} phút</strong></span>
                    <span>Bắt đầu: {new Date(exam.start_time).toLocaleString('vi-VN')}</span>
                    <span>Kết thúc: {new Date(exam.end_time).toLocaleString('vi-VN')}</span>
                  </div>
                </div>
                <div className="flex gap-2 flex-shrink-0">
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      navigate(`/teacher/exams/${exam.id}`);
                    }}
                    className="btn-secondary btn-sm"
                  >
                    Soạn đề
                  </button>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      navigate(`/teacher/exams/${exam.id}/results`);
                    }}
                    className="btn-secondary btn-sm"
                  >
                    Bảng điểm
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {showModal && <CreateExamModal onClose={() => setShowModal(false)} onCreated={handleCreated} />}
    </div>
  );
}
