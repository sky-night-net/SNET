import { useState, useEffect } from 'react';
import { motion, type Variants } from 'framer-motion';
import { Server, Activity, Cpu, ArrowUpRight, ArrowDownLeft, Clock } from 'lucide-react';
import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts';
import { http } from '../lib/api';
import { useTranslation } from 'react-i18next';

const fadeUp: Variants = {
  hidden:  { opacity: 0, y: 16 },
  visible: (i: number) => ({ opacity: 1, y: 0, transition: { delay: i * 0.07, duration: 0.38, ease: 'easeOut' } }),
};

function formatBytes(bytes: number) {
  if (!bytes || bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

function formatUptime(seconds: number) {
  const d = Math.floor(seconds / 86400);
  const h = Math.floor((seconds % 86400) / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  if (d > 0) return `${d}d ${h}h`;
  if (h > 0) return `${h}h ${m}m`;
  return `${m}m`;
}

function StatCard({ label, value, sub, icon: Icon, color, index }: {
  label: string; value: string; sub?: string;
  icon: React.ElementType; color: string; index: number;
}) {
  return (
    <motion.div
      custom={index} initial="hidden" animate="visible" variants={fadeUp}
      style={{
        background: 'var(--bg-card)', border: '1px solid var(--border)',
        borderRadius: 16, padding: '20px 22px', position: 'relative', overflow: 'hidden',
      }}
    >
      <div style={{ position: 'absolute', top: -24, right: -24, width: 90, height: 90, borderRadius: '50%', background: color, opacity: 0.08, filter: 'blur(24px)', pointerEvents: 'none' }} />
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 16 }}>
        <div style={{ width: 40, height: 40, borderRadius: 10, background: `${color}18`, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <Icon size={20} color={color} />
        </div>
        {sub && <span style={{ fontSize: 11, color: 'var(--text-muted)', background: 'rgba(255,255,255,0.04)', padding: '2px 8px', borderRadius: 99, border: '1px solid var(--border)' }}>{sub}</span>}
      </div>
      <div style={{ fontSize: 28, fontWeight: 800, letterSpacing: '-0.6px', lineHeight: 1, marginBottom: 6 }}>{value}</div>
      <div style={{ fontSize: 12, color: 'var(--text-secondary)' }}>{label}</div>
    </motion.div>
  );
}

function ChartCard({ title, sub, children, delay }: { title: string; sub?: string; children: React.ReactNode; delay: number }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 18 }} animate={{ opacity: 1, y: 0 }} transition={{ delay, duration: 0.45, ease: 'easeOut' }}
      style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', borderRadius: 20, padding: '22px 22px 14px', overflow: 'hidden' }}
    >
      <div style={{ marginBottom: 18 }}>
        <div style={{ fontSize: 15, fontWeight: 700 }}>{title}</div>
        {sub && <div style={{ fontSize: 12, color: 'var(--text-muted)', marginTop: 3 }}>{sub}</div>}
      </div>
      {children}
    </motion.div>
  );
}

