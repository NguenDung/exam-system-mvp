import { createContext, useContext, useState, useCallback, useEffect } from 'react';
import { profileAPI } from '../api/endpoints';

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [user, setUser] = useState(() => {
    const token = localStorage.getItem('token');
    if (!token) {
      localStorage.removeItem('user');
      return null;
    }
    try {
      return JSON.parse(localStorage.getItem('user'));
    } catch {
      return null;
    }
  });
  const [hasToken, setHasToken] = useState(() => Boolean(localStorage.getItem('token')));

  const login = useCallback((token, userData) => {
    localStorage.setItem('token', token);
    localStorage.setItem('user', JSON.stringify(userData));
    setHasToken(true);
    setUser(userData);
  }, []);

  const updateUser = useCallback((userData) => {
    localStorage.setItem('user', JSON.stringify(userData));
    setUser(userData);
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    setHasToken(false);
    setUser(null);
  }, []);

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (!token) return;
    profileAPI.getProfile()
      .then(({ data }) => {
        if (data?.user) {
          localStorage.setItem('user', JSON.stringify(data.user));
          setHasToken(true);
          setUser(data.user);
        }
      })
      .catch(() => {});
  }, []);

  return (
    <AuthContext.Provider value={{ user, login, logout, updateUser, isAuthenticated: Boolean(user && hasToken) }}>
      {children}
    </AuthContext.Provider>
  );
}

export const useAuth = () => {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be inside AuthProvider');
  return ctx;
};
