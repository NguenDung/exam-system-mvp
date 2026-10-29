import { useState, useEffect, useRef, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { studentAPI } from '../../api/endpoints';
import Spinner from '../../components/Spinner';

const OPTIONS = ['A', 'B', 'C', 'D'];
const OPTION_FIELDS = { A: 'option_a', B: 'option_b', C: 'option_c', D: 'option_d' };

function CountdownTimer({ seconds, onExpire, onWarning }) {
  const [remaining, setRemaining] = useState(seconds);
  const expiredRef = useRef(false);
  const warnedRef = useRef(false);

  useEffect(() => {
    setRemaining(seconds);
    expiredRef.current = false;
    warnedRef.current = false;
  }, [seconds]);

  useEffect(() => {
    if (remaining > 0 && remaining <= 60 && !warnedRef.current) {
      warnedRef.current = true;
      onWarning(true);
    }
    if (remaining <= 0) {
      if (!expiredRef.current) {
        expiredRef.current = true;
        onWarning(false);
        onExpire();
      }
      return;
    }
    const id = setTimeout(() => setRemaining((r) => r - 1), 1000);
    return () => clearTimeout(id);
  }, [remaining, onExpire, onWarning]);

  const h = Math.floor(remaining / 3600);
  const m = Math.floor((remaining % 3600) / 60);
  const s = remaining % 60;
  const isUrgent = remaining <= 300;
  const isWarning = remaining <= 60;

  return (
    <div
      className={`flex items-center gap-2 px-3.5 py-1.5 rounded font-mono font-bold text-base transition-colors ${
        isWarning
          ? 'bg-red-100 text-red-800'
          : isUrgent
          ? 'bg-amber-100 text-amber-800'
          : 'bg-slate-100 text-slate-800'
      }`}
    >
      <span className="text-xs uppercase font-sans font-medium text-slate-500">Thời gian còn lại:</span>
      <span>
        {h > 0 ? `${String(h).padStart(2, '0')}:` : ''}
        {String(m).padStart(2, '0')}:{String(s).padStart(2, '0')}
      </span>
    </div>
  );
}

function OneMinuteWarningModal({ onDismiss }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div
        role="alertdialog"
        aria-modal="true"
        aria-labelledby="time-warning-title"
        className="w-full max-w-sm rounded-lg bg-white p-6 text-center shadow-xl border border-slate-200"
      >
        <h2 id="time-warning-title" className="text-lg font-bold text-red-700">Thông báo sắp hết giờ</h2>
        <p className="mt-3 text-sm text-slate-600 leading-relaxed">
          Thời gian làm bài còn dưới 1 phút. Hệ thống sẽ tự động khóa và thu bài khi đồng hồ về 00:00.
        </p>
        <button onClick={onDismiss} className="btn-primary mt-5 w-full">
          Tiếp tục làm bài
        </button>
      </div>
    </div>
  );
}

function SubmitConfirmModal({ answeredCount, totalCount, onConfirm, onCancel, submitting }) {
  const unanswered = totalCount - answeredCount;
  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg shadow-xl w-full max-w-sm p-6 text-center border border-slate-200">
        <h3 className="text-lg font-bold text-slate-900 mb-2">Xác nhận nộp bài thi</h3>
        <div className="space-y-1.5 text-sm mb-6 bg-slate-50 p-3.5 rounded border border-slate-100 text-left">
          <p className="text-slate-600">Tổng số câu hỏi: <span className="font-bold text-slate-900">{totalCount}</span></p>
          <p className="text-slate-600">Số câu đã trả lời: <span className="font-bold text-emerald-700">{answeredCount}</span></p>
          {unanswered > 0 && (
            <p className="text-amber-800 font-medium">Lưu ý: Còn {unanswered} câu chưa chọn đáp án.</p>
          )}
        </div>
        <div className="flex gap-3">
          <button onClick={onCancel} disabled={submitting} className="btn-secondary flex-1">
            Quay lại
          </button>
          <button onClick={onConfirm} disabled={submitting} className="btn-primary flex-1">
            {submitting ? <Spinner size="sm" /> : 'Xác nhận nộp'}
          </button>
        </div>
      </div>
    </div>
  );
}

