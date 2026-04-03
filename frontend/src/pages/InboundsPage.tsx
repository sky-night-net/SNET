import { useState, useEffect } from 'react';
import { Plus, Server, ArrowUpRight, ArrowDownLeft, Zap, Shield, Trash2, Edit2, Copy, Check, Wifi, QrCode } from 'lucide-react';
import { api } from '../lib/api';
import { Card, Button, Badge } from '../components/UI';
import InboundModal from '../components/InboundModal';
import QRModal from '../components/QRModal';
import { useTranslation } from 'react-i18next';

function formatBytes(bytes: number) {
  if (!bytes || bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

const PROTOCOL_COLORS: Record<string, string> = {
  vless:       '#818cf8',
  vmess:       '#a78bfa',
  trojan:      '#f59e0b',
  shadowsocks: '#10b981',
  amneziawg:   '#06b6d4',
  openvpn:     '#f97316',
};

const PROTOCOL_ICONS: Record<string, React.ElementType> = {
  vless:       Shield,
  vmess:       Shield,
  trojan:      Zap,
  shadowsocks: Globe,
  amneziawg:   Wifi,
  openvpn:     Globe,
};

function Globe(props: any) {
  return (
    <svg {...props} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/>
      <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/>
    </svg>
  );
}

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);
  const handleCopy = () => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 1800);
  };
  return (
    <button
      onClick={handleCopy}
      title="Copy to clipboard"
      style={{
        padding: '6px 10px', borderRadius: 8, border: '1px solid var(--border)',
        background: 'transparent', color: copied ? 'var(--success)' : 'var(--text-muted)',
        cursor: 'pointer', fontSize: 12, display: 'flex', alignItems: 'center', gap: 4,
        transition: 'all 0.2s',
      }}
    >
      {copied ? <Check size={12} /> : <Copy size={12} />}
      {copied ? 'Copied' : 'Copy'}
    </button>
  );
}

function TrafficBar({ used, total }: { used: number; total: number }) {
  if (total <= 0) return null;
  const pct = Math.min(100, (used / total) * 100);
  const color = pct > 90 ? '#ef4444' : pct > 70 ? '#f59e0b' : 'var(--accent)';
  return (
    <div style={{ marginTop: 8 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 11, color: 'var(--text-muted)', marginBottom: 4 }}>
        <span>{formatBytes(used)} used</span>
        <span>{formatBytes(total)} total</span>
      </div>
      <div style={{ height: 4, borderRadius: 99, background: 'rgba(255,255,255,0.06)', overflow: 'hidden' }}>
        <div style={{ height: '100%', width: `${pct}%`, background: color, borderRadius: 99, transition: 'width 0.4s ease' }} />
      </div>
    </div>
  );
}