const tooltipStyle = {
  contentStyle: { background: 'var(--bg-elevated)', border: '1px solid var(--border)', borderRadius: 10, fontSize: 12 },
  itemStyle: { fontSize: 12 },
};

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
              time:     new Date(p.timestamp * 1000).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' }),
              cpu:      +p.cpu.toFixed(1),
              upload:   +(p.netUp   / 1024).toFixed(1),
              download: +(p.netDown / 1024).toFixed(1),
              mem:      Math.round(p.mem / (1024 * 1024)),
            })));
          }
        }
      } catch { /* ignore */ } finally {
        setLoading(false);
      }
    };
    fetchStats();
    const id = setInterval(fetchStats, 2000);
    return () => clearInterval(id);
  }, []);

  const ramPct = stats ? +(stats.mem.current / stats.mem.total * 100).toFixed(1) : 0;

  const statCards = [
    { label: t('dashboard.cpu_load'),  value: stats ? `${stats.cpu.toFixed(1)}%`   : '—',  sub: 'CPU',                     icon: Cpu,      color: '#a78bfa' },
    { label: t('dashboard.ram_usage'), value: stats ? `${ramPct}%`                  : '—',  sub: stats ? formatBytes(stats.mem.current) : undefined, icon: Activity, color: '#10b981' },
    { label: 'Uptime',                 value: stats ? formatUptime(stats.uptime)     : '—',  sub: undefined,                 icon: Clock,    color: 'var(--accent)' },
    { label: t('dashboard.net_sent'),  value: stats ? formatBytes(stats.net.bytesSent)  : '—', sub: 'TX',                   icon: ArrowUpRight,   color: '#818cf8' },
    { label: t('dashboard.net_recv'),  value: stats ? formatBytes(stats.net.bytesRecv) : '—', sub: 'RX',                    icon: ArrowDownLeft,  color: '#f59e0b' },
    { label: 'Goroutines',             value: stats ? `${stats.go?.goroutines ?? 0}` : '—', sub: stats?.go?.version,        icon: Server,   color: '#06b6d4' },
  ];

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>

      {/* Page title */}
      <motion.div initial={{ opacity: 0, y: -8 }} animate={{ opacity: 1, y: 0 }}>
        <h1 style={{ fontSize: 22, fontWeight: 800, letterSpacing: '-0.4px' }}>{t('dashboard.title')}</h1>
        <p style={{ fontSize: 13, color: 'var(--text-secondary)', marginTop: 4 }}>
          {loading ? t('common.loading') + '…' : t('dashboard.system_status')}
        </p>
      </motion.div>

      {/* Stat cards — always 3 per row on wide screens */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 16 }}>
        {statCards.map((c, i) => <StatCard key={c.label} {...c} index={i} />)}
      </div>

      {/* Charts row */}
      <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: 20 }}>

        {/* Network throughput */}
        <ChartCard
          title={t('dashboard.traffic_stats') + ' (KB/s)'}
          sub={t('dashboard.network_speed')}
          delay={0.32}
        >
          <div style={{ display: 'flex', gap: 20, fontSize: 12, marginBottom: 12 }}>
            <span style={{ display: 'flex', alignItems: 'center', gap: 6, color: 'var(--text-secondary)' }}>
              <span style={{ width: 12, height: 3, borderRadius: 99, background: '#818cf8', display: 'inline-block' }} />
              Upload
            </span>
            <span style={{ display: 'flex', alignItems: 'center', gap: 6, color: 'var(--text-secondary)' }}>
              <span style={{ width: 12, height: 3, borderRadius: 99, background: '#10b981', display: 'inline-block' }} />
              Download
            </span>
          </div>
          <ResponsiveContainer width="100%" height={200}>
            <AreaChart data={history} margin={{ top: 4, right: 0, left: -22, bottom: 0 }}>
              <defs>
                <linearGradient id="gUp"   x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%"   stopColor="#818cf8" stopOpacity={0.28} />
                  <stop offset="100%" stopColor="#818cf8" stopOpacity={0} />
                </linearGradient>
                <linearGradient id="gDown" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%"   stopColor="#10b981" stopOpacity={0.28} />
                  <stop offset="100%" stopColor="#10b981" stopOpacity={0} />
                </linearGradient>
              </defs>
              <XAxis dataKey="time" hide />
              <YAxis tick={{ fill: 'var(--text-muted)', fontSize: 10 }} axisLine={false} tickLine={false} />
              <Tooltip {...tooltipStyle} />
              <Area type="monotone" dataKey="upload"   stroke="#818cf8" strokeWidth={1.8} fill="url(#gUp)"   isAnimationActive={false} name="Upload KB/s" />
              <Area type="monotone" dataKey="download" stroke="#10b981" strokeWidth={1.8} fill="url(#gDown)" isAnimationActive={false} name="Download KB/s" />
            </AreaChart>
          </ResponsiveContainer>
        </ChartCard>

        {/* CPU */}
        <ChartCard title={t('dashboard.cpu_load') + ' (%)'}  sub={t('dashboard.system_status')} delay={0.42}>
          <ResponsiveContainer width="100%" height={200}>
            <AreaChart data={history} margin={{ top: 4, right: 0, left: -22, bottom: 0 }}>
              <defs>
                <linearGradient id="gCpu" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%"   stopColor="var(--accent)" stopOpacity={0.3} />
                  <stop offset="100%" stopColor="var(--accent)" stopOpacity={0} />
                </linearGradient>
              </defs>
              <XAxis dataKey="time" hide />
              <YAxis domain={[0, 100]} tick={{ fill: 'var(--text-muted)', fontSize: 10 }} axisLine={false} tickLine={false} />
              <Tooltip {...tooltipStyle} />
              <Area type="monotone" dataKey="cpu" stroke="var(--accent)" strokeWidth={1.8} fill="url(#gCpu)" isAnimationActive={false} name="CPU %" />
            </AreaChart>
          </ResponsiveContainer>
        </ChartCard>
      </div>

      {/* RAM chart */}
      <ChartCard title={t('dashboard.ram_usage') + ' (MB)'} sub={stats ? `${formatBytes(stats.mem.current)} / ${formatBytes(stats.mem.total)}` : ''} delay={0.52}>
        <ResponsiveContainer width="100%" height={140}>
          <AreaChart data={history} margin={{ top: 4, right: 0, left: -22, bottom: 0 }}>
            <defs>
              <linearGradient id="gMem" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%"   stopColor="#10b981" stopOpacity={0.28} />
                <stop offset="100%" stopColor="#10b981" stopOpacity={0} />
              </linearGradient>
            </defs>
            <XAxis dataKey="time" hide />
            <YAxis tick={{ fill: 'var(--text-muted)', fontSize: 10 }} axisLine={false} tickLine={false} />
            <Tooltip {...tooltipStyle} />
            <Area type="monotone" dataKey="mem" stroke="#10b981" strokeWidth={1.8} fill="url(#gMem)" isAnimationActive={false} name="RAM MB" />
          </AreaChart>
        </ResponsiveContainer>
      </ChartCard>

    </div>
  );
}
