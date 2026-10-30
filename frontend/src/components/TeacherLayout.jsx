import { useNavigate, NavLink, Outlet } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import NotificationDropdown from './NotificationDropdown';

function NavItem({ to, children }) {
  return (
    <NavLink
      to={to}
      end={to === '/teacher'}
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

export default function TeacherLayout() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="min-h-screen bg-slate-50 text-slate-900">
      <header className="bg-slate-900 border-b border-slate-800">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-6">
            <div className="flex items-center gap-3">
              <span className="font-bold text-base tracking-tight text-white uppercase">Hệ Thống Khảo Thí</span>
              <span className="bg-slate-700 text-slate-300 text-xs px-2 py-0.5 rounded font-medium">Giảng viên</span>
            </div>
            <nav className="flex gap-2">
              <NavItem to="/teacher">Quản lý đề thi</NavItem>
              <NavItem to="/teacher/students">Quản lý sinh viên</NavItem>
            </nav>
          </div>
          <div className="flex items-center gap-4">
            <NotificationDropdown />
            <span className="text-slate-300 text-sm hidden sm:block">{user?.full_name}</span>
            <button onClick={handleLogout} className="btn-secondary btn-sm bg-slate-800 text-slate-200 border-slate-700 hover:bg-slate-700">
              Đăng xuất
            </button>
          </div>
        </div>
      </header>
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <Outlet />
      </main>
    </div>
  );
}
