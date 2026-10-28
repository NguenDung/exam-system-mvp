import { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { teacherAPI } from '../../api/endpoints';
import Spinner from '../../components/Spinner';

function StatCard({ icon, label, value, color = 'indigo' }) {
  const colors = {
    indigo: 'bg-indigo-50 text-indigo-700 border-indigo-100',
    emerald: 'bg-emerald-50 text-emerald-700 border-emerald-100',
    amber: 'bg-amber-50 text-amber-700 border-amber-100',
    violet: 'bg-violet-50 text-violet-700 border-violet-100',
  };
  return (
    <div className={`card border ${colors[color]} text-center py-6`}>
      <div className="text-4xl mb-2">{icon}</div>
      <div className="text-3xl font-bold">{value}</div>
      <div className="text-sm mt-1 opacity-80">{label}</div>
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

  useEffect(() => { fetchResults(); }, [fetchResults]);

  if (loading) return <div className="flex justify-center py-20"><Spinner size="lg" /></div>;
  if (!data) return null;

  const { exam, statistics: stats, results } = data;

  return (
    <div>
      {/* Header */}
      <div className="flex items-start gap-4 mb-6 flex-wrap">
        <button onClick={() => navigate(`/teacher/exams/${id}`)} className="btn-secondary btn-sm mt-1">← Quay lại đề thi</button>
        <div className="flex-1">
          <h1 className="text-2xl font-bold text-gray-900">{exam.title}</h1>
          <p className="text-gray-500 mt-1">📊 Bảng điểm & Thống kê kết quả</p>
        </div>
        <button onClick={fetchResults} className="btn-secondary btn-sm">🔄 Làm mới</button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-8">
        <StatCard icon="👥" label="Tổng thí sinh" value={stats.total_students} color="indigo" />
        <StatCard icon="📈" label="Điểm trung bình" value={stats.total_students > 0 ? stats.avg_score.toFixed(2) : '—'} color="emerald" />
        <StatCard icon="🏆" label="Điểm cao nhất" value={stats.total_students > 0 ? stats.max_score.toFixed(2) : '—'} color="amber" />
        <StatCard icon="📉" label="Điểm thấp nhất" value={stats.total_students > 0 ? stats.min_score.toFixed(2) : '—'} color="violet" />
      </div>

      {/* Results table */}
      <div className="card overflow-hidden p-0">
        <div className="px-6 py-4 border-b flex items-center justify-between">
          <h3 className="font-semibold text-gray-800">Danh sách điểm thi</h3>
          <span className="text-sm text-gray-500">{results.length} học sinh</span>
        </div>
        {results.length === 0 ? (
          <div className="text-center py-16">
            <div className="text-4xl mb-3">📭</div>
            <p className="text-gray-500">Chưa có học sinh nào thi đề này</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-gray-50 border-b text-left">
                  <th className="px-6 py-3 font-semibold text-gray-600">#</th>
                  <th className="px-6 py-3 font-semibold text-gray-600">Họ tên</th>
                  <th className="px-6 py-3 font-semibold text-gray-600">Email</th>
                  <th className="px-6 py-3 font-semibold text-gray-600 text-center">Điểm</th>
                  <th className="px-6 py-3 font-semibold text-gray-600 text-center">Đúng</th>
                  <th className="px-6 py-3 font-semibold text-gray-600 text-center">Sai</th>
                  <th className="px-6 py-3 font-semibold text-gray-600 text-center">Trạng thái</th>
                  <th className="px-6 py-3 font-semibold text-gray-600">Nộp lúc</th>
                </tr>
              </thead>
              <tbody>
                {results.map((r, i) => (
                  <tr key={r.attempt_id} className="border-b last:border-0 hover:bg-gray-50 transition-colors">
                    <td className="px-6 py-4 text-gray-400 font-mono">{i + 1}</td>
                    <td className="px-6 py-4 font-medium text-gray-900">{r.full_name}</td>
                    <td className="px-6 py-4 text-gray-500">{r.email}</td>
                    <td className="px-6 py-4 text-center">
                      <span className="font-bold text-lg text-indigo-700">{r.score.toFixed(2)}</span>
                    </td>
                    <td className="px-6 py-4 text-center">
                      <span className="font-semibold text-emerald-600">{r.correct_count}</span>
                    </td>
                    <td className="px-6 py-4 text-center">
                      <span className="font-semibold text-red-500">{r.wrong_count}</span>
                    </td>
                    <td className="px-6 py-4 text-center">
                      <span className={r.status === 'completed' ? 'badge-completed' : 'badge-in_progress'}>
                        {r.status === 'completed' ? '✅ Đã nộp' : '⏳ Đang làm'}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-gray-500 text-xs">
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
