import { useState, type ReactNode } from 'react';
import axios from 'axios';

const BASE = (import.meta.env.VITE_API_URL as string) || '/api';

export const http = axios.create({ baseURL: BASE });

http.interceptors.request.use((cfg) => {
  const t = localStorage.getItem('snet_token');
  if (t) cfg.headers.Authorization = `Bearer ${t}`;
  return cfg;
});

/* ---------- Auth Context ---------- */
import { createContext, useContext } from 'react';

interface AuthCtx {
  token: string | null;
  login: (u: string, p: string) => Promise<void>;
  logout: () => void;
  isAuthenticated: boolean;
}

const AuthContext = createContext<AuthCtx | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(
    () => localStorage.getItem('snet_token')
  );

  const login = async (username: string, password: string) => {
    const { data } = await http.post('/login', { username, password });
    if (!data.success) throw new Error(data.msg || 'Ошибка авторизации');
    localStorage.setItem('snet_token', data.token);
    setToken(data.token);
  };

  const logout = () => {
    localStorage.removeItem('snet_token');
    setToken(null);
  };

  return (
    <AuthContext.Provider value={{ token, login, logout, isAuthenticated: !!token }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be inside AuthProvider');
  return ctx;
}
