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
      setError('Nội dung khiếu nại phải từ 10 đến 1000 ký tự.');
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
      setError(requestError.response?.data?.error || 'Gửi đơn khiếu nại không thành công. Vui lòng thử lại.');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <form
        onSubmit={handleSubmit}
        role="dialog"
        aria-modal="true"
        aria-labelledby="appeal-title"
        className="w-full max-w-lg space-y-4 rounded-lg bg-white p-6 shadow-xl border border-slate-200"
      >
        <div>
          <h2 id="appeal-title" className="text-lg font-bold text-slate-900">Gửi đơn khiếu nại điểm thi</h2>
          <p className="mt-0.5 text-xs text-slate-500">Môn thi: {submission.exam_title}</p>
        </div>
        <div>
          <label htmlFor="appeal-subject" className="label">Tiêu đề (tùy chọn)</label>
          <input
            id="appeal-subject"
            value={title}
            onChange={(event) => setTitle(event.target.value)}
            maxLength={255}
            className="input"
            placeholder="Ví dụ: Xin xem xét lại câu hỏi số 3"
          />
        </div>
        <div>
          <label htmlFor="appeal-content" className="label">Nội dung giải trình *</label>
          <textarea
            id="appeal-content"
            value={content}
            onChange={(event) => setContent(event.target.value)}
            required
            minLength={10}
            maxLength={1000}
            rows={5}
            className="input resize-y"
            placeholder="Nêu rõ lý do hoặc nội dung câu hỏi cần phúc khảo..."
          />
          <p className="mt-1 text-right text-xs text-slate-400">{content.trim().length}/1000 ký tự (tối thiểu 10)</p>
        </div>
        {error && <p role="alert" className="rounded bg-red-50 p-3 text-sm text-red-700 border border-red-200">{error}</p>}
        <div className="flex justify-end gap-3 pt-1">
          <button type="button" onClick={onClose} disabled={submitting} className="btn-secondary">
            Hủy bỏ
          </button>
          <button type="submit" disabled={submitting || content.trim().length < 10} className="btn-primary">
            {submitting ? <Spinner size="sm" /> : 'Gửi đơn khiếu nại'}
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
      setError(requestError.response?.data?.error || 'Không thể tải danh sách bài làm.');
    } finally {
      setLoading(false);
    }
  }, [page, search, status]);

  useEffect(() => {
    fetchHistory();
  }, [fetchHistory, refreshKey]);

  const handleSearch = (event) => {
    event.preventDefault();
    setPage(1);
    setSearch(searchInput.trim());
  };

  const pageCount = Math.max(1, Math.ceil(pagination.total / PAGE_SIZE));

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-bold text-slate-900">Lịch sử bài làm</h1>
        <p className="mt-0.5 text-sm text-slate-500">Danh sách các bài thi đã hoàn thành và trạng thái phúc khảo</p>
      </div>

      <form onSubmit={handleSearch} className="bg-white rounded-lg border border-slate-200 p-4 flex flex-col gap-3 sm:flex-row">
        <input
          value={searchInput}
          onChange={(event) => setSearchInput(event.target.value)}
          className="input flex-1"
          placeholder="Tìm theo tên bài thi..."
          aria-label="Tìm theo tên bài thi"
        />
        <select
          value={status}
          onChange={(event) => {
            setStatus(event.target.value);
            setPage(1);
          }}
          className="input sm:max-w-xs"
          aria-label="Lọc theo trạng thái khiếu nại"
        >
          <option value="">Tất cả trạng thái khiếu nại</option>
          {Object.entries(APPEAL_STATUS).map(([value, label]) => (
            <option key={value} value={value}>{label}</option>
          ))}
        </select>
        <button type="submit" className="btn-primary">Tìm kiếm</button>
      </form>

      {notice && (
        <div role="status" className="rounded border border-emerald-300 bg-emerald-50 p-3 text-sm text-emerald-800">
          {notice}
        </div>
      )}
      {error && (
        <div role="alert" className="rounded border border-red-300 bg-red-50 p-4 text-sm text-red-700">
          {error}
        </div>
      )}

      <div className="bg-white rounded-lg border border-slate-200 overflow-hidden">
        {loading ? (
          <div className="flex justify-center py-16"><Spinner size="lg" /></div>
        ) : items.length === 0 ? (
          <div className="px-6 py-16 text-center text-slate-500">
            <p>{search || status ? 'Không tìm thấy bài làm phù hợp với tiêu chí lọc.' : 'Bạn chưa có bài thi nào đã hoàn thành.'}</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full min-w-[950px] text-sm">
              <thead>
                <tr className="border-b border-slate-200 bg-slate-50 text-left text-slate-600">
                  <th className="px-4 py-3 font-semibold">STT</th>
                  <th className="px-4 py-3 font-semibold">Tên bài thi</th>
                  <th className="px-4 py-3 font-semibold">Bắt đầu làm</th>
                  <th className="px-4 py-3 font-semibold">Nộp bài lúc</th>
                  <th className="px-4 py-3 text-center font-semibold">Điểm số</th>
                  <th className="px-4 py-3 font-semibold">Khiếu nại</th>
                  <th className="px-4 py-3 font-semibold text-right">Thao tác</th>
                </tr>
              </thead>
              <tbody>
                {items.map((item, index) => (
                  <tr key={item.submission_id} className="border-b border-slate-100 last:border-0 hover:bg-slate-50 transition-colors">
                    <td className="px-4 py-3.5 text-slate-400 font-mono">{(page - 1) * PAGE_SIZE + index + 1}</td>
                    <td className="px-4 py-3.5 font-medium text-slate-900">{item.exam_title}</td>
                    <td className="px-4 py-3.5 text-slate-500 text-xs">{formatDate(item.started_at)}</td>
                    <td className="px-4 py-3.5 text-slate-500 text-xs">{formatDate(item.submitted_at)}</td>
                    <td className="px-4 py-3.5 text-center font-bold text-slate-900">
                      {Number(item.score).toFixed(2)} / {Number(item.max_score).toFixed(2)}
                    </td>
                    <td className="px-4 py-3.5">
                      {item.appeal_status ? (
                        <span className="inline-block px-2 py-0.5 rounded text-xs font-medium bg-amber-100 text-amber-800">
                          {APPEAL_STATUS[item.appeal_status] || item.appeal_status}
                        </span>
                      ) : (
                        <span className="text-slate-400 text-xs">Chưa gửi</span>
                      )}
                    </td>
                    <td className="px-4 py-3.5 text-right">
                      <div className="flex justify-end gap-2">
                        <button
                          onClick={() => navigate(`/student/history/${item.submission_id}`)}
                          className="btn-secondary btn-sm"
                        >
                          Chi tiết
                        </button>
                        <button
                          onClick={() => {
                            setNotice('');
                            setSelectedSubmission(item);
                          }}
                          className="btn-secondary btn-sm"
                        >
                          Khiếu nại
                        </button>
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
        <div className="flex flex-wrap items-center justify-between gap-3 text-sm text-slate-600">
          <span>Tổng số: {pagination.total} bài làm · Trang {page}/{pageCount}</span>
          <div className="flex gap-2">
            <button
              disabled={page <= 1}
              onClick={() => setPage((current) => current - 1)}
              className="btn-secondary btn-sm"
            >
              Trước
            </button>
            <button
              disabled={page >= pageCount}
              onClick={() => setPage((current) => current + 1)}
              className="btn-secondary btn-sm"
            >
              Sau
            </button>
          </div>
        </div>
      )}

      {selectedSubmission && (
        <AppealModal
          submission={selectedSubmission}
          onClose={() => setSelectedSubmission(null)}
          onSuccess={() => {
            setSelectedSubmission(null);
            setNotice('Gửi đơn khiếu nại thành công. Giảng viên phụ trách sẽ xem xét và phản hồi.');
            setRefreshKey((key) => key + 1);
          }}
        />
      )}
    </div>
  );
}
