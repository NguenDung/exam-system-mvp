import { useNavigate, NavLink, Outlet } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

function NavItem({ to, end, children }) {
  return (
    <NavLink
      to={to}
      end={end}
      className={({ isActive }) =>
        `px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
          isActive ? 'bg-emerald-700 text-white' : 'text-emerald-100 hover:bg-emerald-700/60'
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
  const handleLogout = () => { logout(); navigate('/login'); };

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-emerald-800 shadow-lg">
        <div className="max-w-5xl mx-auto px-4 sm:px-6 h-16 flex items-center justify-between">
          <div className="flex items-center gap-6">
            <div className="flex items-center gap-2">
              <span className="text-2xl">📝</span>
              <span className="text-white font-bold text-lg tracking-tight">ExamMVP</span>
              <span className="text-emerald-300 text-xs ml-2 font-medium uppercase tracking-wider">Học sinh</span>
            </div>
            <nav className="flex gap-1">
              <NavItem to="/student" end>Làm bài thi</NavItem>
            </nav>
          </div>
          <div className="flex items-center gap-3">
            <span className="text-emerald-200 text-sm hidden sm:block">👋 {user?.full_name}</span>
            <button onClick={handleLogout} className="btn-secondary btn-sm">Đăng xuất</button>
          </div>
        </div>
      </header>
      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-8">
        <Outlet />
      </main>
    </div>
  );
}
