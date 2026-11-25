import { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { classroomAPI } from '../api/endpoints';

export default function ClassDetailPage() {
  const { id } = useParams();
  const { user } = useAuth();
  const navigate = useNavigate();
  const isTeacher = user?.role === 'teacher';

  const [classData, setClassData] = useState(null);
  const [teacher, setTeacher] = useState(null);
  const [members, setMembers] = useState([]);
  const [exams, setExams] = useState([]);

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  const [activeTab, setActiveTab] = useState('members'); // 'members' | 'exams' | 'gradebook'

  // Gradebook State
  const [gradebookData, setGradebookData] = useState(null);
  const [gradebookLoading, setGradebookLoading] = useState(false);
  const [gradebookError, setGradebookError] = useState('');
  const [gradebookSearch, setGradebookSearch] = useState('');
  const [gradebookRankFilter, setGradebookRankFilter] = useState('all');

  // Teacher Add Member Modal
  const [showAddModal, setShowAddModal] = useState(false);
  const [addMode, setAddMode] = useState('email'); // 'email' | 'student_code'
  const [addInputValue, setAddInputValue] = useState('');
  const [adding, setAdding] = useState(false);

  // Copy code feedback
  const [copiedCode, setCopiedCode] = useState(false);

  const fetchClassDetail = useCallback(async () => {
    try {
      setLoading(true);
      setError('');
      const { data } = await classroomAPI.get(id);
      setClassData(data.class);
      setTeacher(data.teacher);
      setMembers(data.members || []);
      setExams(data.exams || []);
    } catch (err) {
      setError(err.response?.data?.error || 'Không thể tải chi tiết lớp học');
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    fetchClassDetail();
  }, [fetchClassDetail]);

  const fetchGradebook = useCallback(async () => {
    if (!isTeacher) return;
    try {
      setGradebookLoading(true);
      setGradebookError('');
      const { data } = await classroomAPI.getGradebook(id);
      setGradebookData(data);
    } catch (err) {
      setGradebookError(err.response?.data?.error || 'Không thể tải sổ điểm lớp học');
    } finally {
      setGradebookLoading(false);
    }
  }, [id, isTeacher]);

  useEffect(() => {
    if (activeTab === 'gradebook' && !gradebookData) {
      fetchGradebook();
    }
  }, [activeTab, gradebookData, fetchGradebook]);

  const handleExportGradebookCSV = () => {
    if (!gradebookData) return;
    const examsList = gradebookData.exams || [];
    const studentsList = gradebookData.students || [];

    let csvContent = '\uFEFF'; // UTF-8 BOM for Excel
    const headers = ['STT', 'Mã SV', 'Họ và tên', 'Email', 'Khoa / Bộ môn'];
    examsList.forEach((ex) => headers.push(`"${ex.title} (${ex.exam_code})"`));
    headers.push('Điểm trung bình', 'Xếp loại');
    csvContent += headers.join(',') + '\r\n';

    studentsList.forEach((st, idx) => {
      const row = [
        idx + 1,
        `"${st.student_code || ''}"`,
        `"${st.full_name}"`,
        `"${st.email}"`,
        `"${st.department || ''}"`,
      ];
      examsList.forEach((ex) => {
        const scoreEntry = st.exam_scores?.[ex.id];
        if (scoreEntry && scoreEntry.score !== null && scoreEntry.score !== undefined) {
          row.push(Number(scoreEntry.score).toFixed(2));
        } else {
          row.push('Chưa thi');
        }
      });
      row.push(Number(st.average_score).toFixed(2));
      row.push(`"${st.academic_rank}"`);
      csvContent += row.join(',') + '\r\n';
    });

    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.setAttribute('href', url);
    link.setAttribute('download', `so-diem-${gradebookData.class_code || 'lop'}-${new Date().toISOString().slice(0, 10)}.csv`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  const handleCopyCode = () => {
    if (!classData?.code) return;
    navigator.clipboard.writeText(classData.code);
    setCopiedCode(true);
    setTimeout(() => setCopiedCode(false), 2000);
  };

  const handleAddMember = async (e) => {
    e.preventDefault();
    const val = addInputValue.trim();
    if (!val) {
      setError('Vui lòng nhập thông tin tìm kiếm sinh viên');
      return;
    }

    try {
      setAdding(true);
      setError('');
      setSuccess('');
      const payload = {};
      if (addMode === 'email') payload.email = val;
      if (addMode === 'student_code') payload.student_code = val;

      const { data } = await classroomAPI.addMember(id, payload);
      setShowAddModal(false);
      setAddInputValue('');
      setSuccess(`Đã thêm sinh viên ${data.member.student?.full_name || val} vào lớp`);
      fetchClassDetail();
    } catch (err) {
      setError(err.response?.data?.error || 'Không tìm thấy sinh viên hoặc sinh viên đã trong lớp');
    } finally {
      setAdding(false);
    }
  };

  const handleRemoveMember = async (member) => {
    if (!window.confirm(`Xác nhận xóa sinh viên "${member.full_name}" khỏi lớp học này?`)) {
      return;
    }

    try {
      setError('');
      setSuccess('');
      await classroomAPI.removeMember(id, member.student_id);
      setSuccess(`Đã xóa sinh viên ${member.full_name} khỏi lớp`);
      setMembers((prev) => prev.filter((m) => m.student_id !== member.student_id));
    } catch (err) {
      setError(err.response?.data?.error || 'Không thể xóa sinh viên');
    }
  };

  if (loading) {
    return (
      <div className="py-12 text-center text-slate-500 text-sm">
        Đang tải thông tin lớp học...
      </div>
    );
  }

  if (error && !classData) {
    return (
      <div className="space-y-4">
        <button
          onClick={() => navigate('/classes')}
          className="text-xs text-indigo-600 hover:text-indigo-800 font-medium"
        >
          Quay lại danh sách lớp học
        </button>
        <div className="p-4 bg-red-50 border border-red-200 text-red-700 text-sm rounded-md">
          {error}
        </div>
      </div>
    );
  }

  const teacherInitial = (teacher?.full_name || 'G').charAt(0).toUpperCase();

  return (
    <div className="space-y-6">
      {/* Breadcrumb / Back button */}
      <div>
        <button
          onClick={() => navigate('/classes')}
          className="text-xs text-indigo-600 hover:text-indigo-800 font-medium hover:underline inline-flex items-center gap-1"
        >
          <span>&larr;</span> Quay lại danh sách lớp học
        </button>
      </div>

      {error && (
        <div className="p-3 bg-red-50 border border-red-200 text-red-700 text-sm rounded-md">
          {error}
        </div>
      )}

      {success && (
        <div className="p-3 bg-emerald-50 border border-emerald-200 text-emerald-700 text-sm rounded-md">
          {success}
        </div>
      )}

      {/* Class Overview Header */}
      <div className="bg-white dark:bg-slate-900 p-6 rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm space-y-4">
        <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4 border-b border-slate-100 dark:border-slate-800 pb-4">
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-xl font-bold text-slate-900 dark:text-slate-100 tracking-tight">
                {classData?.name}
              </h1>
              <div className="flex items-center gap-1.5">
                <span className="font-mono text-xs font-bold bg-indigo-50 text-indigo-700 border border-indigo-200 px-2.5 py-0.5 rounded">
                  {classData?.code}
                </span>
                <button
                  type="button"
                  onClick={handleCopyCode}
                  className="text-xs text-slate-500 hover:text-slate-800 p-1 rounded hover:bg-slate-100"
                  title="Sao chép mã lớp"
                >
                  {copiedCode ? (
                    <span className="text-emerald-600 font-medium text-[11px]">Đã chép</span>
                  ) : (
                    <span className="text-[11px] font-medium underline">Chép</span>
                  )}
                </button>
              </div>
            </div>
            <p className="text-sm text-slate-600 mt-2 max-w-2xl">
              {classData?.description || 'Chưa có mô tả cho lớp học này.'}
            </p>
          </div>

          {/* Teacher Info Card */}
          {teacher && (
            <div className="bg-slate-50 dark:bg-slate-800/60 p-3.5 rounded-lg border border-slate-200 dark:border-slate-700 flex items-center gap-3 min-w-[240px]">
              {teacher.avatar_url ? (
                <img
                  src={teacher.avatar_url}
                  alt={teacher.full_name}
                  className="w-10 h-10 rounded-full object-cover border border-slate-300"
                />
              ) : (
                <div className="w-10 h-10 rounded-full bg-slate-800 text-white flex items-center justify-center font-bold text-sm">
                  {teacherInitial}
                </div>
              )}
              <div className="min-w-0">
                <span className="text-[10px] uppercase font-bold text-slate-500">Giảng viên phụ trách</span>
                <h4 className="text-xs font-bold text-slate-900 dark:text-slate-100 truncate">{teacher.full_name}</h4>
                <p className="text-[11px] text-slate-500 dark:text-slate-400 truncate">{teacher.email}</p>
              </div>
            </div>
          )}
        </div>

        {/* Stats */}
        <div className="flex items-center gap-6 text-xs text-slate-600 dark:text-slate-400">
          <div>
            Số lượng sinh viên: <span className="font-bold text-slate-900 dark:text-slate-100">{members.length}</span>
          </div>
          <div>
            Số đề thi phân bổ: <span className="font-bold text-slate-900 dark:text-slate-100">{exams.length}</span>
          </div>
          <div className="text-slate-400">
            Ngày thành lập: {new Date(classData?.created_at).toLocaleDateString('vi-VN')}
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex items-center justify-between border-b border-slate-200 dark:border-slate-800">
        <div className="flex gap-4">
          <button
            onClick={() => setActiveTab('members')}
            className={`pb-3 text-sm font-semibold border-b-2 transition-colors ${
              activeTab === 'members'
                ? 'border-indigo-600 text-indigo-600'
                : 'border-transparent text-slate-500 hover:text-slate-800'
            }`}
          >
            Danh sách thành viên ({members.length})
          </button>
          <button
            onClick={() => setActiveTab('exams')}
            className={`pb-3 text-sm font-semibold border-b-2 transition-colors ${
              activeTab === 'exams'
                ? 'border-indigo-600 text-indigo-600'
                : 'border-transparent text-slate-500 hover:text-slate-800'
            }`}
          >
            Đề thi lớp học ({exams.length})
          </button>
          {isTeacher && (
            <button
              onClick={() => setActiveTab('gradebook')}
              className={`pb-3 text-sm font-semibold border-b-2 transition-colors ${
                activeTab === 'gradebook'
                  ? 'border-indigo-600 text-indigo-600 dark:border-indigo-400 dark:text-indigo-400'
                  : 'border-transparent text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-slate-200'
              }`}
            >
              Sổ điểm tổng hợp
            </button>
          )}
        </div>

        <div className="flex items-center gap-2">
          {isTeacher && activeTab === 'gradebook' && gradebookData && (
            <button
              type="button"
              onClick={handleExportGradebookCSV}
              className="btn-secondary text-xs font-semibold px-3 py-1.5 flex items-center gap-1.5 border border-slate-300 dark:border-slate-700 rounded-lg text-slate-700 dark:text-slate-200 hover:bg-slate-50 dark:hover:bg-slate-800"
            >
              <svg className="w-4 h-4 text-emerald-600 dark:text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              Xuất Excel (CSV)
            </button>
          )}
          {isTeacher && activeTab === 'members' && (
            <button
              onClick={() => setShowAddModal(true)}
              className="btn-primary text-xs font-semibold px-3 py-1.5 bg-indigo-600 text-white rounded hover:bg-indigo-700"
            >
              Thêm sinh viên
            </button>
          )}
        </div>
      </div>

      {/* Tab 1: Members */}
      {activeTab === 'members' && (
        <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden">
          {members.length === 0 ? (
            <div className="py-12 text-center text-slate-500 dark:text-slate-400 text-sm">
              Lớp học hiện chưa có sinh viên nào ghi danh.
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-slate-200 dark:divide-slate-800 text-left text-sm">
                <thead className="bg-slate-50 dark:bg-slate-800/60 text-slate-700 dark:text-slate-300 uppercase text-xs">
                  <tr>
                    <th className="px-4 py-3 font-semibold">Sinh viên</th>
                    <th className="px-4 py-3 font-semibold">Mã sinh viên</th>
                    <th className="px-4 py-3 font-semibold">Khoa / Bộ môn</th>
                    <th className="px-4 py-3 font-semibold">Ngày tham gia</th>
                    {isTeacher && <th className="px-4 py-3 font-semibold text-right">Thao tác</th>}
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
                  {members.map((m) => {
                    const initial = (m.full_name || 'S').charAt(0).toUpperCase();
                    return (
                      <tr key={m.id} className="hover:bg-slate-50/80 dark:hover:bg-slate-800/40 transition-colors">
                        <td className="px-4 py-3">
                          <div className="flex items-center gap-3">
                            {m.avatar_url ? (
                              <img
                                src={m.avatar_url}
                                alt={m.full_name}
                                className="w-8 h-8 rounded-full object-cover border border-slate-200"
                              />
                            ) : (
                              <div className="w-8 h-8 rounded-full bg-slate-800 text-white flex items-center justify-center font-bold text-xs">
                                {initial}
                              </div>
                            )}
                            <div>
                              <div className="font-semibold text-slate-900 dark:text-slate-100">{m.full_name}</div>
                              <div className="text-xs text-slate-500">{m.email}</div>
                            </div>
                          </div>
                        </td>
                        <td className="px-4 py-3 font-mono text-xs text-slate-700 dark:text-slate-300">
                          {m.student_code || 'Chưa cập nhật'}
                        </td>
                        <td className="px-4 py-3 text-xs text-slate-600 dark:text-slate-400">
                          {m.department || 'Chưa cập nhật'}
                        </td>
                        <td className="px-4 py-3 text-xs text-slate-500 whitespace-nowrap">
                          {new Date(m.joined_at).toLocaleDateString('vi-VN')}
                        </td>
                        {isTeacher && (
                          <td className="px-4 py-3 text-right">
                            <button
                              type="button"
                              onClick={() => handleRemoveMember(m)}
                              className="text-xs text-red-600 hover:text-red-800 font-medium hover:underline"
                            >
                              Xóa khỏi lớp
                            </button>
                          </td>
                        )}
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* Tab 2: Exams */}
      {activeTab === 'exams' && (
        <div className="space-y-4">
          {exams.length === 0 ? (
            <div className="bg-white dark:bg-slate-900 p-8 rounded-xl border border-slate-200 dark:border-slate-800 text-center text-slate-500 dark:text-slate-400 text-sm shadow-sm">
              Chưa có đề thi nào được gán cho lớp học này.
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {exams.map((exam) => (
                <div
                  key={exam.id}
                  className="bg-white dark:bg-slate-900 p-5 rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm flex flex-col justify-between hover:border-slate-300 dark:hover:border-slate-700 transition-all"
                >
                  <div>
                    <div className="flex items-start justify-between gap-3 mb-2">
                      <h3 className="text-sm font-bold text-slate-900 dark:text-slate-100 leading-snug">
                        {exam.title}
                      </h3>
                      <span
                        className={`text-[11px] font-semibold px-2 py-0.5 rounded ${
                          exam.is_published
                            ? 'bg-emerald-50 text-emerald-700 border border-emerald-200'
                            : 'bg-slate-100 text-slate-600 border border-slate-200'
                        }`}
                      >
                        {exam.is_published ? 'Đã công bố' : 'Bản nháp'}
                      </span>
                    </div>
                    <p className="text-xs text-slate-500 line-clamp-2 mb-3">
                      {exam.description || 'Không có mô tả cho đề thi này.'}
                    </p>
                    <div className="flex items-center gap-3 text-xs text-slate-600">
                      <div>
                        Thời gian: <span className="font-semibold text-slate-800 dark:text-slate-200">{exam.duration_minutes} phút</span>
                      </div>
                      <div className="font-mono text-[11px] bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 px-1.5 py-0.5 rounded">
                        Mã: {exam.code}
                      </div>
                    </div>
                  </div>

                  <div className="mt-4 pt-3 border-t border-slate-100 dark:border-slate-800 flex items-center justify-end">
                    {isTeacher ? (
                      <button
                        type="button"
                        onClick={() => navigate(`/teacher/exams/${exam.id}`)}
                        className="btn-primary text-xs font-semibold px-3 py-1.5 bg-indigo-600 text-white rounded hover:bg-indigo-700"
                      >
                        Xem chi tiết đề thi
                      </button>
                    ) : (
                      <button
                        type="button"
                        onClick={() => navigate('/student')}
                        className="btn-primary text-xs font-semibold px-3 py-1.5 bg-indigo-600 text-white rounded hover:bg-indigo-700"
                      >
                        Đi đến cổng thi
                      </button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Tab 3: Gradebook */}
      {isTeacher && activeTab === 'gradebook' && (
        <div className="space-y-5">
          {gradebookLoading ? (
            <div className="py-16 text-center text-slate-500 text-sm">
              Đang tải sổ điểm lớp học...
            </div>
          ) : gradebookError ? (
            <div className="p-4 bg-red-50 dark:bg-rose-950/40 border border-red-200 dark:border-rose-800 text-red-700 dark:text-rose-300 text-sm rounded-lg">
              {gradebookError}
            </div>
          ) : !gradebookData || (gradebookData.students?.length === 0 && gradebookData.exams?.length === 0) ? (
            <div className="bg-white dark:bg-slate-900 p-12 rounded-xl border border-slate-200 dark:border-slate-800 text-center text-slate-500 dark:text-slate-400 text-sm">
              Lớp học hiện chưa có đủ dữ liệu (Cần có ít nhất 1 sinh viên và 1 đề thi).
            </div>
          ) : (
            <>
              {/* Summary Cards */}
              <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
                <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4 shadow-sm">
                  <div className="text-[11px] font-bold text-slate-400 dark:text-slate-500 uppercase tracking-wider">
                    Điểm trung bình lớp
                  </div>
                  <div className="mt-2 flex items-baseline gap-2">
                    <span className="text-2xl font-extrabold text-blue-600 dark:text-blue-400 font-mono">
                      {gradebookData.summary?.class_average !== undefined
                        ? Number(gradebookData.summary.class_average).toFixed(2)
                        : '0.00'}
                    </span>
                    <span className="text-xs text-slate-400 font-medium">/ 10.0</span>
                  </div>
                  <div className="mt-1 text-xs text-slate-500 dark:text-slate-400">
                    Thang điểm hệ 10
                  </div>
                </div>

                <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4 shadow-sm">
                  <div className="text-[11px] font-bold text-slate-400 dark:text-slate-500 uppercase tracking-wider">
                    Biên độ điểm số
                  </div>
                  <div className="mt-2 flex items-baseline gap-2">
                    <span className="text-2xl font-extrabold text-emerald-600 dark:text-emerald-400 font-mono">
                      {Number(gradebookData.summary?.highest_average || 0).toFixed(2)}
                    </span>
                    <span className="text-xs text-slate-400">cao nhất</span>
                    <span className="text-xs text-slate-500 font-mono ml-auto">
                      {Number(gradebookData.summary?.lowest_average || 0).toFixed(2)} min
                    </span>
                  </div>
                  <div className="mt-1 text-xs text-slate-500 dark:text-slate-400">
                    Chênh lệch: {(Number(gradebookData.summary?.highest_average || 0) - Number(gradebookData.summary?.lowest_average || 0)).toFixed(2)} điểm
                  </div>
                </div>

                <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4 shadow-sm">
                  <div className="text-[11px] font-bold text-slate-400 dark:text-slate-500 uppercase tracking-wider">
                    Cảnh báo học vụ
                  </div>
                  <div className="mt-2 flex items-baseline gap-2">
                    <span className={`text-2xl font-extrabold font-mono ${
                      gradebookData.summary?.at_risk_count > 0 ? 'text-rose-600 dark:text-rose-400' : 'text-slate-800 dark:text-slate-200'
                    }`}>
                      {gradebookData.summary?.at_risk_count || 0}
                    </span>
                    <span className="text-xs text-slate-400">sinh viên</span>
                  </div>
                  <div className="mt-1 text-xs text-rose-600 dark:text-rose-400">
                    Điểm TB &lt; 4.0 (nguy cơ thi lại)
                  </div>
                </div>

                <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4 shadow-sm">
                  <div className="text-[11px] font-bold text-slate-400 dark:text-slate-500 uppercase tracking-wider">
                    Quy mô khảo thí
                  </div>
                  <div className="mt-2 flex items-baseline gap-2">
                    <span className="text-2xl font-extrabold text-indigo-600 dark:text-indigo-400 font-mono">
                      {gradebookData.summary?.total_students || 0}
                    </span>
                    <span className="text-xs text-slate-400">SV / {gradebookData.summary?.total_exams || 0} đề thi</span>
                  </div>
                  <div className="mt-1 text-xs text-slate-500 dark:text-slate-400">
                    Tổng lượt tham gia lớp
                  </div>
                </div>
              </div>

              {/* Toolbar */}
              <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-3.5 flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3 shadow-sm">
                <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-3 flex-1">
                  <div className="relative flex-1">
                    <input
                      type="text"
                      placeholder="Tìm theo tên, email hoặc mã số sinh viên..."
                      value={gradebookSearch}
                      onChange={(e) => setGradebookSearch(e.target.value)}
                      className="w-full text-xs px-3 py-2 border border-slate-300 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 placeholder:text-slate-400 focus:outline-none focus:ring-1 focus:ring-blue-500"
                    />
                  </div>

                  <select
                    value={gradebookRankFilter}
                    onChange={(e) => setGradebookRankFilter(e.target.value)}
                    className="text-xs px-3 py-2 border border-slate-300 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-blue-500"
                  >
                    <option value="all">Tất cả xếp loại</option>
                    <option value="Xuất sắc">Xuất sắc (&gt;= 8.5)</option>
                    <option value="Giỏi">Giỏi (7.0 - 8.4)</option>
                    <option value="Khá">Khá (5.5 - 6.9)</option>
                    <option value="Trung bình">Trung bình (4.0 - 5.4)</option>
                    <option value="Yếu / Cảnh báo">Yếu / Cảnh báo (&lt; 4.0)</option>
                  </select>
                </div>

                <button
                  type="button"
                  onClick={handleExportGradebookCSV}
                  className="btn-secondary text-xs px-3 py-2 border border-slate-300 dark:border-slate-700 rounded-lg flex items-center justify-center gap-1.5 font-semibold text-slate-700 dark:text-slate-200 hover:bg-slate-50 dark:hover:bg-slate-800"
                >
                  <svg className="w-4 h-4 text-emerald-600 dark:text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                  Tải bảng điểm CSV (Excel)
                </button>
              </div>

              {/* Gradebook Matrix Table */}
              <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden">
                <div className="overflow-x-auto">
                  <table className="w-full text-xs text-left">
                    <thead className="bg-slate-50 dark:bg-slate-800/60 text-slate-600 dark:text-slate-400 font-semibold border-b border-slate-200 dark:border-slate-800 uppercase tracking-wider text-[11px]">
                      <tr>
                        <th className="px-3 py-3 whitespace-nowrap text-center">STT</th>
                        <th className="px-4 py-3 whitespace-nowrap min-w-[200px]">Sinh viên</th>
                        <th className="px-3 py-3 whitespace-nowrap text-center">Mã SV</th>
                        <th className="px-3 py-3 whitespace-nowrap">Khoa / Bộ môn</th>
                        {(gradebookData.exams || []).map((ex) => (
                          <th key={ex.id} className="px-4 py-3 whitespace-nowrap text-center min-w-[120px]">
                            <div className="font-bold text-slate-800 dark:text-slate-200 truncate max-w-[140px]" title={ex.title}>
                              {ex.title}
                            </div>
                            <div className="font-mono text-[10px] text-slate-400">{ex.exam_code}</div>
                          </th>
                        ))}
                        <th className="px-3 py-3 whitespace-nowrap text-center bg-blue-50/60 dark:bg-blue-950/30 text-blue-900 dark:text-blue-300 font-bold min-w-[90px]">
                          Điểm TB
                        </th>
                        <th className="px-3 py-3 whitespace-nowrap text-center min-w-[110px]">
                          Xếp loại
                        </th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
                      {(() => {
                        const filtered = (gradebookData.students || []).filter((st) => {
                          const q = gradebookSearch.trim().toLowerCase();
                          const matchSearch =
                            !q ||
                            st.full_name?.toLowerCase().includes(q) ||
                            st.email?.toLowerCase().includes(q) ||
                            st.student_code?.toLowerCase().includes(q);
                          const matchRank =
                            gradebookRankFilter === 'all' || st.academic_rank === gradebookRankFilter;
                          return matchSearch && matchRank;
                        });

                        if (filtered.length === 0) {
                          return (
                            <tr>
                              <td
                                colSpan={6 + (gradebookData.exams?.length || 0)}
                                className="py-12 text-center text-slate-500 dark:text-slate-400"
                              >
                                Không tìm thấy sinh viên nào khớp với điều kiện lọc.
                              </td>
                            </tr>
                          );
                        }

                        return filtered.map((st, idx) => {
                          const initial = (st.full_name || 'S').charAt(0).toUpperCase();

                          let rankBadgeClass = 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400';
                          if (st.academic_rank === 'Xuất sắc') rankBadgeClass = 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300 border border-emerald-300 dark:border-emerald-800';
                          else if (st.academic_rank === 'Giỏi') rankBadgeClass = 'bg-blue-100 text-blue-800 dark:bg-blue-950 dark:text-blue-300 border border-blue-300 dark:border-blue-800';
                          else if (st.academic_rank === 'Khá') rankBadgeClass = 'bg-sky-100 text-sky-800 dark:bg-sky-950 dark:text-sky-300 border border-sky-300 dark:border-sky-800';
                          else if (st.academic_rank === 'Trung bình') rankBadgeClass = 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300 border border-amber-300 dark:border-amber-800';
                          else if (st.academic_rank === 'Yếu / Cảnh báo') rankBadgeClass = 'bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300 border border-rose-300 dark:border-rose-800';

                          return (
                            <tr key={st.student_id} className="hover:bg-slate-50/80 dark:hover:bg-slate-800/40 transition-colors">
                              <td className="px-3 py-3 text-center text-slate-400 font-mono">{idx + 1}</td>
                              <td className="px-4 py-3">
                                <div className="flex items-center gap-2.5">
                                  {st.avatar_url ? (
                                    <img
                                      src={st.avatar_url}
                                      alt={st.full_name}
                                      className="w-7 h-7 rounded-full object-cover border border-slate-200 dark:border-slate-700"
                                    />
                                  ) : (
                                    <div className="w-7 h-7 rounded-full bg-slate-800 dark:bg-slate-700 text-white flex items-center justify-center font-bold text-[10px]">
                                      {initial}
                                    </div>
                                  )}
                                  <div className="min-w-0">
                                    <div className="font-semibold text-slate-900 dark:text-slate-100 truncate">{st.full_name}</div>
                                    <div className="text-[11px] text-slate-400 truncate font-mono">{st.email}</div>
                                  </div>
                                </div>
                              </td>
                              <td className="px-3 py-3 text-center font-mono text-slate-700 dark:text-slate-300 whitespace-nowrap">
                                {st.student_code || '—'}
                              </td>
                              <td className="px-3 py-3 text-slate-600 dark:text-slate-400 truncate max-w-[150px]">
                                {st.department || '—'}
                              </td>

                              {/* Exam Columns */}
                              {(gradebookData.exams || []).map((ex) => {
                                const scoreEntry = st.exam_scores?.[ex.id];
                                if (scoreEntry && scoreEntry.score !== null && scoreEntry.score !== undefined) {
                                  const sNum = Number(scoreEntry.score);
                                  const isFail = sNum < 4.0;
                                  return (
                                    <td key={ex.id} className="px-4 py-3 text-center whitespace-nowrap">
                                      <button
                                        type="button"
                                        onClick={() => navigate(`/teacher/exams/${ex.id}`)}
                                        className={`font-mono font-bold hover:underline ${
                                          isFail ? 'text-rose-600 dark:text-rose-400' : 'text-slate-900 dark:text-slate-100'
                                        }`}
                                        title={`Xem bài làm (Vi phạm: ${scoreEntry.violation_count})`}
                                      >
                                        {sNum.toFixed(2)}
                                      </button>
                                      {scoreEntry.violation_count > 0 && (
                                        <span className="ml-1 text-[10px] text-rose-500 font-bold" title={`${scoreEntry.violation_count} lần vi phạm quy chế`}>
                                          (!{scoreEntry.violation_count})
                                        </span>
                                      )}
                                    </td>
                                  );
                                }
                                if (scoreEntry && scoreEntry.status === 'in_progress') {
                                  return (
                                    <td key={ex.id} className="px-4 py-3 text-center whitespace-nowrap">
                                      <span className="px-1.5 py-0.5 rounded text-[10px] font-medium bg-amber-50 text-amber-700 border border-amber-200 dark:bg-amber-950/40 dark:text-amber-300 dark:border-amber-800">
                                        Đang thi
                                      </span>
                                    </td>
                                  );
                                }
                                return (
                                  <td key={ex.id} className="px-4 py-3 text-center text-slate-400 whitespace-nowrap">
                                    —
                                  </td>
                                );
                              })}

                              {/* Average Column */}
                              <td className="px-3 py-3 text-center font-mono font-extrabold text-blue-600 dark:text-blue-400 bg-blue-50/40 dark:bg-blue-950/20 whitespace-nowrap">
                                {st.completed_count > 0 ? Number(st.average_score).toFixed(2) : '—'}
                              </td>

                              {/* Academic Rank Column */}
                              <td className="px-3 py-3 text-center whitespace-nowrap">
                                <span className={`inline-block px-2.5 py-0.5 rounded-full text-[11px] font-semibold whitespace-nowrap ${rankBadgeClass}`}>
                                  {st.academic_rank}
                                </span>
                              </td>
                            </tr>
                          );
                        });
                      })()}
                    </tbody>
                  </table>
                </div>
              </div>
            </>
          )}
        </div>
      )}

      {/* Teacher Add Member Modal */}
      {showAddModal && (
        <div className="fixed inset-0 bg-slate-900/50 backdrop-blur-sm flex items-center justify-center p-4 z-50">
          <div className="bg-white dark:bg-slate-900 rounded-xl shadow-xl max-w-md w-full p-6 space-y-4 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-slate-100">
            <h2 className="text-base font-bold text-slate-900 dark:text-slate-100">Thêm sinh viên vào lớp</h2>
            <form onSubmit={handleAddMember} className="space-y-4">
              <div className="flex gap-4 border-b border-slate-200 dark:border-slate-800 pb-2">
                <button
                  type="button"
                  onClick={() => setAddMode('email')}
                  className={`text-xs font-semibold pb-1 border-b-2 ${
                    addMode === 'email' ? 'border-indigo-600 text-indigo-600 dark:border-indigo-400 dark:text-indigo-400' : 'border-transparent text-slate-500 dark:text-slate-400'
                  }`}
                >
                  Theo địa chỉ Email
                </button>
                <button
                  type="button"
                  onClick={() => setAddMode('student_code')}
                  className={`text-xs font-semibold pb-1 border-b-2 ${
                    addMode === 'student_code' ? 'border-indigo-600 text-indigo-600 dark:border-indigo-400 dark:text-indigo-400' : 'border-transparent text-slate-500 dark:text-slate-400'
                  }`}
                >
                  Theo Mã sinh viên
                </button>
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 uppercase mb-1">
                  {addMode === 'email' ? 'Email sinh viên *' : 'Mã sinh viên *'}
                </label>
                <input
                  type={addMode === 'email' ? 'email' : 'text'}
                  required
                  placeholder={addMode === 'email' ? 'sinhvien@hust.edu.vn' : 'SV2026...'}
                  value={addInputValue}
                  onChange={(e) => setAddInputValue(e.target.value)}
                  className="w-full text-sm px-3 py-2 border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 rounded focus:ring-1 focus:ring-indigo-500 focus:outline-none"
                />
              </div>

              <div className="flex justify-end gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setShowAddModal(false)}
                  className="btn-secondary text-xs px-3 py-2 border border-slate-300 dark:border-slate-700 rounded text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-800"
                >
                  Hủy
                </button>
                <button
                  type="submit"
                  disabled={adding}
                  className="btn-primary text-xs px-4 py-2 bg-indigo-600 text-white font-semibold rounded hover:bg-indigo-700 disabled:opacity-50"
                >
                  {adding ? 'Đang thêm...' : 'Thêm sinh viên'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
