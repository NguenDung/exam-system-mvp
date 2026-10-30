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

function parseBulkQuestionsText(text) {
  const trimmed = text.trim();
  if (trimmed.startsWith('[') || trimmed.startsWith('{')) {
    try {
      const parsed = JSON.parse(trimmed);
      return Array.isArray(parsed) ? parsed : [parsed];
    } catch (e) {
      throw new Error('Định dạng JSON không hợp lệ: ' + e.message);
    }
  }

  const blocks = trimmed.split(/(?:\r?\n){2,}|---+/).filter((b) => b.trim());
  const questions = [];

  for (let i = 0; i < blocks.length; i++) {
    const block = blocks[i].trim();
    if (!block) continue;
    const lines = block.split(/\r?\n/).map((l) => l.trim()).filter(Boolean);

    let prompt = '';
    let option_a = '';
    let option_b = '';
    let option_c = '';
    let option_d = '';
    let correct_option = 'A';
    let points = 1;

    for (const line of lines) {
      if (/^(?:Câu(?:\s*\d+)?[:.]|Prompt[:.]?)\s*(.*)/i.test(line)) {
        prompt = line.replace(/^(?:Câu(?:\s*\d+)?[:.]|Prompt[:.]?)\s*/i, '').trim();
      } else if (/^[A|a][.:)]\s*(.*)/.test(line)) {
        option_a = line.replace(/^[A|a][.:)]\s*/, '').trim();
      } else if (/^[B|b][.:)]\s*(.*)/.test(line)) {
        option_b = line.replace(/^[B|b][.:)]\s*/, '').trim();
      } else if (/^[C|c][.:)]\s*(.*)/.test(line)) {
        option_c = line.replace(/^[C|c][.:)]\s*/, '').trim();
      } else if (/^[D|d][.:)]\s*(.*)/.test(line)) {
        option_d = line.replace(/^[D|d][.:)]\s*/, '').trim();
      } else if (/^(?:Đáp án|Correct)[:.]?\s*([A-D])/i.test(line)) {
        const match = line.match(/^(?:Đáp án|Correct)[:.]?\s*([A-D])/i);
        if (match) correct_option = match[1].toUpperCase();
      } else if (/^(?:Điểm|Points)[:.]?\s*(\d+(?:\.\d+)?)/i.test(line)) {
        const match = line.match(/^(?:Điểm|Points)[:.]?\s*(\d+(?:\.\d+)?)/i);
        if (match) points = parseFloat(match[1]);
      } else if (!prompt) {
        prompt = line;
      }
    }

    if (!prompt || !option_a || !option_b || !option_c || !option_d) {
      throw new Error(`Khối câu hỏi số ${i + 1} thiếu nội dung hoặc không đủ 4 phương án A, B, C, D.`);
    }

    questions.push({
      prompt,
      option_a,
      option_b,
      option_c,
      option_d,
      correct_option,
      points,
    });
  }

  if (questions.length === 0) {
    throw new Error('Không tìm thấy câu hỏi hợp lệ trong nội dung đã nhập.');
  }

  return questions;
}

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
          <input
            type="number"
            step="0.5"
            min="0.5"
            value={q.points}
            onChange={set('points')}
            className="input"
          />
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

