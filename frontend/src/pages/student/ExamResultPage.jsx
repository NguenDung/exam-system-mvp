import { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { studentAPI } from '../../api/endpoints';
import Spinner from '../../components/Spinner';

const OPTIONS = ['A', 'B', 'C', 'D'];
const OPTION_FIELDS = { A: 'option_a', B: 'option_b', C: 'option_c', D: 'option_d' };

export default function ExamResultPage() {
  const { attempt_id } = useParams();
  const navigate = useNavigate();
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const fetchResult = useCallback(async () => {
    try {
      const { data: d } = await studentAPI.getResult(attempt_id);
      setData(d);
    } catch (err) {
      setError(err.response?.data?.error || 'Không thể tải kết quả');
    } finally {
      setLoading(false);
    }
  }, [attempt_id]);

  useEffect(() => { fetchResult(); }, [fetchResult]);

  if (loading) return <div className="flex justify-center py-20"><Spinner size="lg" /></div>;

  if (error) return (
    <div className="card text-center py-16 max-w-md mx-auto">
      <div className="text-4xl mb-3">❌</div>
      <p className="text-red-600 font-medium">{error}</p>
      <button onClick={() => navigate('/student')} className="btn-primary mt-4">← Về trang chủ</button>
    </div>
  );

  const { attempt, exam, answers, show_correct_answers } = data;
  const maxScore = (answers || []).reduce((total, answer) => total + Number(answer.points || 0), 0);
  const scorePercent = answers && answers.length > 0
    ? Math.round((attempt.correct_count / answers.length) * 100)
    : 0;

  const scoreColor = scorePercent >= 80 ? 'text-emerald-600' : scorePercent >= 50 ? 'text-amber-600' : 'text-red-600';
  const scoreRingColor = scorePercent >= 80 ? 'border-emerald-400' : scorePercent >= 50 ? 'border-amber-400' : 'border-red-400';

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4 flex-wrap">
        <button onClick={() => navigate('/student')} className="btn-secondary btn-sm">← Về trang chủ</button>
        <h1 className="text-2xl font-bold text-gray-900">📊 Kết quả bài thi</h1>
      </div>

      {/* Score card */}
      <div className="card max-w-2xl mx-auto text-center py-8">
        <h2 className="text-xl font-bold text-gray-900 mb-1">{exam?.title}</h2>
        <p className="text-gray-500 text-sm mb-6">Mã đề: <code className="font-mono text-indigo-700 font-bold">{exam?.exam_code}</code></p>

        {/* Score circle */}
        <div className={`w-36 h-36 rounded-full border-8 ${scoreRingColor} mx-auto flex flex-col items-center justify-center mb-6`}>
          <span className={`text-3xl font-extrabold ${scoreColor}`}>{Number(attempt.score).toFixed(2)}</span>
          <span className="text-xs font-medium text-gray-400">/ {maxScore.toFixed(2)} điểm</span>
        </div>

        <div className="grid grid-cols-3 gap-4 max-w-sm mx-auto mb-4">
          <div className="text-center">
            <div className="text-2xl font-bold text-emerald-600">{attempt.correct_count}</div>
            <div className="text-xs text-gray-500">Câu đúng</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-red-500">{attempt.wrong_count}</div>
            <div className="text-xs text-gray-500">Câu sai</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-gray-400">
              {(answers?.length || 0) - attempt.correct_count - attempt.wrong_count}
            </div>
            <div className="text-xs text-gray-500">Bỏ qua</div>
          </div>
        </div>

        <div className="bg-gray-100 rounded-full h-3 max-w-xs mx-auto overflow-hidden mb-2">
          <div
            className={`h-full rounded-full transition-all duration-700 ${
              scorePercent >= 80 ? 'bg-emerald-500' : scorePercent >= 50 ? 'bg-amber-500' : 'bg-red-500'
            }`}
            style={{ width: `${scorePercent}%` }}
          />
        </div>
        <p className={`text-sm font-semibold ${scoreColor}`}>{scorePercent}% chính xác</p>

        {attempt.submitted_at && (
          <p className="text-xs text-gray-400 mt-3">
            Nộp lúc: {new Date(attempt.submitted_at).toLocaleString('vi-VN')}
          </p>
        )}

        {!show_correct_answers && (
          <p className="mt-4 text-sm text-amber-600 bg-amber-50 rounded-lg px-4 py-2 inline-block">
            🔒 Đáp án chi tiết sẽ hiển thị sau khi kỳ thi kết thúc
          </p>
        )}
      </div>

      {/* Detail answers */}
      {answers && answers.length > 0 && (
        <div>
          <h3 className="font-bold text-gray-900 text-lg mb-4">
            📝 Chi tiết từng câu
            {!show_correct_answers && <span className="text-sm font-normal text-gray-400 ml-2">(chưa hiển thị đáp án)</span>}
          </h3>
          <div className="space-y-4">
            {answers.map((a, idx) => {
              const isCorrect = a.is_correct;
              const isUnanswered = !a.chosen_option;

              return (
                <div key={a.question_id} className={`card border-2 transition-all ${
                  isUnanswered ? 'border-gray-200' :
                  isCorrect ? 'border-emerald-200 bg-emerald-50/30' :
                  'border-red-200 bg-red-50/30'
                }`}>
                  <div className="flex items-start gap-3 mb-4">
                    <span className={`flex-shrink-0 w-8 h-8 rounded-full font-bold text-sm flex items-center justify-center ${
                      isUnanswered ? 'bg-gray-200 text-gray-500' :
                      isCorrect ? 'bg-emerald-500 text-white' :
                      'bg-red-500 text-white'
                    }`}>
                      {isUnanswered ? '—' : isCorrect ? '✓' : '✗'}
                    </span>
                    <div className="flex-1">
                      <p className="font-medium text-gray-900 leading-relaxed">
                        <span className="text-gray-400 mr-2">Câu {idx + 1}.</span>
                        {a.prompt}
                      </p>
                      <p className="text-xs text-gray-400 mt-0.5">{a.points} điểm</p>
                    </div>
                    <span className="text-sm font-bold">
                      {isCorrect ? <span className="text-emerald-600">+{a.points_awarded.toFixed(1)}</span> :
                       <span className="text-gray-400">0</span>}
                    </span>
                  </div>

                  {/* Options */}
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 ml-11">
                    {OPTIONS.map((opt) => {
                      const field = OPTION_FIELDS[opt];
                      const isChosen = a.chosen_option === opt;
                      const isAnswer = show_correct_answers && a.correct_option === opt;

                      let cls = 'border-gray-200 bg-white text-gray-600';
                      if (isAnswer && show_correct_answers) cls = 'border-emerald-400 bg-emerald-50 text-emerald-800 font-semibold';
                      if (isChosen && !isCorrect && !isAnswer) cls = 'border-red-400 bg-red-50 text-red-800 font-semibold';
                      if (isChosen && isCorrect) cls = 'border-emerald-400 bg-emerald-50 text-emerald-800 font-semibold';

                      return (
                        <div key={opt} className={`flex items-center gap-2.5 px-3 py-2.5 rounded-xl border-2 text-sm ${cls}`}>
                          <span className={`w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold flex-shrink-0 ${
                            isAnswer && show_correct_answers ? 'bg-emerald-500 text-white' :
                            isChosen && !isCorrect ? 'bg-red-500 text-white' :
                            'bg-gray-100 text-gray-500'
                          }`}>{opt}</span>
                          <span>{a[field]}</span>
                          {isChosen && <span className="ml-auto text-xs">{isCorrect ? '✓ Đúng' : '✗ Sai'}</span>}
                          {isAnswer && show_correct_answers && !isChosen && (
                            <span className="ml-auto text-xs text-emerald-600">✓ Đáp án</span>
                          )}
                        </div>
                      );
                    })}
                  </div>
                  {isUnanswered && (
                    <p className="text-xs text-gray-400 mt-2 ml-11 italic">Bỏ qua câu này</p>
                  )}
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* Bottom nav */}
      <div className="text-center pt-4 pb-8">
        <button onClick={() => navigate('/student')} className="btn-primary btn-lg">
          🏠 Về trang chủ
        </button>
      </div>
    </div>
  );
}
