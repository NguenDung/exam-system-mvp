import { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { studentAPI } from '../../api/endpoints';
import Spinner from '../../components/Spinner';

const STATUS_LABELS = {
  not_attempted: { label: 'Chưa thi', badgeClass: 'bg-slate-100 text-slate-700' },
  open: { label: 'Đang mở', badgeClass: 'bg-emerald-100 text-emerald-800' },
  in_progress: { label: 'Đang làm dở', badgeClass: 'bg-amber-100 text-amber-800' },
  submitted: { label: 'Đã nộp bài', badgeClass: 'bg-blue-100 text-blue-800' },
  not_started_yet: { label: 'Chưa mở', badgeClass: 'bg-slate-100 text-slate-600' },
  expired: { label: 'Đã kết thúc', badgeClass: 'bg-red-100 text-red-800' },
  time_up: { label: 'Hết giờ', badgeClass: 'bg-red-100 text-red-800' },
};

export default function StudentHomePage() {
  const [code, setCode] = useState('');
  const [lookupResult, setLookupResult] = useState(null);
  const [lookupLoading, setLookupLoading] = useState(false);
  const [lookupError, setLookupError] = useState('');

  // Available Exams State
  const [availableExams, setAvailableExams] = useState([]);
  const [availableLoading, setAvailableLoading] = useState(true);
  const [availableError, setAvailableError] = useState('');

  // History State
  const [attempts, setAttempts] = useState([]);
  const [attemptsLoading, setAttemptsLoading] = useState(true);
  const [attemptsError, setAttemptsError] = useState('');

  // Starting exam modal
  const [startingExam, setStartingExam] = useState(null);
  const [startingLoading, setStartingLoading] = useState(false);
  const [actionError, setActionError] = useState('');
  const navigate = useNavigate();

  const fetchAvailableExams = useCallback(async () => {
    setAvailableLoading(true);
    setAvailableError('');
    try {
      const { data } = await studentAPI.getAvailableExams();
      setAvailableExams(data.exams || []);
    } catch (err) {
      setAvailableError(err.response?.data?.error || 'Không thể tải danh sách bài thi đang mở.');
    } finally {
      setAvailableLoading(false);
    }
  }, []);

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

  useEffect(() => {
    fetchAvailableExams();
    fetchAttempts();
  }, [fetchAvailableExams, fetchAttempts]);

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
      setLookupError(err.response?.data?.error || 'Không tìm thấy đề thi với mã đã nhập');
    } finally {
      setLookupLoading(false);
    }
  };

  const handleStartExam = async (examId, confirmed = true) => {
    setStartingLoading(true);
    try {
      const { data } = await studentAPI.startExam(examId, { confirmed });
      if (data.auto_submitted || data.attempt?.status === 'completed') {
        navigate(`/student/attempts/${data.attempt.id}/result`, { replace: true });
        return;
      }
      navigate(`/student/exam/${examId}/take`);
    } catch (err) {
      const payload = err.response?.data;
      if (payload?.attempt?.id && payload?.attempt?.status === 'completed') {
        navigate(`/student/attempts/${payload.attempt.id}/result`, { replace: true });
        return;
      }
      setActionError(payload?.error || 'Không thể bắt đầu bài thi');
      setTimeout(() => setActionError(''), 6000);
    } finally {
      setStartingLoading(false);
      setStartingExam(null);
    }
  };

  return (
    <div className="space-y-8">
      {actionError && (
        <div className="rounded-md bg-red-50 border border-red-200 p-3.5 text-sm text-red-800 flex items-center justify-between">
          <span>{actionError}</span>
          <button onClick={() => setActionError('')} className="text-red-500 hover:text-red-700 text-lg leading-none font-bold ml-2">
            &times;
          </button>
        </div>
      )}

      {/* Section 1: Danh sách bài thi đang mở */}
      <div>
        <div className="flex items-center justify-between mb-4 border-b border-slate-200 pb-3">
          <div>
            <h1 className="text-xl font-bold tracking-tight text-slate-900">Đề thi đang mở</h1>
          </div>
          <button onClick={fetchAvailableExams} className="btn-secondary btn-sm">Làm mới</button>
        </div>

        {availableLoading ? (
          <div className="flex justify-center py-12"><Spinner size="lg" /></div>
        ) : availableError ? (
          <div className="p-4 rounded-md bg-red-50 border border-red-200 text-sm text-red-700">
            {availableError}
          </div>
        ) : availableExams.length === 0 ? (
          <div className="text-center py-12 bg-white rounded-lg border border-slate-200 text-slate-500 text-sm">
            Hiện tại không có bài thi nào đang mở hoặc bạn chưa được phân quyền truy cập.
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
            {availableExams.map((ex) => {
              const statusCfg = STATUS_LABELS[ex.attempt_status] || STATUS_LABELS.not_attempted;
              const isNotAttempted = ex.attempt_status === 'not_attempted';
              const isInProgress = ex.attempt_status === 'in_progress';
              const isFinished = ex.attempt_status === 'submitted' || ex.attempt_status === 'time_up';

              return (
                <div
                  key={ex.id}
                  className="bg-white rounded-lg border border-slate-200 p-5 shadow-sm flex flex-col justify-between hover:border-slate-300 transition-colors"
                >
                  <div className="space-y-3">
                    <div className="flex items-start justify-between gap-2">
                      <span className={`px-2.5 py-0.5 rounded text-xs font-semibold ${statusCfg.badgeClass}`}>
                        {statusCfg.label}
                      </span>
                      <span className="font-mono text-xs font-bold bg-slate-100 text-slate-800 px-2 py-0.5 rounded">
                        {ex.exam_code}
                      </span>
                    </div>

                    <div>
                      <h3 className="font-bold text-slate-900 text-base line-clamp-1">{ex.title}</h3>
                      {ex.description && (
                        <p className="text-xs text-slate-500 mt-1 line-clamp-2">{ex.description}</p>
                      )}
                    </div>

                    <div className="space-y-1.5 text-xs text-slate-600 bg-slate-50 p-3 rounded">
                      <div className="flex justify-between">
                        <span>Thời lượng:</span>
                        <strong className="text-slate-900">{ex.duration} phút</strong>
                      </div>
                      <div className="flex justify-between">
                        <span>Số lượng câu hỏi:</span>
                        <strong className="text-slate-900">{ex.questions_count} câu</strong>
                      </div>
                      <div className="flex justify-between">
                        <span>Tổng điểm:</span>
                        <strong className="text-slate-900">{ex.total_points} điểm</strong>
                      </div>
                      <div className="flex justify-between">
                        <span>Hạn nộp bài:</span>
                        <span className="text-slate-800">{new Date(ex.end_time).toLocaleString('vi-VN')}</span>
                      </div>
                    </div>
                  </div>

                  <div className="pt-4 border-t border-slate-100 mt-4">
                    {isNotAttempted && (
                      <button
                        onClick={() => setStartingExam(ex)}
                        className="w-full btn-primary text-sm py-2"
                      >
                        Vào thi ngay
                      </button>
                    )}

                    {isInProgress && (
                      <button
                        onClick={() => navigate(`/student/exam/${ex.id}/take`)}
                        className="w-full btn-primary text-sm py-2 bg-amber-600 hover:bg-amber-700"
                      >
                        Tiếp tục làm bài
                      </button>
                    )}

                    {isFinished && (
                      <button
                        onClick={() => navigate(`/student/attempts/${ex.attempt_id}/result`)}
                        className="w-full btn-secondary text-sm py-2"
                      >
                        Xem kết quả ({ex.score !== undefined ? Number(ex.score).toFixed(2) : '—'} điểm)
                      </button>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* Section 2: Tra cứu kỳ thi thủ công bằng mã */}
      <div className="card max-w-2xl mx-auto border border-slate-200">
        <h2 className="text-base font-bold text-slate-900 mb-1">Vào thi bằng mã đề</h2>
        <p className="text-slate-500 text-xs mb-4">Nhập mã đề thi để vào phòng thi</p>

        <form onSubmit={handleLookup} className="flex gap-2">
          <input
            value={code}
            onChange={(e) => {
              setCode(e.target.value.toUpperCase());
              setLookupResult(null);
              setLookupError('');
            }}
            className="input font-mono text-sm tracking-wider uppercase flex-1"
            placeholder="Ví dụ: IT2026"
            maxLength={20}
          />
          <button type="submit" disabled={lookupLoading || !code.trim()} className="btn-primary px-5">
            {lookupLoading ? <Spinner size="sm" /> : 'Tra cứu'}
          </button>
        </form>

        {lookupError && (
          <div className="mt-3 p-3 bg-red-50 border border-red-200 rounded text-red-700 text-xs">
            {lookupError}
          </div>
        )}

        {lookupResult && (
          <div className="mt-4 p-4 bg-slate-50 border border-slate-200 rounded-lg">
            <div className="flex items-start justify-between gap-4 flex-wrap">
              <div className="space-y-1.5">
                <h3 className="font-bold text-slate-900 text-base">{lookupResult.exam.title}</h3>
                <div className="flex items-center gap-4 text-xs text-slate-600 flex-wrap">
                  <span>Mã đề: <code className="font-mono font-bold text-slate-900">{lookupResult.exam.exam_code}</code></span>
                  <span>Thời lượng: <strong>{lookupResult.exam.duration} phút</strong></span>
                  <span>Hạn nộp: {new Date(lookupResult.exam.end_time).toLocaleString('vi-VN')}</span>
                </div>
                <div className="pt-1">
                  {(() => {
                    const s = STATUS_LABELS[lookupResult.attempt_status] || STATUS_LABELS.not_attempted;
                    return <span className={`px-2 py-0.5 rounded text-xs font-semibold ${s.badgeClass}`}>{s.label}</span>;
                  })()}
                </div>
              </div>

              <div>
                {lookupResult.attempt_status === 'not_attempted' && lookupResult.exam_status === 'open' && (
                  <button
                    onClick={() => setStartingExam(lookupResult.exam)}
                    className="btn-primary btn-sm"
                  >
                    Bắt đầu làm bài
                  </button>
                )}
                {lookupResult.attempt_status === 'in_progress' && (
                  <button
                    onClick={() => navigate(`/student/exam/${lookupResult.exam.id}/take`)}
                    className="btn-primary btn-sm bg-amber-600 hover:bg-amber-700"
                  >
                    Tiếp tục làm bài
                  </button>
                )}
                {(lookupResult.attempt_status === 'submitted' || lookupResult.attempt_status === 'time_up') && lookupResult.attempt && (
                  <button
                    onClick={() => navigate(`/student/attempts/${lookupResult.attempt.id}/result`)}
                    className="btn-secondary btn-sm"
                  >
                    Xem kết quả
                  </button>
                )}
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Section 3: Lịch sử làm bài gần đây */}
      <div>
        <div className="flex items-center justify-between mb-4 border-b border-slate-200 pb-3">
          <h2 className="text-lg font-bold text-slate-900">Lịch sử bài thi gần đây</h2>
          <button onClick={fetchAttempts} className="btn-secondary btn-sm">Làm mới</button>
        </div>

        {attemptsLoading ? (
          <div className="flex justify-center py-10"><Spinner size="lg" /></div>
        ) : attemptsError ? (
          <div className="p-4 rounded-md bg-red-50 border border-red-200 text-sm text-red-700">
            {attemptsError}
          </div>
        ) : attempts.length === 0 ? (
          <div className="text-center py-10 bg-white rounded-lg border border-slate-200 text-slate-500 text-sm">
            Chưa có bản ghi bài thi nào.
          </div>
        ) : (
          <div className="bg-white rounded-lg border border-slate-200 overflow-hidden shadow-sm">
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="bg-slate-50 border-b border-slate-200 text-left text-slate-600">
                    <th className="px-5 py-3 font-semibold">Tên bài thi</th>
                    <th className="px-5 py-3 font-semibold text-center">Mã đề</th>
                    <th className="px-5 py-3 font-semibold text-center">Điểm số</th>
                    <th className="px-5 py-3 font-semibold text-center">Đúng / Sai</th>
                    <th className="px-5 py-3 font-semibold text-center">Trạng thái</th>
                    <th className="px-5 py-3 font-semibold">Thời điểm bắt đầu</th>
                    <th className="px-5 py-3 font-semibold text-right">Thao tác</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {attempts.map((a) => (
                    <tr key={a.id} className="hover:bg-slate-50 transition-colors">
                      <td className="px-5 py-3.5 font-medium text-slate-900">{a.exam_title}</td>
                      <td className="px-5 py-3.5 text-center">
                        <code className="bg-slate-100 px-2 py-0.5 rounded font-mono text-slate-800 text-xs">{a.exam_code}</code>
                      </td>
                      <td className="px-5 py-3.5 text-center font-bold text-slate-900">
                        {a.status === 'completed' ? Number(a.score).toFixed(2) : '-'}
                      </td>
                      <td className="px-5 py-3.5 text-center text-xs">
                        {a.status === 'completed' ? (
                          <span>
                            <span className="text-emerald-700 font-semibold">{a.correct_count}</span>
                            <span className="text-slate-400"> / </span>
                            <span className="text-red-600 font-semibold">{a.wrong_count}</span>
                          </span>
                        ) : '-'}
                      </td>
                      <td className="px-5 py-3.5 text-center">
                        <span className={`px-2.5 py-0.5 rounded text-xs font-semibold ${
                          a.status === 'completed' ? 'bg-blue-100 text-blue-800' : 'bg-amber-100 text-amber-800'
                        }`}>
                          {a.status === 'completed' ? 'Đã nộp' : 'Đang làm'}
                        </span>
                      </td>
                      <td className="px-5 py-3.5 text-slate-500 text-xs">
                        {new Date(a.started_at).toLocaleString('vi-VN')}
                      </td>
                      <td className="px-5 py-3.5 text-right">
                        {a.status === 'completed' ? (
                          <button
                            onClick={() => navigate(`/student/attempts/${a.id}/result`)}
                            className="btn-secondary btn-sm text-xs"
                          >
                            Xem kết quả
                          </button>
                        ) : (
                          <button
                            onClick={() => navigate(`/student/exam/${a.exam_id}/take`)}
                            className="btn-primary btn-sm text-xs bg-amber-600 hover:bg-amber-700"
                          >
                            Làm tiếp
                          </button>
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

      {/* Modal: Xác nhận bắt đầu bài thi */}
      {startingExam && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4">
          <div className="w-full max-w-md rounded-lg bg-white p-6 shadow-xl border border-slate-200 space-y-4">
            <h2 className="text-lg font-bold text-slate-900">Xác nhận vào phòng thi</h2>
            <div className="text-sm text-slate-600 space-y-2 bg-slate-50 p-3.5 rounded border border-slate-100">
              <p>Môn thi: <strong className="text-slate-900">{startingExam.title}</strong></p>
              <p>Thời lượng làm bài: <strong className="text-slate-900">{startingExam.duration} phút</strong></p>
              <p className="text-xs text-amber-800">
                Lưu ý: Đồng hồ tính giờ sẽ bắt đầu đếm ngược ngay khi bạn xác nhận vào thi. Hãy đảm bảo đường truyền mạng ổn định.
              </p>
            </div>
            <div className="flex justify-end gap-2 pt-2 border-t border-slate-100">
              <button
                type="button"
                onClick={() => setStartingExam(null)}
                disabled={startingLoading}
                className="btn-secondary"
              >
                Hủy bỏ
              </button>
              <button
                type="button"
                onClick={() => handleStartExam(startingExam.id)}
                disabled={startingLoading}
                className="btn-primary"
              >
                {startingLoading ? <Spinner size="sm" /> : 'Bắt đầu làm bài'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