function EditQuestionModal({ question, onSave, onCancel, saving }) {
  const [form, setForm] = useState({
    prompt: question.prompt || '',
    option_a: question.option_a || '',
    option_b: question.option_b || '',
    option_c: question.option_c || '',
    option_d: question.option_d || '',
    correct_option: question.correct_option || 'A',
    points: question.points || 1,
  });

  const set = (k) => (e) => setForm((p) => ({ ...p, [k]: e.target.value }));

  const handleSubmit = (e) => {
    e.preventDefault();
    onSave(question.id, {
      ...form,
      points: parseFloat(form.points) || 1,
    });
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full p-6 space-y-4 max-h-[90vh] overflow-y-auto">
        <div className="flex items-center justify-between border-b border-slate-200 pb-3">
          <h3 className="text-base font-bold text-slate-900">Chỉnh sửa câu hỏi #{question.id}</h3>
          <button onClick={onCancel} className="text-slate-400 hover:text-slate-600 text-xl leading-none">&times;</button>
        </div>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="label">Nội dung câu hỏi *</label>
            <textarea
              value={form.prompt}
              onChange={set('prompt')}
              className="input resize-none"
              rows={3}
              required
            />
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            {['a', 'b', 'c', 'd'].map((opt) => (
              <div key={opt}>
                <label className="label">Đáp án {opt.toUpperCase()} *</label>
                <input
                  value={form[`option_${opt}`]}
                  onChange={set(`option_${opt}`)}
                  className="input"
                  required
                />
              </div>
            ))}
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 items-center">
            <div>
              <label className="label">Đáp án đúng *</label>
              <div className="flex gap-2">
                {OPTION_LABELS.map((opt) => (
                  <label
                    key={opt}
                    className={`flex items-center gap-1.5 cursor-pointer px-3 py-1.5 rounded border text-sm font-semibold transition-all ${
                      form.correct_option === opt
                        ? 'border-indigo-600 bg-indigo-50 text-indigo-700'
                        : 'border-slate-300 bg-white text-slate-700 hover:border-slate-400'
                    }`}
                  >
                    <input
                      type="radio"
                      name="modal_correct_option"
                      value={opt}
                      checked={form.correct_option === opt}
                      onChange={set('correct_option')}
                      className="sr-only"
                    />
                    {opt}
                  </label>
                ))}
              </div>
            </div>
            <div>
              <label className="label">Điểm số</label>
              <input
                type="number"
                step="0.5"
                min="0.5"
                value={form.points}
                onChange={set('points')}
                className="input"
                required
              />
            </div>
          </div>
          <div className="flex justify-end gap-2 pt-3 border-t border-slate-100">
            <button type="button" onClick={onCancel} className="btn-secondary">
              Hủy
            </button>
            <button type="submit" disabled={saving} className="btn-primary">
              {saving ? <Spinner size="sm" /> : null} Lưu thay đổi
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

function BulkImportModal({ onImport, onClose, importing }) {
  const [content, setContent] = useState('');
  const [error, setError] = useState('');

  const sampleFormat = `Câu: Thủ đô của Việt Nam là gì?
A. Hà Nội
B. TP. Hồ Chí Minh
C. Đà Nẵng
D. Cần Thơ
Đáp án: A
Điểm: 1

---

Câu: Giao thức mạng nào dùng để truyền tải siêu văn bản an toàn?
A. FTP
B. HTTPS
C. SMTP
D. Telnet
Đáp án: B
Điểm: 1`;

  const handleSubmit = (e) => {
    e.preventDefault();
    setError('');
    try {
      const parsedQuestions = parseBulkQuestionsText(content);
      onImport(parsedQuestions);
    } catch (err) {
      setError(err.message || 'Lỗi định dạng dữ liệu');
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full p-6 space-y-4 max-h-[90vh] overflow-y-auto">
        <div className="flex items-center justify-between border-b border-slate-200 pb-3">
          <h3 className="text-base font-bold text-slate-900">Nhập nhanh câu hỏi hàng loạt</h3>
          <button onClick={onClose} className="text-slate-400 hover:text-slate-600 text-xl leading-none">&times;</button>
        </div>
        <p className="text-xs text-slate-600 leading-relaxed">
          Hỗ trợ định dạng Text phân cách câu bằng dấu gạch ngang (<code className="bg-slate-100 px-1 py-0.5 rounded">---</code>) hoặc mảng JSON chuẩn.
        </p>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <div className="flex justify-between items-center mb-1">
              <label className="label mb-0">Nội dung câu hỏi</label>
              <button
                type="button"
                onClick={() => setContent(sampleFormat)}
                className="text-xs text-indigo-600 hover:underline"
              >
                Tải văn bản mẫu
              </button>
            </div>
            <textarea
              value={content}
              onChange={(e) => setContent(e.target.value)}
              className="input font-mono text-xs resize-none"
              rows={12}
              placeholder={sampleFormat}
              required
            />
          </div>
          {error && (
            <div className="rounded bg-red-50 p-3 text-xs text-red-700 border border-red-200">
              {error}
            </div>
          )}
          <div className="flex justify-end gap-2 pt-2 border-t border-slate-100">
            <button type="button" onClick={onClose} disabled={importing} className="btn-secondary">
              Hủy
            </button>
            <button type="submit" disabled={importing || !content.trim()} className="btn-primary">
              {importing ? <Spinner size="sm" /> : null} Tiến hành nhập dữ liệu
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

function QuestionCard({ q, index, isDraft, onEdit, onDelete }) {
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
          <div className="flex items-center gap-2 flex-shrink-0">
            <button
              onClick={() => onEdit(q)}
              className="btn-secondary btn-sm text-slate-700 hover:bg-slate-100"
            >
              Sửa
            </button>
            {confirming ? (
              <div className="flex gap-2">
                <button onClick={() => onDelete(q.id)} className="btn-danger btn-sm">
                  Xác nhận
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

function EditScoreModal({ attempt, onClose, onUpdated }) {
  const [score, setScore] = useState(attempt.score !== undefined ? attempt.score : 0);
  const [teacherNote, setTeacherNote] = useState(attempt.teacher_note || '');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      await teacherAPI.updateAttemptScore(attempt.attempt_id, {
        score: parseFloat(score),
        teacher_note: teacherNote,
      });
      onUpdated();
      onClose();
    } catch (err) {
      setError(err.response?.data?.error || 'Không thể cập nhật điểm số');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-md w-full p-6 space-y-4">
        <div className="flex items-center justify-between border-b border-slate-200 pb-3">
          <h3 className="text-base font-bold text-slate-900">Điều chỉnh điểm thi thí sinh</h3>
          <button onClick={onClose} className="text-slate-400 hover:text-slate-600 text-xl leading-none">&times;</button>
        </div>
        <p className="text-xs text-slate-600">
          Thí sinh: <span className="font-semibold text-slate-900">{attempt.full_name}</span> ({attempt.email})
        </p>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="label">Điểm số mới *</label>
            <input
              type="number"
              step="0.01"
              min="0"
              max="1000"
              value={score}
              onChange={(e) => setScore(e.target.value)}
              className="input"
              required
            />
          </div>
          <div>
            <label className="label">Ghi chú của giảng viên (Lý do sửa điểm)</label>
            <textarea
              rows={3}
              value={teacherNote}
              onChange={(e) => setTeacherNote(e.target.value)}
              placeholder="Ví dụ: Đính chính điểm sau khi phúc khảo"
              className="input resize-none"
            />
          </div>
          {error && (
            <div className="rounded bg-red-50 p-3 text-xs text-red-700 border border-red-200">
              {error}
            </div>
          )}
          <div className="flex justify-end gap-2 pt-2 border-t border-slate-100">
            <button type="button" onClick={onClose} disabled={loading} className="btn-secondary">
              Hủy
            </button>
            <button type="submit" disabled={loading} className="btn-primary">
              {loading ? <Spinner size="sm" /> : null} Lưu điểm số
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

function AuditLogModal({ onClose }) {
  const [logs, setLogs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    (async () => {
      try {
        const { data } = await teacherAPI.getAuditLogs();
        setLogs(data.logs || []);
      } catch (err) {
        setError(err.response?.data?.error || 'Không thể tải nhật ký thao tác');
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-3xl w-full p-6 space-y-4 max-h-[85vh] flex flex-col">
        <div className="flex items-center justify-between border-b border-slate-200 pb-3">
          <div>
            <h3 className="text-base font-bold text-slate-900">Nhật ký can thiệp điểm số và lượt thi (Audit Log)</h3>
            <p className="text-xs text-slate-500">Lưu vết minh bạch các hành vi sửa điểm hoặc cho phép thi lại</p>
          </div>
          <button onClick={onClose} className="text-slate-400 hover:text-slate-600 text-xl leading-none">&times;</button>
        </div>

        <div className="flex-1 overflow-y-auto">
          {loading ? (
            <div className="flex justify-center py-12"><Spinner size="md" /></div>
          ) : error ? (
            <p className="text-xs text-red-600 p-3 bg-red-50 rounded border border-red-200">{error}</p>
          ) : logs.length === 0 ? (
            <p className="text-center py-10 text-slate-500 text-sm">Chưa có hành vi can thiệp nào được ghi lại.</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-xs">
                <thead>
                  <tr className="bg-slate-50 border-b border-slate-200 text-left text-slate-600">
                    <th className="py-2.5 px-3">Thời gian</th>
                    <th className="py-2.5 px-3">Người thực hiện</th>
                    <th className="py-2.5 px-3">Hành động</th>
                    <th className="py-2.5 px-3">Lượt thi</th>
                    <th className="py-2.5 px-3">Nội dung chi tiết</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {logs.map((log) => (
                    <tr key={log.id} className="hover:bg-slate-50">
                      <td className="py-2.5 px-3 text-slate-500 whitespace-nowrap">
                        {new Date(log.created_at).toLocaleString('vi-VN')}
                      </td>
                      <td className="py-2.5 px-3 font-medium text-slate-800">
                        {log.actor?.full_name || 'Giảng viên'}
                        {log.actor?.email && <span className="block text-[11px] text-slate-400 font-normal">{log.actor.email}</span>}
                      </td>
                      <td className="py-2.5 px-3">
                        <span className={`px-2 py-0.5 rounded font-semibold text-[11px] ${
                          log.action === 'UPDATE_SCORE'
                            ? 'bg-blue-100 text-blue-800'
                            : log.action === 'RESET_ATTEMPT'
                            ? 'bg-amber-100 text-amber-800'
                            : 'bg-slate-100 text-slate-800'
                        }`}>
                          {log.action === 'UPDATE_SCORE' ? 'Sửa điểm' : log.action === 'RESET_ATTEMPT' ? 'Cho thi lại' : log.action}
                        </span>
                      </td>
                      <td className="py-2.5 px-3 font-mono text-slate-600">#{log.target_id}</td>
                      <td className="py-2.5 px-3 text-slate-700 max-w-xs">{log.details}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>

        <div className="flex justify-end pt-3 border-t border-slate-100">
          <button type="button" onClick={onClose} className="btn-secondary btn-sm">
            Đóng
          </button>
        </div>
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
      onClose();
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
          <h3 className="text-base font-bold text-slate-900">Thiết lập thông tin đề thi #{exam.id}</h3>
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
            <label className="label">Mô tả đề thi</label>
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
              Xuất bản đề thi (Cho phép sinh viên thi)
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

  // New states for CRUD features
  const [showEditExamModal, setShowEditExamModal] = useState(false);
  const [editingQuestion, setEditingQuestion] = useState(null);
  const [editingQuestionSaving, setEditingQuestionSaving] = useState(false);
  const [showBulkModal, setShowBulkModal] = useState(false);
  const [bulkImporting, setBulkImporting] = useState(false);
  const [editingAttempt, setEditingAttempt] = useState(null);
  const [resettingAttempt, setResettingAttempt] = useState(null);
  const [resetLoading, setResetLoading] = useState(false);
  const [showDeleteExamModal, setShowDeleteExamModal] = useState(false);
  const [deleteExamLoading, setDeleteExamLoading] = useState(false);
  const [showAuditModal, setShowAuditModal] = useState(false);
  const [exportingCSV, setExportingCSV] = useState(false);

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

  const fetchScores = useCallback(() => {
    setScoreLoading(true);
    setScoreError('');
    teacherAPI
      .getResults(id)
      .then(({ data }) => setScoreData(data))
      .catch((err) => setScoreError(err.response?.data?.error || 'Không thể tải bảng điểm.'))
      .finally(() => setScoreLoading(false));
  }, [id]);

  useEffect(() => {
    if (activeTab === 'scores') {
      fetchScores();
    }
  }, [activeTab, fetchScores]);

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

  const handleUpdateQuestion = async (qId, updatePayload) => {
    setEditingQuestionSaving(true);
    try {
      const { data } = await teacherAPI.updateQuestion(qId, updatePayload);
      setQuestions((prev) => prev.map((q) => (q.id === qId ? data.question : q)));
      setEditingQuestion(null);
      showToast('Đã cập nhật câu hỏi');
    } catch (err) {
      showToast(err.response?.data?.error || 'Cập nhật câu hỏi không thành công', 'error');
    } finally {
      setEditingQuestionSaving(false);
    }
  };

  const handleBulkImport = async (parsedQuestions) => {
    setBulkImporting(true);
    try {
      const { data } = await teacherAPI.bulkAddQuestions(id, { questions: parsedQuestions });
      setQuestions((prev) => [...prev, ...(data.questions || [])]);
      setShowBulkModal(false);
      showToast(`Đã thêm thành công ${data.count || parsedQuestions.length} câu hỏi`);
    } catch (err) {
      showToast(err.response?.data?.error || 'Nhập câu hỏi hàng loạt thất bại', 'error');
    } finally {
      setBulkImporting(false);
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
    if (!window.confirm(`Xác nhận xuất bản đề thi "${exam.title}"? Sau khi xuất bản sinh viên có thể bắt đầu làm bài.`)) return;
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

  const handleDeleteExam = async () => {
    setDeleteExamLoading(true);
    try {
      await teacherAPI.deleteExam(id);
      navigate('/teacher');
    } catch (err) {
      showToast(err.response?.data?.error || 'Lỗi khi xóa bài thi', 'error');
      setDeleteExamLoading(false);
    }
  };

  const handleResetAttempt = async () => {
    if (!resettingAttempt) return;
    setResetLoading(true);
    try {
      await teacherAPI.resetAttempt(resettingAttempt.attempt_id);
      showToast(`Đã xóa lượt thi của thí sinh ${resettingAttempt.full_name}. Sinh viên có thể thi lại.`);
      setResettingAttempt(null);
      fetchScores();
    } catch (err) {
      showToast(err.response?.data?.error || 'Không thể xóa lượt thi', 'error');
    } finally {
      setResetLoading(false);
    }
  };

  const handleExportCSV = async () => {
    setExportingCSV(true);
    try {
      const response = await teacherAPI.exportResultsCSV(id);
      const blob = new Blob([response.data], { type: 'text/csv;charset=utf-8;' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `bang-diem-${exam.exam_code || id}.csv`);
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);
      showToast('Đã tải xuống bảng điểm CSV thành công');
    } catch (err) {
      showToast(err.response?.data?.error || 'Lỗi khi xuất bảng điểm CSV', 'error');
    } finally {
      setExportingCSV(false);
    }
  };

  if (loading) return <div className="flex justify-center py-20"><Spinner size="lg" /></div>;
  if (!exam) return null;

  const isDraft = exam.status === 'draft';
  const tabs = [
    { id: 'info', label: 'Thông tin chung' },
    { id: 'questions', label: 'Câu hỏi & Đáp án' },
    { id: 'scores', label: 'Bảng điểm' },
    { id: 'appeals', label: 'Phúc khảo' },
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
          Quay lại danh sách đề
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
          {exam.description && (
            <p className="text-slate-600 text-sm mt-1">{exam.description}</p>
          )}
          <div className="flex items-center gap-5 mt-1 text-xs text-slate-500 flex-wrap">
            <span>Mã đề: <code className="bg-slate-100 px-2 py-0.5 rounded font-mono font-bold text-slate-800">{exam.exam_code}</code></span>
            <span>Thời lượng: <strong>{exam.duration} phút</strong></span>
            <span>Bắt đầu: {new Date(exam.start_time).toLocaleString('vi-VN')}</span>
            <span>Kết thúc: {new Date(exam.end_time).toLocaleString('vi-VN')}</span>
          </div>
        </div>
        <div className="flex gap-2 flex-wrap items-center">
          <button
            onClick={() => setShowEditExamModal(true)}
            className="btn-secondary btn-sm"
          >
            Sửa đề thi
          </button>
          <button
            onClick={() => setShowDeleteExamModal(true)}
            className="btn-secondary btn-sm text-red-600 hover:bg-red-50 hover:border-red-200"
          >
            Xóa đề thi
          </button>
          <Link to={`/teacher/exams/${id}/results`} className="btn-secondary btn-sm">
            Xem bảng điểm
          </Link>
          {isDraft && (
            <button onClick={handlePublish} disabled={publishing || questions.length === 0} className="btn-primary btn-sm">
              {publishing ? <Spinner size="sm" /> : null} Xuất bản đề
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
              {exam.description && (
                <div className="sm:col-span-2"><dt className="text-slate-500">Mô tả kỳ thi</dt><dd className="text-slate-800 whitespace-pre-wrap">{exam.description}</dd></div>
              )}
              {exam.published_at && (
                <div><dt className="text-slate-500">Thời điểm xuất bản</dt><dd>{new Date(exam.published_at).toLocaleString('vi-VN')}</dd></div>
              )}
            </dl>
          </div>
        </div>
      )}

      {activeTab === 'questions' && (
        <>
          <div className="flex items-center justify-between mb-4 flex-wrap gap-2">
            <h3 className="font-semibold text-slate-900 text-sm">Danh sách câu hỏi ({questions.length})</h3>
            {isDraft && (
              <div className="flex gap-2">
                <button
                  onClick={() => setShowBulkModal(true)}
                  className="btn-secondary btn-sm"
                >
                  Nhập nhanh câu hỏi
                </button>
                <button
                  onClick={() => setShowForm(!showForm)}
                  className="btn-primary btn-sm"
                >
                  {showForm ? 'Đóng form' : 'Thêm câu hỏi mới'}
                </button>
              </div>
            )}
          </div>

          <div className="mb-4 space-y-3">
            {questions.length === 0 && !showForm && (
              <div className="bg-white rounded-lg border border-dashed border-slate-300 py-12 text-center">
                <p className="text-slate-500">Đề thi chưa có câu hỏi nào. {isDraft && 'Bấm nút "Thêm câu hỏi mới" hoặc "Nhập nhanh câu hỏi" để bắt đầu soạn đề.'}</p>
              </div>
            )}
            {questions.map((question, index) => (
              <QuestionCard
                key={question.id}
                q={question}
                index={index}
                isDraft={isDraft}
                onEdit={(q) => setEditingQuestion(q)}
                onDelete={handleDelete}
              />
            ))}
          </div>

          {isDraft && showForm && (
            <QuestionForm onSave={handleAddQuestion} onCancel={() => setShowForm(false)} saving={saving} />
          )}

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
            <div>
              <h2 className="text-base font-bold text-slate-900">Bảng điểm thi sinh viên</h2>
              <p className="text-xs text-slate-500">Giảng viên có quyền hiệu chỉnh điểm thi hoặc cho phép sinh viên thi lại nếu gặp sự cố</p>
            </div>
            <div className="flex gap-2 flex-wrap">
              <button
                onClick={handleExportCSV}
                disabled={exportingCSV}
                className="btn-secondary btn-sm"
              >
                {exportingCSV ? 'Đang xuất CSV...' : 'Xuất bảng điểm (CSV)'}
              </button>
              <button
                onClick={() => setShowAuditModal(true)}
                className="btn-secondary btn-sm"
              >
                Nhật ký can thiệp
              </button>
              <button onClick={fetchScores} className="btn-secondary btn-sm">Làm mới</button>
              <Link to={`/teacher/exams/${id}/results`} className="btn-secondary btn-sm">Xem thống kê phân phối</Link>
            </div>
          </div>
          {scoreLoading ? (
            <div className="flex justify-center py-10"><Spinner size="lg" /></div>
          ) : scoreError ? (
            <p role="alert" className="text-sm text-red-600">{scoreError}</p>
          ) : !scoreData?.results?.length ? (
            <p className="py-8 text-center text-slate-500">Chưa có sinh viên nào tham gia làm bài thi này.</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full min-w-[700px] text-sm">
                <thead>
                  <tr className="border-b border-slate-200 text-left text-slate-500 bg-slate-50">
                    <th className="py-2.5 px-3">Họ và tên</th>
                    <th className="py-2.5 px-3">Điểm số</th>
                    <th className="py-2.5 px-3">Trạng thái</th>
                    <th className="py-2.5 px-3 text-center">Số lần vi phạm</th>
                    <th className="py-2.5 px-3">Thời gian nộp</th>
                    <th className="py-2.5 px-3">Ghi chú</th>
                    <th className="py-2.5 px-3 text-right">Thao tác</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {scoreData.results.map((result) => (
                    <tr key={result.attempt_id} className="hover:bg-slate-50">
                      <td className="py-3 px-3">
                        <p className="font-medium text-slate-900">{result.full_name}</p>
                        <p className="text-xs text-slate-500">{result.email}</p>
                      </td>
                      <td className="py-3 px-3 font-semibold text-slate-900">
                        {result.status === 'completed' ? Number(result.score).toFixed(2) : '-'}
                      </td>
                      <td className="py-3 px-3">
                        <span className={`px-2 py-0.5 rounded text-xs font-semibold ${
                          result.status === 'completed' ? 'bg-blue-100 text-blue-800' : 'bg-amber-100 text-amber-800'
                        }`}>
                          {result.status === 'completed' ? 'Đã nộp' : 'Đang làm'}
                        </span>
                      </td>
                      <td className="py-3 px-3 text-center">
                        {result.violation_count >= 3 ? (
                          <span className="px-2 py-0.5 rounded text-xs font-bold bg-red-100 text-red-800 border border-red-200">
                            {result.violation_count} (Đình chỉ)
                          </span>
                        ) : result.violation_count > 0 ? (
                          <span className="px-2 py-0.5 rounded text-xs font-semibold bg-amber-100 text-amber-800 border border-amber-200">
                            {result.violation_count}
                          </span>
                        ) : (
                          <span className="text-slate-400 text-xs">0</span>
                        )}
                      </td>
                      <td className="py-3 px-3 text-slate-500 text-xs">
                        {result.submitted_at ? new Date(result.submitted_at).toLocaleString('vi-VN') : '-'}
                      </td>
                      <td className="py-3 px-3 text-xs text-slate-500 max-w-xs truncate">
                        {result.teacher_note || '-'}
                      </td>
                      <td className="py-3 px-3 text-right space-x-2">
                        {result.status === 'completed' && (
                          <button
                            onClick={() => setEditingAttempt(result)}
                            className="btn-secondary btn-sm text-xs py-1"
                          >
                            Sửa điểm
                          </button>
                        )}
                        <button
                          onClick={() => setResettingAttempt(result)}
                          className="btn-secondary btn-sm text-xs py-1 text-red-600 hover:bg-red-50 hover:border-red-200"
                        >
                          Cho thi lại
                        </button>
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

      {/* Modal: Edit Exam */}
      {showEditExamModal && (
        <EditExamModal
          exam={exam}
          onClose={() => setShowEditExamModal(false)}
          onUpdated={(updated) => {
            setExam(updated);
            showToast('Đã cập nhật thông tin đề thi');
          }}
        />
      )}

      {/* Modal: Delete Exam */}
      {showDeleteExamModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full p-6 space-y-4">
            <h3 className="text-lg font-bold text-red-600">Xác nhận xóa đề thi</h3>
            <p className="text-sm text-slate-600">
              Bạn có chắc chắn muốn xóa vĩnh viễn đề thi{' '}
              <span className="font-semibold text-slate-900">{exam.title}</span> (Mã: {exam.exam_code})?
            </p>
            <p className="text-xs text-red-600 bg-red-50 p-2.5 rounded border border-red-200">
              Cảnh báo: Hành động này sẽ xóa toàn bộ câu hỏi, kết quả làm bài và đơn phúc khảo của đề thi này.
            </p>
            <div className="flex justify-end gap-2 pt-2 border-t border-slate-100">
              <button
                type="button"
                onClick={() => setShowDeleteExamModal(false)}
                disabled={deleteExamLoading}
                className="btn-secondary"
              >
                Hủy bỏ
              </button>
              <button
                type="button"
                onClick={handleDeleteExam}
                disabled={deleteExamLoading}
                className="btn-danger"
              >
                {deleteExamLoading ? 'Đang xóa...' : 'Xóa vĩnh viễn'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Modal: Edit Question */}
      {editingQuestion && (
        <EditQuestionModal
          question={editingQuestion}
          onSave={handleUpdateQuestion}
          onCancel={() => setEditingQuestion(null)}
          saving={editingQuestionSaving}
        />
      )}

      {/* Modal: Bulk Import Questions */}
      {showBulkModal && (
        <BulkImportModal
          onImport={handleBulkImport}
          onClose={() => setShowBulkModal(false)}
          importing={bulkImporting}
        />
      )}

      {/* Modal: Edit Score */}
      {editingAttempt && (
        <EditScoreModal
          attempt={editingAttempt}
          onClose={() => setEditingAttempt(null)}
          onUpdated={() => {
            showToast('Đã cập nhật điểm thi thành công');
            fetchScores();
          }}
        />
      )}

      {/* Modal: Reset Attempt */}
      {resettingAttempt && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full p-6 space-y-4">
            <h3 className="text-base font-bold text-red-600">Xác nhận cho thi lại (Xóa lượt thi)</h3>
            <p className="text-sm text-slate-600">
              Bạn có chắc chắn muốn xóa toàn bộ câu trả lời và kết quả lượt thi của thí sinh{' '}
              <span className="font-semibold text-slate-900">{resettingAttempt.full_name}</span>?
            </p>
            <p className="text-xs text-slate-500 bg-slate-50 p-2.5 rounded border border-slate-200">
              Sau khi xóa, sinh viên sẽ được làm lại bài thi từ đầu như một lượt thi hoàn toàn mới.
            </p>
            <div className="flex justify-end gap-2 pt-2 border-t border-slate-100">
              <button
                type="button"
                onClick={() => setResettingAttempt(null)}
                disabled={resetLoading}
                className="btn-secondary"
              >
                Hủy
              </button>
              <button
                type="button"
                onClick={handleResetAttempt}
                disabled={resetLoading}
                className="btn-danger"
              >
                {resetLoading ? 'Đang xử lý...' : 'Xác nhận xóa lượt thi'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Modal: Audit Log */}
      {showAuditModal && (
        <AuditLogModal onClose={() => setShowAuditModal(false)} />
      )}
    </div>
  );
}
