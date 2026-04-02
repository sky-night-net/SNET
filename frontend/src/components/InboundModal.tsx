import React, { useState, useEffect } from 'react';
import Modal from './Modal';
import { Input, Select, Button } from './UI';
import { Shield, Globe, Zap } from 'lucide-react';

interface InboundModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: (data: any) => void;
  initialData?: any;
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

export default function InboundModal({ isOpen, onClose, onSave, initialData }: InboundModalProps) {
  const [formData, setFormData] = useState(initialData || {
    remark: '',
    protocol: 'vless',
    port: 443,
    listen: '0.0.0.0',
    enable: true,
    total: 0,
    expiryTime: 0,
    settings: '{"clients": [], "decryption": "none"}',
    streamSettings: '{"network": "tcp", "security": "none"}',
    sniffing: '{"enabled": true, "destOverride": ["http", "tls"]}',
  });

  const [activeTab, setActiveTab] = useState('basic');
  const [streamData, setStreamData] = useState<any>({
    network: 'tcp',
    security: 'none',
    tlsSettings: { serverName: '', certificates: [] },
    realitySettings: { show: false, dest: 'google.com:443', serverNames: ['google.com'], privateKey: '', shortIds: [''] },
    wsSettings: { path: '/', headers: {} },
    grpcSettings: { serviceName: '' }
  });

  useEffect(() => {
    if (initialData?.streamSettings) {
      try {
        setStreamData(JSON.parse(initialData.streamSettings));
      } catch (e) {
        console.error('Failed to parse streamSettings');
      }
    }
  }, [initialData]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const finalData = {
      ...formData,
      streamSettings: JSON.stringify(streamData)
    };
    onSave(finalData);
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={initialData ? 'Редактировать ноду' : 'Добавить новую ноду'}
      width={700}
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>Отмена</Button>
          <Button onClick={handleSubmit}>Сохранить</Button>
        </>
      }
    >
      <div style={{ display: 'flex', gap: 16, borderBottom: '1px solid var(--border)', marginBottom: 24 }}>
        {[
          { id: 'basic', label: 'Основные', icon: Zap },
          { id: 'transport', label: 'Транспорт', icon: Globe },
          { id: 'security', label: 'Безопасность', icon: Shield },
        ].map(tab => (
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
            {tab.label}
          </button>
        ))}
      </div>

      <form onSubmit={handleSubmit} style={{ minHeight: 320 }}>
        {activeTab === 'basic' && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <Input 
              label="Название (Remark)" 
              value={formData.remark} 
              onChange={e => setFormData({...formData, remark: e.target.value})}
              placeholder="e.g. My Secure Node"
            />
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
              <Select 
                label="Протокол" 
                value={formData.protocol} 
                onChange={e => setFormData({...formData, protocol: e.target.value})}
                options={PROTOCOLS}
              />
              <Input 
                label="Порт" 
                type="number"
                value={formData.port} 
                onChange={e => setFormData({...formData, port: parseInt(e.target.value)})}
              />
            </div>
            
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
              <Input 
                label="Лимит трафика (GB)" 
                type="number"
                value={formData.total > 0 ? (formData.total / (1024**3)).toFixed(0) : 0} 
                onChange={e => setFormData({...formData, total: parseInt(e.target.value) * (1024**3)})}
                placeholder="0 = без лимита"
              />
              <Input 
                label="Срок действия (дни)" 
                type="number"
                value={formData.expiryTime > 0 ? Math.round((formData.expiryTime - Date.now()) / (1000*60*60*24)) : 0} 
                onChange={e => setFormData({...formData, expiryTime: Date.now() + parseInt(e.target.value) * (1000*60*60*24)})}
                placeholder="0 = бессрочно"
              />
            </div>
          </div>
        )}

        {activeTab === 'transport' && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <Select 
              label="Тип сети (Transport)" 
              value={streamData.network} 
              onChange={e => setStreamData({...streamData, network: e.target.value})}
              options={NETWORKS}
            />

            {streamData.network === 'ws' && (
              <Input 
                label="WebSocket Path" 
                value={streamData.wsSettings?.path || '/'} 
                onChange={e => setStreamData({
                  ...streamData, 
                  wsSettings: { ...streamData.wsSettings, path: e.target.value }
                })}
              />
            )}

            {streamData.network === 'grpc' && (
              <Input 
                label="gRPC Service Name" 
                value={streamData.grpcSettings?.serviceName || ''} 
                onChange={e => setStreamData({
                  ...streamData, 
                  grpcSettings: { ...streamData.grpcSettings, serviceName: e.target.value }
                })}
              />
            )}

            <Input 
              label="Адрес прослушивания" 
              value={formData.listen} 
              onChange={e => setFormData({...formData, listen: e.target.value})}
            />
          </div>
        )}

        {activeTab === 'security' && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <Select 
              label="Безопасность (Security)" 
              value={streamData.security} 
              onChange={e => setStreamData({...streamData, security: e.target.value})}
              options={SECURITIES}
            />

            {streamData.security === 'tls' && (
              <div style={{ padding: 16, border: '1px solid var(--border)', borderRadius: 12, background: 'rgba(255,255,255,0.02)' }}>
                <Input 
                  label="Server Name (SNI)" 
                  value={streamData.tlsSettings?.serverName || ''} 
                  onChange={e => setStreamData({
                    ...streamData, 
                    tlsSettings: { ...streamData.tlsSettings, serverName: e.target.value }
                  })}
                />
                <p style={{ fontSize: 11, color: 'var(--text-muted)', marginTop: 8 }}>
                  Примечание: сертификаты должны быть установлены в системе.
                </p>
              </div>
            )}

            {streamData.security === 'reality' && (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 12, padding: 16, border: '1px solid var(--border)', borderRadius: 12, background: 'rgba(99,102,241,0.03)' }}>
                <Input 
                  label="Dest (Target Domain:Port)" 
                  value={streamData.realitySettings?.dest || 'google.com:443'} 
                  onChange={e => setStreamData({
                    ...streamData, 
                    realitySettings: { ...streamData.realitySettings, dest: e.target.value }
                  })}
                />
                <Input 
                  label="Server Names (Comma separated)" 
                  value={streamData.realitySettings?.serverNames?.join(',') || 'google.com'} 
                  onChange={e => setStreamData({
                    ...streamData, 
                    realitySettings: { ...streamData.realitySettings, serverNames: e.target.value.split(',') }
                  })}
                />
                <div style={{ position: 'relative' }}>
                  <Input 
                    label="Reality Private Key" 
                    value={streamData.realitySettings?.privateKey || ''} 
                    onChange={e => setStreamData({
                      ...streamData, 
                      realitySettings: { ...streamData.realitySettings, privateKey: e.target.value }
                    })}
                  />
                  <button 
                    type="button"
                    onClick={() => {
                      // Mock generator - would normally call API
                      setStreamData({
                        ...streamData,
                        realitySettings: { ...streamData.realitySettings, privateKey: Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15) }
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

