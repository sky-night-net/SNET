import React, { useState, useEffect } from 'react';
import { Shield, Plus, Trash2, CheckCircle2, XCircle } from 'lucide-react';
import { api } from '../lib/api';
import { Card, Button, Input, Select, Badge } from '../components/UI';
import Modal from '../components/Modal';
import { useTranslation } from 'react-i18next';

interface FirewallRule {
  id: number;
  action: 'allow' | 'deny';
  port: number;
  ip: string;
  protocol: 'tcp' | 'udp' | 'both';
  remark: string;
  enable: boolean;
}

export default function FirewallPage() {
  const { t } = useTranslation();
  const [rules, setRules] = useState<FirewallRule[]>([]);
  const [loading, setLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [newRule, setNewRule] = useState({
    action: 'allow',
    port: 0,
    ip: '0.0.0.0/0',
    protocol: 'tcp',
    remark: '',
    enable: true
  });

  const fetchRules = async () => {
    try {
      const res = await api.get('/firewall');
      if (res.data.success) {
        setRules(res.data.obj || []);
      }
    } catch (e) {
      console.error('Failed to fetch firewall rules');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchRules();
  }, []);

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await api.post('/firewall', newRule);
      setIsModalOpen(false);
      fetchRules();
      setNewRule({ action: 'allow', port: 0, ip: '0.0.0.0/0', protocol: 'tcp', remark: '', enable: true });
    } catch (e) {
      alert('Ошибка при сохранении правила');
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm('Вы уверены, что хотите удалить это правило?')) return;
    try {
      await api.delete(`/firewall/${id}`);
      fetchRules();
    } catch (e) {
      alert('Ошибка при удалении');
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold">{t('firewall.title')}</h1>
          <p className="text-text-muted text-sm">Управление правилами сетевого доступа сервера (iptables)</p>
        </div>
        <Button onClick={() => setIsModalOpen(true)} className="flex items-center gap-2">
          <Plus size={18} />
          {t('firewall.add_rule')}
        </Button>
      </div>

      <Card>
        {loading ? (
          <div className="p-12 text-center text-text-muted">Загрузка правил...</div>
        ) : rules.length === 0 ? (
          <div className="p-12 text-center">
            <Shield size={48} className="mx-auto mb-4 text-text-muted opacity-20" />
            <p className="text-text-muted">Нет активных правил. Сервер использует стандартные настройки.</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left">
              <thead>
                <tr className="border-b border-border text-sm text-text-muted">
                  <th className="px-6 py-4 font-medium">{t('common.status')}</th>
                  <th className="px-6 py-4 font-medium">{t('firewall.action')}</th>
                  <th className="px-6 py-4 font-medium">{t('common.port')}</th>
                  <th className="px-6 py-4 font-medium">{t('common.protocol')}</th>
                  <th className="px-6 py-4 font-medium">{t('firewall.source')}</th>
                  <th className="px-6 py-4 font-medium">{t('common.remark')}</th>
                  <th className="px-6 py-4" />
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {rules.map(rule => (
                  <tr key={rule.id} className="hover:bg-white/5 transition-colors">
                    <td className="px-6 py-4">
                      {rule.enable ? (
                        <Badge variant="success" className="flex items-center gap-1 w-fit">
                          <CheckCircle2 size={10} /> {t('firewall.active')}
                        </Badge>
                      ) : (
                        <Badge variant="secondary" className="flex items-center gap-1 w-fit">
                          <XCircle size={10} /> {t('firewall.disabled')}
                        </Badge>
                      )}
                    </td>
                    <td className="px-6 py-4">
                      <span className={rule.action === 'allow' ? 'text-green-500 font-medium' : 'text-red-500 font-medium'}>
                        {rule.action === 'allow' ? t('firewall.allow') : t('firewall.deny')}
                      </span>
                    </td>
                    <td className="px-6 py-4 font-mono text-sm">
                      {rule.port === 0 ? 'All' : rule.port}
                    </td>
                    <td className="px-6 py-4">
                      <span className="uppercase text-xs font-bold text-text-muted bg-white/5 px-2 py-1 rounded">
                        {rule.protocol}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-sm text-text-muted">
                      {rule.ip || 'Any'}
                    </td>
                    <td className="px-6 py-4 truncate max-w-[150px] text-sm">
                      {rule.remark || '-'}
                    </td>
                    <td className="px-6 py-4 text-right">
                      <button 
                        onClick={() => handleDelete(rule.id)}
                        className="p-2 hover:bg-red-500/10 text-red-500 rounded-lg transition-colors"
                      >
                        <Trash2 size={18} />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      <div className="bg-orange-500/10 border border-orange-500/20 rounded-xl p-4 flex gap-4">
        <div className="w-10 h-10 rounded-full bg-orange-500/20 flex items-center justify-center text-orange-500 shrink-0">
          <Shield size={20} />
        </div>
        <div>
          <h4 className="font-bold text-orange-500">Внимание</h4>
          <p className="text-sm text-orange-500/80">
            Будьте осторожны при настройке правил запрета (DENY). Ошибочное правило на порт 22 (SSH) или порт панели ({window.location.port}) может привести к потере доступа к серверу.
          </p>
        </div>
      </div>

      <Modal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        title="Новое правило Firewall"
        width={500}
        footer={
          <>
            <Button variant="secondary" onClick={() => setIsModalOpen(false)}>Отмена</Button>
            <Button onClick={handleSave}>Применить правило</Button>
          </>
        }
      >
        <form onSubmit={handleSave} className="space-y-4">
          <Select 
            label="Действие"
            value={newRule.action}
            onChange={e => setNewRule({...newRule, action: e.target.value})}
            options={[
              { value: 'allow', label: 'Разрешить (ALLOW)' },
              { value: 'deny', label: 'Запретить (DENY)' },
            ]}
          />
          <div className="grid grid-cols-2 gap-4">
            <Input 
              label="Порт (0 = все)"
              type="number"
              value={newRule.port}
              onChange={e => setNewRule({...newRule, port: parseInt(e.target.value)})}
            />
            <Select 
              label="Протокол"
              value={newRule.protocol}
              onChange={e => setNewRule({...newRule, protocol: e.target.value})}
              options={[
                { value: 'tcp', label: 'TCP' },
                { value: 'udp', label: 'UDP' },
                { value: 'both', label: 'TCP + UDP' },
              ]}
            />
          </div>
          <Input 
            label="Источник IP / CIDR"
            placeholder="0.0.0.0/0"
            value={newRule.ip}
            onChange={e => setNewRule({...newRule, ip: e.target.value})}
          />
          <Input 
            label="Примечание"
            placeholder="например, Доступ к SSH"
            value={newRule.remark}
            onChange={e => setNewRule({...newRule, remark: e.target.value})}
          />
        </form>
      </Modal>
    </div>
  );
}
