import { useState, useEffect, useRef, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { studentAPI } from '../../api/endpoints';
import Spinner from '../../components/Spinner';

const OPTIONS = ['A', 'B', 'C', 'D'];
const OPTION_FIELDS = { A: 'option_a', B: 'option_b', C: 'option_c', D: 'option_d' };
const OPTION_COLORS = {
  base: 'border-gray-200 bg-white hover:border-indigo-300 hover:bg-indigo-50 text-gray-700',
  selected: 'border-indigo-500 bg-indigo-50 text-indigo-900 ring-2 ring-indigo-200 font-semibold',
};
const OPTION_BADGE = {
  A: 'bg-blue-100 text-blue-700',
  B: 'bg-purple-100 text-purple-700',
  C: 'bg-orange-100 text-orange-700',
  D: 'bg-pink-100 text-pink-700',
};

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
      onWarning();
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
  const isUrgent = remaining <= 300; // 5 minutes
  const isWarning = remaining <= 60;

  return (
    <div className={`flex items-center gap-1.5 px-4 py-2 rounded-xl font-mono font-bold text-lg transition-all ${
      isWarning ? 'bg-red-100 text-red-700 animate-pulse' :
      isUrgent ? 'bg-orange-100 text-orange-700' :
      'bg-indigo-100 text-indigo-700'
    }`}>
      <span>⏱</span>
      <span>{h > 0 ? `${String(h).padStart(2,'0')}:` : ''}{String(m).padStart(2,'0')}:{String(s).padStart(2,'0')}</span>
      {isWarning && <span className="text-xs ml-1">⚠️</span>}
    </div>
  );
}

function OneMinuteWarningModal({ onDismiss }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div
        role="alertdialog"
        aria-modal="true"
        aria-labelledby="time-warning-title"
        className="w-full max-w-sm rounded-2xl bg-white p-8 text-center shadow-2xl"
      >
        <div className="mb-3 text-5xl">⏰</div>
        <h2 id="time-warning-title" className="text-xl font-bold text-red-700">Sắp hết giờ làm bài</h2>
        <p className="mt-3 text-sm text-gray-600">
          Thời gian làm bài còn không quá 1 phút. Hãy hoàn tất và nộp bài trước khi hết giờ.
        </p>
        <button onClick={onDismiss} className="btn-primary mt-6 w-full">Tiếp tục làm bài</button>
      </div>
    </div>
  );
}

