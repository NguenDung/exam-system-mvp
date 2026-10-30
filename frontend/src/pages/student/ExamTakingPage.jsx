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
      className={`flex items-center gap-2 px-3 py-1.5 rounded font-mono font-bold text-sm sm:text-base transition-colors ${
        isWarning
          ? 'bg-red-100 text-red-800 border border-red-300'
          : isUrgent
          ? 'bg-amber-100 text-amber-800 border border-amber-300'
          : 'bg-slate-100 text-slate-800 border border-slate-200'
      }`}
    >
      <span className="text-xs uppercase font-sans font-medium text-slate-500 hidden sm:inline">Thời gian:</span>
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

function ViolationWarningModal({ count, onDismiss }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div
        role="alertdialog"
        aria-modal="true"
        className="w-full max-w-md rounded-lg bg-white p-6 text-center shadow-xl border border-red-300 space-y-4"
      >
        <div className="inline-flex items-center justify-center w-12 h-12 rounded-full bg-red-100 text-red-700 font-bold text-lg mb-1">
          !
        </div>
        <h2 className="text-lg font-bold text-red-700">Cảnh báo vi phạm quy chế thi</h2>
        <div className="text-sm text-slate-700 space-y-2 text-left bg-red-50 p-3.5 rounded border border-red-200">
          <p>
            Hệ thống phát hiện bạn vừa rời khỏi màn hình làm bài thi (Chuyển tab hoặc thu nhỏ trình duyệt).
          </p>
          <p className="font-semibold text-red-800">
            Số lần vi phạm đã ghi nhận: {count} / 3 lần.
          </p>
          <p className="text-xs text-slate-500">
            Lưu ý: Nếu vi phạm đạt 3 lần, bài thi sẽ bị hệ thống tự động khóa và thu bài ngay lập tức.
          </p>
        </div>
        <button onClick={onDismiss} className="btn-primary w-full py-2.5">
          Tôi đã hiểu và tiếp tục làm bài
        </button>
      </div>
    </div>
  );
}

