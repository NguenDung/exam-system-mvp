import { useState, useEffect, useRef, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { notificationAPI } from '../api/endpoints';

export default function NotificationDropdown() {
  const [isOpen, setIsOpen] = useState(false);
  const [notifications, setNotifications] = useState([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [loading, setLoading] = useState(false);
  const dropdownRef = useRef(null);
  const navigate = useNavigate();

  const fetchNotifications = useCallback(async () => {
    try {
      const { data } = await notificationAPI.getAll();
      setNotifications(data.notifications || []);
      setUnreadCount(data.unread_count || 0);
    } catch {
      // Polling background error ignored
    }
  }, []);

  useEffect(() => {
    fetchNotifications();
    const interval = setInterval(fetchNotifications, 5000);
    return () => clearInterval(interval);
  }, [fetchNotifications]);

  useEffect(() => {
    const handleClickOutside = (e) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target)) {
        setIsOpen(false);
      }
    };
    const handleEscape = (e) => {
      if (e.key === 'Escape') setIsOpen(false);
    };

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
      document.addEventListener('keydown', handleEscape);
    }
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.removeEventListener('keydown', handleEscape);
    };
  }, [isOpen]);

  const handleToggle = () => {
    if (!isOpen) {
      fetchNotifications();
    }
    setIsOpen((prev) => !prev);
  };

  const handleMarkAsRead = async (id, isRead) => {
    if (isRead) return;
    try {
      await notificationAPI.markAsRead(id);
      setNotifications((prev) =>
        prev.map((n) => (n.id === id ? { ...n, is_read: true } : n))
      );
      setUnreadCount((c) => Math.max(0, c - 1));
    } catch {
      // Ignore
    }
  };

  const handleItemClick = async (item) => {
    if (!item.is_read) {
      await handleMarkAsRead(item.id, false);
    }
    setIsOpen(false);
    if (item.target_url) {
      navigate(item.target_url);
    }
  };

  const handleMarkAllAsRead = async () => {
    setLoading(true);
    try {
      await notificationAPI.markAllAsRead();
      setNotifications((prev) => prev.map((n) => ({ ...n, is_read: true })));
      setUnreadCount(0);
    } catch {
      // Ignore
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteNotification = async (e, id) => {
    e.stopPropagation();
    try {
      await notificationAPI.delete(id);
      const target = notifications.find((n) => n.id === id);
      setNotifications((prev) => prev.filter((n) => n.id !== id));
      if (target && !target.is_read) {
        setUnreadCount((c) => Math.max(0, c - 1));
      }
    } catch {
      // Ignore
    }
  };

  const handleClearRead = async (e) => {
    e.stopPropagation();
    try {
      await notificationAPI.clearRead();
      setNotifications((prev) => prev.filter((n) => !n.is_read));
    } catch {
      // Ignore
    }
  };

  const hasReadNotifications = notifications.some((n) => n.is_read);

  return (
    <div className="relative" ref={dropdownRef}>
      <button
        type="button"
        onClick={handleToggle}
        className="relative flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm font-medium text-slate-300 hover:text-white hover:bg-slate-800 transition-colors"
        aria-label="Xem thông báo"
        aria-expanded={isOpen}
      >
        <span>Thông báo</span>
        {unreadCount > 0 && (
          <span className="bg-red-600 text-white text-[11px] leading-tight font-bold px-1.5 py-0.5 rounded-full min-w-4 text-center">
            {unreadCount > 99 ? '99+' : unreadCount}
          </span>
        )}
      </button>

      {isOpen && (
        <div className="absolute right-0 mt-2 w-80 sm:w-96 rounded-lg bg-white shadow-xl border border-slate-200 py-2 z-50 text-slate-900">
          <div className="flex items-center justify-between px-4 py-2 border-b border-slate-100">
            <div className="flex items-center gap-2">
              <span className="text-sm font-bold text-slate-900">Thông báo hệ thống</span>
              {unreadCount > 0 && (
                <span className="text-xs bg-indigo-50 text-indigo-700 font-semibold px-2 py-0.5 rounded">
                  {unreadCount} mới
                </span>
              )}
            </div>
            <div className="flex items-center gap-2">
              {unreadCount > 0 && (
                <button
                  type="button"
                  onClick={handleMarkAllAsRead}
                  disabled={loading}
                  className="text-xs text-indigo-600 hover:text-indigo-800 font-medium hover:underline disabled:opacity-50"
                >
                  Đọc hết
                </button>
              )}
              {hasReadNotifications && (
                <button
                  type="button"
                  onClick={handleClearRead}
                  className="text-xs text-slate-500 hover:text-red-600 font-medium hover:underline"
                  title="Xóa tất cả thông báo đã đọc"
                >
                  Xóa đã đọc
                </button>
              )}
            </div>
          </div>

          <div className="max-h-80 overflow-y-auto divide-y divide-slate-100">
            {notifications.length === 0 ? (
              <div className="py-8 text-center text-xs text-slate-500">
                Không có thông báo nào.
              </div>
            ) : (
              notifications.map((item) => (
                <div
                  key={item.id}
                  onClick={() => handleItemClick(item)}
                  className={`group relative p-3.5 transition-colors cursor-pointer ${
                    item.is_read
                      ? 'bg-white hover:bg-slate-50'
                      : 'bg-indigo-50/50 hover:bg-indigo-50/80 border-l-2 border-indigo-600'
                  }`}
                >
                  <div className="flex items-start justify-between gap-2">
                    <p className={`text-xs font-semibold leading-snug pr-6 ${item.is_read ? 'text-slate-800' : 'text-slate-900'}`}>
                      {!item.is_read && (
                        <span className="inline-block w-1.5 h-1.5 rounded-full bg-indigo-600 mr-1.5 align-middle" />
                      )}
                      {item.title}
                    </p>
                    <button
                      type="button"
                      onClick={(e) => handleDeleteNotification(e, item.id)}
                      className="absolute top-3 right-3 text-slate-400 hover:text-red-600 p-1 rounded hover:bg-slate-100 transition-colors"
                      title="Xóa thông báo này"
                    >
                      <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    </button>
                  </div>
                  <p className="text-xs text-slate-600 mt-1 leading-relaxed">
                    {item.message}
                  </p>
                  <div className="flex items-center justify-between text-[10px] text-slate-400 mt-1.5">
                    <span>
                      {new Date(item.created_at).toLocaleTimeString('vi-VN', { hour: '2-digit', minute: '2-digit' })}{' '}
                      - {new Date(item.created_at).toLocaleDateString('vi-VN')}
                    </span>
                    {item.target_url && (
                      <span className="text-indigo-600 group-hover:underline">Xem chi tiết</span>
                    )}
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  );
}
