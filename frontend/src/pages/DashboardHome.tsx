import { useState, useEffect } from 'react';
import { motion, type Variants } from 'framer-motion';
import { Server, Activity, Cpu, Globe } from 'lucide-react';
import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts';
import { http } from '../lib/api';
import { useTranslation } from 'react-i18next';

const cardVariants: Variants = {
  hidden: { opacity: 0, y: 20 },
  visible: (i: number) => ({ 
    opacity: 1, 
    y: 0, 
    transition: { delay: i * 0.08, duration: 0.4, ease: 'easeOut' } 
  }),
};

function formatBytes(bytes: number) {
  if (!bytes) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

export default function DashboardHome() {
  const { t } = useTranslation();
  const [stats, setStats] = useState<any>(null);
  const [history, setHistory] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const { data } = await http.get('/system/status');
        if (data.success) {
          setStats(data.obj);
          if (data.obj.history) {
            setHistory(data.obj.history.map((p: any) => ({
              time: new Date(p.timestamp * 1000).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' }),
              cpu: p.cpu,
              upload: p.netUp / 1024, // KB
              download: p.netDown / 1024, // KB
              mem: Math.round(p.mem / (1024 * 1024)) // MB
            })));
          }
        }
      } catch (err) {
        console.error('Failed to fetch system stats');
      } finally {
        setLoading(false);
      }
    };

    fetchStats();
    const timer = setInterval(fetchStats, 2000);
    return () => clearInterval(timer);
  }, []);

  const displayStats = [
    { label: t('dashboard.cpu_load'),     value: stats ? `${stats.cpu.toFixed(1)}%` : '0%', icon: Cpu,      color: '#a78bfa' },
    { label: t('dashboard.ram_usage'),    value: stats ? `${(stats.mem.current / stats.mem.total * 100).toFixed(0)}%` : '0%', icon: Activity, color: 'var(--success)' },
    { label: 'Uptime',                   value: stats ? `${(stats.uptime / 3600).toFixed(1)}h` : '0h', icon: Server,   color: 'var(--accent)' },
    { label: t('dashboard.title') + ' Sent', value: stats ? formatBytes(stats.net.bytesSent) : '0 B', icon: Globe,    color: '#f59e0b' },
  ];

  return (
    <div>
      <motion.div
        initial={{ opacity: 0, y: -10 }}
        animate={{ opacity: 1, y: 0 }}
        style={{ marginBottom: 32 }}
      >
        <h1 style={{ fontSize: 26, fontWeight: 800, letterSpacing: '-0.5px' }}>
          {t('dashboard.title')}
        </h1>
        <p style={{ color: 'var(--text-secondary)', fontSize: 14, marginTop: 4 }}>
          {loading ? 'Загрузка системных данных...' : t('dashboard.system_status')}
        </p>
      </motion.div>

      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
        gap: 16, marginBottom: 28
      }}>
        {displayStats.map(({ label, value, icon: Icon, color }, i) => (
          <motion.div
            key={label}
            custom={i}
            initial="hidden"
            animate="visible"
            variants={cardVariants}
            style={{
              background: 'var(--bg-card)',
              border: '1px solid var(--border)',
              borderRadius: 16, padding: '20px 22px',
              position: 'relative', overflow: 'hidden',
            }}
          >
            <div style={{
              position: 'absolute', top: -20, right: -20, width: 80, height: 80,
              borderRadius: '50%', background: color, opacity: 0.08, filter: 'blur(20px)'
            }} />
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 14 }}>
              <div style={{
                width: 40, height: 40, borderRadius: 10,
                background: `${color}18`, display: 'flex', alignItems: 'center', justifyContent: 'center'
              }}>
                <Icon size={20} color={color} />
              </div>
            </div>
            <div style={{ fontSize: 26, fontWeight: 800, letterSpacing: '-0.5px', marginBottom: 4 }}>
              {value}
            </div>
            <div style={{ fontSize: 13, color: 'var(--text-secondary)' }}>{label}</div>
          </motion.div>
        ))}
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: 24 }}>
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.35, duration: 0.5 }}
          style={{
            background: 'var(--bg-card)', border: '1px solid var(--border)',
            borderRadius: 20, padding: '24px 24px 16px',
          }}
        >
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
            <div>
              <h2 style={{ fontSize: 16, fontWeight: 700 }}>{t('dashboard.traffic_stats')} (KB/s)</h2>
              <p style={{ fontSize: 12, color: 'var(--text-muted)', marginTop: 3 }}>{t('dashboard.network_speed')}</p>
            </div>
            <div style={{ display: 'flex', gap: 16, fontSize: 12 }}>
              <span style={{ display: 'flex', alignItems: 'center', gap: 6, color: 'var(--text-secondary)' }}>
                <span style={{ width: 10, height: 3, borderRadius: 99, background: '#818cf8', display: 'inline-block' }} />
                Upload
              </span>
              <span style={{ display: 'flex', alignItems: 'center', gap: 6, color: 'var(--text-secondary)' }}>
                <span style={{ width: 10, height: 3, borderRadius: 99, background: 'var(--success)', display: 'inline-block' }} />
                Download
              </span>
            </div>
          </div>

          <ResponsiveContainer width="100%" height={220}>
            <AreaChart data={history} margin={{ top: 5, right: 0, left: -20, bottom: 0 }}>
              <defs>
                <linearGradient id="gUpload" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="#818cf8" stopOpacity={0.3} />
                  <stop offset="100%" stopColor="#818cf8" stopOpacity={0} />
                </linearGradient>
                <linearGradient id="gDownload" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="#10b981" stopOpacity={0.3} />
                  <stop offset="100%" stopColor="#10b981" stopOpacity={0} />
                </linearGradient>
              </defs>
              <XAxis dataKey="time" hide />
              <YAxis tick={{ fill: 'var(--text-muted)', fontSize: 11 }} axisLine={false} tickLine={false} />
              <Tooltip 
                contentStyle={{ background: 'var(--bg-elevated)', border: '1px solid var(--border)', borderRadius: '10px' }}
                itemStyle={{ fontSize: '12px' }}
              />
              <Area type="monotone" dataKey="upload" stroke="#818cf8" strokeWidth={2} fill="url(#gUpload)" isAnimationActive={false} />
              <Area type="monotone" dataKey="download" stroke="#10b981" strokeWidth={2} fill="url(#gDownload)" isAnimationActive={false} />
            </AreaChart>
          </ResponsiveContainer>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.45, duration: 0.5 }}
          style={{
            background: 'var(--bg-card)', border: '1px solid var(--border)',
            borderRadius: 20, padding: '24px 24px 16px',
          }}
        >
          <div style={{ marginBottom: 24 }}>
            <h2 style={{ fontSize: 16, fontWeight: 700 }}>{t('dashboard.cpu_load')} (%)</h2>
            <p style={{ fontSize: 12, color: 'var(--text-muted)', marginTop: 3 }}>Динамика процессора</p>
          </div>
          <ResponsiveContainer width="100%" height={220}>
            <AreaChart data={history} margin={{ top: 5, right: 0, left: -20, bottom: 0 }}>
              <defs>
                <linearGradient id="gCpu" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="var(--accent)" stopOpacity={0.3} />
                  <stop offset="100%" stopColor="var(--accent)" stopOpacity={0} />
                </linearGradient>
              </defs>
              <XAxis dataKey="time" hide />
              <YAxis domain={[0, 100]} tick={{ fill: 'var(--text-muted)', fontSize: 11 }} axisLine={false} tickLine={false} />
              <Tooltip 
                contentStyle={{ background: 'var(--bg-elevated)', border: '1px solid var(--border)', borderRadius: '10px' }}
                itemStyle={{ fontSize: '12px' }}
              />
              <Area type="monotone" dataKey="cpu" stroke="var(--accent)" strokeWidth={2} fill="url(#gCpu)" isAnimationActive={false} />
            </AreaChart>
          </ResponsiveContainer>
        </motion.div>
      </div>
    </div>
  );
}

