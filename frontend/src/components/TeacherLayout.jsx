import { useNavigate, NavLink, Outlet } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

function NavItem({ to, children }) {
  return (
    <NavLink
      to={to}
      className={({ isActive }) =>
        `px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
          isActive ? 'bg-indigo-700 text-white' : 'text-indigo-100 hover:bg-indigo-700/60'
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

  const handleLogout = () => { logout(); navigate('/login'); };

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-indigo-800 shadow-lg">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-6">
            <div className="flex items-center gap-2">
              <span className="text-2xl">🎓</span>
              <span className="text-white font-bold text-lg tracking-tight">ExamMVP</span>
              <span className="text-indigo-300 text-xs ml-2 font-medium uppercase tracking-wider">Giáo viên</span>
            </div>
            <nav className="flex gap-1">
              <NavItem to="/teacher">Đề thi của tôi</NavItem>
            </nav>
          </div>
          <div className="flex items-center gap-3">
            <span className="text-indigo-200 text-sm hidden sm:block">👋 {user?.full_name}</span>
            <button onClick={handleLogout} className="btn-secondary btn-sm">Đăng xuất</button>
          </div>
        </div>
      </header>
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <Outlet />
      </main>
    </div>
  );
}
