import React, { useState, useEffect } from 'react';
import { generateUUID } from '../lib/utils';
import Modal from './Modal';
import { Input, Button } from './UI';
import { RefreshCw } from 'lucide-react';

interface ClientModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: (client: any) => void;
  initialData?: any;
  inboundProtocol?: string;
  inboundRemark?: string;
}

export default function ClientModal({ isOpen, onClose, onSave, initialData, inboundProtocol, inboundRemark }: ClientModalProps) {
  const [formData, setFormData] = useState(initialData || {
    id: generateUUID(),
    email: '',
    limitIp: 0,
    totalGB: 0,
    expiryTime: 0,
    enable: true,
    flow: ''
  });

  useEffect(() => {
    if (isOpen) {
      setFormData(initialData || {
        id: generateUUID(),
        email: '',
        limitIp: 0,
        totalGB: 0,
        expiryTime: 0,
        enable: true,
        flow: ''
      });
    }
  }, [isOpen, initialData]);

  const modalTitle = initialData 
    ? `Редактировать клиента${inboundRemark ? ` (${inboundRemark})` : ''}` 
    : `Добавить клиента в ноду ${inboundRemark || ''}`;

  const handleGenerateId = () => {
    setFormData({ ...formData, id: generateUUID() });
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSave(formData);
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={modalTitle}
      width={500}
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>Отмена</Button>
          <Button onClick={handleSubmit}>Сохранить</Button>
        </>
      }
    >
      <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
        {inboundProtocol && (
          <div style={{ padding: '8px 12px', background: 'rgba(99,102,241,0.1)', border: '1px solid var(--border)', borderRadius: 10, fontSize: 12, color: 'var(--accent-light)', display: 'flex', alignItems: 'center', gap: 8 }}>
            <span style={{ fontWeight: 800 }}>{inboundProtocol.toUpperCase()}</span>
            <span>Конфигурация клиента для этого протокола</span>
          </div>
        )}
        <Input 
          label="Email (Имя пользователя)" 
          value={formData.email} 
          onChange={e => setFormData({...formData, email: e.target.value})}
          placeholder="user@example.com"
          required
        />
        
        <div style={{ position: 'relative' }}>
          <Input 
            label="ID / UUID" 
            value={formData.id} 
            onChange={e => setFormData({...formData, id: e.target.value})}
            required
          />
          <button 
            type="button"
            onClick={handleGenerateId}
            style={{ 
              position: 'absolute', right: 10, bottom: 8, 
              background: 'var(--bg-elevated)', border: '1px solid var(--border)',
              borderRadius: 6, padding: '4px 8px', cursor: 'pointer',
              display: 'flex', alignItems: 'center', gap: 4, fontSize: 11
            }}
          >
            <RefreshCw size={12} />
            Сгенерировать
          </button>
        </div>

        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
          <Input 
            label="Лимит IP" 
            type="number"
            value={formData.limitIp} 
            onChange={e => setFormData({...formData, limitIp: parseInt(e.target.value)})}
            placeholder="0 = без лимита"
          />
          <Input 
            label="Лимит трафика (GB)" 
            type="number"
            value={formData.totalGB} 
            onChange={e => setFormData({...formData, totalGB: parseInt(e.target.value)})}
            placeholder="0 = без лимита"
          />
        </div>
      </form>
    </Modal>
  );
}
