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

export default function TeacherResultsPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);

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
                  <th className="px-5 py-3 font-semibold">Thời gian nộp</th>
                </tr>
              </thead>
              <tbody>
                {results.map((r, i) => (
                  <tr key={r.attempt_id} className="border-b border-slate-100 last:border-0 hover:bg-slate-50 transition-colors">
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
                    <td className="px-5 py-3.5 text-slate-500 text-xs">
                      {r.submitted_at ? new Date(r.submitted_at).toLocaleString('vi-VN') : '—'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
