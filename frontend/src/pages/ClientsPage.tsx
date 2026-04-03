import { useState, useEffect, useMemo } from 'react';
import {
  Users, Plus, Trash2, Copy, Check, RefreshCw, QrCode,
  Shield, Activity, ArrowUpRight, ArrowDownLeft, Download
} from 'lucide-react';
import { api } from '../lib/api';
import { Card, Button, Badge, Select } from '../components/UI';
import ClientModal from '../components/ClientModal';
import QRModal from '../components/QRModal';
import { useTranslation } from 'react-i18next';

function formatBytes(bytes: number) {
  if (!bytes || bytes <= 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  if (i < 0) return '0 B';
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

function CopyBtn({ client }: { client: any }) {
  const [copied, setCopied] = useState(false);
  
  const handleCopy = async () => {
    try {
      const res = await api.get(`/inbounds/${client._inboundId}/clients/${client.email || client.id}/config`);
      if (res.data.success) {
        navigator.clipboard.writeText(res.data.obj);
        setCopied(true);
        setTimeout(() => setCopied(false), 1500);
      }
    } catch { /* ignore */ }
  };

  return (
    <button
      onClick={handleCopy}
      title="Copy"
      style={{ padding: '4px 8px', borderRadius: 6, border: '1px solid var(--border)', background: 'transparent', color: copied ? 'var(--success)' : 'var(--text-muted)', cursor: 'pointer', display: 'inline-flex', alignItems: 'center', gap: 4, fontSize: 11 }}
    >
      {copied ? <Check size={11} /> : <Copy size={11} />}
      {copied ? 'OK' : 'Copy'}
    </button>
  );
}

function DownloadBtn({ client, protocol }: { client: any; protocol: string }) {
  const [downloading, setDownloading] = useState(false);
  
  const handleDownload = async () => {
    try {
      setDownloading(true);
      const res = await api.get(`/inbounds/${client._inboundId}/clients/${client.email || client.id}/config`);
      if (res.data.success) {
        const content = res.data.obj;
        const blob = new Blob([content], { type: 'text/plain' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        const ext = protocol.includes('openvpn') ? 'ovpn' : 'conf';
        a.download = `${client.email || client.id || 'config'}_${protocol}.${ext}`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
      }
    } catch { /* ignore */ } finally {
      setDownloading(false);
    }
  };

  return (
    <button
      onClick={handleDownload}
      disabled={downloading}
      title="Download Config File"
      style={{ padding: '4px 8px', borderRadius: 6, border: '1px solid var(--border)', background: 'transparent', color: 'var(--text-muted)', cursor: 'pointer', display: 'inline-flex', alignItems: 'center', gap: 4, fontSize: 11 }}
    >
      <Download size={11} />
      {downloading ? '...' : 'Download'}
    </button>
  );
}

export default function ClientsPage() {
  const { t } = useTranslation();

  // All inbounds (nodes) — needed to know which node to add client to
  const [inbounds, setInbounds] = useState<any[]>([]);
  // Flat list of clients with their parent node info attached
  const [clients, setClients] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  // Modal state
  const [modalOpen, setModalOpen] = useState(false);
  const [selectedInboundId, setSelectedInboundId] = useState<number | null>(null);

  // Filter
  const [filterInbound, setFilterInbound] = useState<string>('all');

  // QR Modal
  const [qrModalOpen, setQrModalOpen] = useState(false);
  const [qrData, setQrData] = useState({ link: '', title: '' });

  const openQr = (link: string, title: string) => {
    setQrData({ link, title });
    setQrModalOpen(true);
  };

  const loadAll = async () => {
    setLoading(true);
    try {
      const res = await api.get('/inbounds');
      if (!res.data.success) return;
      const ibs: any[] = res.data.obj || [];
      setInbounds(ibs);

      // Extract clients from each inbound's settings JSON
      const allClients: any[] = [];
      for (const ib of ibs) {
        try {
          const settings = JSON.parse(ib.settings || '{}');
          const rawClients: any[] = settings.clients || [];
          for (const c of rawClients) {
            allClients.push({
              ...c,
              _inboundId:       ib.id,
              _inboundRemark:   ib.remark,
              _inboundProtocol: ib.protocol,
              _inboundPort:     ib.port,
            });
          }
        } catch { /* ignore malformed */ }
      }
      setClients(allClients);
    } catch { /* ignore */ } finally {
      setLoading(false);
    }
  };

  useEffect(() => { loadAll(); }, []);

  const handleAddClient = (inboundId: number) => {
    setSelectedInboundId(inboundId);
    setModalOpen(true);
  };

  const handleSaveClient = async (clientData: any) => {
    if (!selectedInboundId) return;
    try {
      await api.post(`/inbounds/${selectedInboundId}/clients`, clientData);
      setModalOpen(false);
      await loadAll();
    } catch {
      alert(t('common.error_saving'));
    }
  };

  const handleDeleteClient = async (inboundId: number, clientId: string) => {
    try {
      await api.delete(`/inbounds/${inboundId}/clients/${clientId}`);
      await loadAll();
    } catch {
      alert(t('common.error_deleting'));
    }
  };


  const inboundOptions = useMemo(() => [
    { value: 'all', label: t('clients.all_nodes') },
    ...(inbounds || []).map(ib => ({ value: String(ib.id), label: `${ib.remark || 'Node'} (${(ib.protocol || 'vpn')}:${ib.port || 0})` })),
  ], [inbounds, t]);

  const visibleClients = useMemo(() => {
    const list = clients || [];
    return filterInbound === 'all'
      ? list
      : list.filter(c => String(c._inboundId) === filterInbound);
  }, [clients, filterInbound]);

  const PROTO_COLORS: Record<string, string> = {
    vless: '#818cf8', vmess: '#a78bfa', trojan: '#f59e0b',
    shadowsocks: '#10b981', amneziawg: '#06b6d4', openvpn: '#f97316',
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>

      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <div>
          <h1 style={{ fontSize: 22, fontWeight: 800, letterSpacing: '-0.4px' }}>{t('nav.clients')}</h1>
          <p style={{ fontSize: 13, color: 'var(--text-muted)', marginTop: 4 }}>{t('clients.subtitle')}</p>
        </div>
        <div style={{ display: 'flex', gap: 10, alignItems: 'center' }}>
          <button
            onClick={loadAll}
            title={t('clients.refresh')}
            style={{ padding: '8px 10px', borderRadius: 10, border: '1px solid var(--border)', background: 'transparent', color: 'var(--text-muted)', cursor: 'pointer', display: 'flex', alignItems: 'center' }}
          >
            <RefreshCw size={15} />
          </button>
        </div>
      </div>

      {/* Stats row */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 16 }}>
        {[
          { label: t('clients.total_clients'), value: clients.length, icon: Users, color: '#818cf8' },
          { label: t('nav.nodes'), value: inbounds.length, icon: Shield, color: 'var(--accent)' },
          { label: t('dashboard.active_nodes'), value: inbounds.filter(i => i.enable).length, icon: Activity, color: '#10b981' },
        ].map(({ label, value, icon: Icon, color }) => (
          <div key={label} style={{ background: 'var(--bg-card)', border: '1px solid var(--border)', borderRadius: 14, padding: '18px 20px', display: 'flex', alignItems: 'center', gap: 14 }}>
            <div style={{ width: 40, height: 40, borderRadius: 10, background: `${color}18`, display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
              <Icon size={20} color={color} />
            </div>
            <div>
              <div style={{ fontSize: 26, fontWeight: 800, letterSpacing: '-0.5px', lineHeight: 1 }}>{value}</div>
              <div style={{ fontSize: 12, color: 'var(--text-secondary)', marginTop: 4 }}>{label}</div>
            </div>
          </div>
        ))}
      </div>

      {/* Filter + per-node add buttons */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, flexWrap: 'wrap' }}>
        <div style={{ width: 280 }}>
          <Select
            label=""
            value={filterInbound}
            onChange={e => setFilterInbound(e.target.value)}
            options={inboundOptions}
          />
        </div>
        <div style={{ fontSize: 13, color: 'var(--text-muted)' }}>
          {t('clients.showing', { count: visibleClients.length })}
        </div>
      </div>

      {/* Nodes list with their clients */}
      {loading ? (
        <div style={{ padding: 48, textAlign: 'center', color: 'var(--text-muted)' }}>{t('common.loading')}…</div>
      ) : inbounds.length === 0 ? (
        <Card>
          <div style={{ padding: 48, textAlign: 'center' }}>
            <Shield size={40} style={{ color: 'var(--text-muted)', opacity: 0.2, margin: '0 auto 12px' }} />
            <p style={{ color: 'var(--text-muted)', fontSize: 14 }}>{t('clients.no_nodes')}</p>
          </div>
        </Card>
      ) : (
        inbounds
          .filter(ib => filterInbound === 'all' || String(ib.id) === filterInbound)
          .map(ib => {
            const proto = ib.protocol?.toLowerCase() || 'vless';
            const color = PROTO_COLORS[proto] || 'var(--accent)';
            let ibClients: any[] = [];
            try { ibClients = JSON.parse(ib.settings || '{}').clients || []; } catch { /* ignore */ }

            return (
              <Card key={ib.id} style={{ padding: 0, overflow: 'hidden' }}>
                {/* Node header */}
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '16px 20px', borderBottom: ibClients.length > 0 ? '1px solid var(--border)' : 'none', background: `linear-gradient(90deg, ${color}08, transparent)` }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                    <div style={{ width: 36, height: 36, borderRadius: 10, background: `${color}18`, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                      <Shield size={17} color={color} />
                    </div>
                    <div>
                      <div style={{ fontWeight: 700, fontSize: 14 }}>{ib.remark}</div>
                      <div style={{ display: 'flex', gap: 8, marginTop: 3, alignItems: 'center' }}>
                        <Badge variant="primary" style={{ background: `${color}20`, color, border: `1px solid ${color}30`, fontSize: 10, textTransform: 'uppercase' }}>
                          {ib.protocol}
                        </Badge>
                        <span style={{ fontSize: 12, color: 'var(--text-muted)', fontFamily: 'monospace' }}>:{ib.port}</span>
                        <span style={{ fontSize: 12, color: 'var(--text-muted)' }}>· {ibClients.length} {t('clients.client_count')}</span>
                      </div>
                    </div>
                  </div>
                  <Button onClick={() => handleAddClient(ib.id)} style={{ display: 'flex', alignItems: 'center', gap: 6, padding: '7px 14px', fontSize: 13 }}>
                    <Plus size={14} /> {t('common.add')}
                  </Button>
                </div>

                {/* Clients table */}
                {ibClients.length > 0 ? (
                  <div style={{ overflowX: 'auto' }}>
                    <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left' }}>
                      <thead>
                        <tr style={{ borderBottom: '1px solid var(--border)' }}>
                          {['Email / ID', t('common.traffic'), t('common.expiry'), 'Link', ''].map((h, i) => (
                            <th key={i} style={{ padding: '10px 18px', fontSize: 11, fontWeight: 700, color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.06em', whiteSpace: 'nowrap' }}>
                              {h}
                            </th>
                          ))}
                        </tr>
                      </thead>
                      <tbody>
                        {ibClients.map((client: any) => (
                          <tr
                            key={client.id || client.password}
                            style={{ borderBottom: '1px solid var(--border)', transition: 'background 0.15s' }}
                            onMouseEnter={e => (e.currentTarget.style.background = 'rgba(255,255,255,0.025)')}
                            onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}
                          >
                            <td style={{ padding: '12px 18px' }}>
                              <div style={{ fontWeight: 600, fontSize: 13 }}>{client.email || '—'}</div>
                              <div style={{ fontSize: 11, color: 'var(--text-muted)', fontFamily: 'monospace', marginTop: 2, maxWidth: 220, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                                {client.id || client.password || '—'}
                              </div>
                            </td>
                            <td style={{ padding: '12px 18px' }}>
                              <div style={{ display: 'flex', gap: 12, alignItems: 'center' }}>
                                <span style={{ display: 'flex', alignItems: 'center', gap: 4, fontSize: 12, color: 'var(--text-secondary)' }}>
                                  <ArrowUpRight size={11} color="#818cf8" />
                                  {formatBytes(client.up || 0)}
                                </span>
                                <span style={{ display: 'flex', alignItems: 'center', gap: 4, fontSize: 12, color: 'var(--text-secondary)' }}>
                                  <ArrowDownLeft size={11} color="#10b981" />
                                  {formatBytes(client.down || 0)}
                                </span>
                              </div>
                            </td>
                            <td style={{ padding: '12px 18px', fontSize: 12, color: 'var(--text-secondary)' }}>
                              {client.expiryTime > 0
                                ? new Date(client.expiryTime).toLocaleDateString()
                                : <span style={{ color: 'var(--text-muted)' }}>∞</span>
                              }
                            </td>
                            <td style={{ padding: '12px 18px', display: 'flex', gap: 6 }}>
                              {(proto.startsWith('amneziawg') || proto.includes('openvpn')) ? (
                                <DownloadBtn client={{ ...client, _inboundId: ib.id }} protocol={proto} />
                              ) : (
                                <>
                                  <CopyBtn client={{ ...client, _inboundId: ib.id }} />
                                  <button
                                    onClick={async () => {
                                      try {
                                        const res = await api.get(`/inbounds/${ib.id}/clients/${client.email || client.id}/config`);
                                        if (res.data.success) {
                                          openQr(res.data.obj, client.email || client.id || 'Config');
                                        }
                                      } catch { /* ignore */ }
                                    }}
                                    style={{ padding: '4px 8px', borderRadius: 6, border: '1px solid var(--border)', background: 'transparent', color: 'var(--text-muted)', cursor: 'pointer', display: 'inline-flex', alignItems: 'center', gap: 4, fontSize: 11 }}
                                    title="QR Code"
                                  >
                                    <QrCode size={13} />
                                  </button>
                                </>
                              )}
                            </td>
                            <td style={{ padding: '12px 14px', textAlign: 'right' }}>
                              <button
                                onClick={(e) => { e.stopPropagation(); handleDeleteClient(ib.id, client.id || client.password); }}
                                style={{ padding: '6px 10px', borderRadius: 8, border: 'none', background: 'transparent', color: 'var(--danger)', cursor: 'pointer' }}
                                title={t('common.delete')}
                              >
                                <Trash2 size={14} />
                              </button>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                ) : (
                  <div style={{ padding: '20px 20px', color: 'var(--text-muted)', fontSize: 13, textAlign: 'center' }}>
                    {t('clients.no_clients_in_node')} — <button onClick={() => handleAddClient(ib.id)} style={{ background: 'none', border: 'none', color: 'var(--accent)', cursor: 'pointer', fontWeight: 600, fontSize: 13 }}>{t('clients.add_first')}</button>
                  </div>
                )}
              </Card>
            );
          })
      )}

      {/* Add client modal */}
      <ClientModal
        isOpen={modalOpen}
        onClose={() => setModalOpen(false)}
        onSave={handleSaveClient}
        inboundProtocol={inbounds.find(i => i.id === selectedInboundId)?.protocol}
        inboundRemark={inbounds.find(i => i.id === selectedInboundId)?.remark}
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