export default function InboundsPage() {
  const { t } = useTranslation();
  const [inbounds, setInbounds] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingInbound, setEditingInbound] = useState<any>(null);

  // QR Modal
  const [qrModalOpen, setQrModalOpen] = useState(false);
  const [qrData, setQrData] = useState({ link: '', title: '' });

  const openQr = (link: string, title: string) => {
    setQrData({ link, title });
    setQrModalOpen(true);
  };

  const fetchInbounds = async () => {
    try {
      const res = await api.get('/inbounds');
      if (res.data.success) setInbounds(res.data.obj || []);
    } catch { /* ignore */ } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchInbounds(); }, []);

  const handleCreate = () => { setEditingInbound(null); setIsModalOpen(true); };
  const handleEdit   = (ib: any) => { setEditingInbound(ib);   setIsModalOpen(true); };
  
  const handleDelete = async (id: number) => {
    try {
      await api.delete(`/inbounds/${id}`);
      fetchInbounds();
    } catch {
      alert(t('common.error_deleting'));
    }
  };

  // Returns the correct share/config string for each protocol.
  // Uses the backend-configured EXTERNAL_IP (via API base URL) as the server host.
  const getShareLink = (ib: any) => {
    try {
      const settings = JSON.parse(ib.settings || '{}');
      const stream   = JSON.parse(ib.streamSettings || '{}');
      const client   = settings.clients?.[0];
      if (!client) return `(add a client to ${ib.remark} first)`;

      // Use the API server hostname as the public VPN server IP
      // The backend has EXTERNAL_IP configured in its environment
      const apiBase = (window as any).__SNET_API_HOST__ || window.location.hostname;
      const host = apiBase;
      const port  = ib.port;
      const remark = encodeURIComponent(ib.remark || `node-${ib.id}`);
      const proto  = (ib.protocol || '').toLowerCase();

      // ── VLESS ────────────────────────────────────────────────────────────
      if (proto === 'vless') {
        const uid      = client.id || '';
        const security = stream.security || 'none';
        const net      = stream.network  || 'tcp';
        const params = new URLSearchParams({ security, type: net });
        if (security === 'tls' && stream.tlsSettings?.serverName)
          params.set('sni', stream.tlsSettings.serverName);
        if (security === 'reality' && stream.realitySettings) {
          const rs = stream.realitySettings;
          params.set('pbk', rs.publicKey || '');
          if (rs.serverNames?.[0]) params.set('sni', rs.serverNames[0]);
          if (rs.shortIds?.[0])    params.set('sid', rs.shortIds[0]);
          params.set('fp', 'chrome');
        }
        if (client.flow) params.set('flow', client.flow);
        return `vless://${uid}@${host}:${port}?${params.toString()}#${remark}`;
      }

      // ── VMess ────────────────────────────────────────────────────────────
      if (proto === 'vmess') {
        const obj: Record<string, any> = {
          v: '2', ps: decodeURIComponent(remark),
          add: host, port, id: client.id || '',
          aid: '0', scy: 'auto',
          net: stream.network || 'tcp',
          type: 'none', host: '', path: '', tls: '', sni: '', alpn: '', fp: '',
        };
        if (stream.network === 'ws' && stream.wsSettings) {
          obj.path = stream.wsSettings.path || '';
          obj.host = stream.wsSettings.headers?.Host || '';
        }
        if (stream.network === 'grpc' && stream.grpcSettings)
          obj.path = stream.grpcSettings.serviceName || '';
        if (stream.security && stream.security !== 'none') obj.tls = stream.security;
        return 'vmess://' + btoa(JSON.stringify(obj));
      }

      // ── Trojan ───────────────────────────────────────────────────────────
      if (proto === 'trojan') {
        const pass     = client.password || '';
        const security = stream.security || 'tls';
        const net      = stream.network  || 'tcp';
        const params   = new URLSearchParams({ security, type: net });
        if (stream.tlsSettings?.serverName) params.set('sni', stream.tlsSettings.serverName);
        return `trojan://${pass}@${host}:${port}?${params.toString()}#${remark}`;
      }

      // ── Shadowsocks ──────────────────────────────────────────────────────
      if (proto === 'shadowsocks') {
        const method   = settings.method || 'chacha20-ietf-poly1305';
        const password = client.password || '';
        const creds    = btoa(`${method}:${password}`);
        return `ss://${creds}@${host}:${port}#${remark}`;
      }

      // ── AmneziaWG / OpenVPN — no short URI, show instructions ────────────
      if (proto.startsWith('amneziawg') || proto.startsWith('amnezia')) {
        return `[AmneziaWG] Download the .conf file from the client panel to import into AmneziaVPN app\nServer: ${host}:${port}`;
      }
      if (proto.startsWith('openvpn')) {
        return `[OpenVPN XOR] Download the .ovpn file from the client panel to import into your OpenVPN client\nServer: ${host}:${port}`;
      }

    } catch { /* ignore */ }
    return `(error generating config for ${ib.protocol})`;
  };


  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <div>
          <h1 style={{ fontSize: 22, fontWeight: 800, letterSpacing: '-0.4px' }}>{t('nav.nodes')}</h1>
          <p style={{ fontSize: 13, color: 'var(--text-muted)', marginTop: 4 }}>{t('inbound.subtitle')}</p>
        </div>
        <Button onClick={handleCreate} style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <Plus size={16} /> {t('common.add')}
        </Button>
      </div>

      {/* Grid */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))', gap: 20 }}>
        {loading ? (
          <div style={{ gridColumn: '1/-1', padding: 48, textAlign: 'center', color: 'var(--text-muted)', fontSize: 14 }}>
            {t('common.loading')}...
          </div>
        ) : inbounds.length === 0 ? (
          <div style={{
            gridColumn: '1/-1', padding: 48, textAlign: 'center',
            border: '2px dashed var(--border)', borderRadius: 20,
            display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 12,
          }}>
            <Server size={40} style={{ color: 'var(--text-muted)', opacity: 0.3 }} />
            <p style={{ color: 'var(--text-muted)', fontSize: 14 }}>{t('inbound.empty')}</p>
            <Button onClick={handleCreate}><Plus size={14} /> {t('common.add')}</Button>
          </div>
        ) : (
          inbounds.map(ib => {
            const proto  = ib.protocol?.toLowerCase() || 'vless';
            const color  = PROTOCOL_COLORS[proto] || 'var(--accent)';
            const Icon   = PROTOCOL_ICONS[proto] || Shield;
            const used   = (ib.up || 0) + (ib.down || 0);
            return (
              <Card key={ib.id} style={{ position: 'relative', overflow: 'hidden' }}>
                {/* Glow */}
                <div style={{ position: 'absolute', top: -30, right: -30, width: 100, height: 100, borderRadius: '50%', background: color, opacity: 0.06, filter: 'blur(30px)', pointerEvents: 'none' }} />

                {/* Header row */}
                <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', marginBottom: 16 }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                    <div style={{ width: 40, height: 40, borderRadius: 12, background: `${color}18`, display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                      <Icon size={20} color={color} />
                    </div>
                    <div>
                      <div style={{ fontWeight: 700, fontSize: 15, lineHeight: 1.2 }}>{ib.remark || `Node ${ib.id}`}</div>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginTop: 4 }}>
                        <Badge variant="primary" style={{ background: `${color}20`, color, border: `1px solid ${color}30`, textTransform: 'uppercase', fontSize: 10 }}>
                          {ib.protocol}
                        </Badge>
                        <span style={{ fontSize: 12, color: 'var(--text-muted)', fontFamily: 'monospace' }}>:{ib.port}</span>
                      </div>
                    </div>
                  </div>
                  {/* Status dot */}
                  <div style={{
                    width: 8, height: 8, borderRadius: '50%', flexShrink: 0, marginTop: 4,
                    background: ib.enable ? '#22c55e' : 'var(--text-muted)',
                    boxShadow: ib.enable ? '0 0 8px rgba(34,197,94,0.5)' : 'none',
                  }} />
                </div>

                {/* Traffic stats */}
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10, marginBottom: 12 }}>
                  <div style={{ padding: '10px 12px', borderRadius: 10, background: 'rgba(255,255,255,0.03)', border: '1px solid var(--border)' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 11, color: 'var(--text-muted)', marginBottom: 4 }}>
                      <ArrowUpRight size={11} color="#818cf8" /> Upload
                    </div>
                    <div style={{ fontWeight: 700, fontSize: 13, fontFamily: 'monospace' }}>{formatBytes(ib.up)}</div>
                  </div>
                  <div style={{ padding: '10px 12px', borderRadius: 10, background: 'rgba(255,255,255,0.03)', border: '1px solid var(--border)' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 11, color: 'var(--text-muted)', marginBottom: 4 }}>
                      <ArrowDownLeft size={11} color="#10b981" /> Download
                    </div>
                    <div style={{ fontWeight: 700, fontSize: 13, fontFamily: 'monospace' }}>{formatBytes(ib.down)}</div>
                  </div>
                </div>

                <TrafficBar used={used} total={ib.total} />

                {/* Footer actions */}
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginTop: 14, paddingTop: 12, borderTop: '1px solid var(--border)' }}>
                  <div style={{ display: 'flex', gap: 6 }}>
                    <CopyButton text={getShareLink(ib)} />
                    <button
                      onClick={() => openQr(getShareLink(ib), ib.remark || 'Node')}
                      style={{
                        padding: '6px 10px', borderRadius: 8, border: '1px solid var(--border)',
                        background: 'transparent', color: 'var(--text-muted)',
                        cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 4, transition: 'all 0.2s', fontSize: 12
                      }}
                      title="QR Code"
                    >
                      <QrCode size={12} />
                      QR Code
                    </button>
                  </div>
                  <div style={{ display: 'flex', gap: 4 }}>
                    <button
                      onClick={() => handleEdit(ib)}
                      style={{ padding: '6px 10px', borderRadius: 8, border: 'none', background: 'transparent', color: 'var(--text-muted)', cursor: 'pointer', transition: 'all 0.2s' }}
                      title={t('common.edit')}
                    >
                      <Edit2 size={15} />
                    </button>
                    <button
                      onClick={(e) => { e.stopPropagation(); handleDelete(ib.id); }}
                      style={{ padding: '6px 10px', borderRadius: 8, border: 'none', background: 'transparent', color: 'var(--danger)', cursor: 'pointer', transition: 'all 0.2s' }}
                      title={t('common.delete')}
                    >
                      <Trash2 size={15} />
                    </button>
                  </div>
                </div>
              </Card>
            );
          })
        )}
      </div>

      <InboundModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSuccess={fetchInbounds}
        inbound={editingInbound}
      />

      <QRModal
        isOpen={qrModalOpen}
        onClose={() => setQrModalOpen(false)}
        title={qrData.title}
        link={qrData.link}
      />
    </div>
  );
}
