import React, { useState } from 'react';
import Modal from './Modal';
import { Input, Select, Switch, Button } from './UI';
import { Key } from 'lucide-react';

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
  { value: 'amneziawg-v1', label: 'AmneziaWG v1' },
  { value: 'amneziawg-v2', label: 'AmneziaWG v2' },
  { value: 'openvpn-xor', label: 'OpenVPN XOR' },
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
  });

  const [activeTab, setActiveTab] = useState('basic');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSave(formData);
  };

  const isVPN = formData.protocol.includes('amneziawg') || formData.protocol === 'openvpn-xor';

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={initialData ? 'Редактировать ноду' : 'Добавить новую ноду'}
      width={600}
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>Отмена</Button>
          <Button onClick={handleSubmit}>Сохранить</Button>
        </>
      }
    >
      <div style={{ display: 'flex', gap: 16, borderBottom: '1px solid var(--border)', marginBottom: 24 }}>
        {['basic', 'advanced'].map(tab => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            style={{
              padding: '12px 4px', background: 'none', border: 'none',
              color: activeTab === tab ? 'var(--accent)' : 'var(--text-muted)',
              borderBottom: activeTab === tab ? '2px solid var(--accent)' : '2px solid transparent',
              fontSize: 14, fontWeight: 600, cursor: 'pointer', transition: 'all 0.2s'
            }}
          >
            {tab === 'basic' ? 'Основные' : 'Дополнительно'}
          </button>
        ))}
      </div>

      <form onSubmit={handleSubmit}>
        {activeTab === 'basic' ? (
          <>
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

            {isVPN && (
              <div style={{ 
                padding: '16px', borderRadius: 12, background: 'rgba(99,102,241,0.05)', 
                border: '1px solid var(--border-glow)', marginBottom: 18 
              }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 10, color: 'var(--accent-light)', marginBottom: 8 }}>
                  <Key size={16} />
                  <span style={{ fontSize: 13, fontWeight: 700 }}>VPN Настройки</span>
                </div>
                <p style={{ fontSize: 12, color: 'var(--text-muted)' }}>
                  Настройки ключей и маршрутизации будут сгенерированы автоматически при добавлении клиентов.
                </p>
              </div>
            )}

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
          </>
        ) : (
          <>
            <Input 
              label="Адрес прослушивания" 
              value={formData.listen} 
              onChange={e => setFormData({...formData, listen: e.target.value})}
            />
            <Switch 
              label="Активировать при создании" 
              checked={formData.enable} 
              onChange={val => setFormData({...formData, enable: val})}
            />
            {!isVPN && (
              <div style={{ marginTop: 12 }}>
                <label style={{ fontSize: 13, fontWeight: 600, color: 'var(--text-secondary)', marginBottom: 8, display: 'block' }}>
                  Настройки JSON (Settings)
                </label>
                <textarea
                  style={{
                    width: '100%', height: 120, padding: 12,
                    background: 'var(--bg-elevated)', border: '1px solid var(--border)',
                    borderRadius: 12, color: 'var(--text-primary)', fontSize: 13,
                    fontFamily: 'JetBrains Mono, monospace', outline: 'none'
                  }}
                  placeholder='{"clients": [], "decryption": "none"}'
                />
              </div>
            )}
          </>
        )}
      </form>
    </Modal>
  );
}
