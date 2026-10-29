import { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { teacherAPI } from '../../api/endpoints';
import Spinner from '../../components/Spinner';
import ExamAppealsTab from './ExamAppealsTab';

const OPTION_LABELS = ['A', 'B', 'C', 'D'];

const EMPTY_QUESTION = {
  prompt: '', option_a: '', option_b: '', option_c: '', option_d: '',
  correct_option: 'A', points: 1,
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
    <form onSubmit={handleSubmit} className="bg-indigo-50 border border-indigo-200 rounded-2xl p-5 space-y-4">
      <h4 className="font-semibold text-indigo-800">➕ Thêm câu hỏi mới</h4>
      <div>
        <label className="label">Câu hỏi *</label>
        <textarea value={q.prompt} onChange={set('prompt')} className="input resize-none" rows={3}
          placeholder="Nhập nội dung câu hỏi..." required />
      </div>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        {['a', 'b', 'c', 'd'].map((opt) => (
          <div key={opt}>
            <label className="label">Đáp án {opt.toUpperCase()} *</label>
            <input value={q[`option_${opt}`]} onChange={set(`option_${opt}`)}
              className="input" placeholder={`Nội dung đáp án ${opt.toUpperCase()}`} required />
          </div>
        ))}
      </div>
      <div className="flex gap-4 items-end flex-wrap">
        <div>
          <label className="label">Đáp án đúng *</label>
          <div className="flex gap-2">
            {OPTION_LABELS.map((opt) => (
              <label key={opt} className={`flex items-center gap-1.5 cursor-pointer px-3 py-2 rounded-lg border-2 font-semibold text-sm transition-all ${
                q.correct_option === opt ? 'border-emerald-500 bg-emerald-50 text-emerald-700 ring-2 ring-emerald-200' : 'border-gray-200 bg-white text-gray-600 hover:border-gray-300'
              }`}>
                <input type="radio" name="correct_option" value={opt}
                  checked={q.correct_option === opt} onChange={set('correct_option')} className="sr-only" />
                {opt}
              </label>
            ))}
          </div>
        </div>
        <div className="w-28">
          <label className="label">Điểm</label>
          <input type="number" step="0.5" min="0.5" value={q.points} onChange={set('points')} className="input" />
        </div>
        <div className="flex gap-2 ml-auto">
          <button type="button" onClick={onCancel} className="btn-secondary">Hủy</button>
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
    <div className="card hover:shadow-md transition-shadow">
      <div className="flex items-start gap-3">
        <span className="flex-shrink-0 w-8 h-8 rounded-full bg-indigo-100 text-indigo-700 font-bold text-sm flex items-center justify-center">
          {index + 1}
        </span>
        <div className="flex-1 min-w-0">
          <p className="font-medium text-gray-900 mb-3">{q.prompt}</p>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
            {OPTION_LABELS.map((opt) => (
              <div key={opt} className={`flex items-center gap-2 px-3 py-2 rounded-lg border text-sm ${
                q.correct_option === opt
                  ? 'bg-emerald-50 border-emerald-300 text-emerald-800 font-semibold'
                  : 'bg-gray-50 border-gray-200 text-gray-600'
              }`}>
                <span className={`w-5 h-5 rounded-full flex items-center justify-center text-xs font-bold flex-shrink-0 ${
                  q.correct_option === opt ? 'bg-emerald-500 text-white' : 'bg-gray-200 text-gray-600'
                }`}>{opt}</span>
                {q[`option_${opt.toLowerCase()}`]}
                {q.correct_option === opt && <span className="ml-auto text-emerald-600">✓</span>}
              </div>
            ))}
          </div>
          <div className="flex items-center gap-4 mt-3 text-xs text-gray-500">
            <span>💯 {q.points} điểm</span>
          </div>
        </div>
        {isDraft && (
          <div className="flex-shrink-0">
            {confirming ? (
              <div className="flex gap-2">
                <button onClick={() => onDelete(q.id)} className="btn-danger btn-sm">Xóa</button>
                <button onClick={() => setConfirming(false)} className="btn-secondary btn-sm">Hủy</button>
              </div>
            ) : (
              <button onClick={() => setConfirming(true)} className="btn-sm btn bg-red-50 text-red-600 border border-red-200 hover:bg-red-100">🗑</button>
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

  useEffect(() => { fetchData(); }, [fetchData]);

  useEffect(() => {
    if (activeTab !== 'scores' || scoreData) return;
    let active = true;
    setScoreLoading(true);
    teacherAPI.getResults(id)
      .then(({ data }) => { if (active) setScoreData(data); })
      .catch((err) => { if (active) setScoreError(err.response?.data?.error || 'Không thể tải bảng điểm.'); })
      .finally(() => { if (active) setScoreLoading(false); });
    return () => { active = false; };
  }, [activeTab, id, scoreData]);

  const handleAddQuestion = async (questionData) => {
    setSaving(true);
    try {
      const { data } = await teacherAPI.addQuestions(id, { questions: [questionData] });
      setQuestions((q) => [...q, ...(data.questions || [])]);
      setShowForm(false);
      showToast('✅ Đã thêm câu hỏi!');
    } catch (err) {
      showToast(err.response?.data?.error || 'Thêm câu hỏi thất bại', 'error');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (qId) => {
    try {
      await teacherAPI.deleteQuestion(qId);
      setQuestions((qs) => qs.filter((q) => q.id !== qId));
      showToast('🗑 Đã xóa câu hỏi');
    } catch (err) {
      showToast(err.response?.data?.error || 'Xóa thất bại', 'error');
    }
  };

  const handlePublish = async () => {
    if (!window.confirm(`Xuất bản đề thi "${exam.title}"? Sau khi xuất bản bạn không thể thêm/xóa câu hỏi nữa.`)) return;
    setPublishing(true);
    try {
      await teacherAPI.publishExam(id);
      showToast('🚀 Đề thi đã được xuất bản!');
      fetchData();
    } catch (err) {
      showToast(err.response?.data?.error || 'Xuất bản thất bại', 'error');
    } finally {
      setPublishing(false);
    }
  };

  if (loading) return <div className="flex justify-center py-20"><Spinner size="lg" /></div>;
  if (!exam) return null;

  const isDraft = exam.status === 'draft';
  const tabs = [
    { id: 'info', label: 'Thông tin chung' },
    { id: 'questions', label: 'Câu hỏi - đáp án' },
    { id: 'scores', label: 'Điểm' },
    { id: 'appeals', label: 'Khiếu nại' },
  ];

  return (
    <div>
      {/* Toast */}
      {toast && (
        <div className={`fixed top-4 right-4 z-50 px-5 py-3 rounded-xl shadow-xl text-white text-sm font-medium transition-all ${
          toast.type === 'error' ? 'bg-red-600' : 'bg-emerald-600'
        }`}>
          {toast.msg}
        </div>
      )}

      {/* Header */}
      <div className="flex items-start gap-4 mb-6 flex-wrap">
        <button onClick={() => navigate('/teacher')} className="btn-secondary btn-sm mt-1">← Quay lại</button>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-3 flex-wrap">
            <h1 className="text-2xl font-bold text-gray-900 truncate">{exam.title}</h1>
            <span className={isDraft ? 'badge-draft' : 'badge-published'}>
              {isDraft ? '✏️ Nháp' : '✅ Đã xuất bản'}
            </span>
          </div>
          <div className="flex items-center gap-4 mt-1 text-sm text-gray-500 flex-wrap">
            <span>🔑 Mã: <code className="bg-gray-100 px-2 py-0.5 rounded font-mono font-bold text-indigo-700 text-base">{exam.exam_code}</code></span>
            <span>⏱ {exam.duration} phút</span>
            <span>📅 {new Date(exam.start_time).toLocaleString('vi-VN')} → {new Date(exam.end_time).toLocaleString('vi-VN')}</span>
          </div>
        </div>
        <div className="flex gap-2 flex-wrap">
          <Link to={`/teacher/exams/${id}/results`} className="btn-sm btn bg-violet-600 text-white hover:bg-violet-700">
            📊 Xem kết quả
          </Link>
          {isDraft && (
            <button onClick={handlePublish} disabled={publishing || questions.length === 0} className="btn-success btn-sm">
              {publishing ? <Spinner size="sm" /> : '🚀'} Xuất bản đề thi
            </button>
          )}
        </div>
      </div>

      <nav className="mb-6 flex gap-2 overflow-x-auto border-b" aria-label="Các mục của đề thi">
        {tabs.map((tab) => (
          <button key={tab.id} onClick={() => setActiveTab(tab.id)}
            className={`whitespace-nowrap border-b-2 px-4 py-3 text-sm font-medium ${
              activeTab === tab.id ? 'border-indigo-600 text-indigo-700' : 'border-transparent text-gray-500 hover:text-gray-800'
            }`}>
            {tab.label}
          </button>
        ))}
      </nav>

      {activeTab === 'info' && (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
          {[
            { label: 'Số câu hỏi', value: questions.length, icon: '❓' },
            { label: 'Tổng điểm', value: questions.reduce((sum, question) => sum + (question.points || 1), 0).toFixed(1), icon: '💯' },
            { label: 'Trạng thái', value: isDraft ? 'Bản nháp' : 'Đã xuất bản', icon: isDraft ? '✏️' : '✅' },
          ].map((stat) => (
            <div key={stat.label} className="card text-center py-6">
              <div className="mb-2 text-3xl">{stat.icon}</div>
              <div className="text-xl font-bold text-gray-900">{stat.value}</div>
              <div className="mt-1 text-xs text-gray-500">{stat.label}</div>
            </div>
          ))}
          <div className="card sm:col-span-3">
            <h2 className="font-semibold text-gray-900">Thông tin đề thi</h2>
            <dl className="mt-4 grid gap-4 text-sm sm:grid-cols-2">
              <div><dt className="text-gray-500">Mã đề</dt><dd className="font-mono font-semibold">{exam.exam_code}</dd></div>
              <div><dt className="text-gray-500">Thời lượng</dt><dd>{exam.duration} phút</dd></div>
              <div><dt className="text-gray-500">Bắt đầu</dt><dd>{new Date(exam.start_time).toLocaleString('vi-VN')}</dd></div>
              <div><dt className="text-gray-500">Kết thúc</dt><dd>{new Date(exam.end_time).toLocaleString('vi-VN')}</dd></div>
              {exam.published_at && <div><dt className="text-gray-500">Xuất bản lúc</dt><dd>{new Date(exam.published_at).toLocaleString('vi-VN')}</dd></div>}
            </dl>
          </div>
        </div>
      )}

      {activeTab === 'questions' && (
        <>
          <div className="mb-4 space-y-3">
            {questions.length === 0 && !showForm && (
              <div className="card border-2 border-dashed border-gray-200 py-12 text-center">
                <div className="mb-3 text-4xl">📭</div>
                <p className="text-gray-500">Chưa có câu hỏi nào. {isDraft && 'Thêm câu hỏi đầu tiên!'}</p>
              </div>
            )}
            {questions.map((question, index) => (
              <QuestionCard key={question.id} q={question} index={index} isDraft={isDraft} onDelete={handleDelete} />
            ))}
          </div>
          {isDraft && (showForm
            ? <QuestionForm onSave={handleAddQuestion} onCancel={() => setShowForm(false)} saving={saving} />
            : <button onClick={() => setShowForm(true)}
              className="w-full rounded-2xl border-2 border-dashed border-indigo-300 py-4 font-medium text-indigo-600 transition-all hover:border-indigo-500 hover:bg-indigo-50">
              ➕ Thêm câu hỏi mới
            </button>)}
          {isDraft && questions.length === 0 && <p className="mt-4 rounded-lg bg-amber-50 p-3 text-center text-sm text-amber-600">⚠️ Cần có ít nhất 1 câu hỏi để xuất bản đề thi.</p>}
        </>
      )}

      {activeTab === 'scores' && (
        <section className="card">
          <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
            <h2 className="text-lg font-bold text-gray-900">Bảng điểm</h2>
            <Link to={`/teacher/exams/${id}/results`} className="btn-primary btn-sm">Xem thống kê đầy đủ</Link>
          </div>
          {scoreLoading ? <div className="flex justify-center py-10"><Spinner size="lg" /></div> :
            scoreError ? <p role="alert" className="text-sm text-red-600">{scoreError}</p> :
              !scoreData?.results?.length ? <p className="py-8 text-center text-gray-500">Chưa có học sinh nào làm bài thi này.</p> : (
                <div className="overflow-x-auto">
                  <table className="w-full min-w-[500px] text-sm">
                    <thead><tr className="border-b text-left text-gray-500"><th className="py-3">Student</th><th className="py-3">Điểm</th><th className="py-3">Trạng thái</th><th className="py-3">Nộp lúc</th></tr></thead>
                    <tbody>{scoreData.results.map((result) => (
                      <tr key={result.attempt_id} className="border-b last:border-0">
                        <td className="py-3"><p className="font-medium text-gray-900">{result.full_name}</p><p className="text-xs text-gray-500">{result.email}</p></td>
                        <td className="py-3 font-semibold">{result.status === 'completed' ? Number(result.score).toFixed(2) : '—'}</td>
                        <td className="py-3">{result.status === 'completed' ? 'Đã nộp' : 'Đang làm'}</td>
                        <td className="py-3">{result.submitted_at ? new Date(result.submitted_at).toLocaleString('vi-VN') : '—'}</td>
                      </tr>
                    ))}</tbody>
                  </table>
                </div>
              )}
        </section>
      )}

      {activeTab === 'appeals' && <ExamAppealsTab examId={id} />}
    </div>
  );
}
