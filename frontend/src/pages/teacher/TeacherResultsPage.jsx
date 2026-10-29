import { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { teacherAPI } from '../../api/endpoints';
import Spinner from '../../components/Spinner';

function StatCard({ label, value }) {
  return (
    <div className="bg-white rounded-lg border border-slate-200 p-5 text-center">
      <div className="text-2xl font-bold text-slate-900">{value}</div>
      <div className="text-xs text-slate-500 mt-1 uppercase tracking-wide">{label}</div>
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
              placeholder="Ví dụ: Đính chính điểm câu 4 sau khi rà soát khiếu nại"
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

export default function TeacherResultsPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [editingAttempt, setEditingAttempt] = useState(null);
  const [resettingAttempt, setResettingAttempt] = useState(null);
  const [resetLoading, setResetLoading] = useState(false);
  const [feedbackMsg, setFeedbackMsg] = useState('');

  const fetchResults = useCallback(async () => {
    try {
      const { data: d } = await teacherAPI.getResults(id);
      setData(d);
    } catch {
      navigate('/teacher');
    } finally {
      setLoading(false);
    }
  }, [id, navigate]);

  useEffect(() => {
    fetchResults();
  }, [fetchResults]);

  const handleResetAttempt = async () => {
    if (!resettingAttempt) return;
    setResetLoading(true);
    try {
      await teacherAPI.resetAttempt(resettingAttempt.attempt_id);
      setFeedbackMsg(`Đã xóa lượt thi của thí sinh ${resettingAttempt.full_name}. Sinh viên có thể thi lại.`);
      setResettingAttempt(null);
      fetchResults();
      setTimeout(() => setFeedbackMsg(''), 4000);
    } catch (err) {
      alert(err.response?.data?.error || 'Không thể xóa lượt thi');
    } finally {
      setResetLoading(false);
    }
  };

  if (loading) return <div className="flex justify-center py-20"><Spinner size="lg" /></div>;
  if (!data) return null;

  const { exam, statistics: stats, results } = data;

  return (
    <div>
      {/* Header */}
      <div className="flex items-start gap-4 mb-6 flex-wrap">
        <button onClick={() => navigate(`/teacher/exams/${id}`)} className="btn-secondary btn-sm mt-0.5">
          Quay lại đề thi
        </button>
        <div className="flex-1">
          <h1 className="text-xl font-bold text-slate-900">{exam.title}</h1>
          <p className="text-slate-500 text-sm mt-0.5">Báo cáo bảng điểm và phân bố kết quả thi</p>
        </div>
        <button onClick={fetchResults} className="btn-secondary btn-sm">Làm mới dữ liệu</button>
      </div>

      {feedbackMsg && (
        <div className="mb-4 rounded-md bg-emerald-50 border border-emerald-200 p-3 text-sm text-emerald-800">
          {feedbackMsg}
        </div>
      )}

      {/* Stats */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-6">
        <StatCard label="Tổng số bài nộp" value={stats.total_students} />
        <StatCard label="Điểm trung bình" value={stats.total_students > 0 ? stats.avg_score.toFixed(2) : '—'} />
        <StatCard label="Điểm cao nhất" value={stats.total_students > 0 ? stats.max_score.toFixed(2) : '—'} />
        <StatCard label="Điểm thấp nhất" value={stats.total_students > 0 ? stats.min_score.toFixed(2) : '—'} />
      </div>

      {/* Results table */}
      <div className="bg-white rounded-lg border border-slate-200 overflow-hidden">
        <div className="px-5 py-3.5 border-b border-slate-200 flex items-center justify-between">
          <h3 className="font-semibold text-slate-900 text-sm">Danh sách điểm thí sinh</h3>
          <span className="text-xs text-slate-500">{results.length} bài thi</span>
        </div>
        {results.length === 0 ? (
          <div className="text-center py-16 text-slate-500">
            <p>Chưa có sinh viên nào tham gia làm bài thi này</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-slate-50 border-b border-slate-200 text-left text-slate-600">
                  <th className="px-5 py-3 font-semibold">STT</th>
                  <th className="px-5 py-3 font-semibold">Họ và tên</th>
                  <th className="px-5 py-3 font-semibold">Email</th>
                  <th className="px-5 py-3 font-semibold text-center">Điểm số</th>
                  <th className="px-5 py-3 font-semibold text-center">Đúng</th>
                  <th className="px-5 py-3 font-semibold text-center">Sai / Trống</th>
                  <th className="px-5 py-3 font-semibold text-center">Trạng thái</th>
                  <th className="px-5 py-3 font-semibold">Ghi chú</th>
                  <th className="px-5 py-3 font-semibold text-right">Thao tác</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {results.map((r, i) => (
                  <tr key={r.attempt_id} className="hover:bg-slate-50 transition-colors">
                    <td className="px-5 py-3.5 text-slate-400 font-mono">{i + 1}</td>
                    <td className="px-5 py-3.5 font-medium text-slate-900">{r.full_name}</td>
                    <td className="px-5 py-3.5 text-slate-500">{r.email}</td>
                    <td className="px-5 py-3.5 text-center">
                      <span className="font-bold text-slate-900">{r.score.toFixed(2)}</span>
                    </td>
                    <td className="px-5 py-3.5 text-center text-emerald-700 font-medium">
                      {r.correct_count}
                    </td>
                    <td className="px-5 py-3.5 text-center text-red-600 font-medium">
                      {r.wrong_count}
                    </td>
                    <td className="px-5 py-3.5 text-center">
                      <span className={`px-2.5 py-0.5 rounded text-xs font-semibold ${
                        r.status === 'completed' ? 'bg-blue-100 text-blue-800' : 'bg-amber-100 text-amber-800'
                      }`}>
                        {r.status === 'completed' ? 'Đã nộp' : 'Đang làm'}
                      </span>
                    </td>
                    <td className="px-5 py-3.5 text-slate-500 text-xs max-w-xs truncate">
                      {r.teacher_note || '—'}
                    </td>
                    <td className="px-5 py-3.5 text-right space-x-2">
                      {r.status === 'completed' && (
                        <button
                          onClick={() => setEditingAttempt(r)}
                          className="btn-secondary btn-sm text-xs py-1"
                        >
                          Sửa điểm
                        </button>
                      )}
                      <button
                        onClick={() => setResettingAttempt(r)}
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
      </div>

      {/* Modal: Edit Score */}
      {editingAttempt && (
        <EditScoreModal
          attempt={editingAttempt}
          onClose={() => setEditingAttempt(null)}
          onUpdated={() => {
            setFeedbackMsg('Đã cập nhật điểm thi thành công');
            fetchResults();
            setTimeout(() => setFeedbackMsg(''), 4000);
          }}
        />
      )}

      {/* Modal: Reset Attempt */}
      {resettingAttempt && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full p-6 space-y-4">
            <h3 className="text-base font-bold text-red-600">Xác nhận cho thi lại (Xóa lượt thi)</h3>
            <p className="text-sm text-slate-600">
              Bạn có chắc chắn muốn xóa lượt thi của thí sinh{' '}
              <span className="font-semibold text-slate-900">{resettingAttempt.full_name}</span>?
            </p>
            <p className="text-xs text-slate-500 bg-slate-50 p-2.5 rounded border border-slate-200">
              Toàn bộ bài làm và kết quả của thí sinh sẽ bị xóa. Sinh viên có thể thi lại từ đầu.
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
    </div>
  );
}