export default function ExamTakingPage() {
  const { exam_id } = useParams();
  const navigate = useNavigate();

  const [loading, setLoading] = useState(true);
  const [attempt, setAttempt] = useState(null);
  const [questions, setQuestions] = useState([]);
  const [answers, setAnswers] = useState({});
  const [remainingSeconds, setRemainingSeconds] = useState(0);
  const [saveStatus, setSaveStatus] = useState('');
  const [showConfirm, setShowConfirm] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [showTimeWarning, setShowTimeWarning] = useState(false);

  const saveStatusTimerRef = useRef(null);
  const autoSaveTimerRef = useRef(null);
  const answersRef = useRef(answers);
  const autoSubmittedRef = useRef(false);
  answersRef.current = answers;

  const loadExam = useCallback(async () => {
    try {
      const { data } = await studentAPI.startExam(exam_id);

      if (
        data.auto_submitted ||
        data.attempt?.status === 'completed' ||
        (data.remaining_seconds === 0 && data.attempt?.id && !data.questions?.length)
      ) {
        navigate(`/student/attempts/${data.attempt.id}/result`, { replace: true });
        return;
      }

      setAttempt(data.attempt);
      setQuestions(data.questions || []);
      setRemainingSeconds(Math.max(0, data.remaining_seconds));

      const savedMap = {};
      (data.saved_answers || []).forEach((a) => {
        if (a.chosen_option) savedMap[a.question_id] = a.chosen_option;
      });
      setAnswers(savedMap);
    } catch (err) {
      const payload = err.response?.data;
      if (payload?.attempt?.id && payload?.attempt?.status === 'completed') {
        navigate(`/student/attempts/${payload.attempt.id}/result`, { replace: true });
        return;
      }
      alert(payload?.error || 'Không thể tải đề thi.');
      navigate('/student');
    } finally {
      setLoading(false);
    }
  }, [exam_id, navigate]);

  useEffect(() => {
    loadExam();
  }, [loadExam]);

  const saveAnswersToServer = useCallback(async (answerMap) => {
    if (!attempt) return;
    const answersArray = Object.entries(answerMap).map(([qId, opt]) => ({
      question_id: parseInt(qId),
      chosen_option: opt,
    }));
    if (answersArray.length === 0) return;
    setSaveStatus('saving');
    try {
      await studentAPI.saveAnswers(attempt.id, { answers: answersArray });
      setSaveStatus('saved');
    } catch (err) {
      const payload = err.response?.data;
      if (payload?.auto_submitted && (payload?.attempt?.id || attempt.id)) {
        navigate(`/student/attempts/${payload.attempt?.id || attempt.id}/result`, { replace: true });
        return;
      }
      if (payload?.attempt?.status === 'completed') {
        navigate(`/student/attempts/${payload.attempt.id}/result`, { replace: true });
        return;
      }
      setSaveStatus('error');
    } finally {
      if (saveStatusTimerRef.current) clearTimeout(saveStatusTimerRef.current);
      saveStatusTimerRef.current = setTimeout(() => setSaveStatus(''), 2500);
    }
  }, [attempt, navigate]);

  useEffect(() => {
    if (!attempt) return;
    autoSaveTimerRef.current = setInterval(() => {
      saveAnswersToServer(answersRef.current);
    }, 60000);
    return () => clearInterval(autoSaveTimerRef.current);
  }, [attempt, saveAnswersToServer]);

  const handleSelectAnswer = (questionId, option) => {
    const newAnswers = { ...answersRef.current, [questionId]: option };
    setAnswers(newAnswers);
    saveAnswersToServer(newAnswers);
  };

  const doSubmit = useCallback(async (finalAnswers) => {
    if (!attempt || submitting || autoSubmittedRef.current) return;
    autoSubmittedRef.current = true;
    setSubmitting(true);
    setShowConfirm(false);
    setShowTimeWarning(false);
    const answersArray = Object.entries(finalAnswers || answersRef.current).map(([qId, opt]) => ({
      question_id: parseInt(qId),
      chosen_option: opt,
    }));
    try {
      const { data } = await studentAPI.submitExam(attempt.id, { answers: answersArray });
      const resultId = data?.attempt?.id || attempt.id;
      navigate(`/student/attempts/${resultId}/result`, { replace: true });
    } catch (err) {
      const payload = err.response?.data;
      if (payload?.attempt?.id || attempt.id) {
        navigate(`/student/attempts/${payload?.attempt?.id || attempt.id}/result`, { replace: true });
        return;
      }
      alert(payload?.error || 'Nộp bài không thành công');
      autoSubmittedRef.current = false;
      setSubmitting(false);
    }
  }, [attempt, submitting, navigate]);

  const handleOneMinuteWarning = useCallback((show = true) => {
    setShowTimeWarning(show);
  }, []);

  const handleTimerExpire = useCallback(() => {
    if (autoSubmittedRef.current) return;
    doSubmit(answersRef.current);
  }, [doSubmit]);

  const answeredCount = Object.keys(answers).length;
  const totalCount = questions.length;

  if (loading) {
    return (
      <div className="min-h-screen bg-slate-100 flex flex-col items-center justify-center gap-3">
        <Spinner size="lg" />
        <p className="text-slate-500 text-sm">Đang chuẩn bị đề thi...</p>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-slate-100 text-slate-900 pb-12">
      {/* Sticky header */}
      <div className="sticky top-0 z-30 bg-white border-b border-slate-200 shadow-sm">
        <div className="max-w-4xl mx-auto px-4 py-3 flex items-center justify-between gap-4 flex-wrap">
          <div>
            <h1 className="font-bold text-slate-900 text-sm sm:text-base leading-tight">Đang làm bài thi</h1>
            <p className="text-xs text-slate-500 mt-0.5">
              Đã trả lời: <span className="font-semibold text-slate-900">{answeredCount}</span> / {totalCount} câu
            </p>
          </div>

          <div className="flex-1 max-w-xs hidden md:block">
            <div className="h-2 bg-slate-200 rounded-full overflow-hidden">
              <div
                className="h-full bg-indigo-600 rounded-full transition-all duration-300"
                style={{ width: `${totalCount > 0 ? (answeredCount / totalCount) * 100 : 0}%` }}
              />
            </div>
          </div>

          <div className="flex items-center gap-3">
            {saveStatus === 'saving' && <span className="text-xs text-slate-400">Đang lưu...</span>}
            {saveStatus === 'saved' && <span className="text-xs text-emerald-700 font-medium">Đã tự động lưu</span>}
            {saveStatus === 'error' && <span className="text-xs text-red-600">Lưu thất bại</span>}

            {remainingSeconds > 0 && (
              <CountdownTimer
                seconds={remainingSeconds}
                onExpire={handleTimerExpire}
                onWarning={handleOneMinuteWarning}
              />
            )}

            <button onClick={() => setShowConfirm(true)} disabled={submitting} className="btn-primary btn-sm">
              {submitting ? <Spinner size="sm" /> : 'Nộp bài'}
            </button>
          </div>
        </div>
      </div>

      {/* Questions list */}
      <div className="max-w-4xl mx-auto px-4 py-6 space-y-4">
        {questions.map((q, idx) => {
          const chosen = answers[q.id];
          return (
            <div key={q.id} className="bg-white rounded-lg border border-slate-200 p-5">
              <div className="flex items-start gap-3 mb-4">
                <span className="flex-shrink-0 w-7 h-7 rounded bg-slate-100 text-slate-700 font-bold text-xs flex items-center justify-center">
                  {idx + 1}
                </span>
                <div className="flex-1">
                  <p className="text-slate-900 font-medium leading-relaxed">{q.prompt}</p>
                  <p className="text-xs text-slate-400 mt-1">Trọng số: {q.points} điểm</p>
                </div>
                {chosen && (
                  <span className="text-xs px-2 py-0.5 rounded font-medium bg-slate-100 text-slate-700 border border-slate-200">
                    Đã chọn {chosen}
                  </span>
                )}
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-2.5">
                {OPTIONS.map((opt) => {
                  const field = OPTION_FIELDS[opt];
                  const isSelected = chosen === opt;
                  return (
                    <button
                      key={opt}
                      onClick={() => handleSelectAnswer(q.id, opt)}
                      disabled={submitting}
                      className={`flex items-center gap-3 px-3.5 py-2.5 rounded border text-left text-sm transition-colors ${
                        isSelected
                          ? 'border-indigo-600 bg-indigo-50 text-indigo-900 font-medium'
                          : 'border-slate-200 bg-white text-slate-700 hover:border-slate-300'
                      }`}
                    >
                      <span
                        className={`w-6 h-6 rounded flex items-center justify-center text-xs font-bold flex-shrink-0 ${
                          isSelected ? 'bg-indigo-600 text-white' : 'bg-slate-100 text-slate-600'
                        }`}
                      >
                        {opt}
                      </span>
                      <span className="truncate">{q[field]}</span>
                    </button>
                  );
                })}
              </div>
            </div>
          );
        })}

        <div className="text-center pt-4">
          <button onClick={() => setShowConfirm(true)} disabled={submitting} className="btn-primary px-8 py-2.5">
            {submitting ? <Spinner size="sm" /> : null} Nộp bài thi
          </button>
        </div>
      </div>

      {showConfirm && (
        <SubmitConfirmModal
          answeredCount={answeredCount}
          totalCount={totalCount}
          onConfirm={() => doSubmit(answers)}
          onCancel={() => setShowConfirm(false)}
          submitting={submitting}
        />
      )}
      {showTimeWarning && !submitting && (
        <OneMinuteWarningModal onDismiss={() => setShowTimeWarning(false)} />
      )}
    </div>
  );
}
