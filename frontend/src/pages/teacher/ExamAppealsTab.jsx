import { useCallback, useEffect, useState } from 'react';
import { teacherAPI } from '../../api/endpoints';
import Spinner from '../../components/Spinner';

const PAGE_SIZE = 10;
const STATUSES = [
  ['PENDING', 'Chờ xử lý'],
  ['REPLIED', 'Đã phản hồi'],
  ['REJECTED', 'Đã từ chối'],
  ['RESOLVED', 'Đã xử lý'],
];
const OPTIONS = ['A', 'B', 'C', 'D'];

function AppealDetailModal({ appealId, onClose, onUpdated }) {
  const [detail, setDetail] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [response, setResponse] = useState('');
  const [working, setWorking] = useState(false);

  const loadAppeal = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const { data } = await teacherAPI.getAppeal(appealId);
      setDetail(data);
      setResponse(data.appeal.teacher_response || '');
    } catch (requestError) {
      setError(requestError.response?.data?.error || 'Không thể tải chi tiết phúc khảo.');
    } finally {
      setLoading(false);
    }
  }, [appealId]);

  useEffect(() => {
    loadAppeal();
  }, [loadAppeal]);

  const updateAppeal = async (action) => {
    setWorking(true);
    setError('');
    try {
      const payload = response.trim() ? { response: response.trim() } : {};
      if (action === 'reply') await teacherAPI.replyAppeal(appealId, { response: response.trim() });
      if (action === 'reject') await teacherAPI.rejectAppeal(appealId, payload);
      if (action === 'resolve') await teacherAPI.resolveAppeal(appealId, payload);
      onUpdated();
      await loadAppeal();
    } catch (requestError) {
      setError(requestError.response?.data?.error || 'Cập nhật phúc khảo thất bại.');
    } finally {
      setWorking(false);
    }
  };

  const canRespond = detail && ['PENDING', 'REPLIED'].includes(detail.appeal.status);

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-3 sm:p-5">
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="teacher-appeal-title"
        className="max-h-[92vh] w-full max-w-4xl overflow-y-auto rounded-lg bg-white p-5 shadow-xl border border-slate-200 sm:p-6"
      >
        <div className="mb-5 flex items-start justify-between gap-4 border-b border-slate-200 pb-4">
          <div>
            <h2 id="teacher-appeal-title" className="text-lg font-bold text-slate-900">Chi tiết đơn phúc khảo</h2>
            {detail && <p className="mt-0.5 text-xs text-slate-500">Môn thi: {detail.exam.title}</p>}
          </div>
          <button onClick={onClose} className="btn-secondary btn-sm">Đóng</button>
        </div>

        {loading ? (
          <div className="flex justify-center py-12"><Spinner size="lg" /></div>
        ) : error && !detail ? (
          <div role="alert" className="rounded bg-red-50 p-4 text-sm text-red-700 border border-red-200">{error}</div>
        ) : detail && (
          <div className="space-y-5">
            <div className="grid gap-4 rounded-md bg-slate-50 p-4 border border-slate-200 sm:grid-cols-2 text-sm">
              <div>
                <p className="text-xs text-slate-500">Thí sinh</p>
                <p className="font-semibold text-slate-900">{detail.student.full_name}</p>
                <p className="text-xs text-slate-500">{detail.student.email}</p>
              </div>
              <div>
                <p className="text-xs text-slate-500">Điểm số bài thi</p>
                <p className="text-base font-bold text-slate-900">
                  {Number(detail.submission.score).toFixed(2)} / {Number(detail.submission.max_score).toFixed(2)}
                </p>
              </div>
              <div>
                <p className="text-xs text-slate-500">Thời gian làm bài</p>
                <p className="text-xs text-slate-700">
                  {new Date(detail.submission.started_at).toLocaleString('vi-VN')} -{' '}
                  {detail.submission.submitted_at ? new Date(detail.submission.submitted_at).toLocaleString('vi-VN') : '-'}
                </p>
              </div>
              <div>
                <p className="text-xs text-slate-500">Trạng thái xử lý</p>
                <p className="font-medium text-slate-900">{STATUSES.find(([v]) => v === detail.appeal.status)?.[1] || detail.appeal.status}</p>
              </div>
            </div>

            <section className="border border-slate-200 rounded-md p-4">
              <h3 className="font-semibold text-slate-900 text-sm">{detail.appeal.title || 'Đơn phúc khảo'}</h3>
              <p className="mt-2 whitespace-pre-wrap text-sm text-slate-700 leading-relaxed bg-slate-50 p-3 rounded border border-slate-100">
                {detail.appeal.content}
              </p>
            </section>

            <section>
              <h3 className="mb-3 font-semibold text-slate-900 text-sm">Bài làm chi tiết của thí sinh</h3>
              <div className="max-h-72 space-y-3 overflow-y-auto pr-1">
                {detail.answers.map((answer, index) => (
                  <div key={answer.question_id} className="rounded border border-slate-200 p-3 text-sm">
                    <p className="font-medium text-slate-900">Câu {index + 1}: {answer.prompt}</p>
                    <div className="mt-2 grid gap-2 sm:grid-cols-2">
                      {OPTIONS.map((option) => (
                        <div
                          key={option}
                          className={`rounded border p-2 text-xs ${
                            answer.correct_option === option
                              ? 'border-emerald-300 bg-emerald-50 text-emerald-900 font-medium'
                              : answer.chosen_option === option
                              ? 'border-red-300 bg-red-50 text-red-900 font-medium'
                              : 'border-slate-200 text-slate-600'
                          }`}
                        >
                          <strong>{option}.</strong> {answer[`option_${option.toLowerCase()}`]}
                          {answer.chosen_option === option && <span className="ml-1 text-slate-700 font-normal">(Thí sinh chọn)</span>}
                          {answer.correct_option === option && <span className="ml-1 text-emerald-700 font-semibold">(Đáp án chuẩn)</span>}
                        </div>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            </section>

            {detail.appeal.teacher_response && (
              <div className="rounded-md bg-slate-100 p-4 text-sm text-slate-800 border border-slate-200">
                <strong className="block text-xs uppercase tracking-wide text-slate-500 mb-1">Ý kiến phản hồi của giảng viên:</strong>
                <p className="whitespace-pre-wrap">{detail.appeal.teacher_response}</p>
              </div>
            )}

            {canRespond && (
              <section className="space-y-3 border-t border-slate-200 pt-4">
                <label htmlFor="appeal-response" className="label">Nội dung phản hồi thí sinh</label>
                <textarea
                  id="appeal-response"
                  value={response}
                  onChange={(event) => setResponse(event.target.value)}
                  rows={3}
                  maxLength={2000}
                  className="input resize-y"
                  placeholder="Nhập nội dung giải trình hoặc kết luận phúc khảo..."
                />
                {error && <p role="alert" className="text-sm text-red-600">{error}</p>}
                <div className="flex flex-wrap justify-end gap-2 pt-1">
                  <button disabled={working} onClick={() => updateAppeal('reject')} className="btn-danger btn-sm">
                    Từ chối phúc khảo
                  </button>
                  <button disabled={working} onClick={() => updateAppeal('resolve')} className="btn-secondary btn-sm">
                    Chấp thuận
                  </button>
                  <button
                    disabled={working || !response.trim()}
                    onClick={() => updateAppeal('reply')}
                    className="btn-primary btn-sm"
                  >
                    {working ? <Spinner size="sm" /> : 'Gửi phản hồi'}
                  </button>
                </div>
              </section>
            )}
            {error && detail && <p role="alert" className="text-sm text-red-600">{error}</p>}
          </div>
        )}
      </div>
    </div>
  );
}

export default function ExamAppealsTab({ examId }) {
  const [items, setItems] = useState([]);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [status, setStatus] = useState('');
  const [searchValue, setSearchValue] = useState('');
  const [search, setSearch] = useState('');
  const [selectedAppeal, setSelectedAppeal] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [refreshKey, setRefreshKey] = useState(0);

  const loadAppeals = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const { data } = await teacherAPI.listAppeals(examId, {
        page,
        page_size: PAGE_SIZE,
        status: status || undefined,
        q: search || undefined,
      });
      setItems(data.items || []);
      setTotal(data.pagination?.total || 0);
    } catch (requestError) {
      setError(requestError.response?.data?.error || 'Không thể tải danh sách phúc khảo.');
    } finally {
      setLoading(false);
    }
  }, [examId, page, status, search]);

  useEffect(() => {
    loadAppeals();
  }, [loadAppeals, refreshKey]);

  const handleSearch = (event) => {
    event.preventDefault();
    setPage(1);
    setSearch(searchValue.trim());
  };
  const pageCount = Math.max(1, Math.ceil(total / PAGE_SIZE));

  return (
    <div className="space-y-4">
      <form onSubmit={handleSearch} className="flex flex-col gap-3 sm:flex-row">
        <input
          value={searchValue}
          onChange={(event) => setSearchValue(event.target.value)}
          className="input flex-1"
          placeholder="Tìm theo tên thí sinh, tiêu đề hoặc nội dung..."
          aria-label="Tìm kiếm phúc khảo"
        />
        <select
          value={status}
          onChange={(event) => {
            setStatus(event.target.value);
            setPage(1);
          }}
          className="input sm:max-w-xs"
          aria-label="Lọc theo trạng thái"
        >
          <option value="">Tất cả trạng thái</option>
          {STATUSES.map(([value, label]) => (
            <option key={value} value={value}>{label}</option>
          ))}
        </select>
        <button type="submit" className="btn-primary">Tìm kiếm</button>
      </form>

      {error && <div role="alert" className="rounded bg-red-50 p-4 text-sm text-red-700 border border-red-200">{error}</div>}

      <div className="bg-white rounded-lg border border-slate-200 overflow-hidden">
        {loading ? (
          <div className="flex justify-center py-16"><Spinner size="lg" /></div>
        ) : items.length === 0 ? (
          <div className="py-16 text-center text-slate-500">
            <p>{search || status ? 'Không tìm thấy đơn phúc khảo phù hợp.' : 'Chưa có đơn phúc khảo nào cho đề thi này.'}</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full min-w-[850px] text-sm">
              <thead>
                <tr className="border-b border-slate-200 bg-slate-50 text-left text-slate-600">
                  <th className="px-4 py-3 font-semibold">STT</th>
                  <th className="px-4 py-3 font-semibold">Thí sinh</th>
                  <th className="px-4 py-3 font-semibold">Tiêu đề</th>
                  <th className="px-4 py-3 font-semibold">Nội dung</th>
                  <th className="px-4 py-3 font-semibold">Ngày gửi</th>
                  <th className="px-4 py-3 font-semibold text-center">Trạng thái</th>
                  <th className="px-4 py-3 font-semibold text-right">Thao tác</th>
                </tr>
              </thead>
              <tbody>
                {items.map((item, index) => (
                  <tr key={item.id} className="border-b border-slate-100 last:border-0 hover:bg-slate-50">
                    <td className="px-4 py-3 text-slate-400 font-mono">{(page - 1) * PAGE_SIZE + index + 1}</td>
                    <td className="px-4 py-3">
                      <p className="font-medium text-slate-900">{item.student_name}</p>
                      <p className="text-xs text-slate-500">{item.student_email}</p>
                    </td>
                    <td className="px-4 py-3 font-medium text-slate-800">{item.title || 'Đơn phúc khảo'}</td>
                    <td className="max-w-xs px-4 py-3">
                      <p className="truncate text-slate-600" title={item.content}>{item.content}</p>
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-slate-500 text-xs">
                      {new Date(item.created_at).toLocaleDateString('vi-VN')}
                    </td>
                    <td className="px-4 py-3 text-center">
                      <span className="rounded px-2.5 py-0.5 text-xs font-semibold bg-amber-100 text-amber-800">
                        {STATUSES.find(([value]) => value === item.status)?.[1] || item.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <button onClick={() => setSelectedAppeal(item.id)} className="btn-secondary btn-sm">
                        Xem chi tiết
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {!loading && total > 0 && (
        <div className="flex flex-wrap items-center justify-between gap-3 text-sm text-slate-600">
          <span>Tổng số: {total} đơn phúc khảo · Trang {page}/{pageCount}</span>
          <div className="flex gap-2">
            <button disabled={page <= 1} onClick={() => setPage((v) => v - 1)} className="btn-secondary btn-sm">Trước</button>
            <button disabled={page >= pageCount} onClick={() => setPage((v) => v + 1)} className="btn-secondary btn-sm">Sau</button>
          </div>
        </div>
      )}

      {selectedAppeal && (
        <AppealDetailModal
          appealId={selectedAppeal}
          onClose={() => setSelectedAppeal(null)}
          onUpdated={() => setRefreshKey((k) => k + 1)}
        />
      )}
    </div>
  );
}
