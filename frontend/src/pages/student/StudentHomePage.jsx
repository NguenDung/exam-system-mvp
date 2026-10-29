import { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { studentAPI } from '../../api/endpoints';
import Spinner from '../../components/Spinner';

const STATUS_LABELS = {
  not_attempted: { label: 'Chưa thi', color: 'text-gray-500', icon: '📋' },
  open: { label: 'Đang mở', color: 'text-emerald-600', icon: '🟢' },
  in_progress: { label: 'Đang làm dở', color: 'text-orange-600', icon: '⏳' },
  submitted: { label: 'Đã nộp', color: 'text-blue-600', icon: '✅' },
  not_started_yet: { label: 'Chưa mở', color: 'text-gray-400', icon: '🕐' },
  expired: { label: 'Đã hết hạn', color: 'text-red-500', icon: '🔒' },
  time_up: { label: 'Hết giờ', color: 'text-red-500', icon: '⏰' },
};

export default function StudentHomePage() {
  const [code, setCode] = useState('');
  const [lookupResult, setLookupResult] = useState(null);
  const [lookupLoading, setLookupLoading] = useState(false);
  const [lookupError, setLookupError] = useState('');
  const [attempts, setAttempts] = useState([]);
  const [attemptsLoading, setAttemptsLoading] = useState(true);
  const [attemptsError, setAttemptsError] = useState('');
  const [starting, setStarting] = useState(false);
  const [startConfirmOpen, setStartConfirmOpen] = useState(false);
  const navigate = useNavigate();

  const fetchAttempts = useCallback(async () => {
    setAttemptsLoading(true);
    setAttemptsError('');
    try {
      const { data } = await studentAPI.listAttempts();
      setAttempts(data.attempts || []);
    } catch (err) {
      setAttempts([]);
      setAttemptsError(err.response?.data?.error || 'Không thể tải lịch sử làm bài.');
    } finally {
      setAttemptsLoading(false);
    }
  }, []);

  useEffect(() => { fetchAttempts(); }, [fetchAttempts]);

  const handleLookup = async (e) => {
    e.preventDefault();
    if (!code.trim()) return;
    setLookupLoading(true);
    setLookupError('');
    setLookupResult(null);
    try {
      const { data } = await studentAPI.lookupExam(code.trim().toUpperCase());
      setLookupResult(data);
    } catch (err) {
      setLookupError(err.response?.data?.error || 'Không tìm thấy đề thi');
    } finally {
      setLookupLoading(false);
    }
  };

  const handleStart = async (confirmed = false) => {
    if (!lookupResult?.exam) return;
    setStarting(true);
    try {
      const { data } = await studentAPI.startExam(lookupResult.exam.id, { confirmed });
      // Auto-submitted on start (deadline passed) → go to results
      if (data.auto_submitted || data.attempt?.status === 'completed') {
        navigate(`/student/attempts/${data.attempt.id}/result`, { replace: true });
        return;
      }
      navigate(`/student/exam/${lookupResult.exam.id}/take`);
    } catch (err) {
      const payload = err.response?.data;
      if (payload?.attempt?.id && payload?.attempt?.status === 'completed') {
        navigate(`/student/attempts/${payload.attempt.id}/result`, { replace: true });
        return;
      }
      alert(payload?.error || 'Không thể bắt đầu bài thi');
    } finally {
      setStarting(false);
      setStartConfirmOpen(false);
    }
  };

  const canStart = lookupResult && ['not_attempted', 'in_progress'].includes(lookupResult.attempt_status) && lookupResult.exam_status === 'open';
  const isResume = lookupResult?.attempt_status === 'in_progress';

  return (
    <div className="space-y-8">
      {/* Lookup section */}
      <div className="card max-w-2xl mx-auto">
        <h2 className="text-xl font-bold text-gray-900 mb-1">🔍 Tìm kiếm đề thi</h2>
        <p className="text-gray-500 text-sm mb-4">Nhập mã đề thi để bắt đầu làm bài</p>

        <form onSubmit={handleLookup} className="flex gap-3">
          <div className="flex-1 relative">
            <input
              value={code}
              onChange={(e) => { setCode(e.target.value.toUpperCase()); setLookupResult(null); setLookupError(''); }}
              className="input font-mono text-lg tracking-widest uppercase pr-24"
              placeholder="IT2026"
              maxLength={20}
            />
            <span className="absolute right-3 top-1/2 -translate-y-1/2 text-xs text-gray-400 font-normal">gợi ý: IT2026</span>
          </div>
          <button type="submit" disabled={lookupLoading || !code.trim()} className="btn-primary px-6">
            {lookupLoading ? <Spinner size="sm" /> : '🔍 Tìm'}
          </button>
        </form>

        {lookupError && (
          <div className="mt-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm flex items-center gap-2">
            <span>❌</span> {lookupError}
          </div>
        )}

        {lookupResult && (
          <div className="mt-5 p-5 bg-emerald-50 border border-emerald-200 rounded-2xl">
            <div className="flex items-start justify-between gap-4 flex-wrap">
              <div>
                <h3 className="font-bold text-gray-900 text-lg">{lookupResult.exam.title}</h3>
                <div className="flex items-center gap-3 mt-2 text-sm text-gray-600 flex-wrap">
                  <span>🔑 <code className="font-mono font-bold text-indigo-700">{lookupResult.exam.exam_code}</code></span>
                  <span>⏱ {lookupResult.exam.duration} phút</span>
                  <span>📅 Đến {new Date(lookupResult.exam.end_time).toLocaleDateString('vi-VN')}</span>
                </div>
                <div className="mt-2 flex items-center gap-2">
                  {(() => {
                    const s = STATUS_LABELS[lookupResult.attempt_status] || STATUS_LABELS.not_attempted;
                    return <span className={`text-sm font-medium ${s.color}`}>{s.icon} {s.label}</span>;
                  })()}
                  {lookupResult.exam_status === 'not_started_yet' && (
                    <span className="text-xs text-gray-400">
                      (Mở lúc {new Date(lookupResult.exam.start_time).toLocaleString('vi-VN')})
                    </span>
                  )}
                </div>
              </div>
              <div className="flex flex-col gap-2">
                {canStart && (
                  <button
                    onClick={() => isResume ? handleStart() : setStartConfirmOpen(true)}
                    disabled={starting}
                    className="btn-success btn-lg"
                  >
                    {starting ? <Spinner size="sm" /> : (isResume ? '▶️ Tiếp tục làm bài' : '🚀 Bắt đầu làm bài')}
                  </button>
                )}
                {lookupResult.attempt_status === 'submitted' && lookupResult.attempt && (
                  <button
                    onClick={() => navigate(`/student/attempts/${lookupResult.attempt.id}/result`)}
                    className="btn-primary btn-lg"
                  >
                    📊 Xem kết quả
                  </button>
                )}
                {lookupResult.attempt_status === 'time_up' && lookupResult.attempt && (
                  <button
                    onClick={() => navigate(`/student/attempts/${lookupResult.attempt.id}/result`)}
                    className="btn-primary btn-lg"
                  >
                    📊 Xem kết quả (hết giờ)
                  </button>
                )}
                {lookupResult.exam_status === 'expired' && (
                  <span className="text-red-500 text-sm font-medium">🔒 Đề thi đã hết hạn</span>
                )}
                {lookupResult.exam_status === 'not_started_yet' && (
                  <span className="text-gray-400 text-sm">Đề thi chưa mở</span>
                )}
              </div>
            </div>
          </div>
        )}
      </div>

      {/* History */}
      <div>
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-bold text-gray-900">📚 Lịch sử làm bài</h2>
          <button onClick={fetchAttempts} className="btn-secondary btn-sm">🔄 Làm mới</button>
        </div>

        {attemptsLoading ? (
          <div className="flex justify-center py-10"><Spinner size="lg" /></div>
        ) : attemptsError ? (
          <div role="alert" className="card flex flex-col items-center gap-3 py-10 text-center">
            <p className="text-sm text-red-700">{attemptsError}</p>
            <button onClick={fetchAttempts} className="btn-secondary btn-sm">Thử lại</button>
          </div>
        ) : attempts.length === 0 ? (
          <div className="card text-center py-12">
            <div className="text-4xl mb-3">📭</div>
            <p className="text-gray-500">Bạn chưa tham gia bài thi nào</p>
          </div>
        ) : (
          <div className="card overflow-hidden p-0">
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="bg-gray-50 border-b text-left">
                    <th className="px-5 py-3 font-semibold text-gray-600">Đề thi</th>
                    <th className="px-5 py-3 font-semibold text-gray-600 text-center">Mã đề</th>
                    <th className="px-5 py-3 font-semibold text-gray-600 text-center">Điểm</th>
                    <th className="px-5 py-3 font-semibold text-gray-600 text-center">Đúng/Sai</th>
                    <th className="px-5 py-3 font-semibold text-gray-600 text-center">Trạng thái</th>
                    <th className="px-5 py-3 font-semibold text-gray-600">Thời gian</th>
                    <th className="px-5 py-3 font-semibold text-gray-600"></th>
                  </tr>
                </thead>
                <tbody>
                  {attempts.map((a) => (
                    <tr key={a.id} className="border-b last:border-0 hover:bg-gray-50 transition-colors">
                      <td className="px-5 py-4 font-medium text-gray-900">{a.exam_title}</td>
                      <td className="px-5 py-4 text-center">
                        <code className="bg-gray-100 px-2 py-0.5 rounded font-mono text-indigo-700">{a.exam_code}</code>
                      </td>
                      <td className="px-5 py-4 text-center font-bold text-indigo-700">
                        {a.status === 'completed' ? a.score.toFixed(2) : '—'}
                      </td>
                      <td className="px-5 py-4 text-center text-xs">
                        {a.status === 'completed' ? (
                          <span className="text-emerald-600 font-semibold">{a.correct_count}✓</span>
                        ) : '—'}
                        {a.status === 'completed' && a.wrong_count > 0 && (
                          <span className="text-red-500 font-semibold ml-1">{a.wrong_count}✗</span>
                        )}
                      </td>
                      <td className="px-5 py-4 text-center">
                        <span className={a.status === 'completed' ? 'badge-completed' : 'badge-in_progress'}>
                          {a.status === 'completed' ? '✅ Đã nộp' : '⏳ Đang làm'}
                        </span>
                      </td>
                      <td className="px-5 py-4 text-gray-500 text-xs">
                        {new Date(a.started_at).toLocaleString('vi-VN')}
                      </td>
                      <td className="px-5 py-4">
                        {a.status === 'completed' ? (
                          <button onClick={() => navigate(`/student/attempts/${a.id}/result`)}
                            className="btn-primary btn-sm">Kết quả</button>
                        ) : (
                          <button onClick={() => navigate(`/student/exam/${a.exam_id}/take`)}
                            className="btn-success btn-sm">Tiếp tục</button>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </div>

      {startConfirmOpen && lookupResult?.exam && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <div
            role="dialog"
            aria-modal="true"
            aria-labelledby="start-exam-title"
            className="w-full max-w-md rounded-2xl bg-white p-7 text-center shadow-2xl"
          >
            <div className="mb-3 text-5xl">📝</div>
            <h2 id="start-exam-title" className="text-xl font-bold text-gray-900">Xác nhận tham gia bài thi?</h2>
            <p className="mt-3 text-sm text-gray-600">
              Thời gian làm bài {lookupResult.exam.duration} phút sẽ bắt đầu đếm ngược ngay khi bạn đồng ý.
              Khi hết giờ, bài thi được tự động nộp và khóa.
            </p>
            <div className="mt-6 flex gap-3">
              <button
                onClick={() => setStartConfirmOpen(false)}
                disabled={starting}
                className="btn-secondary flex-1"
              >
                Để sau
              </button>
              <button
                onClick={() => handleStart(true)}
                disabled={starting}
                className="btn-success flex-1"
              >
                {starting ? <Spinner size="sm" /> : 'Đồng ý, vào thi'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
