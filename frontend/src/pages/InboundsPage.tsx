import { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { http } from '../lib/api';
import {
  Plus, Trash2, Download,
  RefreshCw, Shield, Key, ChevronDown, ChevronRight,
  Wifi, Lock, Globe, Server
} from 'lucide-react';
import InboundModal from '../components/InboundModal';

interface Inbound {
  id: number;
  remark: string;
  protocol: string;
  port: number;
  enable: boolean;
  up: number;
  down: number;
  total: number;
  expiryTime: number;
  clientStats?: { email: string; enable: boolean; up: number; down: number }[];
}

const PROTOCOL_META: Record<string, { label: string; color: string; icon: any }> = {
  'vmess':          { label: 'VMess',          color: '#6366f1', icon: Shield },
  'vless':          { label: 'VLESS',          color: '#8b5cf6', icon: Shield },
  'trojan':         { label: 'Trojan',         color: '#06b6d4', icon: Lock },
  'shadowsocks':    { label: 'Shadowsocks',    color: '#f59e0b', icon: Globe },
  'amneziawg-v1':   { label: 'AmneziaWG v1',  color: '#10b981', icon: Wifi },
  'amneziawg-v2':   { label: 'AmneziaWG v2',  color: '#34d399', icon: Wifi },
  'openvpn-xor':    { label: 'OpenVPN XOR',   color: '#f97316', icon: Key },
};

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024, sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

function InboundCard({ inbound, onDelete }: { inbound: Inbound; onDelete: (id: number) => void }) {
  const [expanded, setExpanded] = useState(false);
  const meta = PROTOCOL_META[inbound.protocol] ?? { label: inbound.protocol, color: '#94a3b8', icon: Globe };
  const Icon = meta.icon;
  const usedPercent = inbound.total > 0 ? Math.round(((inbound.up + inbound.down) / inbound.total) * 100) : 0;

  return (
    <motion.div
      layout
      initial={{ opacity: 0, y: 16 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -10 }}
      style={{
        background: 'var(--bg-card)', border: '1px solid var(--border)',
        borderRadius: 16, overflow: 'hidden',
        transition: 'border-color 0.2s',
      }}
      whileHover={{ borderColor: 'rgba(99,102,241,0.3)' }}
    >
      <div
        style={{ display: 'flex', alignItems: 'center', padding: '16px 20px', cursor: 'pointer', gap: 14 }}
        onClick={() => setExpanded(!expanded)}
      >
        <div style={{
          width: 40, height: 40, borderRadius: 10, flexShrink: 0,
          background: `${meta.color}18`,
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          border: `1px solid ${meta.color}30`
        }}>
          <Icon size={18} color={meta.color} />
        </div>

        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{ fontWeight: 600, fontSize: 14, marginBottom: 3 }}>{inbound.remark || `Node #${inbound.id}`}</div>
          <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
            <span style={{
              fontSize: 11, fontWeight: 500, padding: '2px 8px', borderRadius: 99,
              background: `${meta.color}18`, color: meta.color, border: `1px solid ${meta.color}30`
            }}>{meta.label}</span>
            <span style={{ fontSize: 12, color: 'var(--text-muted)' }}>:{inbound.port}</span>
          </div>
        </div>

        <div style={{ textAlign: 'right', fontSize: 12, color: 'var(--text-secondary)', display: 'flex', gap: 16, marginRight: 8 }}>
          <div>
            <div style={{ color: 'var(--text-muted)', fontSize: 10, marginBottom: 2 }}>↑ Upload</div>
            <div>{formatBytes(inbound.up)}</div>
          </div>
          <div>
            <div style={{ color: 'var(--text-muted)', fontSize: 10, marginBottom: 2 }}>↓ Download</div>
            <div>{formatBytes(inbound.down)}</div>
          </div>
        </div>

        <div style={{
          width: 8, height: 8, borderRadius: '50%', flexShrink: 0,
          background: inbound.enable ? 'var(--success)' : 'var(--text-muted)',
          boxShadow: inbound.enable ? '0 0 8px rgba(16,185,129,0.6)' : 'none'
        }} />

        <div
          style={{ display: 'flex', gap: 6 }}
          onClick={e => e.stopPropagation()}
        >
          <ActionBtn icon={<Trash2 size={14} />} color="#ef4444" onClick={() => onDelete(inbound.id)} title="Удалить" />
        </div>

        {expanded ? <ChevronDown size={16} color="var(--text-muted)" /> : <ChevronRight size={16} color="var(--text-muted)" />}
      </div>

      {inbound.total > 0 && (
        <div style={{ padding: '0 20px 4px' }}>
          <div style={{ height: 3, borderRadius: 99, background: 'var(--bg-elevated)', overflow: 'hidden' }}>
            <motion.div
              initial={{ width: 0 }}
              animate={{ width: `${usedPercent}%` }}
              transition={{ duration: 0.8, ease: 'easeOut' }}
              style={{ height: '100%', background: `linear-gradient(90deg, ${meta.color}, ${meta.color}99)`, borderRadius: 99 }}
            />
          </div>
        </div>
      )}

      <AnimatePresence>
        {expanded && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: 'auto', opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.25, ease: 'easeInOut' }}
            style={{ overflow: 'hidden' }}
          >
            <div style={{ padding: '4px 20px 20px', borderTop: '1px solid var(--border)', marginTop: 4 }}>
              <div style={{ fontSize: 12, fontWeight: 600, color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.8px', marginBottom: 12, paddingTop: 12 }}>
                Клиенты ({inbound.clientStats?.length ?? 0})
              </div>
              {!inbound.clientStats?.length ? (
                <p style={{ fontSize: 13, color: 'var(--text-muted)' }}>Нет клиентов</p>
              ) : (
                <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                  {inbound.clientStats?.map(c => (
                    <div key={c.email} style={{
                      display: 'flex', alignItems: 'center', gap: 12,
                      padding: '9px 14px', borderRadius: 10,
                      background: 'var(--bg-elevated)', border: '1px solid var(--border)'
                    }}>
                      <div style={{
                        width: 7, height: 7, borderRadius: '50%',
                        background: c.enable ? 'var(--success)' : 'var(--text-muted)',
                        flexShrink: 0
                      }} />
                      <span style={{ fontSize: 13, flex: 1 }}>{c.email}</span>
                      <span style={{ fontSize: 11, color: 'var(--text-muted)' }}>{formatBytes(c.up)} / {formatBytes(c.down)}</span>
                      <button style={{
                        background: 'none', border: 'none', cursor: 'pointer',
                        color: 'var(--text-muted)', display: 'flex', padding: 4, borderRadius: 6
                      }}
                        title="Скачать конфигурацию"
                        onClick={() => alert(`Скачать конфиг для ${c.email}`)}
                      >
                        <Download size={14} />
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </motion.div>
  );
}

function ActionBtn({ icon, color, onClick, title }: { icon: any; color: string; onClick: () => void; title: string }) {
  return (
    <button
      onClick={onClick}
      title={title}
      style={{
        background: `${color}12`, border: `1px solid ${color}25`,
        color, borderRadius: 8, width: 30, height: 30,
        cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center',
        transition: 'all 0.15s ease'
      }}
      onMouseEnter={e => { e.currentTarget.style.background = `${color}25`; }}
      onMouseLeave={e => { e.currentTarget.style.background = `${color}12`; }}
    >
      {icon}
    </button>
  );
}

export default function InboundsPage() {
  const [inbounds, setInbounds] = useState<Inbound[]>([]);
  const [loading, setLoading] = useState(true);
  const [modalOpen, setModalOpen] = useState(false);

  const fetchInbounds = async () => {
    setLoading(true);
    try {
      const { data } = await http.get('/inbounds');
      if (data.success) {
        setInbounds(data.obj || []);
      }
    } catch {
      console.error('Failed to fetch inbounds');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchInbounds();
  }, []);

  const handleDelete = async (id: number) => {
    if (!confirm('Вы уверены, что хотите удалить эту ноду?')) return;
    try {
      const { data } = await http.delete(`/inbounds/${id}`);
      if (data.success) fetchInbounds();
    } catch (err) {
      alert('Ошибка при удалении');
    }
  };

  const handleSave = async (payload: any) => {
    try {
      const { data } = await http.post('/inbounds', payload);
      if (data.success) {
        setModalOpen(false);
        fetchInbounds();
      }
    } catch (err) {
      alert('Ошибка при сохранении');
    }
  };

  return (
    <div>
      <motion.div
        initial={{ opacity: 0, y: -10 }}
        animate={{ opacity: 1, y: 0 }}
        style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 28 }}
      >
        <div>
          <h1 style={{ fontSize: 26, fontWeight: 800, letterSpacing: '-0.5px' }}>Ноды</h1>
          <p style={{ color: 'var(--text-secondary)', fontSize: 14, marginTop: 4 }}>
            Управление VPN и Proxy серверами
          </p>
        </div>

        <div style={{ display: 'flex', gap: 10 }}>
          <button
            onClick={fetchInbounds}
            style={{
              display: 'flex', alignItems: 'center', gap: 8,
              padding: '9px 16px', borderRadius: 10, border: '1px solid var(--border)',
              background: 'var(--bg-card)', color: 'var(--text-secondary)',
              cursor: 'pointer', fontSize: 13, fontFamily: 'inherit'
            }}
          >
            <motion.span animate={loading ? { rotate: 360 } : { rotate: 0 }} transition={{ duration: 0.8, repeat: loading ? Infinity : 0, ease: 'linear' }}>
              <RefreshCw size={14} />
            </motion.span>
            Обновить
          </button>

          <motion.button
            whileTap={{ scale: 0.97 }}
            className="glow-btn"
            onClick={() => setModalOpen(true)}
            style={{
              display: 'flex', alignItems: 'center', gap: 8,
              padding: '9px 18px', borderRadius: 10, border: 'none',
              color: 'white', cursor: 'pointer', fontSize: 13,
              fontFamily: 'inherit', fontWeight: 600
            }}
          >
            <Plus size={16} />
            Добавить ноду
          </motion.button>
        </div>
      </motion.div>

      <motion.div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <AnimatePresence>
          {inbounds.map(inbound => (
            <InboundCard key={inbound.id} inbound={inbound} onDelete={handleDelete} />
          ))}
        </AnimatePresence>

        {!loading && !inbounds.length && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            style={{
              textAlign: 'center', padding: '60px 20px',
              background: 'var(--bg-card)', borderRadius: 16, border: '1px dashed var(--border)'
            }}
          >
            <Server size={40} color="var(--text-muted)" style={{ margin: '0 auto 16px' }} />
            <p style={{ color: 'var(--text-secondary)', fontSize: 15 }}>Нет нод</p>
            <p style={{ color: 'var(--text-muted)', fontSize: 13, marginTop: 6 }}>Добавьте первую ноду для начала работы</p>
          </motion.div>
        )}
      </motion.div>

      <InboundModal 
        isOpen={modalOpen} 
        onClose={() => setModalOpen(false)} 
        onSave={handleSave}
      />
    </div>
  );
}
