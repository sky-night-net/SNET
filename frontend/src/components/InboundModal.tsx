import React, { useState, useEffect } from 'react';
import Modal from './Modal';
import { Input, Select, Button } from './UI';
import { Shield, Globe, Zap, Copy, RefreshCw } from 'lucide-react';
import { api } from '../lib/api';
import { useTranslation } from 'react-i18next';

interface InboundModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
  inbound?: any;
}

const PROTOCOLS = [
  { value: 'vless',       label: 'VLESS' },
  { value: 'vmess',       label: 'VMess' },
  { value: 'trojan',      label: 'Trojan' },
  { value: 'shadowsocks', label: 'Shadowsocks' },
  { value: 'amneziawg',   label: 'AmneziaWG' },
  { value: 'openvpn',     label: 'OpenVPN XOR' },
];

const NETWORKS = [
  { value: 'tcp', label: 'TCP' },
  { value: 'ws',  label: 'WebSocket' },
  { value: 'grpc', label: 'gRPC' },
  { value: 'http', label: 'HTTP/2' },
];

const SECURITIES = [
  { value: 'none',    label: 'None' },
  { value: 'tls',     label: 'TLS' },
  { value: 'reality', label: 'REALITY' },
];

const SS_METHODS = [
  { value: 'aes-256-gcm',       label: 'aes-256-gcm' },
  { value: 'aes-128-gcm',       label: 'aes-128-gcm' },
  { value: 'chacha20-poly1305', label: 'chacha20-poly1305' },
  { value: 'none',              label: 'none (no encryption)' },
];

function generateUUID() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = Math.random() * 16 | 0;
    const v = c === 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
}

function generatePassword(len = 16) {
  const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz23456789!@#';
  return Array.from({ length: len }, () => chars[Math.floor(Math.random() * chars.length)]).join('');
}

function generateWgPrivKey() {
  const bytes = new Uint8Array(32);
  crypto.getRandomValues(bytes);
  bytes[0] &= 248; bytes[31] &= 127; bytes[31] |= 64;
  return btoa(String.fromCharCode(...bytes));
}

function generateShortId() {
  return Array.from({ length: 8 }, () => Math.floor(Math.random() * 16).toString(16)).join('');
}

/** Small helper: copy-to-clipboard button */
function CopyBtn({ value }: { value: string }) {
  const [copied, setCopied] = useState(false);
  const copy = () => {
    navigator.clipboard.writeText(value);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };
  return (
    <button type="button" onClick={copy} style={{
      position: 'absolute', right: 44, bottom: 8,
      background: 'transparent', border: 'none', color: copied ? 'var(--success)' : 'var(--text-muted)',
      cursor: 'pointer', padding: '4px',
    }}>
      <Copy size={14} />
    </button>
  );
}

/** Generate-button sitting inside an Input */
function GenBtn({ onClick }: { onClick: () => void }) {
  return (
    <button type="button" onClick={onClick} style={{
      position: 'absolute', right: 10, bottom: 8,
      background: 'var(--accent)', border: 'none', color: '#fff',
      borderRadius: 6, fontSize: 10, padding: '4px 8px', cursor: 'pointer',
      display: 'flex', alignItems: 'center', gap: 4,
    }}>
      <RefreshCw size={10} /> Gen
    </button>
  );
}

/** Section card wrapper */
function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div style={{
      padding: 16, borderRadius: 12,
      background: 'rgba(255,255,255,0.02)',
      border: '1px solid var(--border)',
    }}>
      <div style={{ fontSize: 11, fontWeight: 700, color: 'var(--text-muted)', marginBottom: 12, textTransform: 'uppercase', letterSpacing: '0.08em' }}>
        {title}
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        {children}
      </div>
    </div>
  );
}

function grid2(children: React.ReactNode) {
  return <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>{children}</div>;
}

