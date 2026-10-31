import { useState, useEffect, useMemo } from 'react';
import { Link } from 'react-router-dom';
import { studentAPI, classroomAPI } from '../../api/endpoints';
import Spinner from '../../components/Spinner';

// Quy chế đào tạo tín chỉ: Quy đổi điểm hệ 10 sang Điểm chữ & Thang 4
function calculateScale(score10) {
  if (score10 === null || score10 === undefined || isNaN(score10)) {
    return { letter: '—', scale4: null, status: 'Chưa có điểm', badge: 'bg-slate-100 text-slate-600 border border-slate-200 dark:bg-slate-800 dark:text-slate-300 dark:border-slate-700' };
  }
  const s = Number(score10);
  if (s >= 8.5) return { letter: 'A', scale4: 4.0, status: 'Đạt (Xuất sắc)', badge: 'bg-emerald-50 text-emerald-700 border border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-800' };
  if (s >= 8.0) return { letter: 'B+', scale4: 3.5, status: 'Đạt (Giỏi)', badge: 'bg-blue-50 text-blue-700 border border-blue-200 dark:bg-blue-950/40 dark:text-blue-300 dark:border-blue-800' };
  if (s >= 7.0) return { letter: 'B', scale4: 3.0, status: 'Đạt (Khá)', badge: 'bg-blue-50 text-blue-700 border border-blue-200 dark:bg-blue-950/40 dark:text-blue-300 dark:border-blue-800' };
  if (s >= 6.5) return { letter: 'C+', scale4: 2.5, status: 'Đạt (Trung bình khá)', badge: 'bg-amber-50 text-amber-700 border border-amber-200 dark:bg-amber-950/40 dark:text-amber-300 dark:border-amber-800' };
  if (s >= 5.5) return { letter: 'C', scale4: 2.0, status: 'Đạt (Trung bình)', badge: 'bg-amber-50 text-amber-700 border border-amber-200 dark:bg-amber-950/40 dark:text-amber-300 dark:border-amber-800' };
  if (s >= 5.0) return { letter: 'D+', scale4: 1.5, status: 'Đạt (Trung bình yếu)', badge: 'bg-orange-50 text-orange-700 border border-orange-200 dark:bg-orange-950/40 dark:text-orange-300 dark:border-orange-800' };
  if (s >= 4.0) return { letter: 'D', scale4: 1.0, status: 'Đạt (Yếu)', badge: 'bg-orange-50 text-orange-700 border border-orange-200 dark:bg-orange-950/40 dark:text-orange-300 dark:border-orange-800' };
  return { letter: 'F', scale4: 0.0, status: 'Không đạt (Thi lại)', badge: 'bg-rose-50 text-rose-700 border border-rose-200 dark:bg-rose-950/40 dark:text-rose-300 dark:border-rose-800' };
}

