import { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate, useLocation } from 'react-router-dom';
import { studentAPI } from '../../api/endpoints';
import Spinner from '../../components/Spinner';

const OPTIONS = ['A', 'B', 'C', 'D'];
const OPTION_FIELDS = { A: 'option_a', B: 'option_b', C: 'option_c', D: 'option_d' };

export default function ExamResultPage() {
  const { attempt_id } = useParams();
  const navigate = useNavigate();
  const location = useLocation();
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const fetchResult = useCallback(async () => {
    try {
      const { data: d } = await studentAPI.getResult(attempt_id);
      setData(d);
    } catch (err) {
      setError(err.response?.data?.error || 'Không thể tải kết quả bài thi');
    } finally {
      setLoading(false);
    }
  }, [attempt_id]);

  useEffect(() => {
    fetchResult();
  }, [fetchResult]);

  if (loading) return <div className="flex justify-center py-20"><Spinner size="lg" /></div>;

  if (error) {
    return (
      <div className="bg-white dark:bg-slate-900 rounded-xl border border-red-200 dark:border-rose-900/60 text-center py-12 px-6 max-w-md mx-auto shadow-sm">
        <p className="text-red-700 dark:text-rose-300 font-medium text-sm">{error}</p>
        <button onClick={() => navigate('/student')} className="btn-primary mt-4">
          Quay về trang chủ
        </button>
      </div>
    );
  }

  const { attempt, exam, answers, show_correct_answers } = data;
  const maxScore = (answers || []).reduce((total, answer) => total + Number(answer.points || 0), 0);
  const scorePercent =
    answers && answers.length > 0 ? Math.round((attempt.correct_count / answers.length) * 100) : 0;

  return (
    <div className="space-y-6">
      {(location.state?.violationNotice || attempt.violation_count >= 3) && (
        <div className="rounded-lg bg-red-50 dark:bg-rose-950/40 border border-red-300 dark:border-rose-800 p-4 text-sm text-red-800 dark:text-rose-200 space-y-1">
          <p className="font-bold text-red-900 dark:text-rose-100">Thông báo từ ban giám thị / phòng thi:</p>
          <p>{location.state?.violationNotice || 'Bài thi đã bị khóa và thu bài tự động do phát hiện rời khỏi màn hình làm bài thi 3 lần.'}</p>
        </div>
      )}

      <div className="flex items-center justify-between gap-4 flex-wrap">
        <div>
          <h1 className="text-xl font-bold text-slate-900 dark:text-slate-100">Kết quả bài thi</h1>
          <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">Bảng điểm và thống kê bài làm</p>
        </div>
        <button onClick={() => navigate('/student')} className="btn-secondary btn-sm">
          Quay lại trang chủ
        </button>
      </div>

      {/* Score overview card */}
      <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-6 max-w-2xl mx-auto text-center shadow-sm">
        <h2 className="text-lg font-bold text-slate-900 dark:text-slate-100">{exam?.title}</h2>
        <p className="text-slate-500 dark:text-slate-400 text-xs mt-1 mb-6">
          Mã đề: <code className="font-mono text-slate-800 dark:text-slate-200 font-bold">{exam?.exam_code}</code>
        </p>

        <div className="inline-block p-6 rounded-lg bg-slate-50 dark:bg-slate-800/60 border border-slate-200 dark:border-slate-700 mb-6 min-w-[200px]">
          <span className="text-4xl font-extrabold text-slate-900 dark:text-slate-100 block">{Number(attempt.score).toFixed(2)}</span>
          <span className="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">
            trên tổng {maxScore.toFixed(2)} điểm
          </span>
        </div>

        <div className="grid grid-cols-3 gap-4 max-w-sm mx-auto mb-4 border-t border-slate-100 dark:border-slate-800 pt-4 text-sm">
          <div>
            <div className="text-xl font-bold text-emerald-700 dark:text-emerald-400">{attempt.correct_count}</div>
            <div className="text-xs text-slate-500 dark:text-slate-400">Số câu đúng</div>
          </div>
          <div>
            <div className="text-xl font-bold text-red-600 dark:text-rose-400">{attempt.wrong_count}</div>
            <div className="text-xs text-slate-500 dark:text-slate-400">Số câu sai</div>
          </div>
          <div>
            <div className="text-xl font-bold text-slate-400 dark:text-slate-500">
              {(answers?.length || 0) - attempt.correct_count - attempt.wrong_count}
            </div>
            <div className="text-xs text-slate-500 dark:text-slate-400">Bỏ qua</div>
          </div>
        </div>

        <p className="text-xs text-slate-500 dark:text-slate-400">Tỷ lệ chính xác: <strong className="text-slate-800 dark:text-slate-200">{scorePercent}%</strong></p>

        {attempt.submitted_at && (
          <p className="text-xs text-slate-400 mt-2">
            Thời điểm nộp bài: {new Date(attempt.submitted_at).toLocaleString('vi-VN')}
          </p>
        )}

        {!show_correct_answers && (
          <div className="mt-4 text-xs text-amber-800 bg-amber-50 border border-amber-200 rounded p-2.5 max-w-md mx-auto">
            Đáp án chuẩn sẽ được hiển thị sau khi kỳ thi chính thức kết thúc thời gian làm bài.
          </div>
        )}
      </div>

      {/* Answers list */}
      {answers && answers.length > 0 && (
        <div className="max-w-2xl mx-auto space-y-4">
          <h3 className="font-bold text-slate-900 dark:text-slate-100 text-sm uppercase tracking-wide">
            Chi tiết các câu hỏi
            {!show_correct_answers && (
              <span className="text-xs font-normal text-slate-400 dark:text-slate-500 ml-2">(Chưa mở đáp án đúng)</span>
            )}
          </h3>

          <div className="space-y-3">
            {answers.map((a, idx) => {
              const isCorrect = a.is_correct;
              const isUnanswered = !a.chosen_option;

              return (
                <div
                  key={a.question_id}
                  className={`bg-white dark:bg-slate-900 rounded-xl border p-4 shadow-sm ${
                    isUnanswered
                      ? 'border-slate-200 dark:border-slate-800'
                      : isCorrect
                      ? 'border-emerald-200 dark:border-emerald-800'
                      : 'border-red-200 dark:border-rose-800'
                  }`}
                >
                  <div className="flex items-start gap-3 mb-3">
                    <span className="flex-shrink-0 w-6 h-6 rounded bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 font-bold text-xs flex items-center justify-center">
                      {idx + 1}
                    </span>
                    <div className="flex-1">
                      <p className="font-medium text-slate-900 dark:text-slate-100 text-sm">{a.prompt}</p>
                      <p className="text-xs text-slate-400 dark:text-slate-500 mt-0.5">Trọng số: {a.points} điểm</p>
                    </div>
                    <span className="text-xs font-semibold">
                      {isCorrect ? (
                        <span className="text-emerald-700 dark:text-emerald-400">+{Number(a.points_awarded).toFixed(1)}</span>
                      ) : (
                        <span className="text-slate-400 dark:text-slate-500">0</span>
                      )}
                    </span>
                  </div>

                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-xs">
                    {OPTIONS.map((opt) => {
                      const field = OPTION_FIELDS[opt];
                      const isChosen = a.chosen_option === opt;
                      const isAnswer = show_correct_answers && a.correct_option === opt;

                      let cls = 'border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800/80 text-slate-700 dark:text-slate-200';
                      if (isAnswer && show_correct_answers) cls = 'border-emerald-300 dark:border-emerald-700 bg-emerald-50 dark:bg-emerald-950/40 text-emerald-900 dark:text-emerald-200 font-semibold';
                      if (isChosen && !isCorrect && !isAnswer) cls = 'border-red-300 dark:border-rose-700 bg-red-50 dark:bg-rose-950/40 text-red-900 dark:text-rose-200 font-semibold';
                      if (isChosen && isCorrect) cls = 'border-emerald-300 dark:border-emerald-700 bg-emerald-50 dark:bg-emerald-950/40 text-emerald-900 dark:text-emerald-200 font-semibold';

                      return (
                        <div key={opt} className={`flex items-center gap-2 px-3 py-2 rounded border ${cls}`}>
                          <span
                            className={`w-5 h-5 rounded flex items-center justify-center text-xs font-bold flex-shrink-0 ${
                              isAnswer && show_correct_answers
                                ? 'bg-emerald-600 text-white'
                                : isChosen && !isCorrect
                                ? 'bg-red-600 text-white'
                                : 'bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300'
                            }`}
                          >
                            {opt}
                          </span>
                          <span className="truncate">{a[field]}</span>
                          {isChosen && (
                            <span className="ml-auto text-xs font-medium">
                              {isCorrect ? '(Đúng)' : '(Bạn chọn)'}
                            </span>
                          )}
                          {isAnswer && show_correct_answers && !isChosen && (
                            <span className="ml-auto text-xs text-emerald-700 font-semibold">(Đáp án)</span>
                          )}
                        </div>
                      );
                    })}
                  </div>
                  {isUnanswered && (
                    <p className="text-xs text-slate-400 mt-2 italic">Câu hỏi này chưa được trả lời</p>
                  )}
                </div>
              );
            })}
          </div>
        </div>
      )}

      <div className="text-center pt-4 pb-8">
        <button onClick={() => navigate('/student')} className="btn-secondary">
          Quay lại trang chủ
        </button>
      </div>
    </div>
  );
}
