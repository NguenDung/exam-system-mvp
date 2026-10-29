import { useCallback, useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { studentAPI } from '../../api/endpoints';
import Spinner from '../../components/Spinner';

const OPTIONS = ['A', 'B', 'C', 'D'];
const APPEAL_STATUS = {
  PENDING: 'Chờ xử lý',
  REPLIED: 'Đã phản hồi',
  REJECTED: 'Đã từ chối',
  RESOLVED: 'Đã xử lý',
};

function formatDate(value) {
  return value ? new Date(value).toLocaleString('vi-VN') : '—';
}

export default function StudentHistoryDetailPage() {
  const { submissionId } = useParams();
  const navigate = useNavigate();
  const [detail, setDetail] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const loadDetail = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const { data } = await studentAPI.getHistoryDetail(submissionId);
      setDetail(data);
    } catch (requestError) {
      setError(requestError.response?.data?.error || 'Không thể tải chi tiết bài làm.');
    } finally {
      setLoading(false);
    }
  }, [submissionId]);

  useEffect(() => { loadDetail(); }, [loadDetail]);

  if (loading) return <div className="flex justify-center py-20"><Spinner size="lg" /></div>;
  if (error) return (
    <div role="alert" className="card mx-auto max-w-xl py-12 text-center">
      <p className="text-red-600">{error}</p>
      <button onClick={() => navigate('/student/history')} className="btn-primary mt-4">Về lịch sử bài làm</button>
    </div>
  );
  if (!detail) return null;

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center gap-3">
        <button onClick={() => navigate('/student/history')} className="btn-secondary btn-sm">← Lịch sử bài làm</button>
        <h1 className="text-2xl font-bold text-gray-900">{detail.exam.title}</h1>
      </div>

      <section className="card grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
        <div><p className="text-xs text-gray-500">Điểm đạt được</p><p className="mt-1 text-2xl font-bold text-indigo-700">{Number(detail.score).toFixed(2)}/{Number(detail.max_score).toFixed(2)}</p></div>
        <div><p className="text-xs text-gray-500">Thời gian bắt đầu</p><p className="mt-1 font-medium">{formatDate(detail.started_at)}</p></div>
        <div><p className="text-xs text-gray-500">Thời gian kết thúc</p><p className="mt-1 font-medium">{formatDate(detail.submitted_at)}</p></div>
        <div><p className="text-xs text-gray-500">Mã bài thi</p><p className="mt-1 font-mono font-semibold text-indigo-700">{detail.exam.exam_code}</p></div>
      </section>

      <div>
        <h2 className="mb-4 text-lg font-bold text-gray-900">Chi tiết câu trả lời</h2>
        {detail.answers.length === 0 ? (
          <div className="card py-10 text-center text-gray-500">Bài thi không có câu hỏi.</div>
        ) : (
          <div className="space-y-4">
            {detail.answers.map((answer, index) => (
              <article key={answer.question_id} className={`card border-2 ${answer.is_correct ? 'border-emerald-200' : 'border-red-200'}`}>
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div>
                    <p className="font-semibold text-gray-900">Câu {index + 1}: {answer.prompt}</p>
                    <p className="mt-1 text-xs text-gray-500">{answer.points} điểm</p>
                  </div>
                  <span className={`rounded-full px-3 py-1 text-sm font-semibold ${answer.is_correct ? 'bg-emerald-100 text-emerald-700' : 'bg-red-100 text-red-700'}`}>
                    {answer.is_correct ? '✅ Đúng' : '❌ Sai'}
                  </span>
                </div>
                <div className="mt-4 grid gap-2 sm:grid-cols-2">
                  {OPTIONS.map((option) => {
                    const isCorrect = answer.correct_option === option;
                    const isChosen = answer.chosen_option === option;
                    return (
                      <div key={option} className={`rounded-lg border p-3 text-sm ${
                        isCorrect ? 'border-emerald-300 bg-emerald-50 text-emerald-900' :
                          isChosen ? 'border-red-300 bg-red-50 text-red-900' : 'border-gray-200 text-gray-700'
                      }`}>
                        <span className="mr-2 font-bold">{option}.</span>
                        {answer[`option_${option.toLowerCase()}`]}
                        {isChosen && <span className="ml-2 text-xs font-semibold">(Bạn chọn)</span>}
                        {isCorrect && <span className="ml-2 text-xs font-semibold">(Đáp án đúng)</span>}
                      </div>
                    );
                  })}
                </div>
                <div className="mt-4 flex flex-wrap gap-x-6 gap-y-1 text-sm">
                  <span>Student chọn: <strong>{answer.chosen_option || 'Bỏ qua'}</strong></span>
                  <span>Đáp án đúng: <strong className="text-emerald-700">{answer.correct_option}</strong></span>
                </div>
              </article>
            ))}
          </div>
        )}
      </div>

      {detail.appeals?.length > 0 && (
        <section className="card">
          <h2 className="mb-3 font-bold text-gray-900">Khiếu nại đã gửi</h2>
          <div className="space-y-3">
            {detail.appeals.map((appeal) => (
              <div key={appeal.id} className="rounded-xl border border-gray-200 p-4">
                <div className="flex flex-wrap justify-between gap-2">
                  <strong>{appeal.title || 'Khiếu nại điểm thi'}</strong>
                  <span className="text-sm text-indigo-700">{APPEAL_STATUS[appeal.status] || appeal.status}</span>
                </div>
                <p className="mt-2 whitespace-pre-wrap text-sm text-gray-600">{appeal.content}</p>
                {appeal.teacher_response && (
                  <p className="mt-3 rounded-lg bg-indigo-50 p-3 text-sm text-indigo-900">
                    Phản hồi giáo viên: {appeal.teacher_response}
                  </p>
                )}
              </div>
            ))}
          </div>
        </section>
      )}
    </div>
  );
}
