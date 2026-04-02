import { useState, useEffect } from 'react';
import { motion, type Variants } from 'framer-motion';
import { Server, Activity, Cpu, Globe } from 'lucide-react';
import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts';
import { http } from '../lib/api';

// Mock traffic for chart (can be improved later with historical data)
const generateMockTraffic = () => Array.from({ length: 12 }, (_, i) => ({
  time: `${String(i * 2).padStart(2, '0')}:00`,
  upload: Math.floor(Math.random() * 80 + 20),
  download: Math.floor(Math.random() * 120 + 40),
}));

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
  const [stats, setStats] = useState<any>(null);
  const [trafficData] = useState(generateMockTraffic());
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const { data } = await http.get('/system/status');
        if (data.success) {
          setStats(data.obj);
        }
      } catch (err) {
        console.error('Failed to fetch system stats');
      } finally {
        setLoading(false);
      }
    };

    fetchStats();
    const timer = setInterval(fetchStats, 5000);
    return () => clearInterval(timer);
  }, []);

  const displayStats = [
    { label: 'CPU Usage',     value: stats ? `${stats.cpu.toFixed(1)}%` : '0%', icon: Cpu,      color: '#a78bfa' },
    { label: 'Memory',        value: stats ? `${(stats.mem.current / stats.mem.total * 100).toFixed(0)}%` : '0%', icon: Activity, color: 'var(--success)' },
    { label: 'Uptime',        value: stats ? `${(stats.uptime / 3600).toFixed(1)}h` : '0h', icon: Server,   color: 'var(--accent)' },
    { label: 'Total Sent',    value: stats ? formatBytes(stats.net.bytesSent) : '0 B', icon: Globe,    color: '#f59e0b' },
  ];

  return (
    <div>
      <motion.div
        initial={{ opacity: 0, y: -10 }}
        animate={{ opacity: 1, y: 0 }}
        style={{ marginBottom: 32 }}
      >
        <h1 style={{ fontSize: 26, fontWeight: 800, letterSpacing: '-0.5px' }}>
          Дашборд
        </h1>
        <p style={{ color: 'var(--text-secondary)', fontSize: 14, marginTop: 4 }}>
          {loading ? 'Загрузка системных данных...' : 'Живой мониторинг вашей инфраструктуры'}
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
            <h2 style={{ fontSize: 16, fontWeight: 700 }}>Сетевая активность</h2>
            <p style={{ fontSize: 12, color: 'var(--text-muted)', marginTop: 3 }}>Общая нагрузка интерфейса</p>
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
          <AreaChart data={trafficData} margin={{ top: 5, right: 0, left: -20, bottom: 0 }}>
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
            <XAxis dataKey="time" tick={{ fill: 'var(--text-muted)', fontSize: 11 }} axisLine={false} tickLine={false} />
            <YAxis tick={{ fill: 'var(--text-muted)', fontSize: 11 }} axisLine={false} tickLine={false} />
            <Tooltip 
              contentStyle={{ background: 'var(--bg-elevated)', border: '1px solid var(--border)', borderRadius: '10px' }}
              itemStyle={{ fontSize: '12px' }}
            />
            <Area type="monotone" dataKey="upload" stroke="#818cf8" strokeWidth={2} fill="url(#gUpload)" dot={false} />
            <Area type="monotone" dataKey="download" stroke="#10b981" strokeWidth={2} fill="url(#gDownload)" dot={false} />
          </AreaChart>
        </ResponsiveContainer>
      </motion.div>
    </div>
  );
}
