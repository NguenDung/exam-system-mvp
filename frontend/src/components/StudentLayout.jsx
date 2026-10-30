import { useNavigate, NavLink, Outlet, Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import NotificationDropdown from './NotificationDropdown';

function NavItem({ to, end, children }) {
  return (
    <NavLink
      to={to}
      end={end}
      className={({ isActive }) =>
        `px-3 py-2 rounded-md text-sm font-medium transition-colors ${
          isActive ? 'bg-slate-800 text-white' : 'text-slate-300 hover:bg-slate-700 hover:text-white'
        }`
      }
    >
      {children}
    </NavLink>
  );
}

export default function StudentLayout() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const initial = (user?.full_name || 'S').charAt(0).toUpperCase();

  return (
    <div className="min-h-screen bg-slate-50 text-slate-900">
      <header className="bg-slate-900 border-b border-slate-800">
        <div className="mx-auto flex min-h-16 max-w-6xl flex-wrap items-center justify-between gap-3 px-4 py-2 sm:px-6">
          <div className="flex flex-wrap items-center gap-3 sm:gap-6">
            <div className="flex items-center gap-3">
              <span className="font-bold text-base tracking-tight text-white uppercase">Cổng Thi Trực Tuyến</span>
              <span className="bg-slate-700 text-slate-300 text-xs px-2 py-0.5 rounded font-medium">Sinh viên</span>
            </div>
            <nav className="flex flex-wrap gap-1 sm:gap-2">
              <NavItem to="/student" end>Đề thi</NavItem>
              <NavItem to="/schedule">Lịch thi</NavItem>
              <NavItem to="/classes">Lớp học</NavItem>
              <NavItem to="/student/history">Lịch sử thi</NavItem>
              <NavItem to="/directory">Danh bạ</NavItem>
              <NavItem to="/profile">Hồ sơ</NavItem>
            </nav>
          </div>
          <div className="flex items-center gap-3">
            <NotificationDropdown />
            <Link
              to="/profile"
              className="flex items-center gap-2 px-2 py-1 rounded-md text-slate-300 hover:text-white hover:bg-slate-800 transition-colors"
              title="Xem và chỉnh sửa hồ sơ"
            >
              {user?.avatar_url ? (
                <img
                  src={user.avatar_url}
                  alt={user.full_name}
                  className="w-7 h-7 rounded-full object-cover border border-slate-600 bg-slate-800"
                />
              ) : (
                <div className="w-7 h-7 rounded-full bg-slate-700 text-slate-200 flex items-center justify-center font-bold text-xs border border-slate-600">
                  {initial}
                </div>
              )}
              <span className="text-sm font-medium hidden md:block max-w-[140px] truncate">{user?.full_name}</span>
            </Link>
            <button onClick={handleLogout} className="btn-secondary btn-sm bg-slate-800 text-slate-200 border-slate-700 hover:bg-slate-700">
              Đăng xuất
            </button>
          </div>
        </div>
      </header>
      <main className="max-w-6xl mx-auto px-4 sm:px-6 py-8">
        <Outlet />
      </main>
    </div>
  );
}
