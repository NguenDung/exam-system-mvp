import { useNavigate, NavLink, Outlet } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

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

  return (
    <div className="min-h-screen bg-slate-50 text-slate-900">
      <header className="bg-slate-900 border-b border-slate-800">
        <div className="mx-auto flex min-h-16 max-w-5xl flex-wrap items-center justify-between gap-3 px-4 py-2 sm:px-6">
          <div className="flex flex-wrap items-center gap-3 sm:gap-6">
            <div className="flex items-center gap-3">
              <span className="font-bold text-base tracking-tight text-white uppercase">Cổng Thi Trực Tuyến</span>
              <span className="bg-slate-700 text-slate-300 text-xs px-2 py-0.5 rounded font-medium">Sinh viên</span>
            </div>
            <nav className="flex flex-wrap gap-2">
              <NavItem to="/student" end>Làm bài thi</NavItem>
              <NavItem to="/student/history">Lịch sử bài làm</NavItem>
            </nav>
          </div>
          <div className="flex items-center gap-4">
            <span className="text-slate-300 text-sm hidden sm:block">{user?.full_name}</span>
            <button onClick={handleLogout} className="btn-secondary btn-sm bg-slate-800 text-slate-200 border-slate-700 hover:bg-slate-700">
              Đăng xuất
            </button>
          </div>
        </div>
      </header>
      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-8">
        <Outlet />
      </main>
    </div>
  );
}
