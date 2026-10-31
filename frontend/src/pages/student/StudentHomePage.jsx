import { useState, useEffect, useCallback, useMemo } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { studentAPI, scheduleAPI, classroomAPI } from '../../api/endpoints';
import Spinner from '../../components/Spinner';

const STATUS_LABELS = {
  not_attempted: { label: 'Chưa thi', badgeClass: 'bg-slate-100 text-slate-700 border-slate-200 dark:bg-slate-800 dark:text-slate-300 dark:border-slate-700' },
  open: { label: 'Đang mở', badgeClass: 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-800' },
  in_progress: { label: 'Đang làm dở', badgeClass: 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950/40 dark:text-amber-300 dark:border-amber-800' },
  submitted: { label: 'Đã nộp bài', badgeClass: 'bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950/40 dark:text-blue-300 dark:border-blue-800' },
  not_started_yet: { label: 'Chưa mở', badgeClass: 'bg-slate-100 text-slate-600 border-slate-200 dark:bg-slate-800 dark:text-slate-400 dark:border-slate-700' },
  expired: { label: 'Đã kết thúc', badgeClass: 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950/40 dark:text-rose-300 dark:border-rose-800' },
  time_up: { label: 'Hết giờ', badgeClass: 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950/40 dark:text-rose-300 dark:border-rose-800' },
};

export default function StudentHomePage() {
  const navigate = useNavigate();

  // Quick lookup state
  const [code, setCode] = useState('');
  const [lookupResult, setLookupResult] = useState(null);
  const [lookupLoading, setLookupLoading] = useState(false);
  const [lookupError, setLookupError] = useState('');

  // Available Exams State
  const [availableExams, setAvailableExams] = useState([]);
  const [availableLoading, setAvailableLoading] = useState(true);
  const [availableError, setAvailableError] = useState('');

  // Filter & Search
  const [activeTab, setActiveTab] = useState('all'); // all, open, in_progress, submitted
  const [searchQuery, setSearchQuery] = useState('');

  // History / Attempts State
  const [attempts, setAttempts] = useState([]);
  const [attemptsLoading, setAttemptsLoading] = useState(true);

  // Widget states: Schedule & Classes
  const [upcomingEvents, setUpcomingEvents] = useState([]);
  const [classes, setClasses] = useState([]);

  // Starting exam modal
  const [startingExam, setStartingExam] = useState(null);
  const [startingLoading, setStartingLoading] = useState(false);
  const [actionError, setActionError] = useState('');

  const fetchDashboardData = useCallback(async () => {
    setAvailableLoading(true);
    setAttemptsLoading(true);
    setAvailableError('');

    // Fetch exams
    try {
      const { data } = await studentAPI.getAvailableExams();
      setAvailableExams(data.exams || []);
    } catch (err) {
      setAvailableError(err.response?.data?.error || 'Không thể tải danh sách bài thi đang mở.');
    } finally {
      setAvailableLoading(false);
    }

    // Fetch attempts
    try {
      const { data } = await studentAPI.listAttempts();
      setAttempts(data.attempts || []);
    } catch {
      setAttempts([]);
    } finally {
      setAttemptsLoading(false);
    }

    // Fetch upcoming schedule
    try {
      const { data } = await scheduleAPI.getSchedule();
      const items = data.schedule || data.events || [];
      setUpcomingEvents(Array.isArray(items) ? items.slice(0, 3) : []);
    } catch {
      setUpcomingEvents([]);
    }

    // Fetch classrooms
    try {
      const { data } = await classroomAPI.list();
      const list = data.classrooms || data.classes || [];
      setClasses(Array.isArray(list) ? list.slice(0, 3) : []);
    } catch {
      setClasses([]);
    }
  }, []);

  useEffect(() => {
    fetchDashboardData();
  }, [fetchDashboardData]);

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

  // Filtered exams
  const filteredExams = useMemo(() => {
    return availableExams.filter((ex) => {
      // Tab filter
      if (activeTab === 'open' && ex.attempt_status !== 'not_attempted') return false;
      if (activeTab === 'in_progress' && ex.attempt_status !== 'in_progress') return false;
      if (activeTab === 'submitted' && ex.attempt_status !== 'submitted' && ex.attempt_status !== 'time_up') return false;

      // Text search
      if (searchQuery.trim()) {
        const q = searchQuery.toLowerCase().trim();
        const matchTitle = ex.title?.toLowerCase().includes(q);
        const matchCode = ex.exam_code?.toLowerCase().includes(q);
        return matchTitle || matchCode;
      }
      return true;
    });
  }, [availableExams, activeTab, searchQuery]);

  // Calculations for stats widget
  const stats = useMemo(() => {
    const completed = attempts.filter((a) => a.status === 'completed');
    const totalCompleted = completed.length;
    const avgScore = totalCompleted > 0
      ? (completed.reduce((acc, cur) => acc + (Number(cur.score) || 0), 0) / totalCompleted).toFixed(1)
      : '—';
    const passed = completed.filter((a) => (Number(a.score) || 0) >= 5).length;
    const passRate = totalCompleted > 0 ? Math.round((passed / totalCompleted) * 100) : 0;

    return { totalCompleted, avgScore, passRate };
  }, [attempts]);

  return (
    <div className="space-y-6">
      {/* Top Banner / Alert */}
      {actionError && (
        <div className="rounded-lg bg-rose-50 dark:bg-rose-950/40 border border-rose-200 dark:border-rose-800 p-3.5 text-sm text-rose-800 dark:text-rose-300 flex items-center justify-between shadow-sm">
          <span>{actionError}</span>
          <button
            onClick={() => setActionError('')}
            className="text-rose-500 hover:text-rose-700 text-lg leading-none font-bold ml-2"
          >
            &times;
          </button>
        </div>
      )}

      {/* Main 2-Column Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left Column: Primary Workspace (2 Cols) */}
        <div className="lg:col-span-2 space-y-6">
          {/* Quick Exam Code Entry Card */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-5 shadow-sm">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-3">
              <div>
                <h2 className="text-sm font-bold text-slate-900 dark:text-slate-100 tracking-tight">Vào thi nhanh bằng mã đề</h2>
                <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
                  Nhập mã phòng thi (ví dụ: <span className="font-mono font-semibold text-slate-700 dark:text-slate-300">MIDTERM-01</span>) do giảng viên cung cấp
                </p>
              </div>
              <span className="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-medium bg-blue-50 text-blue-700 border border-blue-200 dark:bg-blue-950/60 dark:text-blue-300 dark:border-blue-800 self-start sm:self-auto">
                Khảo thí bảo mật
              </span>
            </div>

            <form onSubmit={handleLookup} className="flex gap-2">
              <div className="relative flex-1">
                <input
                  value={code}
                  onChange={(e) => {
                    setCode(e.target.value.toUpperCase());
                    setLookupResult(null);
                    setLookupError('');
                  }}
                  className="input font-mono text-sm tracking-wider uppercase"
                  placeholder="Nhập mã đề thi..."
                  maxLength={20}
                />
              </div>
              <button
                type="submit"
                disabled={lookupLoading || !code.trim()}
                className="btn-primary h-10 px-5 text-sm font-medium flex items-center gap-2"
              >
                {lookupLoading ? <Spinner size="sm" /> : 'Tra cứu'}
              </button>
            </form>

            {lookupError && (
              <div className="mt-3 p-3 bg-rose-50 dark:bg-rose-950/40 border border-rose-200 dark:border-rose-800 rounded-lg text-rose-700 dark:text-rose-300 text-xs">
                {lookupError}
              </div>
            )}

            {lookupResult && (
              <div className="mt-4 p-4 bg-slate-50/80 dark:bg-slate-800/60 border border-slate-200 dark:border-slate-700 rounded-lg">
                <div className="flex items-start justify-between gap-4 flex-wrap">
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <span className="font-mono text-xs font-bold bg-blue-100 text-blue-800 dark:bg-blue-950 dark:text-blue-300 px-2 py-0.5 rounded">
                        {lookupResult.exam?.exam_code}
                      </span>
                      <h3 className="font-bold text-slate-900 dark:text-slate-100 text-sm">{lookupResult.exam?.title}</h3>
                    </div>
                    <div className="flex items-center gap-4 text-xs text-slate-600 dark:text-slate-400 flex-wrap pt-1">
                      <span>Thời lượng: <strong>{lookupResult.exam?.duration} phút</strong></span>
                      <span>Hạn nộp: {new Date(lookupResult.exam?.end_time).toLocaleString('vi-VN')}</span>
                      <span>
                        Trạng thái:{' '}
                        <span className="font-semibold text-slate-800 dark:text-slate-200">
                          {STATUS_LABELS[lookupResult.attempt_status]?.label || 'Chưa làm'}
                        </span>
                      </span>
                    </div>
                  </div>

                  <div className="pt-1">
                    {lookupResult.attempt_status === 'not_attempted' && lookupResult.exam_status === 'open' && (
                      <button
                        onClick={() => setStartingExam(lookupResult.exam)}
                        className="btn-primary btn-sm text-xs"
                      >
                        Vào thi ngay
                      </button>
                    )}
                    {lookupResult.attempt_status === 'in_progress' && (
                      <button
                        onClick={() => navigate(`/student/exam/${lookupResult.exam.id}/take`)}
                        className="btn-primary btn-sm text-xs bg-amber-600 hover:bg-amber-700"
                      >
                        Tiếp tục làm bài
                      </button>
                    )}
                    {(lookupResult.attempt_status === 'submitted' || lookupResult.attempt_status === 'time_up') && lookupResult.attempt && (
                      <button
                        onClick={() => navigate(`/student/attempts/${lookupResult.attempt.id}/result`)}
                        className="btn-secondary btn-sm text-xs"
                      >
                        Xem kết quả
                      </button>
                    )}
                  </div>
                </div>
              </div>
            )}
          </div>

          {/* Exam Hub & Filter Tabs */}
          <div className="space-y-4">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <div>
                <h1 className="text-base font-bold text-slate-900 dark:text-slate-100 tracking-tight">Đề thi của tôi</h1>
                <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">Danh sách các bài khảo thí phân quyền cho tài khoản của bạn</p>
              </div>

              {/* Action: Refresh */}
              <button
                onClick={fetchDashboardData}
                className="btn-secondary btn-sm self-start sm:self-auto text-xs"
              >
                Làm mới dữ liệu
              </button>
            </div>

            {/* Filter Bar: Tabs + Search input */}
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 bg-white dark:bg-slate-900 p-2.5 rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm">
              <div className="flex items-center gap-1 overflow-x-auto pb-1 sm:pb-0">
                {[
                  { id: 'all', label: 'Tất cả' },
                  { id: 'open', label: 'Chưa làm' },
                  { id: 'in_progress', label: 'Đang làm dở' },
                  { id: 'submitted', label: 'Đã nộp bài' },
                ].map((tab) => (
                  <button
                    key={tab.id}
                    onClick={() => setActiveTab(tab.id)}
                    className={`px-3 py-1.5 rounded-lg text-xs font-semibold whitespace-nowrap transition-colors ${
                      activeTab === tab.id
                        ? 'bg-blue-600 text-white shadow-sm'
                        : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100 hover:bg-slate-100 dark:hover:bg-slate-800'
                    }`}
                  >
                    {tab.label}
                  </button>
                ))}
              </div>

              <div className="w-full sm:w-60">
                <input
                  type="text"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  placeholder="Lọc theo tên hoặc mã đề..."
                  className="input h-8 px-3 text-xs"
                />
              </div>
            </div>

            {/* Exam Cards Grid */}
            {availableLoading ? (
              <div className="flex justify-center py-16 bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800">
                <Spinner size="lg" />
              </div>
            ) : availableError ? (
              <div className="p-4 rounded-xl bg-rose-50 dark:bg-rose-950/40 border border-rose-200 dark:border-rose-800 text-sm text-rose-700 dark:text-rose-300">
                {availableError}
              </div>
            ) : filteredExams.length === 0 ? (
              <div className="text-center py-12 bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 text-slate-500 dark:text-slate-400 text-xs shadow-sm">
                Không tìm thấy bài thi nào phù hợp với bộ lọc hiện tại.
              </div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {filteredExams.map((ex) => {
                  const statusCfg = STATUS_LABELS[ex.attempt_status] || STATUS_LABELS.not_attempted;
                  const isNotAttempted = ex.attempt_status === 'not_attempted';
                  const isInProgress = ex.attempt_status === 'in_progress';
                  const isFinished = ex.attempt_status === 'submitted' || ex.attempt_status === 'time_up';

                  return (
                    <div
                      key={ex.id}
                      className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4 shadow-sm flex flex-col justify-between hover:border-blue-400 dark:hover:border-blue-600 hover:shadow transition-all"
                    >
                      <div className="space-y-3">
                        <div className="flex items-center justify-between gap-2">
                          <span className={`px-2 py-0.5 rounded text-[11px] font-semibold border ${statusCfg.badgeClass}`}>
                            {statusCfg.label}
                          </span>
                          <span className="font-mono text-xs font-bold bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 px-2 py-0.5 rounded border border-slate-200 dark:border-slate-700">
                            {ex.exam_code}
                          </span>
                        </div>

                        <div>
                          <h3 className="font-bold text-slate-900 dark:text-slate-100 text-sm line-clamp-1">{ex.title}</h3>
                          {ex.description && (
                            <p className="text-xs text-slate-500 dark:text-slate-400 mt-1 line-clamp-2">{ex.description}</p>
                          )}
                        </div>

                        <div className="space-y-1.5 text-xs text-slate-600 dark:text-slate-400 bg-slate-50/70 dark:bg-slate-800/60 p-2.5 rounded-lg border border-slate-100 dark:border-slate-800">
                          <div className="flex justify-between">
                            <span className="text-slate-500 dark:text-slate-400">Thời lượng:</span>
                            <strong className="text-slate-800 dark:text-slate-200">{ex.duration} phút</strong>
                          </div>
                          <div className="flex justify-between">
                            <span className="text-slate-500 dark:text-slate-400">Số lượng:</span>
                            <strong className="text-slate-800 dark:text-slate-200">{ex.questions_count} câu ({ex.total_points} điểm)</strong>
                          </div>
                          <div className="flex justify-between">
                            <span className="text-slate-500 dark:text-slate-400">Hạn chót:</span>
                            <span className="text-slate-700 dark:text-slate-300 font-medium">
                              {new Date(ex.end_time).toLocaleDateString('vi-VN', {
                                hour: '2-digit',
                                minute: '2-digit',
                                day: '2-digit',
                                month: '2-digit',
                              })}
                            </span>
                          </div>
                        </div>
                      </div>

                      <div className="pt-3.5 border-t border-slate-100 dark:border-slate-800 mt-3">
                        {isNotAttempted && (
                          <button
                            onClick={() => setStartingExam(ex)}
                            className="w-full btn-primary text-xs py-2"
                          >
                            Bắt đầu làm bài
                          </button>
                        )}

                        {isInProgress && (
                          <button
                            onClick={() => navigate(`/student/exam/${ex.id}/take`)}
                            className="w-full btn-primary text-xs py-2 bg-amber-600 hover:bg-amber-700 focus:ring-amber-500"
                          >
                            Tiếp tục làm bài
                          </button>
                        )}

                        {isFinished && (
                          <button
                            onClick={() => navigate(`/student/attempts/${ex.attempt_id}/result`)}
                            className="w-full btn-secondary text-xs py-2 font-medium"
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

          {/* Recent Attempts Quick View */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 overflow-hidden shadow-sm">
            <div className="px-5 py-3.5 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between">
              <div>
                <h2 className="text-sm font-bold text-slate-900 dark:text-slate-100">Bài thi đã làm gần đây</h2>
                <p className="text-[11px] text-slate-500 dark:text-slate-400">Xem lại các lần làm bài và kết quả khảo thí</p>
              </div>
              <Link
                to="/student/history"
                className="text-xs font-semibold text-blue-600 dark:text-blue-400 hover:text-blue-700 transition-colors"
              >
                Xem tất cả lịch sử &rarr;
              </Link>
            </div>

            {attemptsLoading ? (
              <div className="py-8 flex justify-center"><Spinner size="md" /></div>
            ) : attempts.length === 0 ? (
              <div className="p-6 text-center text-xs text-slate-500 dark:text-slate-400">
                Chưa có bản ghi kết quả nào. Hãy hoàn thành bài thi đầu tiên của bạn!
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-xs">
                  <thead>
                    <tr className="bg-slate-50 dark:bg-slate-800/50 border-b border-slate-200 dark:border-slate-800 text-left text-slate-600 dark:text-slate-400 font-semibold">
                      <th className="px-4 py-2.5">Tên đề thi</th>
                      <th className="px-4 py-2.5 text-center">Mã đề</th>
                      <th className="px-4 py-2.5 text-center">Điểm số</th>
                      <th className="px-4 py-2.5 text-center">Trạng thái</th>
                      <th className="px-4 py-2.5 text-right">Chi tiết</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
                    {attempts.slice(0, 4).map((a) => (
                      <tr key={a.id} className="hover:bg-slate-50/70 dark:hover:bg-slate-800/40 transition-colors">
                        <td className="px-4 py-3 font-medium text-slate-900 dark:text-slate-100 max-w-[180px] truncate">
                          {a.exam_title}
                        </td>
                        <td className="px-4 py-3 text-center">
                          <code className="font-mono bg-slate-100 dark:bg-slate-800 px-1.5 py-0.5 rounded text-slate-700 dark:text-slate-300">
                            {a.exam_code}
                          </code>
                        </td>
                        <td className="px-4 py-3 text-center font-bold text-slate-900 dark:text-slate-100">
                          {a.status === 'completed' ? Number(a.score).toFixed(2) : '—'}
                        </td>
                        <td className="px-4 py-3 text-center">
                          <span
                            className={`px-2 py-0.5 rounded font-semibold text-[11px] ${
                              a.status === 'completed'
                                ? 'bg-blue-50 text-blue-700 border border-blue-200 dark:bg-blue-950 dark:text-blue-300 dark:border-blue-800'
                                : 'bg-amber-50 text-amber-700 border border-amber-200 dark:bg-amber-950 dark:text-amber-300 dark:border-amber-800'
                            }`}
                          >
                            {a.status === 'completed' ? 'Đã nộp' : 'Đang làm'}
                          </span>
                        </td>
                        <td className="px-4 py-3 text-right">
                          {a.status === 'completed' ? (
                            <Link
                              to={`/student/attempts/${a.id}/result`}
                              className="text-blue-600 dark:text-blue-400 hover:text-blue-800 font-semibold"
                            >
                              Xem điểm
                            </Link>
                          ) : (
                            <Link
                              to={`/student/exam/${a.exam_id}/take`}
                              className="text-amber-600 dark:text-amber-400 hover:text-amber-800 font-semibold"
                            >
                              Làm tiếp
                            </Link>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </div>

        {/* Right Column: Widgets (1 Col) */}
        <div className="space-y-5">
          {/* Widget 1: Personal Performance / Stats */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-5 shadow-sm">
            <h3 className="text-xs font-bold text-slate-400 dark:text-slate-500 uppercase tracking-wider mb-3">
              Tiến độ học tập
            </h3>
            <div className="grid grid-cols-2 gap-3 mb-4">
              <div className="p-3 rounded-lg bg-slate-50 dark:bg-slate-800/60 border border-slate-100 dark:border-slate-800 text-center">
                <div className="text-2xl font-bold text-slate-900 dark:text-slate-100">{stats.totalCompleted}</div>
                <div className="text-[11px] text-slate-500 dark:text-slate-400 font-medium mt-0.5">Bài đã hoàn thành</div>
              </div>
              <div className="p-3 rounded-lg bg-blue-50/60 dark:bg-blue-950/40 border border-blue-100 dark:border-blue-900 text-center">
                <div className="text-2xl font-bold text-blue-700 dark:text-blue-300">{stats.avgScore}</div>
                <div className="text-[11px] text-blue-600 dark:text-blue-400 font-medium mt-0.5">Điểm trung bình</div>
              </div>
            </div>

            <div className="p-3 rounded-lg bg-slate-50 dark:bg-slate-800/60 border border-slate-100 dark:border-slate-800 space-y-2">
              <div className="flex items-center justify-between text-xs">
                <span className="text-slate-600 dark:text-slate-300 font-medium">Tỷ lệ đạt (&ge; 5.0):</span>
                <span className="font-bold text-slate-800 dark:text-slate-200">{stats.passRate}%</span>
              </div>
              <div className="w-full bg-slate-200 dark:bg-slate-700 h-1.5 rounded-full overflow-hidden">
                <div
                  className="bg-blue-600 h-full rounded-full transition-all duration-500"
                  style={{ width: `${stats.passRate}%` }}
                />
              </div>
            </div>

            <div className="mt-4 pt-3 border-t border-slate-100 dark:border-slate-800">
              <Link
                to="/student/grades"
                className="w-full flex items-center justify-center gap-1.5 py-2 px-3 rounded-lg bg-blue-50 dark:bg-blue-950/60 hover:bg-blue-100 dark:hover:bg-blue-900 text-blue-700 dark:text-blue-300 text-xs font-semibold transition-colors border border-blue-200 dark:border-blue-800"
              >
                <span>Xem bảng điểm &amp; GPA tích lũy &rarr;</span>
              </Link>
            </div>
          </div>

          {/* Widget 2: Upcoming Schedule */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-5 shadow-sm">
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-xs font-bold text-slate-400 dark:text-slate-500 uppercase tracking-wider">
                Lịch thi sắp tới
              </h3>
              <Link to="/schedule" className="text-[11px] font-semibold text-blue-600 dark:text-blue-400 hover:text-blue-700">
                Toàn bộ &rarr;
              </Link>
            </div>

            {upcomingEvents.length === 0 ? (
              <div className="p-4 text-center text-xs text-slate-500 dark:text-slate-400 bg-slate-50 dark:bg-slate-800/40 rounded-lg border border-slate-100 dark:border-slate-800">
                Không có lịch thi nào được lên kế hoạch trong tuần này.
              </div>
            ) : (
              <div className="space-y-2.5">
                {upcomingEvents.map((ev, idx) => (
                  <div
                    key={idx}
                    className="p-3 rounded-lg bg-slate-50 dark:bg-slate-800/60 border border-slate-100 dark:border-slate-800 hover:border-slate-200 dark:hover:border-slate-700 transition-colors"
                  >
                    <div className="font-semibold text-xs text-slate-900 dark:text-slate-100 line-clamp-1">
                      {ev.title || ev.exam_title || 'Bài thi định kỳ'}
                    </div>
                    <div className="text-[11px] text-slate-500 dark:text-slate-400 mt-1 flex items-center justify-between">
                      <span>{ev.start_time ? new Date(ev.start_time).toLocaleDateString('vi-VN') : 'Sắp diễn ra'}</span>
                      <span className="font-mono text-[10px] bg-slate-200 dark:bg-slate-700 px-1 rounded text-slate-700 dark:text-slate-300">
                        {ev.exam_code || 'EXAM'}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Widget 3: Enrolled Classes */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-5 shadow-sm">
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-xs font-bold text-slate-400 dark:text-slate-500 uppercase tracking-wider">
                Lớp học phần
              </h3>
              <Link to="/classes" className="text-[11px] font-semibold text-blue-600 dark:text-blue-400 hover:text-blue-700">
                Xem tất cả &rarr;
              </Link>
            </div>

            {classes.length === 0 ? (
              <div className="p-4 text-center text-xs text-slate-500 dark:text-slate-400 bg-slate-50 dark:bg-slate-800/40 rounded-lg border border-slate-100 dark:border-slate-800">
                Chưa có lớp học phần nào. Tham gia lớp bằng mã lớp tại trang Lớp học.
              </div>
            ) : (
              <div className="space-y-2">
                {classes.map((c) => (
                  <Link
                    key={c.id}
                    to={`/classes/${c.id}`}
                    className="block p-2.5 rounded-lg border border-slate-100 dark:border-slate-800 hover:bg-slate-50 dark:hover:bg-slate-800/60 hover:border-slate-200 dark:hover:border-slate-700 transition-colors"
                  >
                    <div className="font-semibold text-xs text-slate-800 dark:text-slate-200 line-clamp-1">{c.name}</div>
                    <div className="text-[11px] text-slate-400 font-mono mt-0.5">Mã: {c.code}</div>
                  </Link>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Modal: Xác nhận bắt đầu bài thi */}
      {startingExam && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 p-4 backdrop-blur-xs">
          <div className="w-full max-w-md rounded-xl bg-white dark:bg-slate-900 p-6 shadow-xl border border-slate-200 dark:border-slate-800 space-y-4">
            <h2 className="text-base font-bold text-slate-900 dark:text-slate-100">Xác nhận vào phòng thi</h2>
            <div className="text-xs text-slate-600 dark:text-slate-300 space-y-2 bg-slate-50 dark:bg-slate-800/60 p-3.5 rounded-lg border border-slate-200 dark:border-slate-700">
              <p>Môn thi: <strong className="text-slate-900 dark:text-slate-100">{startingExam.title}</strong></p>
              <p>Thời lượng làm bài: <strong className="text-slate-900 dark:text-slate-100">{startingExam.duration} phút</strong></p>
              <p>Số lượng câu hỏi: <strong className="text-slate-900 dark:text-slate-100">{startingExam.questions_count} câu</strong></p>
              <p className="text-amber-800 dark:text-amber-400 font-medium pt-1 border-t border-slate-200 dark:border-slate-700">
                Lưu ý: Đồng hồ tính giờ sẽ bắt đầu đếm ngược ngay khi bạn xác nhận vào thi. Hãy đảm bảo kết nối mạng ổn định.
              </p>
            </div>
            <div className="flex justify-end gap-2 pt-2 border-t border-slate-100 dark:border-slate-800">
              <button
                type="button"
                onClick={() => setStartingExam(null)}
                disabled={startingLoading}
                className="btn-secondary text-xs"
              >
                Hủy bỏ
              </button>
              <button
                type="button"
                onClick={() => handleStartExam(startingExam.id)}
                disabled={startingLoading}
                className="btn-primary text-xs"
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
