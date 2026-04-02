import React, { useState, useEffect } from 'react';
import { Shield, Plus, Trash2, CheckCircle2, XCircle, RefreshCw, AlertTriangle } from 'lucide-react';
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

const ACTION_OPTIONS  = [
  { value: 'allow', label: 'ALLOW — разрешить' },
  { value: 'deny',  label: 'DENY — запретить' },
];
const PROTO_OPTIONS = [
  { value: 'tcp',  label: 'TCP' },
  { value: 'udp',  label: 'UDP' },
  { value: 'both', label: 'TCP + UDP' },
];

const EMPTY_RULE = { action: 'allow', port: 0, ip: '0.0.0.0/0', protocol: 'tcp', remark: '', enable: true };

export default function FirewallPage() {
  const { t } = useTranslation();
  const [rules, setRules]         = useState<FirewallRule[]>([]);
  const [loading, setLoading]     = useState(true);
  const [syncing, setSyncing]     = useState(false);
  const [isModalOpen, setModal]   = useState(false);
  const [newRule, setNewRule]     = useState<any>(EMPTY_RULE);
  const [saveLoading, setSave]    = useState(false);

  const fetchRules = async () => {
    try {
      const res = await api.get('/firewall');
      if (res.data.success) setRules(res.data.obj || []);
    } catch { /* ignore */ } finally { setLoading(false); }
  };

  useEffect(() => { fetchRules(); }, []);

  const handleSync = async () => {
    setSyncing(true);
    try {
      await api.post('/firewall/sync', {});
    } catch { /* ignore */ } finally {
      setSyncing(false);
      fetchRules();
    }
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    setSave(true);
    try {
      await api.post('/firewall', newRule);
      setModal(false);
      setNewRule(EMPTY_RULE);
      fetchRules();
    } catch {
      alert(t('common.error_saving'));
    } finally {
      setSave(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm(t('common.delete') + '?')) return;
    try {
      await api.delete(`/firewall/${id}`);
      fetchRules();
    } catch {
      alert(t('common.error_deleting'));
    }
  };

  const openModal = () => { setNewRule(EMPTY_RULE); setModal(true); };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>

      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <div>
          <h1 style={{ fontSize: 22, fontWeight: 800, letterSpacing: '-0.4px' }}>{t('firewall.title')}</h1>
          <p style={{ fontSize: 13, color: 'var(--text-muted)', marginTop: 4 }}>{t('firewall.subtitle')}</p>
        </div>
        <div style={{ display: 'flex', gap: 10 }}>
          <Button
            variant="secondary"
            onClick={handleSync}
            style={{ display: 'flex', alignItems: 'center', gap: 8 }}
          >
            <RefreshCw size={15} style={{ animation: syncing ? 'spin 1s linear infinite' : 'none' }} />
            {t('firewall.sync')}
          </Button>
          <Button onClick={openModal} style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <Plus size={15} /> {t('firewall.add_rule')}
          </Button>
        </div>
      </div>

      {/* Warning banner */}
      <div style={{
        display: 'flex', gap: 14, padding: '14px 18px',
        background: 'rgba(245,158,11,0.07)', border: '1px solid rgba(245,158,11,0.2)',
        borderRadius: 14, alignItems: 'flex-start',
      }}>
        <AlertTriangle size={18} color="#f59e0b" style={{ flexShrink: 0, marginTop: 1 }} />
        <div>
          <div style={{ fontWeight: 700, color: '#f59e0b', fontSize: 13 }}>{t('firewall.warning_title')}</div>
          <div style={{ fontSize: 12, color: 'rgba(245,158,11,0.8)', marginTop: 3, lineHeight: 1.6 }}>
            {t('firewall.warning_body', { port: window.location.port || '80' })}
          </div>
        </div>
      </div>

      {/* Rules table */}
      <Card style={{ padding: 0, overflow: 'hidden' }}>
        {loading ? (
          <div style={{ padding: 48, textAlign: 'center', color: 'var(--text-muted)', fontSize: 14 }}>
            {t('common.loading')}…
          </div>
        ) : rules.length === 0 ? (
          <div style={{ padding: 48, textAlign: 'center', display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 12 }}>
            <Shield size={40} style={{ color: 'var(--text-muted)', opacity: 0.2 }} />
            <p style={{ color: 'var(--text-muted)', fontSize: 14 }}>{t('firewall.empty')}</p>
          </div>
        ) : (
          <div style={{ overflowX: 'auto' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left' }}>
              <thead>
                <tr style={{ borderBottom: '1px solid var(--border)' }}>
                  {[t('common.status'), t('firewall.action'), t('common.port'), t('common.protocol'), t('firewall.source'), t('common.remark'), ''].map((h, i) => (
                    <th key={i} style={{ padding: '14px 20px', fontSize: 11, fontWeight: 700, color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.06em', whiteSpace: 'nowrap' }}>
                      {h}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {rules.map(rule => (
                  <tr key={rule.id} style={{ borderBottom: '1px solid var(--border)', transition: 'background 0.15s' }}
                    onMouseEnter={e => (e.currentTarget.style.background = 'rgba(255,255,255,0.025)')}
                    onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}
                  >
                    <td style={{ padding: '14px 20px' }}>
                      {rule.enable ? (
                        <Badge variant="success" style={{ display: 'inline-flex', alignItems: 'center', gap: 5 }}>
                          <CheckCircle2 size={10} /> {t('firewall.active')}
                        </Badge>
                      ) : (
                        <Badge variant="secondary" style={{ display: 'inline-flex', alignItems: 'center', gap: 5 }}>
                          <XCircle size={10} /> {t('firewall.disabled')}
                        </Badge>
                      )}
                    </td>
                    <td style={{ padding: '14px 20px' }}>
                      <span style={{ fontWeight: 700, fontSize: 13, color: rule.action === 'allow' ? '#22c55e' : '#ef4444' }}>
                        {rule.action === 'allow' ? t('firewall.allow') : t('firewall.deny')}
                      </span>
                    </td>
                    <td style={{ padding: '14px 20px', fontFamily: 'monospace', fontSize: 13 }}>
                      {rule.port === 0 ? <span style={{ color: 'var(--text-muted)' }}>All</span> : rule.port}
                    </td>
                    <td style={{ padding: '14px 20px' }}>
                      <span style={{ fontSize: 11, fontWeight: 700, textTransform: 'uppercase', background: 'rgba(255,255,255,0.05)', border: '1px solid var(--border)', padding: '3px 8px', borderRadius: 6 }}>
                        {rule.protocol}
                      </span>
                    </td>
                    <td style={{ padding: '14px 20px', fontSize: 13, color: 'var(--text-secondary)', fontFamily: 'monospace' }}>
                      {rule.ip || 'Any'}
                    </td>
                    <td style={{ padding: '14px 20px', fontSize: 13, maxWidth: 180, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                      {rule.remark || <span style={{ color: 'var(--text-muted)' }}>—</span>}
                    </td>
                    <td style={{ padding: '14px 16px', textAlign: 'right' }}>
                      <button
                        onClick={() => handleDelete(rule.id)}
                        style={{ padding: '6px 10px', borderRadius: 8, border: 'none', background: 'transparent', color: 'var(--danger)', cursor: 'pointer', transition: 'background 0.15s' }}
                        title={t('common.delete')}
                      >
                        <Trash2 size={15} />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      {/* Add rule modal */}
      <Modal
        isOpen={isModalOpen}
        onClose={() => setModal(false)}
        title={t('firewall.add_rule')}
        width={500}
        footer={
          <>
            <Button variant="secondary" onClick={() => setModal(false)}>{t('common.cancel')}</Button>
            <Button onClick={handleSave} disabled={saveLoading}>
              {saveLoading ? t('common.loading') + '…' : t('firewall.apply_rule')}
            </Button>
          </>
        }
      >
        <form onSubmit={handleSave} style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          <Select
            label={t('firewall.action')}
            value={newRule.action}
            onChange={e => setNewRule({ ...newRule, action: e.target.value })}
            options={ACTION_OPTIONS}
          />
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 14 }}>
            <Input
              label={t('common.port') + ' (0 = all)'}
              type="number"
              value={newRule.port}
              onChange={e => setNewRule({ ...newRule, port: parseInt(e.target.value) || 0 })}
            />
            <Select
              label={t('common.protocol')}
              value={newRule.protocol}
              onChange={e => setNewRule({ ...newRule, protocol: e.target.value })}
              options={PROTO_OPTIONS}
            />
          </div>
          <Input
            label={t('firewall.source') + ' IP / CIDR'}
            placeholder="0.0.0.0/0"
            value={newRule.ip}
            onChange={e => setNewRule({ ...newRule, ip: e.target.value })}
          />
          <Input
            label={t('common.remark')}
            placeholder={t('firewall.remark_placeholder')}
            value={newRule.remark}
            onChange={e => setNewRule({ ...newRule, remark: e.target.value })}
          />
        </form>
      </Modal>

      {/* Spin keyframe */}
      <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
    </div>
  );
}