export default function StudentGradesPage() {
  const [loading, setLoading] = useState(true);
  const [attempts, setAttempts] = useState([]);
  const [availableExams, setAvailableExams] = useState([]);
  const [classes, setClasses] = useState([]);
  const [chartView, setChartView] = useState('subjects'); // 'subjects' | 'distribution'

  // Simulator State
  const [simCredits, setSimCredits] = useState(12);
  const [simTargetGrade, setSimTargetGrade] = useState('3.5');

  useEffect(() => {
    async function loadAcademicData() {
      setLoading(true);
      try {
        const [attRes, exRes, classRes] = await Promise.allSettled([
          studentAPI.listAttempts(),
          studentAPI.getAvailableExams(),
          classroomAPI.list(),
        ]);

        if (attRes.status === 'fulfilled') {
          setAttempts(attRes.value.data?.attempts || []);
        }
        if (exRes.status === 'fulfilled') {
          setAvailableExams(exRes.value.data?.exams || []);
        }
        if (classRes.status === 'fulfilled') {
          setClasses(classRes.value.data?.classes || classRes.value.data?.classrooms || []);
        }
      } finally {
        setLoading(false);
      }
    }

    loadAcademicData();
  }, []);

  // Process Real Course List from DB
  const processedCourses = useMemo(() => {
    const now = new Date();
    const map = new Map();

    // 1. Populate from available exams
    availableExams.forEach((ex) => {
      const isPastDeadline = new Date(ex.end_time) < now;
      map.set(ex.id, {
        examId: ex.id,
        code: ex.exam_code || `EX-${ex.id}`,
        title: ex.title,
        duration: ex.duration,
        endTime: ex.end_time,
        credits: 3,
        attemptStatus: ex.attempt_status,
        score10: ex.score !== undefined && ex.score !== null ? Number(ex.score) : null,
        attemptId: ex.attempt_id || null,
        isMissed: isPastDeadline && (ex.attempt_status === 'not_attempted' || !ex.attempt_status),
        isCompleted: ex.attempt_status === 'submitted' || ex.attempt_status === 'time_up',
      });
    });

    // 2. Overlay with attempts
    attempts.forEach((att) => {
      const existing = map.get(att.exam_id) || {
        examId: att.exam_id,
        code: att.exam_code || `EX-${att.exam_id}`,
        title: att.exam_title || `Đề thi #${att.exam_id}`,
        credits: 3,
        duration: 45,
      };

      const score = att.score !== undefined && att.score !== null ? Number(att.score) : existing.score10;
      const isCompleted = att.status === 'completed';

      map.set(att.exam_id, {
        ...existing,
        attemptId: att.id,
        score10: score,
        isCompleted,
        isMissed: false,
        attemptStatus: isCompleted ? 'submitted' : 'in_progress',
      });
    });

    // Calculate grades
    return Array.from(map.values()).map((course) => {
      let finalScore = course.score10;
      if (course.isMissed) {
        finalScore = 0.0;
      }

      const gradeInfo = calculateScale(finalScore);
      const isFailed = finalScore !== null && finalScore < 4.0;

      return {
        ...course,
        finalScore,
        ...gradeInfo,
        isFailed,
      };
    });
  }, [availableExams, attempts]);

  // Overall Statistics from Real Data
  const stats = useMemo(() => {
    const evaluatedCourses = processedCourses.filter((c) => c.isCompleted || c.isMissed);
    const completedPassed = evaluatedCourses.filter((c) => c.finalScore !== null && c.finalScore >= 4.0);
    const failedCourses = evaluatedCourses.filter((c) => c.isFailed || c.isMissed);

    const totalCredits = evaluatedCourses.reduce((acc, c) => acc + c.credits, 0);
    const earnedCredits = completedPassed.reduce((acc, c) => acc + c.credits, 0);

    const totalWeightedPoints4 = evaluatedCourses.reduce((acc, c) => acc + (c.scale4 * c.credits), 0);
    const cpa4 = totalCredits > 0 ? (totalWeightedPoints4 / totalCredits).toFixed(2) : '0.00';

    const totalWeightedPoints10 = evaluatedCourses.reduce((acc, c) => acc + ((c.finalScore || 0) * c.credits), 0);
    const cpa10 = totalCredits > 0 ? (totalWeightedPoints10 / totalCredits).toFixed(2) : '0.00';

    // 1. Tự động xếp loại học lực tốt nghiệp
    let graduationRank = 'Chưa đủ điều kiện';
    let rankBadge = 'bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300';
    const cpaNum = Number(cpa4);

    if (totalCredits > 0) {
      if (cpaNum >= 3.6) {
        graduationRank = 'Bằng Xuất sắc';
        rankBadge = 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/50 dark:text-emerald-300 dark:border-emerald-800';
      } else if (cpaNum >= 3.2) {
        graduationRank = 'Bằng Giỏi';
        rankBadge = 'bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950/50 dark:text-blue-300 dark:border-blue-800';
      } else if (cpaNum >= 2.5) {
        graduationRank = 'Bằng Khá';
        rankBadge = 'bg-indigo-50 text-indigo-700 border-indigo-200 dark:bg-indigo-950/50 dark:text-indigo-300 dark:border-indigo-800';
      } else if (cpaNum >= 2.0) {
        graduationRank = 'Bằng Trung bình';
        rankBadge = 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950/50 dark:text-amber-300 dark:border-amber-800';
      } else {
        graduationRank = 'Cảnh báo: Dưới chuẩn 2.0';
        rankBadge = 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950/50 dark:text-rose-300 dark:border-rose-800';
      }
    }

    // 2. Hệ thống cảnh báo học vụ
    let warningLevel = 0;
    let warningText = 'Tình trạng học vụ: Bình thường';

    const failCount = failedCourses.length;

    if (failCount >= 3 || (cpaNum > 0 && cpaNum < 1.2)) {
      warningLevel = 3;
      warningText = 'CẢNH BÁO HỌC VỤ MỨC 3 (Nguy cơ buộc thôi học)';
    } else if (failCount === 2 || (cpaNum > 0 && cpaNum < 1.6)) {
      warningLevel = 2;
      warningText = 'Cảnh báo học vụ Mức 2 (Nguy cơ đình chỉ học phần)';
    } else if (failCount === 1 || (cpaNum > 0 && cpaNum < 2.0)) {
      warningLevel = 1;
      warningText = 'Cảnh báo học vụ Mức 1 (Cần cải thiện điểm)';
    }

    const distribution = [
      { tier: 'Xuất sắc (A)', letter: 'A', scale4: '4.0', count: processedCourses.filter((c) => c.letter === 'A').length, color: 'bg-emerald-500' },
      { tier: 'Giỏi (B+)', letter: 'B+', scale4: '3.5', count: processedCourses.filter((c) => c.letter === 'B+').length, color: 'bg-blue-500' },
      { tier: 'Khá (B)', letter: 'B', scale4: '3.0', count: processedCourses.filter((c) => c.letter === 'B').length, color: 'bg-indigo-500' },
      { tier: 'TB Khá (C+)', letter: 'C+', scale4: '2.5', count: processedCourses.filter((c) => c.letter === 'C+').length, color: 'bg-amber-500' },
      { tier: 'Trung bình (C)', letter: 'C', scale4: '2.0', count: processedCourses.filter((c) => c.letter === 'C').length, color: 'bg-amber-600' },
      { tier: 'Yếu (D/D+)', letter: 'D/D+', scale4: '1.0-1.5', count: processedCourses.filter((c) => c.letter === 'D' || c.letter === 'D+').length, color: 'bg-orange-500' },
      { tier: 'Không đạt / Bỏ lỡ (F)', letter: 'F', scale4: '0.0', count: failedCourses.length, color: 'bg-rose-500' },
      { tier: 'Chưa có điểm / Đang học', letter: '—', scale4: '—', count: processedCourses.filter((c) => c.finalScore === null).length, color: 'bg-slate-400' },
    ];

    return {
      cpa4,
      cpa10,
      totalCredits,
      earnedCredits,
      failedCourses,
      failCount,
      graduationRank,
      rankBadge,
      warningLevel,
      warningText,
      evaluatedCount: evaluatedCourses.length,
      evaluatedCourses,
      distribution,
    };
  }, [processedCourses]);

  // GPA Simulator
  const simulatedCPA = useMemo(() => {
    const currentCredits = stats.totalCredits;
    const currentPoints = Number(stats.cpa4) * currentCredits;
    const newCredits = Number(simCredits) || 0;
    const targetScale = Number(simTargetGrade) || 0;

    const totalSimCredits = currentCredits + newCredits;
    if (totalSimCredits === 0) return '0.00';

    return ((currentPoints + targetScale * newCredits) / totalSimCredits).toFixed(2);
  }, [stats.totalCredits, stats.cpa4, simCredits, simTargetGrade]);

  const handlePrint = () => {
    window.print();
  };

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-2 border-b border-slate-200 dark:border-slate-800">
        <div>
          <h1 className="text-xl font-bold text-slate-900 dark:text-slate-100 tracking-tight">
            Bảng Điểm &amp; Học Bạ Sinh Viên
          </h1>
          <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
            Dữ liệu kết quả khảo thí thực tế từ cơ sở dữ liệu hệ thống &middot; Thang điểm 4 và Thang điểm 10
          </p>
        </div>

        <div className="flex items-center gap-2.5">
          <button
            onClick={handlePrint}
            className="btn-secondary btn-sm flex items-center gap-1.5"
            title="In bảng điểm học vụ"
          >
            <svg className="w-4 h-4 text-slate-500 dark:text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.75} d="M6.72 13.829c-.24.03-.48.062-.72.096m.72-.096a42.415 42.415 0 0110.56 0m-10.56 0L6.34 18m10.94-4.171c.24.03.48.062.72.096m-.72-.096L17.66 18m0 0l.229 2.523a1.125 1.125 0 01-1.12 1.227H7.231c-.662 0-1.18-.568-1.12-1.227L6.34 18m11.318 0h1.091A2.25 2.25 0 0021 15.75V9.456c0-1.081-.768-2.015-1.837-2.175a48.055 48.055 0 00-1.913-.247M6.34 18H5.25A2.25 2.25 0 013 15.75V9.456c0-1.081.768-2.015 1.837-2.175a48.041 48.041 0 011.913-.247m10.5 0a48.536 48.536 0 00-10.5 0m10.5 0V3.375c0-.621-.504-1.125-1.125-1.125h-8.25c-.621 0-1.125.504-1.125 1.125v3.659M18 10.5h.008v.008H18V10.5zm-3 0h.008v.008H15V10.5z" />
            </svg>
            <span>In bảng điểm</span>
          </button>
        </div>
      </div>

      {loading ? (
        <div className="flex justify-center py-20"><Spinner size="lg" /></div>
      ) : (
        <>
          {/* Academic Warning Notification Banner */}
          {stats.warningLevel > 0 ? (
            <div className={`p-4 rounded-xl border flex items-start justify-between gap-3 shadow-sm ${
              stats.warningLevel === 3
                ? 'bg-rose-50 border-rose-300 dark:bg-rose-950/60 dark:border-rose-800 text-rose-900 dark:text-rose-200'
                : stats.warningLevel === 2
                ? 'bg-orange-50 border-orange-200 dark:bg-orange-950/50 dark:border-orange-800 text-orange-900 dark:text-orange-200'
                : 'bg-amber-50 border-amber-200 dark:bg-amber-950/50 dark:border-amber-800 text-amber-900 dark:text-amber-200'
            }`}>
              <div className="flex items-start gap-3">
                <svg className="w-5 h-5 flex-shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </svg>
                <div>
                  <h4 className="font-bold text-sm">{stats.warningText}</h4>
                  <p className="text-xs mt-1 leading-relaxed opacity-90">
                    Bạn hiện có <strong>{stats.failCount} bài thi</strong> chưa đạt (&lt; 4.0 điểm hoặc đã bỏ lỡ kỳ thi). 
                    Theo quy chế đào tạo tín chỉ, sinh viên cần sớm đăng ký thi lại hoặc liên hệ giảng viên học phần để gỡ điểm.
                  </p>
                </div>
              </div>
            </div>
          ) : (
            <div className="p-3.5 bg-emerald-50 dark:bg-emerald-950/30 border border-emerald-200 dark:border-emerald-800 rounded-xl text-xs text-emerald-800 dark:text-emerald-300 flex items-center justify-between shadow-sm">
              <div className="flex items-center gap-2">
                <svg className="w-4 h-4 text-emerald-600 dark:text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <span className="font-semibold">Tình trạng học vụ: Bình thường</span>
                <span className="text-emerald-600 dark:text-emerald-400">&mdash; Bạn đang duy trì tiến độ học tập tốt và không nợ môn thi nào.</span>
              </div>
            </div>
          )}

          {/* Top 4 Performance Cards */}
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
            {/* Card 1: CPA */}
            <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4 shadow-sm">
              <div className="text-[11px] font-bold text-slate-400 dark:text-slate-500 uppercase tracking-wider">
                Điểm tích lũy (CPA)
              </div>
              <div className="mt-2 flex items-baseline gap-2">
                <span className="text-3xl font-extrabold text-blue-600 dark:text-blue-400">
                  {stats.cpa4}
                </span>
                <span className="text-xs text-slate-400 font-medium">/ 4.0</span>
                <span className="text-xs font-semibold text-slate-600 dark:text-slate-300 ml-auto">
                  ({stats.cpa10} / 10)
                </span>
              </div>
              <div className="mt-2 text-xs font-medium text-slate-500 dark:text-slate-400">
                Dựa trên <strong className="text-slate-800 dark:text-slate-200">{stats.evaluatedCount} môn thi</strong> trong hệ thống
              </div>
            </div>

            {/* Card 2: Tự động Xếp loại Học Lực */}
            <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4 shadow-sm">
              <div className="text-[11px] font-bold text-slate-400 dark:text-slate-500 uppercase tracking-wider">
                Xếp loại học lực dự kiến
              </div>
              <div className="mt-2">
                <span className={`inline-flex px-2.5 py-1 rounded-md text-xs font-bold border ${stats.rankBadge}`}>
                  {stats.graduationRank}
                </span>
              </div>
              <p className="mt-2 text-[11px] text-slate-500 dark:text-slate-400">
                Xếp hạng tự động theo Quy chế tín chỉ Bộ GD&amp;ĐT
              </p>
            </div>

            {/* Card 3: Tín chỉ tích lũy */}
            <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4 shadow-sm">
              <div className="text-[11px] font-bold text-slate-400 dark:text-slate-500 uppercase tracking-wider">
                Tín chỉ tích lũy (Đạt)
              </div>
              <div className="mt-2 flex items-baseline gap-2">
                <span className="text-3xl font-extrabold text-slate-900 dark:text-slate-100">
                  {stats.earnedCredits}
                </span>
                <span className="text-xs text-slate-400 font-medium">/ {stats.totalCredits || 120} TC</span>
              </div>
              <div className="mt-2 text-xs font-medium text-slate-500 dark:text-slate-400">
                Môn đạt: <strong className="text-slate-800 dark:text-slate-200">{stats.totalCredits > 0 ? Math.round((stats.earnedCredits / stats.totalCredits) * 100) : 100}%</strong>
              </div>
            </div>

            {/* Card 4: Số môn trượt / cần thi lại */}
            <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4 shadow-sm">
              <div className="text-[11px] font-bold text-slate-400 dark:text-slate-500 uppercase tracking-wider">
                Môn nợ / Cần thi lại
              </div>
              <div className="mt-2 flex items-baseline gap-2">
                <span className={`text-3xl font-extrabold ${stats.failCount > 0 ? 'text-rose-600 dark:text-rose-400' : 'text-emerald-600 dark:text-emerald-400'}`}>
                  {stats.failCount}
                </span>
                <span className="text-xs text-slate-400 font-medium">học phần</span>
              </div>
              <div className="mt-2 text-xs font-medium text-slate-500 dark:text-slate-400">
                {stats.failCount > 0 ? 'Cần cải thiện / thi lại' : 'Không có môn nào trượt'}
              </div>
            </div>
          </div>

          {/* Biểu đồ trực quan GPA & Điểm Học Phần (Interactive SVG Chart & Distribution) */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-5 shadow-sm">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-5">
              <div>
                <h2 className="text-sm font-bold text-slate-900 dark:text-slate-100 flex items-center gap-2">
                  <span>Biểu Đồ Xu Hướng Điểm &amp; Học Lực Khảo Thí</span>
                  <span className="text-[11px] font-semibold px-2 py-0.5 rounded bg-blue-50 text-blue-700 dark:bg-blue-950/60 dark:text-blue-300 border border-blue-200 dark:border-blue-800">
                    Thang 4.0 &amp; Hệ 10
                  </span>
                </h2>
                <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
                  Dữ liệu trực tiếp từ cơ sở dữ liệu các bài thi đã làm và bài thi được phân quyền
                </p>
              </div>

              {/* View Mode Switcher */}
              <div className="flex items-center gap-1.5 p-1 bg-slate-100 dark:bg-slate-800 rounded-lg self-start sm:self-auto">
                <button
                  type="button"
                  onClick={() => setChartView('subjects')}
                  className={`px-3 py-1 text-xs font-semibold rounded-md transition-all ${
                    chartView === 'subjects'
                      ? 'bg-white dark:bg-slate-900 text-blue-600 dark:text-blue-400 shadow-sm'
                      : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'
                  }`}
                >
                  Từng môn học
                </button>
                <button
                  type="button"
                  onClick={() => setChartView('distribution')}
                  className={`px-3 py-1 text-xs font-semibold rounded-md transition-all ${
                    chartView === 'distribution'
                      ? 'bg-white dark:bg-slate-900 text-blue-600 dark:text-blue-400 shadow-sm'
                      : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'
                  }`}
                >
                  Phân bố học lực
                </button>
              </div>
            </div>

            {chartView === 'subjects' ? (
              <>
                <div className="flex items-center gap-4 text-xs font-medium flex-wrap mb-3 pb-3 border-b border-slate-100 dark:border-slate-800">
                  <div className="flex items-center gap-1.5 text-slate-600 dark:text-slate-300">
                    <span className="w-3 h-3 rounded bg-emerald-500 inline-block"></span>
                    <span>Đạt Giỏi/Xuất sắc (&ge; 3.5)</span>
                  </div>
                  <div className="flex items-center gap-1.5 text-slate-600 dark:text-slate-300">
                    <span className="w-3 h-3 rounded bg-blue-500 inline-block"></span>
                    <span>Đạt Khá/TB (&ge; 2.0)</span>
                  </div>
                  <div className="flex items-center gap-1.5 text-slate-600 dark:text-slate-300">
                    <span className="w-3 h-3 rounded bg-rose-500 inline-block"></span>
                    <span>Không đạt / Bỏ lỡ (&lt; 1.0)</span>
                  </div>
                  <div className="flex items-center gap-1.5 text-slate-500 dark:text-slate-400">
                    <span className="w-3 h-3 rounded border border-dashed border-slate-400 bg-slate-200/50 dark:bg-slate-800 inline-block"></span>
                    <span>Chưa có điểm / Chờ thi</span>
                  </div>
                </div>

                {/* SVG Visual Chart */}
                <div className="w-full overflow-x-auto">
                  <div className="min-w-[620px] h-64 pt-2">
                    <svg viewBox="0 0 620 210" className="w-full h-full">
                      {/* Grid lines for 4.0, 3.2, 2.0, 1.0 */}
                      <line x1="50" y1="20" x2="590" y2="20" stroke="currentColor" strokeDasharray="3 3" className="text-slate-200 dark:text-slate-800" />
                      <text x="35" y="24" fontSize="10" className="fill-slate-400 font-mono" textAnchor="end">4.0</text>

                      {/* Line 3.2 Good */}
                      <line x1="50" y1="50" x2="590" y2="50" stroke="#10b981" strokeDasharray="4 4" strokeWidth="1.2" opacity="0.7" />
                      <text x="35" y="54" fontSize="9" fill="#10b981" fontWeight="bold" textAnchor="end">3.2</text>

                      {/* Line 2.0 Warning */}
                      <line x1="50" y1="95" x2="590" y2="95" stroke="#f59e0b" strokeDasharray="3 3" opacity="0.6" />
                      <text x="35" y="99" fontSize="9" fill="#f59e0b" textAnchor="end">2.0</text>

                      {/* Line 1.0 Pass */}
                      <line x1="50" y1="132" x2="590" y2="132" stroke="#ef4444" strokeDasharray="3 3" opacity="0.6" />
                      <text x="35" y="136" fontSize="9" fill="#ef4444" textAnchor="end">1.0</text>

                      {/* Baseline 0.0 */}
                      <line x1="50" y1="170" x2="590" y2="170" stroke="currentColor" className="text-slate-300 dark:text-slate-700" strokeWidth="1.5" />
                      <text x="35" y="174" fontSize="10" className="fill-slate-400 font-mono" textAnchor="end">0.0</text>

                      {/* Render Columns for Processed Courses */}
                      {processedCourses.length === 0 ? (
                        <text x="320" y="100" textAnchor="middle" className="fill-slate-400 text-xs">Chưa có bài thi nào</text>
                      ) : (
                        processedCourses.map((c, idx) => {
                          const totalItems = processedCourses.length;
                          const step = Math.min(130, 500 / Math.max(1, totalItems));
                          const startX = 50 + (540 - (totalItems - 1) * step) / 2;
                          const x = startX + idx * step;
                          const isPending = c.finalScore === null || c.finalScore === undefined;

                          if (isPending) {
                            // Render visible dashed outline column for pending course
                            const phY = 85;
                            const phHeight = 170 - phY;
                            return (
                              <g key={c.examId} className="group">
                                <rect
                                  x={x - 24}
                                  y={phY}
                                  width="48"
                                  height={phHeight}
                                  rx="6"
                                  strokeDasharray="4 4"
                                  className="stroke-slate-300 dark:stroke-slate-600 fill-slate-100/40 dark:fill-slate-800/40"
                                  strokeWidth="1.5"
                                />
                                <text
                                  x={x}
                                  y={phY - 7}
                                  fontSize="10"
                                  fontWeight="600"
                                  textAnchor="middle"
                                  className="fill-slate-400 dark:fill-slate-500"
                                >
                                  Chờ điểm
                                </text>
                                <text
                                  x={x}
                                  y={phY + 45}
                                  fontSize="9"
                                  textAnchor="middle"
                                  className="fill-slate-400 dark:fill-slate-500 font-medium"
                                >
                                  Đang học
                                </text>
                                <text
                                  x={x}
                                  y="188"
                                  fontSize="10"
                                  fontWeight="bold"
                                  fontFamily="monospace"
                                  textAnchor="middle"
                                  className="fill-slate-600 dark:fill-slate-400"
                                >
                                  {c.code}
                                </text>
                              </g>
                            );
                          }

                          const val = Number(c.scale4 || 0);
                          const y = 170 - (val / 4.0) * 150;
                          const barHeight = Math.max(6, 170 - y);

                          let barColor = 'fill-blue-500';
                          let topColor = 'fill-blue-600';
                          if (c.isMissed || c.isFailed) {
                            barColor = 'fill-rose-400';
                            topColor = 'fill-rose-600';
                          } else if (val >= 3.5) {
                            barColor = 'fill-emerald-400';
                            topColor = 'fill-emerald-600';
                          } else if (val < 2.0) {
                            barColor = 'fill-orange-400';
                            topColor = 'fill-orange-500';
                          }

                          return (
                            <g key={c.examId} className="cursor-pointer group">
                              <rect
                                x={x - 22}
                                y={y}
                                width="44"
                                height={barHeight}
                                rx="5"
                                className={`${barColor} hover:opacity-90 transition-opacity`}
                              />
                              <rect
                                x={x - 22}
                                y={y}
                                width="44"
                                height="3"
                                rx="1.5"
                                className={topColor}
                              />
                              <text
                                x={x}
                                y={y - 7}
                                fontSize="11"
                                fontWeight="bold"
                                textAnchor="middle"
                                className="fill-slate-900 dark:fill-slate-100"
                              >
                                {c.isMissed ? '0.0 (Bỏ lỡ)' : `${val.toFixed(1)} (${c.letter})`}
                              </text>
                              <text
                                x={x}
                                y="188"
                                fontSize="10"
                                fontWeight="bold"
                                fontFamily="monospace"
                                textAnchor="middle"
                                className="fill-slate-700 dark:fill-slate-300"
                              >
                                {c.code}
                              </text>
                            </g>
                          );
                        })
                      )}
                    </svg>
                  </div>
                </div>
              </>
            ) : (
              /* Grade Tier Distribution View */
              <div className="space-y-3 pt-2">
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
                  {stats.distribution.map((tier) => {
                    const percent = processedCourses.length > 0
                      ? Math.round((tier.count / processedCourses.length) * 100)
                      : 0;

                    return (
                      <div
                        key={tier.tier}
                        className="p-3 rounded-lg border border-slate-100 dark:border-slate-800 bg-slate-50/60 dark:bg-slate-800/40 flex items-center justify-between"
                      >
                        <div>
                          <div className="flex items-center gap-2">
                            <span className={`w-2.5 h-2.5 rounded-full ${tier.color}`}></span>
                            <span className="text-xs font-bold text-slate-800 dark:text-slate-200">
                              {tier.tier}
                            </span>
                          </div>
                          <span className="text-[11px] text-slate-500 dark:text-slate-400 mt-1 block">
                            Hệ 4: {tier.scale4}
                          </span>
                        </div>
                        <div className="text-right">
                          <span className="text-lg font-extrabold text-slate-900 dark:text-slate-100">
                            {tier.count}
                          </span>
                          <span className="text-[11px] text-slate-400 block font-mono">
                            {percent}%
                          </span>
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            )}
          </div>

          {/* Action Required: Failed & Missed Exams List */}
          {stats.failedCourses.length > 0 && (
            <div className="bg-white dark:bg-slate-900 rounded-xl border border-rose-200 dark:border-rose-900/60 p-5 shadow-sm">
              <div className="flex items-center gap-2 mb-3">
                <svg className="w-5 h-5 text-rose-600 dark:text-rose-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </svg>
                <h3 className="text-sm font-bold text-rose-700 dark:text-rose-400">
                  Danh Sách Môn Thi Cần Đăng Ký Thi Lại Hoặc Phúc Khảo
                </h3>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full text-xs">
                  <thead>
                    <tr className="border-b border-rose-100 dark:border-rose-900/40 text-left font-semibold text-rose-800 dark:text-rose-300">
                      <th className="py-2 px-3 whitespace-nowrap">Mã đề</th>
                      <th className="py-2 px-3">Tên môn thi</th>
                      <th className="py-2 px-3 text-center whitespace-nowrap">Tình trạng</th>
                      <th className="py-2 px-3 text-center whitespace-nowrap">Điểm số</th>
                      <th className="py-2 px-3 text-right whitespace-nowrap">Thao tác</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-rose-50 dark:divide-rose-950/60">
                    {stats.failedCourses.map((c) => (
                      <tr key={c.examId}>
                        <td className="py-2.5 px-3 font-mono font-bold text-slate-800 dark:text-slate-200 whitespace-nowrap">{c.code}</td>
                        <td className="py-2.5 px-3 font-medium text-slate-900 dark:text-slate-100">{c.title}</td>
                        <td className="py-2.5 px-3 text-center whitespace-nowrap">
                          {c.isMissed ? (
                            <span className="inline-block px-2.5 py-0.5 rounded text-[11px] font-semibold bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300 whitespace-nowrap">
                              Bỏ lỡ kỳ thi (Quá hạn)
                            </span>
                          ) : (
                            <span className="inline-block px-2.5 py-0.5 rounded text-[11px] font-semibold bg-rose-50 text-rose-700 dark:bg-rose-950/60 dark:text-rose-300 border border-rose-200 dark:border-rose-800 whitespace-nowrap">
                              Không đạt (&lt; 4.0 điểm)
                            </span>
                          )}
                        </td>
                        <td className="py-2.5 px-3 text-center font-bold text-rose-600 dark:text-rose-400 font-mono whitespace-nowrap">
                          {c.finalScore !== null ? Number(c.finalScore).toFixed(2) : '0.00'}
                        </td>
                        <td className="py-2.5 px-3 text-right whitespace-nowrap">
                          {c.attemptId ? (
                            <Link
                              to={`/student/attempts/${c.attemptId}/result`}
                              className="text-xs font-semibold text-blue-600 dark:text-blue-400 hover:underline"
                            >
                              Xem bài làm &amp; Phúc khảo &rarr;
                            </Link>
                          ) : (
                            <span className="text-slate-400 text-xs">Liên hệ phòng Đào tạo</span>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Interactive GPA Simulator (Dự Báo Kéo Điểm) */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-5 shadow-sm">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-4">
              <div>
                <h3 className="text-sm font-bold text-slate-900 dark:text-slate-100">
                  Công Cụ Dự Báo Kéo Điểm &amp; Nâng Hạng Tốt Nghiệp (GPA Simulator)
                </h3>
                <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
                  Thử tính xem nếu kỳ tới bạn đạt mức điểm mục tiêu thì CPA tích lũy sẽ tăng lên bao nhiêu
                </p>
              </div>
              <div className="p-2.5 rounded-lg bg-blue-50/70 dark:bg-blue-950/40 border border-blue-200 dark:border-blue-900 text-center">
                <span className="text-xs text-slate-600 dark:text-slate-300 font-medium">CPA Dự kiến: </span>
                <span className="text-base font-extrabold text-blue-700 dark:text-blue-300 ml-1">
                  {simulatedCPA} / 4.0
                </span>
              </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 bg-slate-50 dark:bg-slate-800/40 p-4 rounded-lg border border-slate-100 dark:border-slate-800">
              <div>
                <label className="label">Số tín chỉ dự kiến học kỳ tới</label>
                <input
                  type="number"
                  min="3"
                  max="30"
                  value={simCredits}
                  onChange={(e) => setSimCredits(parseInt(e.target.value, 10) || 0)}
                  className="input h-9 text-xs"
                />
              </div>
              <div>
                <label className="label">Điểm trung bình kỳ vọng kỳ tới</label>
                <select
                  value={simTargetGrade}
                  onChange={(e) => setSimTargetGrade(e.target.value)}
                  className="input h-9 text-xs"
                >
                  <option value="4.0">Điểm A (4.0 - Xuất sắc / 8.5 - 10)</option>
                  <option value="3.5">Điểm B+ (3.5 - Giỏi / 8.0 - 8.4)</option>
                  <option value="3.0">Điểm B (3.0 - Khá / 7.0 - 7.9)</option>
                  <option value="2.5">Điểm C+ (2.5 - Trung bình khá / 6.5 - 6.9)</option>
                  <option value="2.0">Điểm C (2.0 - Trung bình / 5.5 - 6.4)</option>
                </select>
              </div>
            </div>
          </div>

          {/* Full Academic Transcript Table */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 overflow-hidden shadow-sm">
            <div className="px-5 py-3.5 bg-slate-50/80 dark:bg-slate-800/50 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between">
              <div>
                <h3 className="text-sm font-bold text-slate-900 dark:text-slate-100">
                  Bảng Điểm Chi Tiết Tất Cả Các Môn Khảo Thí
                </h3>
                <p className="text-[11px] text-slate-500 dark:text-slate-400">
                  Tổng hợp từ cơ sở dữ liệu các bài thi đã làm và bài thi được phân quyền
                </p>
              </div>
              <span className="text-xs text-slate-500 dark:text-slate-400 font-medium">
                Tổng cộng: <strong className="text-slate-800 dark:text-slate-200">{processedCourses.length} môn</strong>
              </span>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-xs">
                <thead>
                  <tr className="bg-slate-100/60 dark:bg-slate-800/30 text-left text-slate-600 dark:text-slate-400 font-semibold border-b border-slate-200 dark:border-slate-800">
                    <th className="px-4 py-2.5 whitespace-nowrap">Mã đề</th>
                    <th className="px-4 py-2.5">Tên môn học / kỳ thi</th>
                    <th className="px-4 py-2.5 text-center whitespace-nowrap">Số TC</th>
                    <th className="px-4 py-2.5 text-center whitespace-nowrap">Thời lượng</th>
                    <th className="px-4 py-2.5 text-center whitespace-nowrap">Điểm hệ 10</th>
                    <th className="px-4 py-2.5 text-center whitespace-nowrap">Điểm chữ</th>
                    <th className="px-4 py-2.5 text-center whitespace-nowrap">Hệ 4</th>
                    <th className="px-4 py-2.5 text-center whitespace-nowrap min-w-[160px]">Trạng thái</th>
                    <th className="px-4 py-2.5 text-right whitespace-nowrap">Chi tiết</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
                  {processedCourses.length === 0 ? (
                    <tr>
                      <td colSpan={9} className="text-center py-8 text-slate-500">
                        Chưa có dữ liệu bài thi nào trên tài khoản này.
                      </td>
                    </tr>
                  ) : (
                    processedCourses.map((c) => (
                      <tr key={c.examId} className="hover:bg-slate-50/70 dark:hover:bg-slate-800/40 transition-colors">
                        <td className="px-4 py-3 font-mono font-bold text-slate-700 dark:text-slate-300 whitespace-nowrap">
                          {c.code}
                        </td>
                        <td className="px-4 py-3 font-medium text-slate-900 dark:text-slate-100">
                          {c.title}
                        </td>
                        <td className="px-4 py-3 text-center text-slate-600 dark:text-slate-400 whitespace-nowrap">
                          {c.credits}
                        </td>
                        <td className="px-4 py-3 text-center text-slate-600 dark:text-slate-400 whitespace-nowrap">
                          {c.duration} phút
                        </td>
                        <td className="px-4 py-3 text-center font-mono font-bold text-blue-700 dark:text-blue-400 whitespace-nowrap">
                          {c.finalScore !== null ? Number(c.finalScore).toFixed(2) : '—'}
                        </td>
                        <td className="px-4 py-3 text-center font-bold whitespace-nowrap">
                          {c.letter}
                        </td>
                        <td className="px-4 py-3 text-center font-mono font-bold text-slate-800 dark:text-slate-200 whitespace-nowrap">
                          {c.scale4 ? Number(c.scale4).toFixed(1) : '—'}
                        </td>
                        <td className="px-4 py-3 text-center whitespace-nowrap min-w-[160px]">
                          <span className={`inline-flex items-center justify-center px-3 py-1 rounded-md text-xs font-semibold whitespace-nowrap leading-none ${c.badge}`}>
                            {c.isMissed ? 'Bỏ lỡ kỳ thi' : c.status}
                          </span>
                        </td>
                        <td className="px-4 py-3 text-right whitespace-nowrap">
                          {c.attemptId ? (
                            <Link
                              to={`/student/attempts/${c.attemptId}/result`}
                              className="text-blue-600 dark:text-blue-400 hover:underline font-semibold"
                            >
                              Xem điểm
                            </Link>
                          ) : (
                            <span className="text-slate-400">—</span>
                          )}
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
