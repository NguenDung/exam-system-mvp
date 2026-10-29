import { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { studentAPI } from '../../api/endpoints';
import Spinner from '../../components/Spinner';

const PAGE_SIZE = 10;

const APPEAL_STATUS = {
  PENDING: 'Chờ xử lý',
  REPLIED: 'Đã phản hồi',
  REJECTED: 'Đã từ chối',
  RESOLVED: 'Đã xử lý',
};

function formatDate(value) {
  return value ? new Date(value).toLocaleString('vi-VN') : '—';
}

function AppealModal({ submission, onClose, onSuccess }) {
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (event) => {
    event.preventDefault();
    const trimmed = content.trim();
    if (trimmed.length < 10 || trimmed.length > 1000) {
      setError('Nội dung phải có từ 10 đến 1000 ký tự.');
      return;
    }

    setSubmitting(true);
    setError('');
    try {
      await studentAPI.createAppeal({
        submissionId: submission.submission_id,
        title: title.trim(),
        content: trimmed,
      });
      onSuccess();
    } catch (requestError) {
      setError(requestError.response?.data?.error || 'Không thể gửi khiếu nại. Vui lòng thử lại.');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <form onSubmit={handleSubmit} role="dialog" aria-modal="true" aria-labelledby="appeal-title"
        className="w-full max-w-lg space-y-4 rounded-2xl bg-white p-6 shadow-2xl">
        <div>
          <h2 id="appeal-title" className="text-xl font-bold text-gray-900">Khiếu nại điểm thi</h2>
          <p className="mt-1 text-sm text-gray-500">{submission.exam_title}</p>
        </div>
        <div>
          <label htmlFor="appeal-subject" className="label">Tiêu đề khiếu nại (không bắt buộc)</label>
          <input id="appeal-subject" value={title} onChange={(event) => setTitle(event.target.value)}
            maxLength={255} className="input" placeholder="Ví dụ: Phúc khảo câu 4" />
        </div>
        <div>
          <label htmlFor="appeal-content" className="label">Nội dung khiếu nại *</label>
          <textarea id="appeal-content" value={content} onChange={(event) => setContent(event.target.value)}
            required minLength={10} maxLength={1000} rows={5} className="input resize-y"
            placeholder="Mô tả câu hỏi hoặc điểm số bạn muốn được xem xét..." />
          <p className="mt-1 text-right text-xs text-gray-400">{content.trim().length}/1000 ký tự (tối thiểu 10)</p>
        </div>
        {error && <p role="alert" className="rounded-lg bg-red-50 p-3 text-sm text-red-700">{error}</p>}
        <div className="flex justify-end gap-3">
          <button type="button" onClick={onClose} disabled={submitting} className="btn-secondary">Hủy</button>
          <button type="submit" disabled={submitting || content.trim().length < 10} className="btn-primary">
            {submitting ? <Spinner size="sm" /> : 'Gửi khiếu nại'}
          </button>
        </div>
      </form>
    </div>
  );
}

export default function StudentHistoryPage() {
  const navigate = useNavigate();
  const [items, setItems] = useState([]);
  const [pagination, setPagination] = useState({ page: 1, page_size: PAGE_SIZE, total: 0 });
  const [page, setPage] = useState(1);
  const [searchInput, setSearchInput] = useState('');
  const [search, setSearch] = useState('');
  const [status, setStatus] = useState('');
  const [selectedSubmission, setSelectedSubmission] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');
  const [refreshKey, setRefreshKey] = useState(0);

  const fetchHistory = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const { data } = await studentAPI.getHistory({
        page,
        page_size: PAGE_SIZE,
        q: search || undefined,
        appeal_status: status || undefined,
      });
      setItems(data.items || []);
      setPagination(data.pagination || { page, page_size: PAGE_SIZE, total: 0 });
    } catch (requestError) {
      setError(requestError.response?.data?.error || 'Không thể tải lịch sử bài làm.');
    } finally {
      setLoading(false);
    }
  }, [page, search, status]);

  useEffect(() => { fetchHistory(); }, [fetchHistory, refreshKey]);

  const handleSearch = (event) => {
    event.preventDefault();
    setPage(1);
    setSearch(searchInput.trim());
  };

  const pageCount = Math.max(1, Math.ceil(pagination.total / PAGE_SIZE));

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Lịch sử bài làm</h1>
        <p className="mt-1 text-sm text-gray-500">Các bài thi bạn đã hoàn thành và điểm số tương ứng.</p>
      </div>

      <form onSubmit={handleSearch} className="card flex flex-col gap-3 p-4 sm:flex-row">
        <input value={searchInput} onChange={(event) => setSearchInput(event.target.value)}
          className="input flex-1" placeholder="Tìm theo tên bài thi..." aria-label="Tìm theo tên bài thi" />
        <select value={status} onChange={(event) => { setStatus(event.target.value); setPage(1); }}
          className="input sm:max-w-xs" aria-label="Lọc theo trạng thái khiếu nại">
          <option value="">Tất cả trạng thái khiếu nại</option>
          {Object.entries(APPEAL_STATUS).map(([value, label]) => <option key={value} value={value}>{label}</option>)}
        </select>
        <button type="submit" className="btn-primary">Tìm kiếm</button>
      </form>

      {notice && <div role="status" className="rounded-lg border border-emerald-200 bg-emerald-50 p-3 text-sm text-emerald-800">{notice}</div>}
      {error && <div role="alert" className="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700">{error}</div>}

      <div className="card overflow-hidden p-0">
        {loading ? (
          <div className="flex justify-center py-16"><Spinner size="lg" /></div>
        ) : items.length === 0 ? (
          <div className="px-6 py-16 text-center">
            <div className="mb-3 text-4xl">📭</div>
            <p className="text-gray-500">{search || status ? 'Không tìm thấy bài làm phù hợp.' : 'Bạn chưa thực hiện bài thi nào.'}</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full min-w-[1050px] text-sm">
              <thead>
                <tr className="border-b bg-gray-50 text-left">
                  <th className="px-4 py-3 font-semibold text-gray-600">STT</th>
                  <th className="px-4 py-3 font-semibold text-gray-600">Tên bài thi</th>
                  <th className="px-4 py-3 font-semibold text-gray-600">Xuất bản</th>
                  <th className="px-4 py-3 font-semibold text-gray-600">Bắt đầu làm</th>
                  <th className="px-4 py-3 font-semibold text-gray-600">Kết thúc làm</th>
                  <th className="px-4 py-3 text-center font-semibold text-gray-600">Điểm</th>
                  <th className="px-4 py-3 font-semibold text-gray-600">Khiếu nại</th>
                  <th className="px-4 py-3 font-semibold text-gray-600">Thao tác</th>
                </tr>
              </thead>
              <tbody>
                {items.map((item, index) => (
                  <tr key={item.submission_id} className="border-b last:border-0 hover:bg-gray-50">
                    <td className="px-4 py-4 text-gray-500">{(page - 1) * PAGE_SIZE + index + 1}</td>
                    <td className="px-4 py-4 font-medium text-gray-900">{item.exam_title}</td>
                    <td className="px-4 py-4 text-gray-600">{formatDate(item.published_at)}</td>
                    <td className="px-4 py-4 text-gray-600">{formatDate(item.started_at)}</td>
                    <td className="px-4 py-4 text-gray-600">{formatDate(item.submitted_at)}</td>
                    <td className="px-4 py-4 text-center font-bold text-indigo-700">
                      {Number(item.score).toFixed(2)}/{Number(item.max_score).toFixed(2)}
                    </td>
                    <td className="px-4 py-4">
                      {item.appeal_status ? (
                        <span className="text-xs font-medium text-amber-700">{APPEAL_STATUS[item.appeal_status] || item.appeal_status}</span>
                      ) : <span className="text-gray-400">Chưa gửi</span>}
                    </td>
                    <td className="px-4 py-4">
                      <div className="flex gap-2">
                        <button onClick={() => navigate(`/student/history/${item.submission_id}`)}
                          className="btn-primary btn-sm">Chi tiết</button>
                        <button onClick={() => { setNotice(''); setSelectedSubmission(item); }}
                          className="btn-secondary btn-sm">Khiếu nại</button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {!loading && pagination.total > 0 && (
        <div className="flex flex-wrap items-center justify-between gap-3 text-sm text-gray-600">
          <span>{pagination.total} bài làm · Trang {page}/{pageCount}</span>
          <div className="flex gap-2">
            <button disabled={page <= 1} onClick={() => setPage((current) => current - 1)} className="btn-secondary btn-sm">Trước</button>
            <button disabled={page >= pageCount} onClick={() => setPage((current) => current + 1)} className="btn-secondary btn-sm">Sau</button>
          </div>
        </div>
      )}

      {selectedSubmission && (
        <AppealModal
          submission={selectedSubmission}
          onClose={() => setSelectedSubmission(null)}
          onSuccess={() => {
            setSelectedSubmission(null);
            setNotice('Gửi khiếu nại thành công.');
            setRefreshKey((key) => key + 1);
          }}
        />
      )}
    </div>
  );
}
