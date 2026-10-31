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
  return value ? new Date(value).toLocaleString('vi-VN') : '-';
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

  useEffect(() => {
    loadDetail();
  }, [loadDetail]);

  if (loading) return <div className="flex justify-center py-20"><Spinner size="lg" /></div>;
  if (error) {
    return (
      <div role="alert" className="bg-white dark:bg-slate-900 rounded-xl border border-red-200 dark:border-rose-900/60 p-8 mx-auto max-w-xl text-center shadow-sm">
        <p className="text-red-700 dark:text-rose-300 text-sm">{error}</p>
        <button onClick={() => navigate('/student/history')} className="btn-primary mt-4">
          Quay lại lịch sử bài làm
        </button>
      </div>
    );
  }
  if (!detail) return null;

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center gap-3">
        <button onClick={() => navigate('/student/history')} className="btn-secondary btn-sm">
          Quay lại danh sách
        </button>
        <h1 className="text-xl font-bold text-slate-900 dark:text-slate-100">{detail.exam.title}</h1>
      </div>

      <section className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-5 grid gap-5 sm:grid-cols-2 lg:grid-cols-4 shadow-sm">
        <div>
          <p className="text-xs text-slate-500 dark:text-slate-400 uppercase tracking-wide">Điểm đạt được</p>
          <p className="mt-1 text-2xl font-bold text-slate-900 dark:text-slate-100">
            {Number(detail.score).toFixed(2)} / {Number(detail.max_score).toFixed(2)}
          </p>
        </div>
        <div>
          <p className="text-xs text-slate-500 dark:text-slate-400 uppercase tracking-wide">Thời gian bắt đầu</p>
          <p className="mt-1 text-sm font-medium text-slate-800 dark:text-slate-200">{formatDate(detail.started_at)}</p>
        </div>
        <div>
          <p className="text-xs text-slate-500 dark:text-slate-400 uppercase tracking-wide">Thời gian nộp bài</p>
          <p className="mt-1 text-sm font-medium text-slate-800 dark:text-slate-200">{formatDate(detail.submitted_at)}</p>
        </div>
        <div>
          <p className="text-xs text-slate-500 dark:text-slate-400 uppercase tracking-wide">Mã đề</p>
          <p className="mt-1 font-mono text-sm font-bold text-slate-900 dark:text-slate-100">{detail.exam.exam_code}</p>
        </div>
      </section>

      <div>
        <h2 className="mb-4 text-base font-bold text-slate-900 dark:text-slate-100 uppercase tracking-wide">Chi tiết bài làm</h2>
        {detail.answers.length === 0 ? (
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-8 text-center text-slate-500 dark:text-slate-400 shadow-sm">
            Bài thi không có câu hỏi.
          </div>
        ) : (
          <div className="space-y-4">
            {detail.answers.map((answer, index) => (
              <article
                key={answer.question_id}
                className={`bg-white dark:bg-slate-900 rounded-xl border p-5 shadow-sm ${
                  answer.is_correct ? 'border-emerald-200 dark:border-emerald-800/80' : 'border-red-200 dark:border-rose-800/80'
                }`}
              >
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div>
                    <p className="font-semibold text-slate-900 dark:text-slate-100 text-sm sm:text-base">
                      Câu {index + 1}: {answer.prompt}
                    </p>
                    <p className="mt-0.5 text-xs text-slate-500 dark:text-slate-400">Trọng số: {answer.points} điểm</p>
                  </div>
                  <span
                    className={`rounded px-2.5 py-0.5 text-xs font-semibold ${
                      answer.is_correct
                        ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950/60 dark:text-emerald-300 dark:border dark:border-emerald-800'
                        : 'bg-red-100 text-red-800 dark:bg-rose-950/60 dark:text-rose-300 dark:border dark:border-rose-800'
                    }`}
                  >
                    {answer.is_correct ? 'Đúng' : 'Chưa chính xác'}
                  </span>
                </div>
                <div className="mt-4 grid gap-2 sm:grid-cols-2">
                  {OPTIONS.map((option) => {
                    const isCorrect = answer.correct_option === option;
                    const isChosen = answer.chosen_option === option;
                    return (
                      <div
                        key={option}
                        className={`rounded-lg border p-2.5 text-xs sm:text-sm ${
                          isCorrect
                            ? 'border-emerald-300 dark:border-emerald-700 bg-emerald-50 dark:bg-emerald-950/40 text-emerald-900 dark:text-emerald-200 font-medium'
                            : isChosen
                            ? 'border-red-300 dark:border-rose-700 bg-red-50 dark:bg-rose-950/40 text-red-900 dark:text-rose-200 font-medium'
                            : 'border-slate-200 dark:border-slate-700/80 bg-slate-50/50 dark:bg-slate-800/40 text-slate-700 dark:text-slate-300'
                        }`}
                      >
                        <span className="mr-2 font-bold">{option}.</span>
                        {answer[`option_${option.toLowerCase()}`]}
                        {isChosen && <span className="ml-2 text-xs font-medium text-slate-600 dark:text-slate-400">(Lựa chọn của bạn)</span>}
                        {isCorrect && <span className="ml-2 text-xs font-semibold text-emerald-700 dark:text-emerald-400">(Đáp án đúng)</span>}
                      </div>
                    );
                  })}
                </div>
                <div className="mt-3 flex flex-wrap gap-x-6 gap-y-1 text-xs text-slate-600 dark:text-slate-400 border-t border-slate-100 dark:border-slate-800 pt-2">
                  <span>Bạn đã chọn: <strong className="text-slate-800 dark:text-slate-200">{answer.chosen_option || 'Chưa chọn'}</strong></span>
                  <span>Đáp án chuẩn: <strong className="text-emerald-700 dark:text-emerald-400">{answer.correct_option}</strong></span>
                </div>
              </article>
            ))}
          </div>
        )}
      </div>

      {detail.appeals?.length > 0 && (
        <section className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-5 shadow-sm">
          <h2 className="mb-3 font-bold text-slate-900 dark:text-slate-100 text-sm uppercase tracking-wide">Lịch sử phúc khảo</h2>
          <div className="space-y-3">
            {detail.appeals.map((appeal) => (
              <div key={appeal.id} className="rounded-lg border border-slate-200 dark:border-slate-750 p-4 bg-slate-50 dark:bg-slate-800/50">
                <div className="flex flex-wrap justify-between gap-2 text-sm">
                  <strong className="text-slate-900 dark:text-slate-100">{appeal.title || 'Đơn phúc khảo'}</strong>
                  <span className="rounded px-2 py-0.5 text-xs font-semibold bg-amber-100 text-amber-800 dark:bg-amber-950/60 dark:text-amber-300 dark:border dark:border-amber-800">
                    {APPEAL_STATUS[appeal.status] || appeal.status}
                  </span>
                </div>
                <p className="mt-2 whitespace-pre-wrap text-sm text-slate-700 dark:text-slate-300">{appeal.content}</p>
                {appeal.teacher_response && (
                  <div className="mt-3 rounded-lg bg-white dark:bg-slate-900 p-3 text-sm text-slate-800 dark:text-slate-200 border border-slate-200 dark:border-slate-700">
                    <span className="block text-xs font-semibold uppercase text-slate-500 dark:text-slate-400 mb-1">Ý kiến phản hồi từ giảng viên:</span>
                    <p className="whitespace-pre-wrap">{appeal.teacher_response}</p>
                  </div>
                )}
              </div>
            ))}
          </div>
        </section>
      )}
    </div>
  );
}
