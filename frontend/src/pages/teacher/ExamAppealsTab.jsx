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
      setError(requestError.response?.data?.error || 'Không thể tải chi tiết khiếu nại.');
    } finally {
      setLoading(false);
    }
  }, [appealId]);

  useEffect(() => { loadAppeal(); }, [loadAppeal]);

  const updateAppeal = async (action) => {
    setWorking(true);
    setError('');
    try {
      if (action === 'reply') await teacherAPI.replyAppeal(appealId, { response });
      if (action === 'reject') await teacherAPI.rejectAppeal(appealId);
      if (action === 'resolve') await teacherAPI.resolveAppeal(appealId);
      onUpdated();
      await loadAppeal();
    } catch (requestError) {
      setError(requestError.response?.data?.error || 'Không thể cập nhật khiếu nại.');
    } finally {
      setWorking(false);
    }
  };

  const canRespond = detail && ['PENDING', 'REPLIED'].includes(detail.appeal.status);

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-3 sm:p-5">
      <div role="dialog" aria-modal="true" aria-labelledby="teacher-appeal-title"
        className="max-h-[92vh] w-full max-w-4xl overflow-y-auto rounded-2xl bg-white p-5 shadow-2xl sm:p-7">
        <div className="mb-5 flex items-start justify-between gap-4">
          <div>
            <h2 id="teacher-appeal-title" className="text-xl font-bold text-gray-900">Chi tiết khiếu nại</h2>
            {detail && <p className="mt-1 text-sm text-gray-500">{detail.exam.title}</p>}
          </div>
          <button onClick={onClose} className="btn-secondary btn-sm">Đóng</button>
        </div>

        {loading ? <div className="flex justify-center py-12"><Spinner size="lg" /></div> : error && !detail ? (
          <div role="alert" className="rounded-lg bg-red-50 p-4 text-sm text-red-700">{error}</div>
        ) : detail && (
          <div className="space-y-5">
            <div className="grid gap-4 rounded-xl bg-gray-50 p-4 sm:grid-cols-2">
              <div><p className="text-xs text-gray-500">Student</p><p className="font-semibold">{detail.student.full_name}</p><p className="text-sm text-gray-500">{detail.student.email}</p></div>
              <div><p className="text-xs text-gray-500">Điểm hiện tại</p><p className="text-xl font-bold text-indigo-700">{Number(detail.submission.score).toFixed(2)}/{Number(detail.submission.max_score).toFixed(2)}</p></div>
              <div><p className="text-xs text-gray-500">Thời gian làm bài</p><p className="text-sm">{new Date(detail.submission.started_at).toLocaleString('vi-VN')} – {detail.submission.submitted_at ? new Date(detail.submission.submitted_at).toLocaleString('vi-VN') : '—'}</p></div>
              <div><p className="text-xs text-gray-500">Trạng thái</p><p className="font-semibold">{detail.appeal.status}</p></div>
            </div>

            <section>
              <h3 className="font-semibold text-gray-900">{detail.appeal.title || 'Khiếu nại điểm thi'}</h3>
              <p className="mt-2 whitespace-pre-wrap rounded-xl border p-4 text-sm text-gray-700">{detail.appeal.content}</p>
            </section>

            <section>
              <h3 className="mb-3 font-semibold text-gray-900">Bài làm của Student</h3>
              <div className="max-h-72 space-y-3 overflow-y-auto">
                {detail.answers.map((answer, index) => (
                  <div key={answer.question_id} className="rounded-xl border p-4">
                    <p className="font-medium">Câu {index + 1}: {answer.prompt}</p>
                    <div className="mt-3 grid gap-2 sm:grid-cols-2">
                      {OPTIONS.map((option) => (
                        <div key={option} className={`rounded-lg border p-2 text-sm ${
                          answer.correct_option === option ? 'border-emerald-300 bg-emerald-50' :
                            answer.chosen_option === option ? 'border-red-300 bg-red-50' : 'border-gray-200'
                        }`}>
                          <strong>{option}.</strong> {answer[`option_${option.toLowerCase()}`]}
                          {answer.chosen_option === option && <span className="ml-1 text-xs">(Student chọn)</span>}
                          {answer.correct_option === option && <span className="ml-1 text-xs">(Đáp án đúng)</span>}
                        </div>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            </section>

            {detail.appeal.teacher_response && (
              <div className="rounded-xl bg-indigo-50 p-4 text-sm text-indigo-900">
                <strong>Phản hồi:</strong>
                <p className="mt-1 whitespace-pre-wrap">{detail.appeal.teacher_response}</p>
              </div>
            )}

            {canRespond && (
              <section className="space-y-3 border-t pt-4">
                <label htmlFor="appeal-response" className="label">Phản hồi cho Student</label>
                <textarea id="appeal-response" value={response} onChange={(event) => setResponse(event.target.value)}
                  rows={3} maxLength={2000} className="input resize-y" placeholder="Nhập phản hồi..." />
                {error && <p role="alert" className="text-sm text-red-600">{error}</p>}
                <div className="flex flex-wrap justify-end gap-2">
                  <button disabled={working} onClick={() => updateAppeal('reject')} className="btn-danger">Từ chối</button>
                  <button disabled={working} onClick={() => updateAppeal('resolve')} className="btn-secondary">Đã xử lý</button>
                  <button disabled={working || !response.trim()} onClick={() => updateAppeal('reply')} className="btn-primary">
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
        page, page_size: PAGE_SIZE, status: status || undefined, q: search || undefined,
      });
      setItems(data.items || []);
      setTotal(data.pagination?.total || 0);
    } catch (requestError) {
      setError(requestError.response?.data?.error || 'Không thể tải danh sách khiếu nại.');
    } finally {
      setLoading(false);
    }
  }, [examId, page, status, search]);

  useEffect(() => { loadAppeals(); }, [loadAppeals, refreshKey]);

  const handleSearch = (event) => {
    event.preventDefault();
    setPage(1);
    setSearch(searchValue.trim());
  };
  const pageCount = Math.max(1, Math.ceil(total / PAGE_SIZE));

  return (
    <div className="space-y-4">
      <form onSubmit={handleSearch} className="flex flex-col gap-3 sm:flex-row">
        <input value={searchValue} onChange={(event) => setSearchValue(event.target.value)}
          className="input flex-1" placeholder="Tìm Student, tiêu đề hoặc nội dung..." aria-label="Tìm khiếu nại" />
        <select value={status} onChange={(event) => { setStatus(event.target.value); setPage(1); }}
          className="input sm:max-w-xs" aria-label="Lọc trạng thái khiếu nại">
          <option value="">Tất cả trạng thái</option>
          {STATUSES.map(([value, label]) => <option key={value} value={value}>{label}</option>)}
        </select>
        <button type="submit" className="btn-primary">Tìm kiếm</button>
      </form>

      {error && <div role="alert" className="rounded-lg bg-red-50 p-4 text-sm text-red-700">{error}</div>}
      <div className="card overflow-hidden p-0">
        {loading ? <div className="flex justify-center py-16"><Spinner size="lg" /></div> : items.length === 0 ? (
          <div className="py-16 text-center">
            <div className="mb-3 text-4xl">📭</div>
            <p className="text-gray-500">{search || status ? 'Không tìm thấy khiếu nại phù hợp.' : 'Chưa có khiếu nại nào cho đề thi này.'}</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full min-w-[850px] text-sm">
              <thead><tr className="border-b bg-gray-50 text-left">
                <th className="px-4 py-3">STT</th><th className="px-4 py-3">Student</th>
                <th className="px-4 py-3">Tiêu đề</th><th className="px-4 py-3">Nội dung</th>
                <th className="px-4 py-3">Ngày gửi</th><th className="px-4 py-3">Trạng thái</th>
                <th className="px-4 py-3">Thao tác</th>
              </tr></thead>
              <tbody>
                {items.map((item, index) => (
                  <tr key={item.id} className="border-b last:border-0">
                    <td className="px-4 py-4">{(page - 1) * PAGE_SIZE + index + 1}</td>
                    <td className="px-4 py-4"><p className="font-medium">{item.student_name}</p><p className="text-xs text-gray-500">{item.student_email}</p></td>
                    <td className="px-4 py-4">{item.title || 'Khiếu nại điểm thi'}</td>
                    <td className="max-w-xs px-4 py-4"><p className="truncate" title={item.content}>{item.content}</p></td>
                    <td className="whitespace-nowrap px-4 py-4">{new Date(item.created_at).toLocaleDateString('vi-VN')}</td>
                    <td className="px-4 py-4"><span className="rounded-full bg-amber-50 px-2.5 py-1 text-xs font-semibold text-amber-800">{STATUSES.find(([value]) => value === item.status)?.[1] || item.status}</span></td>
                    <td className="px-4 py-4"><button onClick={() => setSelectedAppeal(item.id)} className="btn-primary btn-sm">Xem</button></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {!loading && total > 0 && (
        <div className="flex flex-wrap items-center justify-between gap-3 text-sm text-gray-600">
          <span>{total} khiếu nại · Trang {page}/{pageCount}</span>
          <div className="flex gap-2">
            <button disabled={page <= 1} onClick={() => setPage((value) => value - 1)} className="btn-secondary btn-sm">Trước</button>
            <button disabled={page >= pageCount} onClick={() => setPage((value) => value + 1)} className="btn-secondary btn-sm">Sau</button>
          </div>
        </div>
      )}

      {selectedAppeal && <AppealDetailModal appealId={selectedAppeal} onClose={() => setSelectedAppeal(null)}
        onUpdated={() => setRefreshKey((key) => key + 1)} />}
    </div>
  );
}