export default function InboundModal({ isOpen, onClose, onSuccess, inbound }: InboundModalProps) {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState('basic');

  const [form, setForm] = useState<any>({
    remark: '', protocol: 'vless', port: 443,
    listen: '0.0.0.0', enable: true, total: 0, expiryTime: 0,
    sniffing: '{"enabled":true,"destOverride":["http","tls"]}',
  });

  const [settings, setSettings] = useState<any>({
    clients: [{ id: generateUUID(), email: 'user@snet', flow: '' }],
    decryption: 'none', fallbacks: [],
  });

  const [stream, setStream] = useState<any>({
    network: 'tcp', security: 'none',
    tlsSettings:     { serverName: '', minVersion: '1.2', alpn: ['h2', 'http/1.1'] },
    realitySettings: { dest: 'google.com:443', serverNames: ['google.com'], privateKey: '', shortIds: [generateShortId()] },
    wsSettings:      { path: '/', headers: {} },
    grpcSettings:    { serviceName: 'grpc' },
    httpSettings:    { path: '/', host: [] },
  });

  // AmneziaWG specific
  const [awg, setAwg] = useState<any>({
    privateKey: '', publicKey: '', presharedKey: '', address: '10.8.0.1/24',
    dns: '1.1.1.1', listenPort: 51820,
    jc: 4, jmin: 50, jmax: 1000, s1: 0, s2: 0, h1: 1, h2: 2, h3: 3, h4: 4,
    peers: [],
  });

  // OpenVPN XOR specific
  const [ovpn, setOvpn] = useState<any>({
    port: 1194, proto: 'udp', cipher: 'AES-256-GCM', auth: 'SHA256',
    xorKey: generatePassword(8), scramble: 'obfuscate',
    server: '10.9.0.0', netmask: '255.255.255.0',
    keepalive: '10 120', compress: 'lz4-v2', dns: '1.1.1.1 8.8.8.8',
  });

  useEffect(() => {
    if (!isOpen) return;
    if (inbound) {
      setForm({
        id: inbound.id, remark: inbound.remark, protocol: inbound.protocol,
        port: inbound.port, listen: inbound.listen, enable: inbound.enable,
        total: inbound.total, expiryTime: inbound.expiryTime, sniffing: inbound.sniffing,
      });
      try {
        if (inbound.settings)       setSettings(JSON.parse(inbound.settings));
        if (inbound.streamSettings) setStream(JSON.parse(inbound.streamSettings));
      } catch { /* ignore */ }
    } else {
      const randomPort = Math.floor(Math.random() * (55000 - 10000) + 10000);
      setForm({ remark: '', protocol: 'vless', port: randomPort, listen: '0.0.0.0', enable: true, total: 0, expiryTime: 0, sniffing: '{"enabled":true,"destOverride":["http","tls"]}' });
      applyProtocolDefaults('vless');
    }
    setActiveTab('basic');
  }, [inbound, isOpen]);

  function applyProtocolDefaults(p: string) {
    if (p === 'shadowsocks') {
      setSettings({ method: 'aes-256-gcm', password: generatePassword(16), network: 'tcp,udp' });
    } else if (p === 'trojan') {
      setSettings({ clients: [{ password: generatePassword(16), email: 'user@snet' }], fallbacks: [] });
    } else if (p === 'amneziawg' || p === 'openvpn') {
      // no stream settings for WG/OVPN
    } else {
      setSettings({ clients: [{ id: generateUUID(), email: 'user@snet', flow: '' }], decryption: 'none', fallbacks: [] });
    }
  }

  function handleProtocolChange(p: string) {
    setForm((prev: any) => ({ ...prev, protocol: p }));
    applyProtocolDefaults(p);
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    try {
      let payload: any = { ...form };
      if (form.protocol === 'amneziawg') {
        payload.settings = JSON.stringify(awg);
        payload.streamSettings = '{}';
      } else if (form.protocol === 'openvpn') {
        payload.settings = JSON.stringify(ovpn);
        payload.streamSettings = '{}';
      } else {
        payload.settings = JSON.stringify(settings);
        payload.streamSettings = JSON.stringify(stream);
      }
      if (inbound) {
        await api.put(`/inbounds/${inbound.id}`, payload);
      } else {
        await api.post('/inbounds', payload);
      }
      onSuccess();
      onClose();
    } catch {
      alert(t('common.error_saving'));
    }
  }

  const isWireGuard = form.protocol === 'amneziawg';
  const isOpenVPN   = form.protocol === 'openvpn';
  const isXray      = !isWireGuard && !isOpenVPN;

  const tabs = [
    { id: 'basic',     label: t('inbound.tab_basic'),     icon: Zap },
    ...(isXray ? [
      { id: 'transport', label: t('inbound.tab_transport'), icon: Globe },
      { id: 'security',  label: t('inbound.tab_security'),  icon: Shield },
    ] : []),
    ...(isWireGuard ? [
      { id: 'awg', label: 'AmneziaWG', icon: Shield },
    ] : []),
    ...(isOpenVPN ? [
      { id: 'ovpn', label: 'OpenVPN', icon: Globe },
    ] : []),
  ];

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={inbound ? t('common.edit') : t('common.add')}
      width={700}
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>{t('common.cancel')}</Button>
          <Button onClick={handleSubmit}>{t('common.save')}</Button>
        </>
      }
    >
      {/* Tabs */}
      <div style={{ display: 'flex', gap: 4, borderBottom: '1px solid var(--border)', marginBottom: 20 }}>
        {tabs.map(tab => (
          <button
            key={tab.id}
            type="button"
            onClick={() => setActiveTab(tab.id)}
            style={{
              display: 'flex', alignItems: 'center', gap: 6,
              padding: '10px 14px', background: 'none', border: 'none',
              color: activeTab === tab.id ? 'var(--accent)' : 'var(--text-muted)',
              borderBottom: activeTab === tab.id ? '2px solid var(--accent)' : '2px solid transparent',
              marginBottom: -1, fontSize: 13, fontWeight: 600, cursor: 'pointer', transition: 'color 0.2s',
            }}
          >
            <tab.icon size={14} />
            {tab.label}
          </button>
        ))}
      </div>

      <form onSubmit={handleSubmit} style={{ minHeight: 380, display: 'flex', flexDirection: 'column', gap: 16 }}>

        {/* ============ BASIC TAB ============ */}
        {activeTab === 'basic' && (
          <>
            {grid2(<>
              <Input
                label={t('common.remark')}
                value={form.remark}
                onChange={e => setForm({ ...form, remark: e.target.value })}
                placeholder="My Node"
              />
              <Select
                label={t('common.protocol')}
                value={form.protocol}
                onChange={e => handleProtocolChange(e.target.value)}
                options={PROTOCOLS}
              />
            </>)}

            {grid2(<>
              <Input
                label={t('common.port')}
                type="number"
                value={form.port}
                onChange={e => setForm({ ...form, port: parseInt(e.target.value) })}
              />
              <Input
                label={t('common.listen')}
                value={form.listen}
                onChange={e => setForm({ ...form, listen: e.target.value })}
                placeholder="0.0.0.0"
              />
            </>)}

            {/* Protocol-specific basic settings */}
            {form.protocol === 'shadowsocks' && (
              <Section title="Shadowsocks">
                {grid2(<>
                  <Select
                    label="Encryption Method"
                    value={settings.method}
                    onChange={e => setSettings({ ...settings, method: e.target.value })}
                    options={SS_METHODS}
                  />
                  <div style={{ position: 'relative' }}>
                    <Input
                      label="Password"
                      value={settings.password}
                      onChange={e => setSettings({ ...settings, password: e.target.value })}
                    />
                    <CopyBtn value={settings.password} />
                    <GenBtn onClick={() => setSettings({ ...settings, password: generatePassword(16) })} />
                  </div>
                </>)}
              </Section>
            )}

            {form.protocol === 'trojan' && (
              <Section title="Trojan">
                <div style={{ position: 'relative' }}>
                  <Input
                    label="Password"
                    value={settings.clients?.[0]?.password || ''}
                    onChange={e => {
                      const cls = [...(settings.clients || [{}])];
                      cls[0] = { ...cls[0], password: e.target.value };
                      setSettings({ ...settings, clients: cls });
                    }}
                  />
                  <CopyBtn value={settings.clients?.[0]?.password || ''} />
                  <GenBtn onClick={() => {
                    const cls = [...(settings.clients || [{}])];
                    cls[0] = { ...cls[0], password: generatePassword(16) };
                    setSettings({ ...settings, clients: cls });
                  }} />
                </div>
                <Input
                  label="Email (label)"
                  value={settings.clients?.[0]?.email || ''}
                  onChange={e => {
                    const cls = [...(settings.clients || [{}])];
                    cls[0] = { ...cls[0], email: e.target.value };
                    setSettings({ ...settings, clients: cls });
                  }}
                />
              </Section>
            )}

            {(form.protocol === 'vless' || form.protocol === 'vmess') && (
              <Section title="Client">
                <div style={{ position: 'relative' }}>
                  <Input
                    label="UUID"
                    value={settings.clients?.[0]?.id || ''}
                    onChange={e => {
                      const cls = [...(settings.clients || [{}])];
                      cls[0] = { ...cls[0], id: e.target.value };
                      setSettings({ ...settings, clients: cls });
                    }}
                  />
                  <CopyBtn value={settings.clients?.[0]?.id || ''} />
                  <GenBtn onClick={() => {
                    const cls = [...(settings.clients || [{}])];
                    cls[0] = { ...cls[0], id: generateUUID() };
                    setSettings({ ...settings, clients: cls });
                  }} />
                </div>
                {grid2(<>
                  <Input
                    label="Email (label)"
                    value={settings.clients?.[0]?.email || ''}
                    onChange={e => {
                      const cls = [...(settings.clients || [{}])];
                      cls[0] = { ...cls[0], email: e.target.value };
                      setSettings({ ...settings, clients: cls });
                    }}
                  />
                  {form.protocol === 'vless' && (
                    <Select
                      label="Flow"
                      value={settings.clients?.[0]?.flow || ''}
                      onChange={e => {
                        const cls = [...(settings.clients || [{}])];
                        cls[0] = { ...cls[0], flow: e.target.value };
                        setSettings({ ...settings, clients: cls });
                      }}
                      options={[
                        { value: '', label: 'None' },
                        { value: 'xtls-rprx-vision', label: 'xtls-rprx-vision' },
                      ]}
                    />
                  )}
                </>)}
              </Section>
            )}

            {/* Traffic & expiry */}
            {grid2(<>
              <Input
                label={t('common.traffic') + ' (GB, 0=∞)'}
                type="number"
                value={form.total > 0 ? +(form.total / 1024 ** 3).toFixed(2) : 0}
                onChange={e => setForm({ ...form, total: parseFloat(e.target.value) * 1024 ** 3 })}
              />
              <Input
                label={t('common.expiry') + ' (days, 0=∞)'}
                type="number"
                value={form.expiryTime > 0 ? Math.round((form.expiryTime - Date.now()) / 86400000) : 0}
                onChange={e => setForm({ ...form, expiryTime: Date.now() + (parseInt(e.target.value) || 0) * 86400000 })}
              />
            </>)}
          </>
        )}

        {/* ============ TRANSPORT TAB (Xray) ============ */}
        {activeTab === 'transport' && isXray && (
          <>
            <Select
              label="Network / Transport"
              value={stream.network}
              onChange={e => setStream({ ...stream, network: e.target.value })}
              options={NETWORKS}
            />

            {stream.network === 'ws' && (
              <Section title="WebSocket Settings">
                <Input
                  label="Path"
                  value={stream.wsSettings?.path || '/'}
                  onChange={e => setStream({ ...stream, wsSettings: { ...stream.wsSettings, path: e.target.value } })}
                  placeholder="/ws"
                />
                <Input
                  label="Host Header"
                  value={stream.wsSettings?.headers?.Host || ''}
                  onChange={e => setStream({ ...stream, wsSettings: { ...stream.wsSettings, headers: { Host: e.target.value } } })}
                  placeholder="example.com"
                />
              </Section>
            )}

            {stream.network === 'grpc' && (
              <Section title="gRPC Settings">
                <Input
                  label="Service Name"
                  value={stream.grpcSettings?.serviceName || 'grpc'}
                  onChange={e => setStream({ ...stream, grpcSettings: { ...stream.grpcSettings, serviceName: e.target.value } })}
                />
              </Section>
            )}

            {stream.network === 'http' && (
              <Section title="HTTP/2 Settings">
                <Input
                  label="Path"
                  value={stream.httpSettings?.path || '/'}
                  onChange={e => setStream({ ...stream, httpSettings: { ...stream.httpSettings, path: e.target.value } })}
                />
              </Section>
            )}
          </>
        )}

        {/* ============ SECURITY TAB (Xray) ============ */}
        {activeTab === 'security' && isXray && (
          <>
            <Select
              label="Security Layer"
              value={stream.security}
              onChange={e => setStream({ ...stream, security: e.target.value })}
              options={SECURITIES}
            />

            {stream.security === 'tls' && (
              <Section title="TLS Settings">
                {grid2(<>
                  <Input
                    label="Server Name (SNI)"
                    value={stream.tlsSettings?.serverName || ''}
                    onChange={e => setStream({ ...stream, tlsSettings: { ...stream.tlsSettings, serverName: e.target.value } })}
                    placeholder="example.com"
                  />
                  <Select
                    label="Min TLS Version"
                    value={stream.tlsSettings?.minVersion || '1.2'}
                    onChange={e => setStream({ ...stream, tlsSettings: { ...stream.tlsSettings, minVersion: e.target.value } })}
                    options={[
                      { value: '1.1', label: 'TLS 1.1' },
                      { value: '1.2', label: 'TLS 1.2' },
                      { value: '1.3', label: 'TLS 1.3' },
                    ]}
                  />
                </>)}
                <Input
                  label="ALPN (comma-separated)"
                  value={(stream.tlsSettings?.alpn || []).join(',')}
                  onChange={e => setStream({ ...stream, tlsSettings: { ...stream.tlsSettings, alpn: e.target.value.split(',').map((s: string) => s.trim()) } })}
                  placeholder="h2,http/1.1"
                />
              </Section>
            )}

            {stream.security === 'reality' && (
              <Section title="REALITY Settings">
                {grid2(<>
                  <Input
                    label="Dest (target)"
                    value={stream.realitySettings?.dest || 'google.com:443'}
                    onChange={e => setStream({ ...stream, realitySettings: { ...stream.realitySettings, dest: e.target.value } })}
                  />
                  <Input
                    label="Server Names (comma)"
                    value={(stream.realitySettings?.serverNames || []).join(',')}
                    onChange={e => setStream({ ...stream, realitySettings: { ...stream.realitySettings, serverNames: e.target.value.split(',') } })}
                  />
                </>)}
                <div style={{ position: 'relative' }}>
                  <Input
                    label="Private Key"
                    value={stream.realitySettings?.privateKey || ''}
                    onChange={e => setStream({ ...stream, realitySettings: { ...stream.realitySettings, privateKey: e.target.value } })}
                    placeholder="Auto-generate or paste here"
                  />
                  <CopyBtn value={stream.realitySettings?.privateKey || ''} />
                  <GenBtn onClick={() => setStream({ ...stream, realitySettings: { ...stream.realitySettings, privateKey: generateWgPrivKey() } })} />
                </div>
                <div style={{ position: 'relative' }}>
                  <Input
                    label="Short IDs (comma)"
                    value={(stream.realitySettings?.shortIds || []).join(',')}
                    onChange={e => setStream({ ...stream, realitySettings: { ...stream.realitySettings, shortIds: e.target.value.split(',') } })}
                    placeholder="e.g. abc123,def456"
                  />
                  <GenBtn onClick={() => setStream({ ...stream, realitySettings: { ...stream.realitySettings, shortIds: [generateShortId(), generateShortId()] } })} />
                </div>
              </Section>
            )}
          </>
        )}

        {/* ============ AMNEZIAWG TAB ============ */}
        {activeTab === 'awg' && isWireGuard && (
          <>
            <Section title="Interface">
              {grid2(<>
                <div style={{ position: 'relative' }}>
                  <Input
                    label="Private Key"
                    value={awg.privateKey}
                    onChange={e => setAwg({ ...awg, privateKey: e.target.value })}
                    placeholder="Auto-generate"
                  />
                  <CopyBtn value={awg.privateKey} />
                  <GenBtn onClick={() => setAwg({ ...awg, privateKey: generateWgPrivKey() })} />
                </div>
                <Input
                  label="Listen Port"
                  type="number"
                  value={awg.listenPort}
                  onChange={e => setAwg({ ...awg, listenPort: parseInt(e.target.value) })}
                />
              </>)}
              {grid2(<>
                <Input
                  label="Address"
                  value={awg.address}
                  onChange={e => setAwg({ ...awg, address: e.target.value })}
                  placeholder="10.8.0.1/24"
                />
                <Input
                  label="DNS"
                  value={awg.dns}
                  onChange={e => setAwg({ ...awg, dns: e.target.value })}
                  placeholder="1.1.1.1"
                />
              </>)}
            </Section>

            <Section title="AmneziaWG Obfuscation Parameters">
              <div style={{ padding: '8px 12px', background: 'rgba(99,102,241,0.05)', borderRadius: 8, fontSize: 12, color: 'var(--text-muted)', lineHeight: 1.6 }}>
                ℹ️ Параметры Jc, Jmin, Jmax добавляют «мусорные» пакеты для обхода DPI. S1/S2 — смещения заголовков. H1-H4 — инициирующие пакеты.
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 12 }}>
                <Input label="Jc (garbage count)" type="number" value={awg.jc} onChange={e => setAwg({ ...awg, jc: parseInt(e.target.value) })} />
                <Input label="Jmin (ms)" type="number" value={awg.jmin} onChange={e => setAwg({ ...awg, jmin: parseInt(e.target.value) })} />
                <Input label="Jmax (ms)" type="number" value={awg.jmax} onChange={e => setAwg({ ...awg, jmax: parseInt(e.target.value) })} />
                <Input label="S1 (init pkt size)" type="number" value={awg.s1} onChange={e => setAwg({ ...awg, s1: parseInt(e.target.value) })} />
                <Input label="S2 (resp pkt size)" type="number" value={awg.s2} onChange={e => setAwg({ ...awg, s2: parseInt(e.target.value) })} />
              </div>
            </Section>
          </>
        )}

        {/* ============ OPENVPN TAB ============ */}
        {activeTab === 'ovpn' && isOpenVPN && (
          <>
            <Section title="OpenVPN Settings">
              {grid2(<>
                <Input label="Port" type="number" value={ovpn.port} onChange={e => setOvpn({ ...ovpn, port: parseInt(e.target.value) })} />
                <Select
                  label="Protocol"
                  value={ovpn.proto}
                  onChange={e => setOvpn({ ...ovpn, proto: e.target.value })}
                  options={[{ value: 'udp', label: 'UDP' }, { value: 'tcp', label: 'TCP' }]}
                />
              </>)}
              {grid2(<>
                <Select
                  label="Cipher"
                  value={ovpn.cipher}
                  onChange={e => setOvpn({ ...ovpn, cipher: e.target.value })}
                  options={[
                    { value: 'AES-256-GCM', label: 'AES-256-GCM' },
                    { value: 'AES-128-GCM', label: 'AES-128-GCM' },
                    { value: 'CHACHA20-POLY1305', label: 'CHACHA20-POLY1305' },
                  ]}
                />
                <Select
                  label="Auth Hash"
                  value={ovpn.auth}
                  onChange={e => setOvpn({ ...ovpn, auth: e.target.value })}
                  options={[
                    { value: 'SHA256', label: 'SHA256' },
                    { value: 'SHA512', label: 'SHA512' },
                  ]}
                />
              </>)}
              {grid2(<>
                <Input label="Server Network" value={ovpn.server} onChange={e => setOvpn({ ...ovpn, server: e.target.value })} placeholder="10.9.0.0" />
                <Input label="Netmask" value={ovpn.netmask} onChange={e => setOvpn({ ...ovpn, netmask: e.target.value })} placeholder="255.255.255.0" />
              </>)}
            </Section>

            <Section title="XOR Scramble (Traffic Obfuscation)">
              <div style={{ padding: '8px 12px', background: 'rgba(245,158,11,0.05)', borderRadius: 8, fontSize: 12, color: 'var(--text-muted)' }}>
                ℹ️ XOR Scramble шифрует заголовки OpenVPN для обхода DPI. Ключ должен совпадать на клиенте и сервере.
              </div>
              {grid2(<>
                <Select
                  label="Scramble Mode"
                  value={ovpn.scramble}
                  onChange={e => setOvpn({ ...ovpn, scramble: e.target.value })}
                  options={[
                    { value: 'none',       label: 'None (standard)' },
                    { value: 'xormask',    label: 'xormask' },
                    { value: 'xorptrpos',  label: 'xorptrpos' },
                    { value: 'reverse',    label: 'reverse' },
                    { value: 'obfuscate',  label: 'obfuscate (recommended)' },
                  ]}
                />
                <div style={{ position: 'relative' }}>
                  <Input
                    label="XOR Key"
                    value={ovpn.xorKey}
                    onChange={e => setOvpn({ ...ovpn, xorKey: e.target.value })}
                    placeholder="Random key"
                  />
                  <CopyBtn value={ovpn.xorKey} />
                  <GenBtn onClick={() => setOvpn({ ...ovpn, xorKey: generatePassword(12) })} />
                </div>
              </>)}
            </Section>
          </>
        )}
      </form>
    </Modal>
  );
}
