import { useState, useEffect, useCallback, useMemo } from 'react';
import { Link } from 'react-router-dom';
import { scheduleAPI } from '../api/endpoints';
import { useAuth } from '../contexts/AuthContext';

function formatCountdown(totalSeconds) {
  if (totalSeconds <= 0) return '00:00:00';
  const days = Math.floor(totalSeconds / 86400);
  const hours = Math.floor((totalSeconds % 86400) / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;

  if (days > 0) {
    return `${days} ngày ${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
  }
  return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
}

function formatDateTime(isoString) {
  if (!isoString) return '';
  const d = new Date(isoString);
  return d.toLocaleString('vi-VN', {
    hour: '2-digit',
    minute: '2-digit',
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  });
}

export default function ExamSchedulePage() {
  const { user } = useAuth();
  const [data, setData] = useState({ ongoing: [], upcoming: [], past: [], summary: {} });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [activeTab, setActiveTab] = useState('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [now, setNow] = useState(Date.now());

  const fetchSchedule = useCallback(async () => {
    try {
      setLoading(true);
      setError('');
      const res = await scheduleAPI.getSchedule();
      setData(res.data);
    } catch {
      setError('Không thể tải lịch thi. Vui lòng thử lại sau.');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchSchedule();
  }, [fetchSchedule]);

  // Tick every second to update countdowns smoothly
  useEffect(() => {
    const timer = setInterval(() => {
      setNow(Date.now());
    }, 1000);
    return () => clearInterval(timer);
  }, []);

  const isTeacher = user?.role === 'teacher';

  const filterList = (list) => {
    if (!searchQuery.trim()) return list;
    const q = searchQuery.toLowerCase();
    return list.filter(
      (item) =>
        item.title.toLowerCase().includes(q) ||
        item.exam_code.toLowerCase().includes(q) ||
        (item.classroom_name && item.classroom_name.toLowerCase().includes(q)) ||
        (item.teacher_name && item.teacher_name.toLowerCase().includes(q))
    );
  };

  const ongoingList = useMemo(() => filterList(data.ongoing || []), [data.ongoing, searchQuery]);
  const upcomingList = useMemo(() => filterList(data.upcoming || []), [data.upcoming, searchQuery]);
  const pastList = useMemo(() => filterList(data.past || []), [data.past, searchQuery]);

  const renderCountdown = (item) => {
    if (item.time_status === 'upcoming') {
      const remainingSec = Math.max(0, Math.floor((new Date(item.start_time).getTime() - now) / 1000));
      return (
        <span className="font-mono text-indigo-700 bg-indigo-50 px-2 py-0.5 rounded font-semibold text-xs">
          Mở sau: {formatCountdown(remainingSec)}
        </span>
      );
    }
    if (item.time_status === 'ongoing') {
      const remainingSec = Math.max(0, Math.floor((new Date(item.end_time).getTime() - now) / 1000));
      return (
        <span className="font-mono text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded font-semibold text-xs">
          Còn lại: {formatCountdown(remainingSec)}
        </span>
      );
    }
    return (
      <span className="text-slate-500 bg-slate-100 px-2 py-0.5 rounded text-xs">
        Đã kết thúc
      </span>
    );
  };

  const renderCard = (item) => {
    return (
      <div
        key={item.id}
        className={`bg-white dark:bg-slate-900 rounded-xl border p-5 shadow-sm transition-shadow hover:shadow-md flex flex-col justify-between ${
          item.time_status === 'ongoing'
            ? 'border-emerald-300 dark:border-emerald-700 ring-1 ring-emerald-100 dark:ring-emerald-950'
            : item.time_status === 'upcoming'
            ? 'border-indigo-200 dark:border-indigo-800'
            : 'border-slate-200 dark:border-slate-800 opacity-90'
        }`}
      >
        <div>
          <div className="flex items-start justify-between gap-3 mb-2">
            <span className="text-xs font-mono font-bold px-2 py-0.5 bg-slate-100 text-slate-800 rounded border border-slate-200">
              {item.exam_code}
            </span>
            {renderCountdown(item)}
          </div>

          <h3 className="text-base font-bold text-slate-900 leading-snug mb-1">
            {item.title}
          </h3>

          {item.description && (
            <p className="text-xs text-slate-600 line-clamp-2 mb-3">
              {item.description}
            </p>
          )}

          <div className="space-y-1.5 text-xs text-slate-600 mt-3 border-t border-slate-100 pt-3">
            <div className="flex items-center justify-between">
              <span className="text-slate-500">Phạm vi:</span>
              <span className="font-medium text-slate-800">
                {item.classroom_name} {item.classroom_code && `(${item.classroom_code})`}
              </span>
            </div>
            {!isTeacher && (
              <div className="flex items-center justify-between">
                <span className="text-slate-500">Giảng viên:</span>
                <span className="font-medium text-slate-800">{item.teacher_name}</span>
              </div>
            )}
            <div className="flex items-center justify-between">
              <span className="text-slate-500">Thời lượng:</span>
              <span className="font-medium text-slate-800">{item.duration} phút</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-slate-500">Quy mô:</span>
              <span className="font-medium text-slate-800">
                {item.question_count} câu ({item.total_points} điểm)
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-slate-500">Thời gian mở:</span>
              <span className="font-medium text-slate-800 text-[11px]">
                {formatDateTime(item.start_time)}
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-slate-500">Thời gian đóng:</span>
              <span className="font-medium text-slate-800 text-[11px]">
                {formatDateTime(item.end_time)}
              </span>
            </div>
          </div>
        </div>

        <div className="mt-4 pt-3 border-t border-slate-100">
          {isTeacher ? (
            <div className="flex items-center gap-2">
              <Link
                to={`/teacher/exams/${item.id}`}
                className="flex-1 text-center text-xs font-semibold px-3 py-2 rounded bg-slate-100 hover:bg-slate-200 text-slate-800 transition-colors"
              >
                Cấu hình đề
              </Link>
              <Link
                to={`/teacher/exams/${item.id}/results`}
                className="flex-1 text-center text-xs font-semibold px-3 py-2 rounded bg-indigo-600 hover:bg-indigo-700 text-white transition-colors"
              >
                Kết quả ({item.total_attempts})
              </Link>
            </div>
          ) : (
            <div>
              {item.time_status === 'ongoing' && (
                <>
                  {item.attempt_status === 'submitted' ? (
                    <div className="flex items-center justify-between">
                      <span className="text-xs text-emerald-700 font-semibold">
                        Đã nộp bài (Điểm: {item.score ?? 'N/A'})
                      </span>
                      {item.attempt_id && (
                        <Link
                          to={`/student/attempts/${item.attempt_id}/result`}
                          className="text-xs font-medium text-indigo-600 hover:underline"
                        >
                          Xem chi tiết
                        </Link>
                      )}
                    </div>
                  ) : item.attempt_status === 'in_progress' ? (
                    <Link
                      to={`/student/exam/${item.id}/take`}
                      className="block text-center text-xs font-bold px-3 py-2 rounded bg-emerald-600 hover:bg-emerald-700 text-white transition-colors animate-pulse"
                    >
                      Tiếp tục làm bài
                    </Link>
                  ) : item.attempt_status === 'time_up' ? (
                    <div className="text-center text-xs text-slate-500 font-medium py-1">
                      Hết thời gian làm bài
                    </div>
                  ) : (
                    <Link
                      to={`/student/exam/${item.id}/take`}
                      className="block text-center text-xs font-bold px-3 py-2 rounded bg-indigo-600 hover:bg-indigo-700 text-white transition-colors"
                    >
                      Vào phòng thi ngay
                    </Link>
                  )}
                </>
              )}

              {item.time_status === 'upcoming' && (
                <div className="text-center text-xs text-slate-500 font-medium py-1.5 bg-slate-50 rounded border border-dashed border-slate-200">
                  Chưa tới giờ thi
                </div>
              )}

              {item.time_status === 'past' && (
                <>
                  {item.attempt_status === 'submitted' ? (
                    <div className="flex items-center justify-between">
                      <span className="text-xs text-slate-700 font-semibold">
                        Điểm: {item.score ?? 'N/A'}
                      </span>
                      {item.attempt_id && (
                        <Link
                          to={`/student/attempts/${item.attempt_id}/result`}
                          className="text-xs font-medium text-indigo-600 hover:underline"
                        >
                          Xem bài làm
                        </Link>
                      )}
                    </div>
                  ) : (
                    <div className="text-center text-xs text-slate-400 font-medium py-1">
                      Không tham gia ca thi này
                    </div>
                  )}
                </>
              )}
            </div>
          )}
        </div>
      </div>
    );
  };

  return (
    <div className="max-w-6xl mx-auto px-4 py-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-5 border-b border-slate-200">
        <div>
          <h1 className="text-xl font-bold text-slate-900">
            Lịch thi & Thời khóa biểu
          </h1>
          <p className="text-xs text-slate-500 mt-1">
            Theo dõi danh sách ca thi sắp diễn ra, đang mở và lịch sử kỳ thi
          </p>
        </div>

        <div className="flex items-center gap-2">
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Tìm theo tên, mã đề, lớp..."
            className="text-xs border border-slate-300 rounded-md px-3 py-1.5 focus:outline-none focus:ring-1 focus:ring-indigo-500 w-56"
          />
          <button
            type="button"
            onClick={fetchSchedule}
            className="text-xs font-medium px-3 py-1.5 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-md transition-colors"
          >
            Làm mới
          </button>
        </div>
      </div>

      {error && (
        <div className="mt-4 p-3 bg-red-50 border border-red-200 text-red-700 text-xs rounded-md">
          {error}
        </div>
      )}

      {/* Tabs / Filter Badges */}
      <div className="flex items-center gap-2 mt-4 border-b border-slate-200 pb-2">
        <button
          type="button"
          onClick={() => setActiveTab('all')}
          className={`text-xs font-semibold px-3 py-1.5 rounded-md transition-colors ${
            activeTab === 'all'
              ? 'bg-slate-900 text-white'
              : 'text-slate-600 hover:bg-slate-100'
          }`}
        >
          Tất cả ({data.summary?.total || 0})
        </button>
        <button
          type="button"
          onClick={() => setActiveTab('ongoing')}
          className={`text-xs font-semibold px-3 py-1.5 rounded-md transition-colors flex items-center gap-1.5 ${
            activeTab === 'ongoing'
              ? 'bg-emerald-600 text-white'
              : 'text-slate-600 hover:bg-slate-100'
          }`}
        >
          <span>Đang diễn ra</span>
          <span className={`text-[10px] px-1.5 py-0.2 rounded-full font-bold ${
            activeTab === 'ongoing' ? 'bg-emerald-700 text-white' : 'bg-emerald-100 text-emerald-800'
          }`}>
            {data.summary?.ongoing_count || 0}
          </span>
        </button>
        <button
          type="button"
          onClick={() => setActiveTab('upcoming')}
          className={`text-xs font-semibold px-3 py-1.5 rounded-md transition-colors flex items-center gap-1.5 ${
            activeTab === 'upcoming'
              ? 'bg-indigo-600 text-white'
              : 'text-slate-600 hover:bg-slate-100'
          }`}
        >
          <span>Sắp diễn ra</span>
          <span className={`text-[10px] px-1.5 py-0.2 rounded-full font-bold ${
            activeTab === 'upcoming' ? 'bg-indigo-700 text-white' : 'bg-indigo-100 text-indigo-800'
          }`}>
            {data.summary?.upcoming_count || 0}
          </span>
        </button>
        <button
          type="button"
          onClick={() => setActiveTab('past')}
          className={`text-xs font-semibold px-3 py-1.5 rounded-md transition-colors flex items-center gap-1.5 ${
            activeTab === 'past'
              ? 'bg-slate-700 text-white'
              : 'text-slate-600 hover:bg-slate-100'
          }`}
        >
          <span>Đã kết thúc</span>
          <span className={`text-[10px] px-1.5 py-0.2 rounded-full font-bold ${
            activeTab === 'past' ? 'bg-slate-800 text-white' : 'bg-slate-200 text-slate-800'
          }`}>
            {data.summary?.past_count || 0}
          </span>
        </button>
      </div>

      {loading ? (
        <div className="py-16 text-center text-xs text-slate-500">
          Đang tải lịch thi...
        </div>
      ) : (
        <div className="mt-6 space-y-8">
          {(activeTab === 'all' || activeTab === 'ongoing') && (
            <section>
              <div className="flex items-center gap-2 mb-3">
                <span className="w-2.5 h-2.5 rounded-full bg-emerald-500 animate-pulse" />
                <h2 className="text-sm font-bold text-slate-900 dark:text-slate-100 uppercase tracking-wide">
                  Ca thi đang diễn ra ({ongoingList.length})
                </h2>
              </div>
              {ongoingList.length === 0 ? (
                <div className="p-6 bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 text-center text-xs text-slate-500 dark:text-slate-400 shadow-sm">
                  Hiện tại không có ca thi nào đang diễn ra.
                </div>
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {ongoingList.map(renderCard)}
                </div>
              )}
            </section>
          )}

          {(activeTab === 'all' || activeTab === 'upcoming') && (
            <section>
              <div className="flex items-center gap-2 mb-3">
                <span className="w-2.5 h-2.5 rounded-full bg-indigo-500" />
                <h2 className="text-sm font-bold text-slate-900 dark:text-slate-100 uppercase tracking-wide">
                  Ca thi sắp diễn ra ({upcomingList.length})
                </h2>
              </div>
              {upcomingList.length === 0 ? (
                <div className="p-6 bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 text-center text-xs text-slate-500 dark:text-slate-400 shadow-sm">
                  Không có ca thi nào sắp diễn ra trong danh sách.
                </div>
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {upcomingList.map(renderCard)}
                </div>
              )}
            </section>
          )}

          {(activeTab === 'all' || activeTab === 'past') && (
            <section>
              <div className="flex items-center gap-2 mb-3">
                <span className="w-2.5 h-2.5 rounded-full bg-slate-400" />
                <h2 className="text-sm font-bold text-slate-900 dark:text-slate-100 uppercase tracking-wide">
                  Ca thi đã kết thúc ({pastList.length})
                </h2>
              </div>
              {pastList.length === 0 ? (
                <div className="p-6 bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 text-center text-xs text-slate-500 dark:text-slate-400 shadow-sm">
                  Chưa có ca thi nào trong danh sách kết thúc.
                </div>
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {pastList.map(renderCard)}
                </div>
              )}
            </section>
          )}
        </div>
      )}
    </div>
  );
}
