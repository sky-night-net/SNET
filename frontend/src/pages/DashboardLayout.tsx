import { Outlet, NavLink, useNavigate } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import { useAuth } from '../lib/api';
import { useTranslation } from 'react-i18next';
import {
  LayoutDashboard, Server, Users, Settings,
  LogOut, ShieldCheck, Menu, Shield
} from 'lucide-react';
import { useState, useEffect } from 'react';

const NAV_ITEMS = [
  { label: 'nav.dashboard', to: '/',         icon: LayoutDashboard },
  { label: 'nav.nodes',     to: '/inbounds', icon: Server },
  { label: 'nav.clients',  to: '/clients',  icon: Users },
  { label: 'nav.firewall', to: '/firewall', icon: Shield },
  { label: 'nav.settings',to: '/settings', icon: Settings },
];

export default function DashboardLayout() {
  const { t, i18n } = useTranslation();
  const { logout } = useAuth();
  const navigate = useNavigate();
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [isMobile, setIsMobile] = useState(window.innerWidth < 768);

  useEffect(() => {
    const handleResize = () => setIsMobile(window.innerWidth < 768);
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  const handleLogout = () => { logout(); navigate('/login'); };

  const Sidebar = ({ mobile = false }: { mobile?: boolean }) => (
    <div style={{
      width: mobile ? '100%' : 240, height: '100%',
      display: 'flex', flexDirection: 'column',
      background: mobile ? 'var(--bg-card)' : 'rgba(13,15,23,0.6)',
      backdropFilter: 'blur(20px)',
      borderRight: mobile ? 'none' : '1px solid var(--border)',
      padding: mobile ? '20px 16px' : '28px 16px',
    }}>
      {/* Logo */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, paddingLeft: 8, marginBottom: 36 }}>
        <div style={{
          width: 36, height: 36, borderRadius: 10,
          background: 'linear-gradient(135deg, var(--accent), #4f46e5)',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          boxShadow: '0 0 20px var(--accent-glow)'
        }}>
          <ShieldCheck size={20} color="white" />
        </div>
        <div>
          <div style={{ fontWeight: 800, fontSize: 16, letterSpacing: '-0.3px' }}>SNET</div>
          <div style={{ fontSize: 10, color: 'var(--text-muted)', letterSpacing: '0.5px' }}>v3.0 NATIVE</div>
        </div>
      </div>

      {/* Nav */}
      <nav style={{ flex: 1 }}>
        <div style={{ fontSize: 10, fontWeight: 600, color: 'var(--text-muted)', letterSpacing: '1.5px', textTransform: 'uppercase', paddingLeft: 10, marginBottom: 10 }}>
          {t('common.actions')}
        </div>
        {NAV_ITEMS.map(({ label, to, icon: Icon }) => (
          <NavLink
            key={to}
            to={to}
            end={to === '/'}
            onClick={() => setSidebarOpen(false)}
            style={({ isActive }) => ({
              display: 'flex', alignItems: 'center', gap: 12,
              padding: '10px 12px', borderRadius: 10,
              fontSize: 13.5, fontWeight: 500,
              color: isActive ? 'white' : 'var(--text-secondary)',
              background: isActive ? 'linear-gradient(135deg, rgba(99,102,241,0.2), rgba(99,102,241,0.08))' : 'transparent',
              border: isActive ? '1px solid rgba(99,102,241,0.25)' : '1px solid transparent',
              textDecoration: 'none', marginBottom: 4,
              transition: 'all 0.15s ease',
            })}
          >
            {({ isActive }) => (
              <>
                <Icon size={17} color={isActive ? 'var(--accent-light)' : 'currentColor'} />
                {t(label)}
                {isActive && (
                  <motion.div
                    layoutId="nav-pill"
                    style={{
                      marginLeft: 'auto', width: 6, height: 6, borderRadius: '50%',
                      background: 'var(--accent)'
                    }}
                  />
                )}
              </>
            )}
          </NavLink>
        ))}
      </nav>

      {/* Logout */}
      <button
        onClick={handleLogout}
        style={{
          display: 'flex', alignItems: 'center', gap: 12,
          padding: '10px 12px', borderRadius: 10,
          fontSize: 13.5, fontWeight: 500, color: 'var(--text-muted)',
          background: 'none', border: '1px solid transparent',
          cursor: 'pointer', width: '100%',
          transition: 'all 0.15s ease',
          fontFamily: 'inherit'
        }}
        onMouseEnter={e => { e.currentTarget.style.color = '#fca5a5'; e.currentTarget.style.background = 'rgba(239,68,68,0.08)'; }}
        onMouseLeave={e => { e.currentTarget.style.color = 'var(--text-muted)'; e.currentTarget.style.background = 'none'; }}
      >
        <LogOut size={17} />
        {t('nav.logout')}
      </button>
    </div>
  );

  return (
    <div style={{ display: 'flex', height: '100vh', overflow: 'hidden', background: 'var(--bg-deep)' }}>
      {/* Desktop sidebar */}
      {!isMobile && (
        <div style={{ width: 240, flexShrink: 0 }}>
          <Sidebar />
        </div>
      )}

      {/* Mobile overlay */}
      <AnimatePresence>
        {isMobile && sidebarOpen && (
          <>
            <motion.div
              initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}
              onClick={() => setSidebarOpen(false)}
              style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.6)', zIndex: 40, backdropFilter: 'blur(4px)' }}
            />
            <motion.div
              initial={{ x: -260 }} animate={{ x: 0 }} exit={{ x: -260 }}
              transition={{ type: 'spring', stiffness: 300, damping: 30 }}
              style={{ position: 'fixed', left: 0, top: 0, bottom: 0, zIndex: 50, width: 260 }}
            >
              <Sidebar mobile />
            </motion.div>
          </>
        )}
      </AnimatePresence>

      {/* Main area */}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
        {/* Top bar */}
        <div style={{
          height: 60, display: 'flex', alignItems: 'center', padding: '0 24px',
          borderBottom: '1px solid var(--border)',
          background: 'rgba(13,15,23,0.5)', backdropFilter: 'blur(12px)',
          flexShrink: 0, gap: 12
        }}>
          {isMobile && (
            <button
              onClick={() => setSidebarOpen(!sidebarOpen)}
              style={{
                background: 'none', border: 'none', color: 'var(--text-secondary)',
                cursor: 'pointer', display: 'flex', padding: 6, borderRadius: 8
              }}
            >
              <Menu size={20} />
            </button>
          )}

          <div style={{ flex: 1 }} />

          {/* Language Switcher */}
          <div style={{ display: 'flex', gap: 4, background: 'rgba(255,255,255,0.05)', padding: 4, borderRadius: 10 }}>
            {['ru', 'en'].map(lang => (
              <button
                key={lang}
                onClick={() => i18n.changeLanguage(lang)}
                style={{
                  padding: '4px 8px', borderRadius: 6, fontSize: 11, fontWeight: 700,
                  textTransform: 'uppercase', border: 'none', cursor: 'pointer',
                  background: i18n.language === lang ? 'var(--accent)' : 'transparent',
                  color: i18n.language === lang ? 'white' : 'var(--text-muted)',
                  transition: 'all 0.2s'
                }}
              >
                {lang}
              </button>
            ))}
          </div>

          {/* Status indicator */}
          <div className="glassmorphism" style={{
            display: 'flex', alignItems: 'center', gap: 8,
            padding: '6px 14px', borderRadius: 99, fontSize: 12, fontWeight: 500
          }}>
            <motion.div
              animate={{ scale: [1, 1.3, 1], opacity: [1, 0.7, 1] }}
              transition={{ duration: 2, repeat: Infinity }}
              style={{ width: 7, height: 7, borderRadius: '50%', background: 'var(--success)' }}
            />
            <span style={{ color: 'var(--text-secondary)' }}>{t('dashboard.system_status')}</span>
          </div>
        </div>

        <main style={{
          flex: 1, overflow: 'auto',
          background: 'radial-gradient(ellipse 70% 50% at 50% 0%, rgba(99,102,241,0.06), transparent 60%), var(--bg-deep)',
          padding: isMobile ? '20px' : '32px 28px'
        }}>
          <Outlet />
        </main>
      </div>
    </div>
  );
}
