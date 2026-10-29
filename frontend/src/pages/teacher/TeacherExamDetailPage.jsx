import { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { teacherAPI } from '../../api/endpoints';
import Spinner from '../../components/Spinner';
import ExamAppealsTab from './ExamAppealsTab';

const OPTION_LABELS = ['A', 'B', 'C', 'D'];

const EMPTY_QUESTION = {
  prompt: '',
  option_a: '',
  option_b: '',
  option_c: '',
  option_d: '',
  correct_option: 'A',
  points: 1,
};

function QuestionForm({ onSave, onCancel, saving }) {
  const [q, setQ] = useState(EMPTY_QUESTION);
  const set = (k) => (e) => setQ((p) => ({ ...p, [k]: e.target.value }));

  const handleSubmit = (e) => {
    e.preventDefault();
    onSave({ ...q, points: parseFloat(q.points) || 1 });
    setQ(EMPTY_QUESTION);
  };

  return (
    <form onSubmit={handleSubmit} className="bg-slate-50 border border-slate-300 rounded-lg p-5 space-y-4">
      <h4 className="font-semibold text-slate-900">Thêm câu hỏi mới</h4>
      <div>
        <label className="label">Nội dung câu hỏi *</label>
        <textarea
          value={q.prompt}
          onChange={set('prompt')}
          className="input resize-none"
          rows={3}
          placeholder="Nhập nội dung câu hỏi..."
          required
        />
      </div>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        {['a', 'b', 'c', 'd'].map((opt) => (
          <div key={opt}>
            <label className="label">Đáp án {opt.toUpperCase()} *</label>
            <input
              value={q[`option_${opt}`]}
              onChange={set(`option_${opt}`)}
              className="input"
              placeholder={`Nội dung đáp án ${opt.toUpperCase()}`}
              required
            />
          </div>
        ))}
      </div>
      <div className="flex gap-4 items-end flex-wrap">
        <div>
          <label className="label">Đáp án đúng *</label>
          <div className="flex gap-2">
            {OPTION_LABELS.map((opt) => (
              <label
                key={opt}
                className={`flex items-center gap-1.5 cursor-pointer px-3 py-1.5 rounded border text-sm font-semibold transition-all ${
                  q.correct_option === opt
                    ? 'border-indigo-600 bg-indigo-50 text-indigo-700'
                    : 'border-slate-300 bg-white text-slate-700 hover:border-slate-400'
                }`}
              >
                <input
                  type="radio"
                  name="correct_option"
                  value={opt}
                  checked={q.correct_option === opt}
                  onChange={set('correct_option')}
                  className="sr-only"
                />
                {opt}
              </label>
            ))}
          </div>
        </div>
        <div className="w-28">
          <label className="label">Điểm số</label>
          <input type="number" step="0.5" min="0.5" value={q.points} onChange={set('points')} className="input" />
        </div>
        <div className="flex gap-2 ml-auto">
          <button type="button" onClick={onCancel} className="btn-secondary">
            Hủy
          </button>
          <button type="submit" disabled={saving} className="btn-primary">
            {saving ? <Spinner size="sm" /> : null} Lưu câu hỏi
          </button>
        </div>
      </div>
    </form>
  );
}

function QuestionCard({ q, index, isDraft, onDelete }) {
  const [confirming, setConfirming] = useState(false);
  return (
    <div className="bg-white rounded-lg border border-slate-200 p-5">
      <div className="flex items-start gap-3">
        <span className="flex-shrink-0 w-7 h-7 rounded bg-slate-100 text-slate-700 font-bold text-xs flex items-center justify-center">
          {index + 1}
        </span>
        <div className="flex-1 min-w-0">
          <p className="font-medium text-slate-900 mb-3">{q.prompt}</p>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
            {OPTION_LABELS.map((opt) => (
              <div
                key={opt}
                className={`flex items-center gap-2 px-3 py-2 rounded border text-xs sm:text-sm ${
                  q.correct_option === opt
                    ? 'bg-emerald-50 border-emerald-300 text-emerald-900 font-semibold'
                    : 'bg-slate-50 border-slate-200 text-slate-700'
                }`}
              >
                <span
                  className={`w-5 h-5 rounded flex items-center justify-center text-xs font-bold flex-shrink-0 ${
                    q.correct_option === opt ? 'bg-emerald-600 text-white' : 'bg-slate-200 text-slate-700'
                  }`}
                >
                  {opt}
                </span>
                <span className="truncate">{q[`option_${opt.toLowerCase()}`]}</span>
                {q.correct_option === opt && <span className="ml-auto text-xs text-emerald-700 font-medium">Đáp án đúng</span>}
              </div>
            ))}
          </div>
          <div className="flex items-center gap-4 mt-3 text-xs text-slate-500">
            <span>Trọng số: <strong>{q.points} điểm</strong></span>
          </div>
        </div>
        {isDraft && (
          <div className="flex-shrink-0">
            {confirming ? (
              <div className="flex gap-2">
                <button onClick={() => onDelete(q.id)} className="btn-danger btn-sm">
                  Xác nhận xóa
                </button>
                <button onClick={() => setConfirming(false)} className="btn-secondary btn-sm">
                  Hủy
                </button>
              </div>
            ) : (
              <button
                onClick={() => setConfirming(true)}
                className="btn-secondary btn-sm text-red-600 hover:bg-red-50 hover:border-red-200"
              >
                Xóa
              </button>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

export default function TeacherExamDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [exam, setExam] = useState(null);
  const [questions, setQuestions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [saving, setSaving] = useState(false);
  const [publishing, setPublishing] = useState(false);
  const [toast, setToast] = useState(null);
  const [activeTab, setActiveTab] = useState('info');
  const [scoreData, setScoreData] = useState(null);
  const [scoreLoading, setScoreLoading] = useState(false);
  const [scoreError, setScoreError] = useState('');

  const showToast = useCallback((msg, type = 'success') => {
    setToast({ msg, type });
    setTimeout(() => setToast(null), 3000);
  }, []);

  const fetchData = useCallback(async () => {
    try {
      const { data } = await teacherAPI.getExam(id);
      setExam(data.exam);
      setQuestions(data.exam.questions || []);
    } catch {
      navigate('/teacher');
    } finally {
      setLoading(false);
    }
  }, [id, navigate]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  useEffect(() => {
    if (activeTab !== 'scores' || scoreData) return;
    let active = true;
    setScoreLoading(true);
    teacherAPI
      .getResults(id)
      .then(({ data }) => {
        if (active) setScoreData(data);
      })
      .catch((err) => {
        if (active) setScoreError(err.response?.data?.error || 'Không thể tải bảng điểm.');
      })
      .finally(() => {
        if (active) setScoreLoading(false);
      });
    return () => {
      active = false;
    };
  }, [activeTab, id, scoreData]);

  const handleAddQuestion = async (questionData) => {
    setSaving(true);
    try {
      const { data } = await teacherAPI.addQuestions(id, { questions: [questionData] });
      setQuestions((q) => [...q, ...(data.questions || [])]);
      setShowForm(false);
      showToast('Đã thêm câu hỏi thành công');
    } catch (err) {
      showToast(err.response?.data?.error || 'Thêm câu hỏi không thành công', 'error');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (qId) => {
    try {
      await teacherAPI.deleteQuestion(qId);
      setQuestions((qs) => qs.filter((q) => q.id !== qId));
      showToast('Đã xóa câu hỏi');
    } catch (err) {
      showToast(err.response?.data?.error || 'Xóa câu hỏi không thành công', 'error');
    }
  };

  const handlePublish = async () => {
    if (!window.confirm(`Xác nhận xuất bản đề thi "${exam.title}"? Sau khi xuất bản bạn sẽ không thể chỉnh sửa hoặc xóa câu hỏi.`)) return;
    setPublishing(true);
    try {
      await teacherAPI.publishExam(id);
      showToast('Đề thi đã được xuất bản');
      fetchData();
    } catch (err) {
      showToast(err.response?.data?.error || 'Xuất bản không thành công', 'error');
    } finally {
      setPublishing(false);
    }
  };

  if (loading) return <div className="flex justify-center py-20"><Spinner size="lg" /></div>;
  if (!exam) return null;

  const isDraft = exam.status === 'draft';
  const tabs = [
    { id: 'info', label: 'Thông tin chung' },
    { id: 'questions', label: 'Câu hỏi & Đáp án' },
    { id: 'scores', label: 'Bảng điểm' },
    { id: 'appeals', label: 'Quản lý khiếu nại' },
  ];

  return (
    <div>
      {/* Toast */}
      {toast && (
        <div
          className={`fixed top-4 right-4 z-50 px-4 py-2.5 rounded-md shadow-lg text-white text-sm font-medium transition-all ${
            toast.type === 'error' ? 'bg-red-700' : 'bg-emerald-700'
          }`}
        >
          {toast.msg}
        </div>
      )}

      {/* Header */}
      <div className="flex items-start gap-4 mb-6 flex-wrap">
        <button onClick={() => navigate('/teacher')} className="btn-secondary btn-sm mt-0.5">
          Quay lại danh mục
        </button>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-3 flex-wrap">
            <h1 className="text-xl font-bold text-slate-900 truncate">{exam.title}</h1>
            <span
              className={`px-2.5 py-0.5 rounded text-xs font-semibold ${
                isDraft ? 'bg-amber-100 text-amber-800' : 'bg-emerald-100 text-emerald-800'
              }`}
            >
              {isDraft ? 'Bản nháp' : 'Đã xuất bản'}
            </span>
          </div>
          <div className="flex items-center gap-5 mt-1 text-xs text-slate-500 flex-wrap">
            <span>Mã đề: <code className="bg-slate-100 px-2 py-0.5 rounded font-mono font-bold text-slate-800">{exam.exam_code}</code></span>
            <span>Thời lượng: <strong>{exam.duration} phút</strong></span>
            <span>Bắt đầu: {new Date(exam.start_time).toLocaleString('vi-VN')}</span>
            <span>Kết thúc: {new Date(exam.end_time).toLocaleString('vi-VN')}</span>
          </div>
        </div>
        <div className="flex gap-2 flex-wrap">
          <Link to={`/teacher/exams/${id}/results`} className="btn-secondary btn-sm">
            Bảng điểm chi tiết
          </Link>
          {isDraft && (
            <button onClick={handlePublish} disabled={publishing || questions.length === 0} className="btn-primary btn-sm">
              {publishing ? <Spinner size="sm" /> : null} Xuất bản đề thi
            </button>
          )}
        </div>
      </div>

      <nav className="mb-6 flex gap-2 overflow-x-auto border-b border-slate-200" aria-label="Mục chi tiết đề thi">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`whitespace-nowrap border-b-2 px-4 py-2.5 text-sm font-medium ${
              activeTab === tab.id
                ? 'border-indigo-600 text-indigo-700'
                : 'border-transparent text-slate-500 hover:text-slate-800'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </nav>

      {activeTab === 'info' && (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
          {[
            { label: 'Số lượng câu hỏi', value: questions.length },
            { label: 'Tổng điểm đề thi', value: questions.reduce((sum, question) => sum + (question.points || 1), 0).toFixed(1) },
            { label: 'Trạng thái phát hành', value: isDraft ? 'Bản nháp (Đang soạn)' : 'Đã xuất bản chính thức' },
          ].map((stat) => (
            <div key={stat.label} className="bg-white rounded-lg border border-slate-200 p-5 text-center">
              <div className="text-2xl font-bold text-slate-900">{stat.value}</div>
              <div className="mt-1 text-xs text-slate-500 uppercase tracking-wide">{stat.label}</div>
            </div>
          ))}
          <div className="bg-white rounded-lg border border-slate-200 p-5 sm:col-span-3">
            <h2 className="font-semibold text-slate-900 text-sm uppercase tracking-wide">Chi tiết thiết lập kỳ thi</h2>
            <dl className="mt-4 grid gap-4 text-sm sm:grid-cols-2">
              <div><dt className="text-slate-500">Mã kỳ thi</dt><dd className="font-mono font-semibold text-slate-900">{exam.exam_code}</dd></div>
              <div><dt className="text-slate-500">Thời lượng làm bài</dt><dd className="font-medium">{exam.duration} phút</dd></div>
              <div><dt className="text-slate-500">Thời điểm mở đề</dt><dd>{new Date(exam.start_time).toLocaleString('vi-VN')}</dd></div>
              <div><dt className="text-slate-500">Thời điểm đóng đề</dt><dd>{new Date(exam.end_time).toLocaleString('vi-VN')}</dd></div>
              {exam.published_at && (
                <div><dt className="text-slate-500">Thời điểm xuất bản</dt><dd>{new Date(exam.published_at).toLocaleString('vi-VN')}</dd></div>
              )}
            </dl>
          </div>
        </div>
      )}

      {activeTab === 'questions' && (
        <>
          <div className="mb-4 space-y-3">
            {questions.length === 0 && !showForm && (
              <div className="bg-white rounded-lg border border-dashed border-slate-300 py-12 text-center">
                <p className="text-slate-500">Đề thi chưa có câu hỏi nào. {isDraft && 'Bấm nút thêm câu hỏi để bắt đầu soạn đề.'}</p>
              </div>
            )}
            {questions.map((question, index) => (
              <QuestionCard key={question.id} q={question} index={index} isDraft={isDraft} onDelete={handleDelete} />
            ))}
          </div>
          {isDraft &&
            (showForm ? (
              <QuestionForm onSave={handleAddQuestion} onCancel={() => setShowForm(false)} saving={saving} />
            ) : (
              <button
                onClick={() => setShowForm(true)}
                className="w-full rounded-lg border border-dashed border-indigo-400 py-3.5 font-medium text-indigo-700 bg-white hover:bg-indigo-50 transition-colors"
              >
                Thêm câu hỏi mới
              </button>
            ))}
          {isDraft && questions.length === 0 && (
            <p className="mt-4 rounded-md bg-amber-50 border border-amber-200 p-3 text-center text-sm text-amber-800">
              Cần có tối thiểu 1 câu hỏi trước khi xuất bản đề thi.
            </p>
          )}
        </>
      )}

      {activeTab === 'scores' && (
        <section className="bg-white rounded-lg border border-slate-200 p-5">
          <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
            <h2 className="text-base font-bold text-slate-900">Bảng điểm thi sinh viên</h2>
            <Link to={`/teacher/exams/${id}/results`} className="btn-secondary btn-sm">Xem thống kê phân phối</Link>
          </div>
          {scoreLoading ? (
            <div className="flex justify-center py-10"><Spinner size="lg" /></div>
          ) : scoreError ? (
            <p role="alert" className="text-sm text-red-600">{scoreError}</p>
          ) : !scoreData?.results?.length ? (
            <p className="py-8 text-center text-slate-500">Chưa có sinh viên nào hoàn thành bài thi này.</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full min-w-[500px] text-sm">
                <thead>
                  <tr className="border-b border-slate-200 text-left text-slate-500">
                    <th className="py-2.5">Họ và tên</th>
                    <th className="py-2.5">Điểm số</th>
                    <th className="py-2.5">Trạng thái</th>
                    <th className="py-2.5">Thời gian nộp</th>
                  </tr>
                </thead>
                <tbody>
                  {scoreData.results.map((result) => (
                    <tr key={result.attempt_id} className="border-b border-slate-100 last:border-0">
                      <td className="py-3">
                        <p className="font-medium text-slate-900">{result.full_name}</p>
                        <p className="text-xs text-slate-500">{result.email}</p>
                      </td>
                      <td className="py-3 font-semibold text-slate-900">
                        {result.status === 'completed' ? Number(result.score).toFixed(2) : '—'}
                      </td>
                      <td className="py-3">
                        <span className={`px-2 py-0.5 rounded text-xs font-semibold ${
                          result.status === 'completed' ? 'bg-blue-100 text-blue-800' : 'bg-amber-100 text-amber-800'
                        }`}>
                          {result.status === 'completed' ? 'Đã nộp' : 'Đang làm'}
                        </span>
                      </td>
                      <td className="py-3 text-slate-500 text-xs">
                        {result.submitted_at ? new Date(result.submitted_at).toLocaleString('vi-VN') : '—'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>
      )}

      {activeTab === 'appeals' && <ExamAppealsTab examId={id} />}
    </div>
  );
}