function SubmitConfirmModal({ answeredCount, totalCount, onConfirm, onCancel, submitting }) {
  const unanswered = totalCount - answeredCount;
  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-2xl shadow-2xl w-full max-w-sm p-8 text-center">
        <div className="text-5xl mb-4">📤</div>
        <h3 className="text-xl font-bold text-gray-900 mb-2">Xác nhận nộp bài?</h3>
        <div className="space-y-1 text-sm mb-6">
          <p className="text-gray-600">Tổng số câu: <span className="font-bold text-gray-900">{totalCount}</span></p>
          <p className="text-emerald-600">Đã trả lời: <span className="font-bold">{answeredCount}</span></p>
          {unanswered > 0 && (
            <p className="text-amber-600 font-medium">⚠️ Còn {unanswered} câu chưa làm!</p>
          )}
        </div>
        <div className="flex gap-3">
          <button onClick={onCancel} disabled={submitting} className="btn-secondary flex-1">Làm tiếp</button>
          <button onClick={onConfirm} disabled={submitting} className="btn-success flex-1">
            {submitting ? <Spinner size="sm" /> : '✅ Nộp bài'}
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
  const [answers, setAnswers] = useState({}); // { question_id: 'A'|'B'|'C'|'D' }
  const [remainingSeconds, setRemainingSeconds] = useState(0);
  const [saveStatus, setSaveStatus] = useState(''); // 'saving' | 'saved' | 'error' | ''
  const [showConfirm, setShowConfirm] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [showTimeWarning, setShowTimeWarning] = useState(false);

  const saveStatusTimerRef = useRef(null);
  const autoSaveTimerRef = useRef(null);
  const answersRef = useRef(answers);
  const autoSubmittedRef = useRef(false);
  answersRef.current = answers;

  // ── Load exam on mount (also handles F5 resume) ───────────────────────────
  const loadExam = useCallback(async () => {
    try {
      const { data } = await studentAPI.startExam(exam_id);

      // Auto-submitted / already completed → go straight to results
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

      // Rebuild answers from saved_answers
      const savedMap = {};
      (data.saved_answers || []).forEach((a) => {
        if (a.chosen_option) savedMap[a.question_id] = a.chosen_option;
      });
      setAnswers(savedMap);
    } catch (err) {
      const payload = err.response?.data;
      // Backend may return 400 with completed attempt when already submitted
      if (payload?.attempt?.id && payload?.attempt?.status === 'completed') {
        navigate(`/student/attempts/${payload.attempt.id}/result`, { replace: true });
        return;
      }
      alert(payload?.error || 'Không thể tải bài thi');
      navigate('/student');
    } finally {
      setLoading(false);
    }
  }, [exam_id, navigate]);

  useEffect(() => { loadExam(); }, [loadExam]);

  // ── Save to server ────────────────────────────────────────────────────────
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
      // 409 Conflict / auto-submit on timeout → redirect to results
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
      saveStatusTimerRef.current = setTimeout(() => setSaveStatus(''), 3000);
    }
  }, [attempt, navigate]);

  // ── Periodic auto-save every 60s ─────────────────────────────────────────
  useEffect(() => {
    if (!attempt) return;
    autoSaveTimerRef.current = setInterval(() => {
      saveAnswersToServer(answersRef.current);
    }, 60000);
    return () => clearInterval(autoSaveTimerRef.current);
  }, [attempt, saveAnswersToServer]);

  // ── Select answer ─────────────────────────────────────────────────────────
  const handleSelectAnswer = (questionId, option) => {
    const newAnswers = { ...answersRef.current, [questionId]: option };
    setAnswers(newAnswers);
    // Immediate save on each selection
    saveAnswersToServer(newAnswers);
  };

  // ── Submit ────────────────────────────────────────────────────────────────
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
      // Already submitted / auto-graded → still go to results if we have an attempt id
      if (payload?.attempt?.id || attempt.id) {
        navigate(`/student/attempts/${payload?.attempt?.id || attempt.id}/result`, { replace: true });
        return;
      }
      alert(payload?.error || 'Nộp bài thất bại');
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
      <div className="min-h-screen bg-gray-50 flex flex-col items-center justify-center gap-4">
        <Spinner size="lg" />
        <p className="text-gray-500">Đang tải bài thi...</p>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-100">
      {/* Sticky header */}
      <div className="sticky top-0 z-30 bg-white shadow-md border-b">
        <div className="max-w-4xl mx-auto px-4 py-3 flex items-center justify-between gap-4 flex-wrap">
          <div>
            <h1 className="font-bold text-gray-900 text-base leading-tight">📝 Đang làm bài thi</h1>
            <p className="text-xs text-gray-500 mt-0.5">
              Đã trả lời: <span className="font-semibold text-emerald-600">{answeredCount}</span> / {totalCount} câu
            </p>
          </div>

          {/* Progress bar */}
          <div className="flex-1 max-w-xs hidden sm:block">
            <div className="h-2 bg-gray-200 rounded-full overflow-hidden">
              <div
                className="h-full bg-emerald-500 rounded-full transition-all duration-300"
                style={{ width: `${totalCount > 0 ? (answeredCount / totalCount) * 100 : 0}%` }}
              />
            </div>
            <p className="text-xs text-gray-400 text-right mt-0.5">{totalCount > 0 ? Math.round((answeredCount / totalCount) * 100) : 0}%</p>
          </div>

          <div className="flex items-center gap-3">
            {/* Auto-save status */}
            {saveStatus === 'saving' && <span className="text-xs text-gray-400 animate-pulse">💾 Đang lưu...</span>}
            {saveStatus === 'saved' && <span className="text-xs text-emerald-600">✅ Đã tự động lưu</span>}
            {saveStatus === 'error' && <span className="text-xs text-red-500">⚠️ Lưu thất bại</span>}

            {/* Timer */}
            {remainingSeconds > 0 && (
              <CountdownTimer
                seconds={remainingSeconds}
                onExpire={handleTimerExpire}
                onWarning={handleOneMinuteWarning}
              />
            )}

            <button
              onClick={() => setShowConfirm(true)}
              disabled={submitting}
              className="btn-success"
            >
              {submitting ? <Spinner size="sm" /> : '📤 Nộp bài'}
            </button>
          </div>
        </div>
      </div>

      {/* Questions */}
      <div className="max-w-4xl mx-auto px-4 py-6 space-y-5">
        {questions.map((q, idx) => {
          const chosen = answers[q.id];
          return (
            <div key={q.id} className="bg-white rounded-2xl shadow-sm border border-gray-100 overflow-hidden">
              {/* Question header */}
              <div className="flex items-start gap-3 p-5 border-b bg-gray-50">
                <span className="flex-shrink-0 w-8 h-8 rounded-full bg-indigo-600 text-white font-bold text-sm flex items-center justify-center">
                  {idx + 1}
                </span>
                <div className="flex-1">
                  <p className="text-gray-900 font-medium leading-relaxed">{q.prompt}</p>
                  <p className="text-xs text-gray-400 mt-1">{q.points} điểm</p>
                </div>
                {chosen && (
                  <span className={`text-xs px-2 py-0.5 rounded-full font-semibold ${OPTION_BADGE[chosen]}`}>
                    Đã chọn {chosen}
                  </span>
                )}
              </div>

              {/* Options */}
              <div className="p-5 grid grid-cols-1 sm:grid-cols-2 gap-3">
                {OPTIONS.map((opt) => {
                  const field = OPTION_FIELDS[opt];
                  const isSelected = chosen === opt;
                  return (
                    <button
                      key={opt}
                      onClick={() => handleSelectAnswer(q.id, opt)}
                      disabled={submitting}
                      className={`flex items-center gap-3 px-4 py-3 rounded-xl border-2 text-left transition-all duration-150 active:scale-95 ${
                        isSelected ? OPTION_COLORS.selected : OPTION_COLORS.base
                      }`}
                    >
                      <span className={`w-7 h-7 rounded-full flex items-center justify-center text-xs font-bold flex-shrink-0 ${
                        isSelected ? 'bg-indigo-600 text-white' : OPTION_BADGE[opt]
                      }`}>
                        {opt}
                      </span>
                      <span className="text-sm leading-snug">{q[field]}</span>
                    </button>
                  );
                })}
              </div>
            </div>
          );
        })}

        {/* Bottom submit button */}
        <div className="text-center pt-4">
          <button onClick={() => setShowConfirm(true)} disabled={submitting} className="btn-success btn-lg px-10">
            {submitting ? <Spinner size="sm" /> : '📤'} Nộp bài thi
          </button>
        </div>
      </div>

      {/* Submit confirm modal */}
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
