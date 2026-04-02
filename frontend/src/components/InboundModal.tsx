import React, { useState, useEffect } from 'react';
import Modal from './Modal';
import { Input, Select, Button } from './UI';
import { Shield, Globe, Zap } from 'lucide-react';
import { api } from '../lib/api';
import { useTranslation } from 'react-i18next';

interface InboundModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
  inbound?: any;
}

const PROTOCOLS = [
  { value: 'vless', label: 'VLESS' },
  { value: 'vmess', label: 'VMess' },
  { value: 'trojan', label: 'Trojan' },
  { value: 'shadowsocks', label: 'Shadowsocks' },
];

const NETWORKS = [
  { value: 'tcp', label: 'TCP' },
  { value: 'ws', label: 'WebSocket' },
  { value: 'grpc', label: 'gRPC' },
];

const SECURITIES = [
  { value: 'none', label: 'None' },
  { value: 'tls', label: 'TLS' },
  { value: 'reality', label: 'REALITY' },
];

const SS_METHODS = [
  { value: 'aes-256-gcm', label: 'aes-256-gcm' },
  { value: 'aes-128-gcm', label: 'aes-128-gcm' },
  { value: 'chacha20-poly1305', label: 'chacha20-poly1305' },
  { value: 'none', label: 'none' },
];

function generateUUID() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    var r = Math.random() * 16 | 0, v = c === 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
}