function SubmitConfirmModal({ answeredCount, totalCount, flaggedCount, onConfirm, onCancel, submitting }) {
  const unanswered = totalCount - answeredCount;
  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg shadow-xl w-full max-w-sm p-6 text-center border border-slate-200">
        <h3 className="text-lg font-bold text-slate-900 mb-2">Xác nhận nộp bài thi</h3>
        <div className="space-y-1.5 text-sm mb-6 bg-slate-50 p-3.5 rounded border border-slate-100 text-left">
          <p className="text-slate-600">Tổng số câu hỏi: <span className="font-bold text-slate-900">{totalCount}</span></p>
          <p className="text-slate-600">Số câu đã làm: <span className="font-bold text-emerald-700">{answeredCount}</span></p>
          {flaggedCount > 0 && (
            <p className="text-amber-700">Số câu đang đặt cờ: <span className="font-bold">{flaggedCount}</span></p>
          )}
          {unanswered > 0 && (
            <p className="text-red-700 font-medium">Lưu ý: Còn {unanswered} câu chưa chọn đáp án.</p>
          )}
        </div>
        <div className="flex gap-3">
          <button onClick={onCancel} disabled={submitting} className="btn-secondary flex-1">
            Quay lại làm tiếp
          </button>
          <button onClick={onConfirm} disabled={submitting} className="btn-primary flex-1">
            {submitting ? <Spinner size="sm" /> : 'Xác nhận nộp bài'}
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
  const [flagged, setFlagged] = useState({});
  const [currentIndex, setCurrentIndex] = useState(0);

  const [remainingSeconds, setRemainingSeconds] = useState(0);
  const [saveStatus, setSaveStatus] = useState('');
  const [showConfirm, setShowConfirm] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [showTimeWarning, setShowTimeWarning] = useState(false);
  const [violationCount, setViolationCount] = useState(0);
  const [showViolationWarning, setShowViolationWarning] = useState(false);
  const [loadError, setLoadError] = useState('');
  const [submitError, setSubmitError] = useState('');

  const saveStatusTimerRef = useRef(null);
  const autoSaveTimerRef = useRef(null);
  const answersRef = useRef(answers);
  const autoSubmittedRef = useRef(false);
  const lastViolationReportRef = useRef(0);
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
      setViolationCount(data.attempt?.violation_count || 0);
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
      setLoadError(payload?.error || 'Không thể tải đề thi hoặc chưa được cấp quyền truy cập.');
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
      question_id: parseInt(qId, 10),
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

  const toggleFlag = (questionId) => {
    setFlagged((prev) => ({
      ...prev,
      [questionId]: !prev[questionId],
    }));
  };

  const reportViolation = useCallback(async (reason) => {
    const now = Date.now();
    // Throttle repeated triggers within 2 seconds
    if (now - lastViolationReportRef.current < 2000) return;
    lastViolationReportRef.current = now;

    if (!attempt || submitting || autoSubmittedRef.current) return;

    try {
      const { data } = await studentAPI.recordViolation(attempt.id, { reason });
      const currentViolations = data.violation_count !== undefined ? data.violation_count : (violationCount + 1);
      setViolationCount(currentViolations);

      if (data.auto_submitted || currentViolations >= 3) {
        autoSubmittedRef.current = true;
        navigate(`/student/attempts/${attempt.id}/result`, {
          replace: true,
          state: { violationNotice: 'Bài thi đã tự động khóa và thu bài do thí sinh vi phạm quy chế rời màn hình làm bài 3 lần.' }
        });
        return;
      }

      setShowViolationWarning(true);
    } catch (err) {
      console.error('Không thể ghi nhận vi phạm:', err);
    }
  }, [attempt, submitting, violationCount, navigate]);

  useEffect(() => {
    if (!attempt || attempt.status === 'completed') return;

    const handleVisibilityChange = () => {
      if (document.hidden) {
        reportViolation('visibilitychange');
      }
    };

    const handleBlur = () => {
      reportViolation('blur');
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
    window.addEventListener('blur', handleBlur);

    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange);
      window.removeEventListener('blur', handleBlur);
    };
  }, [attempt, reportViolation]);

  const doSubmit = useCallback(async (finalAnswers) => {
    if (!attempt || submitting || autoSubmittedRef.current) return;
    autoSubmittedRef.current = true;
    setSubmitting(true);
    setShowConfirm(false);
    setShowTimeWarning(false);
    const answersArray = Object.entries(finalAnswers || answersRef.current).map(([qId, opt]) => ({
      question_id: parseInt(qId, 10),
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
      setSubmitError(payload?.error || 'Nộp bài không thành công. Vui lòng thử lại.');
      setTimeout(() => setSubmitError(''), 6000);
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
  const flaggedCount = Object.values(flagged).filter(Boolean).length;
  const currentQuestion = questions[currentIndex];

  if (loading) {
    return (
      <div className="min-h-screen bg-slate-100 flex flex-col items-center justify-center gap-3">
        <Spinner size="lg" />
        <p className="text-slate-500 text-sm">Đang kết nối vào phòng thi...</p>
      </div>
    );
  }

  if (loadError) {
    return (
      <div className="min-h-screen bg-slate-100 flex flex-col items-center justify-center gap-3 p-4">
        <div className="bg-white rounded-lg border border-red-200 p-6 max-w-md w-full text-center shadow-sm space-y-4">
          <h2 className="text-base font-bold text-red-700">Thông báo từ phòng thi</h2>
          <p className="text-sm text-slate-700">{loadError}</p>
          <button onClick={() => navigate('/student')} className="btn-primary w-full">
            Quay lại trang chủ sinh viên
          </button>
        </div>
      </div>
    );
  }

  if (!currentQuestion) {
    return (
      <div className="min-h-screen bg-slate-100 flex flex-col items-center justify-center gap-3 p-4">
        <p className="text-slate-600">Đề thi chưa có câu hỏi nào để hiển thị.</p>
        <button onClick={() => navigate('/student')} className="btn-secondary btn-sm">Quay lại trang chủ</button>
      </div>
    );
  }

  const chosenOption = answers[currentQuestion.id];
  const isCurrentFlagged = !!flagged[currentQuestion.id];

  return (
    <div className="min-h-screen bg-slate-100 text-slate-900 pb-12">
      {submitError && (
        <div className="fixed top-4 right-4 z-50 px-4 py-2.5 rounded-md shadow-lg text-white text-sm font-medium bg-red-700">
          {submitError}
        </div>
      )}

      {/* Top Navigation Bar */}
      <header className="sticky top-0 z-30 bg-slate-900 text-white border-b border-slate-800 shadow-sm">
        <div className="max-w-7xl mx-auto px-4 py-3 flex items-center justify-between gap-4 flex-wrap">
          <div className="flex items-center gap-3">
            <span className="font-bold text-sm tracking-tight uppercase">Phòng thi trực tuyến</span>
            <span className="text-xs text-slate-400 font-mono hidden sm:inline">#{attempt?.id}</span>
          </div>

          <div className="flex items-center gap-4">
            {violationCount > 0 && (
              <span className="text-xs px-2 py-0.5 rounded font-medium bg-red-100 text-red-800 border border-red-200">
                Vi phạm: {violationCount}/3
              </span>
            )}
            {saveStatus === 'saving' && <span className="text-xs text-slate-300">Đang lưu...</span>}
            {saveStatus === 'saved' && <span className="text-xs text-emerald-400 font-medium">Đã lưu tự động</span>}
            {saveStatus === 'error' && <span className="text-xs text-red-400">Lưu lỗi mạng</span>}

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
              className="px-4 py-1.5 bg-emerald-600 hover:bg-emerald-700 text-white rounded font-medium text-xs sm:text-sm transition-colors"
            >
              {submitting ? <Spinner size="sm" /> : 'Nộp bài'}
            </button>
          </div>
        </div>
      </header>

      {/* Main Content: 2 Columns */}
      <main className="max-w-7xl mx-auto px-4 py-6">
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
          {/* Left Column: Current Question View (3 cols on lg) */}
          <section className="lg:col-span-3 space-y-4">
            <div className="bg-white rounded-lg border border-slate-200 p-6 shadow-sm">
              {/* Question Header */}
              <div className="flex items-center justify-between border-b border-slate-100 pb-4 mb-4 gap-2 flex-wrap">
                <div className="flex items-center gap-3">
                  <span className="text-xs font-bold uppercase tracking-wider text-slate-500 bg-slate-100 px-2.5 py-1 rounded">
                    Câu {currentIndex + 1} / {totalCount}
                  </span>
                  <span className="text-xs text-slate-500">
                    Điểm số: <strong className="text-slate-800">{currentQuestion.points} điểm</strong>
                  </span>
                </div>

                {/* Flag Question Toggle */}
                <button
                  onClick={() => toggleFlag(currentQuestion.id)}
                  className={`inline-flex items-center gap-1.5 px-3 py-1 rounded text-xs font-medium border transition-colors ${
                    isCurrentFlagged
                      ? 'bg-amber-100 text-amber-900 border-amber-300'
                      : 'bg-white text-slate-600 border-slate-200 hover:bg-slate-50'
                  }`}
                >
                  <span className={`inline-block w-2 h-2 rounded-full ${isCurrentFlagged ? 'bg-amber-500' : 'bg-slate-300'}`} />
                  {isCurrentFlagged ? 'Đã đặt cờ xem lại' : 'Đặt cờ câu này'}
                </button>
              </div>

              {/* Question Prompt */}
              <div className="py-2 mb-6">
                <p className="text-slate-900 text-base sm:text-lg font-medium leading-relaxed whitespace-pre-wrap">
                  {currentQuestion.prompt}
                </p>
              </div>

              {/* Options List */}
              <div className="space-y-3">
                {OPTIONS.map((opt) => {
                  const field = OPTION_FIELDS[opt];
                  const isSelected = chosenOption === opt;
                  return (
                    <button
                      key={opt}
                      onClick={() => handleSelectAnswer(currentQuestion.id, opt)}
                      disabled={submitting}
                      className={`w-full flex items-start gap-3 p-4 rounded-lg border text-left text-sm sm:text-base transition-colors ${
                        isSelected
                          ? 'border-indigo-600 bg-indigo-50/70 text-indigo-950 font-medium'
                          : 'border-slate-200 bg-white text-slate-700 hover:border-slate-300 hover:bg-slate-50'
                      }`}
                    >
                      <span
                        className={`w-7 h-7 rounded flex items-center justify-center text-xs font-bold flex-shrink-0 mt-0.5 ${
                          isSelected ? 'bg-indigo-600 text-white' : 'bg-slate-100 text-slate-600 border border-slate-200'
                        }`}
                      >
                        {opt}
                      </span>
                      <span className="flex-1 leading-snug">{currentQuestion[field]}</span>
                    </button>
                  );
                })}
              </div>

              {/* Question Navigation Controls */}
              <div className="flex items-center justify-between pt-6 mt-6 border-t border-slate-100">
                <button
                  onClick={() => setCurrentIndex((idx) => Math.max(0, idx - 1))}
                  disabled={currentIndex === 0}
                  className="btn-secondary text-sm disabled:opacity-40"
                >
                  &larr; Câu trước
                </button>

                <div className="text-xs text-slate-500 hidden sm:block">
                  {chosenOption ? (
                    <span className="text-emerald-700 font-medium">Đã lưu phương án: {chosenOption}</span>
                  ) : (
                    <span className="text-slate-400">Chưa chọn đáp án</span>
                  )}
                </div>

                {currentIndex < totalCount - 1 ? (
                  <button
                    onClick={() => setCurrentIndex((idx) => Math.min(totalCount - 1, idx + 1))}
                    className="btn-primary text-sm"
                  >
                    Câu tiếp theo &rarr;
                  </button>
                ) : (
                  <button
                    onClick={() => setShowConfirm(true)}
                    className="btn-primary text-sm bg-emerald-600 hover:bg-emerald-700"
                  >
                    Nộp bài thi
                  </button>
                )}
              </div>
            </div>
          </section>

          {/* Right Column: Question Palette (1 col on lg) */}
          <aside className="lg:col-span-1 space-y-4">
            <div className="bg-white rounded-lg border border-slate-200 p-5 shadow-sm sticky top-20">
              <h3 className="font-bold text-slate-900 text-sm mb-3">Bảng câu hỏi ({totalCount})</h3>

              {/* Legend */}
              <div className="grid grid-cols-2 gap-2 text-xs text-slate-600 mb-4 pb-4 border-b border-slate-100">
                <div className="flex items-center gap-1.5">
                  <span className="w-3.5 h-3.5 rounded bg-emerald-600 inline-block" />
                  <span>Đã trả lời</span>
                </div>
                <div className="flex items-center gap-1.5">
                  <span className="w-3.5 h-3.5 rounded bg-amber-500 inline-block" />
                  <span>Đặt cờ</span>
                </div>
                <div className="flex items-center gap-1.5">
                  <span className="w-3.5 h-3.5 rounded bg-slate-100 border border-slate-300 inline-block" />
                  <span>Chưa làm</span>
                </div>
                <div className="flex items-center gap-1.5">
                  <span className="w-3.5 h-3.5 rounded border-2 border-indigo-600 bg-white inline-block" />
                  <span>Đang xem</span>
                </div>
              </div>

              {/* Question Buttons Grid */}
              <div className="grid grid-cols-5 gap-2 mb-5">
                {questions.map((q, idx) => {
                  const isAnswered = !!answers[q.id];
                  const isQFlagged = !!flagged[q.id];
                  const isCurrent = idx === currentIndex;

                  let colorClass = 'bg-slate-100 text-slate-700 hover:bg-slate-200 border border-slate-300';
                  if (isQFlagged) {
                    colorClass = 'bg-amber-500 text-white font-bold border-amber-600';
                  } else if (isAnswered) {
                    colorClass = 'bg-emerald-600 text-white font-bold border-emerald-700';
                  }

                  const activeRing = isCurrent ? 'ring-2 ring-indigo-500 ring-offset-1 font-extrabold' : '';

                  return (
                    <button
                      key={q.id}
                      onClick={() => setCurrentIndex(idx)}
                      className={`h-9 rounded text-xs flex items-center justify-center transition-all ${colorClass} ${activeRing}`}
                      title={`Câu ${idx + 1}: ${isAnswered ? 'Đã trả lời' : 'Chưa trả lời'}${isQFlagged ? ' (Đã đặt cờ)' : ''}`}
                    >
                      {idx + 1}
                    </button>
                  );
                })}
              </div>

              {/* Summary Stats */}
              <div className="space-y-1.5 text-xs text-slate-500 bg-slate-50 p-3 rounded mb-4">
                <div className="flex justify-between">
                  <span>Tiến độ:</span>
                  <span className="font-semibold text-slate-800">{answeredCount} / {totalCount} ({Math.round((answeredCount / (totalCount || 1)) * 100)}%)</span>
                </div>
                <div className="flex justify-between">
                  <span>Số câu đặt cờ:</span>
                  <span className="font-semibold text-amber-700">{flaggedCount}</span>
                </div>
              </div>

              {/* Big Submit Button */}
              <button
                onClick={() => setShowConfirm(true)}
                disabled={submitting}
                className="w-full py-2.5 px-4 bg-emerald-600 hover:bg-emerald-700 text-white rounded font-medium text-sm transition-colors shadow-sm"
              >
                Hoàn thành & Nộp bài
              </button>
            </div>
          </aside>
        </div>
      </main>

      {/* Confirmation Modal */}
      {showConfirm && (
        <SubmitConfirmModal
          answeredCount={answeredCount}
          totalCount={totalCount}
          flaggedCount={flaggedCount}
          onConfirm={() => doSubmit(answers)}
          onCancel={() => setShowConfirm(false)}
          submitting={submitting}
        />
      )}

      {/* 1-Minute Warning Modal */}
      {showTimeWarning && !submitting && (
        <OneMinuteWarningModal onDismiss={() => setShowTimeWarning(false)} />
      )}

      {/* Violation Warning Modal */}
      {showViolationWarning && !submitting && (
        <ViolationWarningModal
          count={violationCount}
          onDismiss={() => setShowViolationWarning(false)}
        />
      )}
    </div>
  );
}