export default function InboundModal({ isOpen, onClose, onSuccess, inbound }: InboundModalProps) {
  const { t } = useTranslation();
  const [formData, setFormData] = useState<any>({
    remark: '',
    protocol: 'vless',
    port: 443,
    listen: '0.0.0.0',
    enable: true,
    total: 0,
    expiryTime: 0,
    sniffing: '{"enabled": true, "destOverride": ["http", "tls"]}',
  });

  const [activeTab, setActiveTab] = useState('basic');
  
  const [settingsData, setSettingsData] = useState<any>({
    clients: [{ id: generateUUID(), email: 'admin@snet', flow: '' }],
    decryption: 'none',
    fallbacks: []
  });

  const [streamData, setStreamData] = useState<any>({
    network: 'tcp',
    security: 'none',
    tlsSettings: { serverName: '', certificates: [] },
    realitySettings: { show: false, dest: 'google.com:443', serverNames: ['google.com'], privateKey: '', shortIds: [''] },
    wsSettings: { path: '/', headers: {} },
    grpcSettings: { serviceName: '' }
  });

  useEffect(() => {
    if (inbound) {
      setFormData({
        id: inbound.id,
        remark: inbound.remark,
        protocol: inbound.protocol,
        port: inbound.port,
        listen: inbound.listen,
        enable: inbound.enable,
        total: inbound.total,
        expiryTime: inbound.expiryTime,
        sniffing: inbound.sniffing,
      });
      try {
        if (inbound.settings) setSettingsData(JSON.parse(inbound.settings));
        if (inbound.streamSettings) setStreamData(JSON.parse(inbound.streamSettings));
      } catch (e) {
        console.error('Failed to parse inbound settings');
      }
    } else {
      setFormData({
        remark: '',
        protocol: 'vless',
        port: Math.floor(Math.random() * (65000 - 10000) + 10000),
        listen: '0.0.0.0',
        enable: true,
        total: 0,
        expiryTime: 0,
        sniffing: '{"enabled": true, "destOverride": ["http", "tls"]}',
      });
      handleProtocolChange('vless');
    }
    setActiveTab('basic');
  }, [inbound, isOpen]);

  const handleProtocolChange = (p: string) => {
    setFormData((prev: any) => ({ ...prev, protocol: p }));
    if (p === 'shadowsocks') {
      setSettingsData({ method: 'aes-256-gcm', password: Math.random().toString(36).slice(-8), network: 'tcp,udp' });
    } else if (p === 'trojan') {
      setSettingsData({ clients: [{ password: Math.random().toString(36).slice(-8), email: 'admin@snet' }] });
    } else {
      setSettingsData({ clients: [{ id: generateUUID(), email: 'admin@snet', flow: '' }], decryption: 'none' });
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const payload = {
        ...formData,
        settings: JSON.stringify(settingsData),
        streamSettings: JSON.stringify(streamData)
      };
      
      if (inbound) {
        await api.put(`/inbounds/${inbound.id}`, payload);
      } else {
        await api.post('/inbounds', payload);
      }
      onSuccess();
      onClose();
    } catch (e) {
      alert('Ошибка сохранения');
    }
  };

  const tabs = [
    { id: 'basic', label: 'Zap', icon: Zap, key: 'dashboard.system_status' },
    { id: 'transport', label: 'Globe', icon: Globe, key: 'dashboard.network_speed' },
    { id: 'security', label: 'Shield', icon: Shield, key: 'firewall.title' },
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
      <div style={{ display: 'flex', gap: 16, borderBottom: '1px solid var(--border)', marginBottom: 24 }}>
        {tabs.map(tab => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            style={{
              display: 'flex', alignItems: 'center', gap: 8,
              padding: '12px 16px', background: 'none', border: 'none',
              color: activeTab === tab.id ? 'var(--accent)' : 'var(--text-muted)',
              borderBottom: activeTab === tab.id ? '2px solid var(--accent)' : '2px solid transparent',
              fontSize: 14, fontWeight: 600, cursor: 'pointer', transition: 'all 0.2s'
            }}
          >
            <tab.icon size={16} />
            {t(tab.key)}
          </button>
        ))}
      </div>

      <form onSubmit={handleSubmit} style={{ minHeight: 400 }}>
        {activeTab === 'basic' && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <Input 
              label={t('common.remark')}
              value={formData.remark} 
              onChange={e => setFormData({...formData, remark: e.target.value})}
              placeholder="My Node"
            />
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
              <Select 
                label={t('common.protocol')} 
                value={formData.protocol} 
                onChange={e => handleProtocolChange(e.target.value)}
                options={PROTOCOLS}
              />
              <Input 
                label={t('common.port')} 
                type="number"
                value={formData.port} 
                onChange={e => setFormData({...formData, port: parseInt(e.target.value)})}
              />
            </div>

            <div style={{ padding: 16, borderRadius: 12, background: 'rgba(255,255,255,0.02)', border: '1px solid var(--border)' }}>
              <div style={{ fontSize: 12, fontWeight: 700, color: 'var(--text-muted)', marginBottom: 12, textTransform: 'uppercase' }}>Settings</div>
              
              {formData.protocol === 'shadowsocks' ? (
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
                  <Select 
                    label="Method" 
                    value={settingsData.method}
                    onChange={e => setSettingsData({...settingsData, method: e.target.value})}
                    options={SS_METHODS}
                  />
                  <Input 
                    label="Password" 
                    value={settingsData.password}
                    onChange={e => setSettingsData({...settingsData, password: e.target.value})}
                  />
                </div>
              ) : formData.protocol === 'trojan' ? (
                <Input 
                  label="Password" 
                  value={settingsData.clients?.[0]?.password}
                  onChange={e => {
                    const cls = [...settingsData.clients];
                    cls[0].password = e.target.value;
                    setSettingsData({...settingsData, clients: cls});
                  }}
                />
              ) : (
                <div style={{ position: 'relative' }}>
                  <Input 
                    label="UUID (Client ID)" 
                    value={settingsData.clients?.[0]?.id}
                    onChange={e => {
                      const cls = [...settingsData.clients];
                      cls[0].id = e.target.value;
                      setSettingsData({...settingsData, clients: cls});
                    }}
                  />
                  <button 
                    type="button" 
                    onClick={() => {
                        const cls = [...settingsData.clients];
                        cls[0].id = generateUUID();
                        setSettingsData({...settingsData, clients: cls});
                    }}
                    style={{ position: 'absolute', right: 12, bottom: 8, background: 'var(--accent)', border: 'none', color: 'white', borderRadius: 6, fontSize: 10, padding: '4px 8px', cursor: 'pointer' }}
                  >
                    Generate
                  </button>
                </div>
              )}
            </div>
            
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
              <Input 
                label={t('common.traffic') + " (GB)"}
                type="number"
                value={formData.total > 0 ? (formData.total / (1024**3)).toFixed(0) : 0} 
                onChange={e => setFormData({...formData, total: parseInt(e.target.value) * (1024**3)})}
              />
              <Input 
                label={t('common.expiry')} 
                type="number"
                value={formData.expiryTime > 0 ? Math.round((formData.expiryTime - Date.now()) / (1000*60*60*24)) : 0} 
                onChange={e => setFormData({...formData, expiryTime: Date.now() + (parseInt(e.target.value) || 0) * (1000*60*60*24)})}
              />
            </div>
          </div>
        )}

        {activeTab === 'transport' && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <Select 
              label="Network" 
              value={streamData.network} 
              onChange={e => setStreamData({...streamData, network: e.target.value})}
              options={NETWORKS}
            />
            <Input 
              label={t('common.listen')} 
              value={formData.listen} 
              onChange={e => setFormData({...formData, listen: e.target.value})}
            />
          </div>
        )}

        {activeTab === 'security' && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <Select 
              label="Security" 
              value={streamData.security} 
              onChange={e => setStreamData({...streamData, security: e.target.value})}
              options={SECURITIES}
            />

            {streamData.security === 'reality' && (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 12, padding: 16, border: '1px solid var(--border)', borderRadius: 12, background: 'rgba(99,102,241,0.03)' }}>
                <Input 
                  label="Dest" 
                  value={streamData.realitySettings?.dest || 'google.com:443'} 
                  onChange={e => setStreamData({
                    ...streamData, 
                    realitySettings: { ...streamData.realitySettings, dest: e.target.value }
                  })}
                />
                <Input 
                  label="Server Names" 
                  value={streamData.realitySettings?.serverNames?.join(',') || 'google.com'} 
                  onChange={e => setStreamData({
                    ...streamData, 
                    realitySettings: { ...streamData.realitySettings, serverNames: e.target.value.split(',') }
                  })}
                />
                <div style={{ position: 'relative' }}>
                  <Input 
                    label="Private Key" 
                    value={streamData.realitySettings?.privateKey || ''} 
                    onChange={e => setStreamData({
                      ...streamData, 
                      realitySettings: { ...streamData.realitySettings, privateKey: e.target.value }
                    })}
                  />
                  <button 
                    type="button"
                    onClick={() => {
                      setStreamData({
                        ...streamData,
                        realitySettings: { ...streamData.realitySettings, privateKey: Math.random().toString(36).substring(2, 15) }
                      });
                    }}
                    style={{ position: 'absolute', right: 12, bottom: 8, background: 'var(--accent)', border: 'none', color: 'white', borderRadius: 6, fontSize: 10, padding: '4px 8px', cursor: 'pointer' }}
                  >
                    Generate
                  </button>
                </div>
              </div>
            )}
          </div>
        )}
      </form>
    </Modal>
  );
}
