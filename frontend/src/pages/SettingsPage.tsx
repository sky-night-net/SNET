import { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { Shield, Save, Info, Globe, Layout, RefreshCw } from 'lucide-react';
import { http } from '../lib/api';
import { Input, Button, Card } from '../components/UI';

export default function SettingsPage() {
  const [settings, setSettings] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [passForm, setPassForm] = useState({ oldPassword: '', newPassword: '', confirmPassword: '' });

  useEffect(() => {
    fetchSettings();
  }, []);

  const fetchSettings = async () => {
    try {
      const { data } = await http.get('/settings');
      if (data.success) setSettings(data.obj);
    } catch (err) {
      console.error('Failed to fetch settings');
    } finally {
      setLoading(false);
    }
  };

  const handleSaveSettings = async () => {
    setSaving(true);
    try {
      const { data } = await http.put('/settings', settings);
      if (data.success) alert('Настройки сохранены');
    } catch (err) {
      alert('Ошибка при сохранении');
    } finally {
      setSaving(false);
    }
  };

  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    if (passForm.newPassword !== passForm.confirmPassword) {
      alert('Пароли не совпадают');
      return;
    }
    try {
      const { data } = await http.post('/settings/password', {
        oldPassword: passForm.oldPassword,
        newPassword: passForm.newPassword
      });
      if (data.success) {
        alert('Пароль успешно изменен');
        setPassForm({ oldPassword: '', newPassword: '', confirmPassword: '' });
      }
    } catch (err: any) {
      alert(err.response?.data?.msg || 'Ошибка при смене пароля');
    }
  };

  if (loading) return <div style={{ padding: 40, textAlign: 'center', color: 'var(--text-muted)' }}>Загрузка настроек...</div>;

  return (
    <div style={{ maxWidth: 1000, margin: '0 auto' }}>
      <motion.div
        initial={{ opacity: 0, y: -10 }}
        animate={{ opacity: 1, y: 0 }}
        style={{ marginBottom: 32 }}
      >
        <h1 style={{ fontSize: 26, fontWeight: 800, letterSpacing: '-0.5px' }}>Настройки</h1>
        <p style={{ color: 'var(--text-secondary)', fontSize: 14, marginTop: 4 }}>
          Управление параметрами панели и безопасностью системы
        </p>
      </motion.div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 24 }}>
        {/* General Settings */}
        <motion.div initial={{ opacity: 0, x: -20 }} animate={{ opacity: 1, x: 0 }} transition={{ delay: 0.1 }}>
          <Card 
            title="Панель и сеть" 
            icon={Globe}
            footer={
              <Button 
                onClick={handleSaveSettings} 
                loading={saving}
                style={{ width: '100%', gap: 8 }}
              >
                <Save size={16} /> Сохранить изменения
              </Button>
            }
          >
            <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
              <Input 
                label="Порт панели" 
                value={settings.port || '8080'} 
                onChange={e => setSettings({...settings, port: e.target.value})}
                placeholder="8080"
              />
              <div style={{ padding: '12px 14px', borderRadius: 12, background: 'rgba(99,102,241,0.05)', border: '1px solid var(--border)' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, color: 'var(--accent)', marginBottom: 6 }}>
                  <Info size={14} />
                  <span style={{ fontSize: 13, fontWeight: 700 }}>Совет</span>
                </div>
                <p style={{ fontSize: 11, color: 'var(--text-muted)', lineHeight: 1.5 }}>
                  Смена порта потребует перезапуска панели. Убедитесь, что новый порт открыт в вашем фаерволе.
                </p>
              </div>
            </div>
          </Card>
        </motion.div>

        {/* Security */}
        <motion.div initial={{ opacity: 0, x: 20 }} animate={{ opacity: 1, x: 0 }} transition={{ delay: 0.2 }}>
          <Card title="Безопасность" icon={Shield}>
            <form onSubmit={handleChangePassword} style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
              <Input 
                label="Текущий пароль" 
                type="password"
                value={passForm.oldPassword} 
                onChange={e => setPassForm({...passForm, oldPassword: e.target.value})}
                required
              />
              <Input 
                label="Новый пароль" 
                type="password"
                value={passForm.newPassword} 
                onChange={e => setPassForm({...passForm, newPassword: e.target.value})}
                required
              />
              <Input 
                label="Подтвердите пароль" 
                type="password"
                value={passForm.confirmPassword} 
                onChange={e => setPassForm({...passForm, confirmPassword: e.target.value})}
                required
              />
              <Button type="submit" variant="secondary" style={{ marginTop: 8 }}>
                Обновить пароль
              </Button>
            </form>
          </Card>
        </motion.div>

        {/* System Info */}
        <motion.div 
          initial={{ opacity: 0, y: 20 }} 
          animate={{ opacity: 1, y: 0 }} 
          transition={{ delay: 0.3 }}
          style={{ gridColumn: 'span 2' }}
        >
          <Card title="О системе" icon={Layout}>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 16 }}>
              {[
                { label: 'Панель', value: 'v3.0.0 Stable' },
                { label: 'Ядро Xray', value: 'v1.8.4' },
                { label: 'ОС', value: 'Ubuntu 24.04 Focal' },
                { label: 'Архитектура', value: 'amd64 / x86_64' },
                { label: 'Uptime', value: '18d 04h 22m' },
                { label: 'База данных', value: 'SQLite 3' },
              ].map(item => (
                <div key={item.label} style={{ padding: '12px 16px', border: '1px solid var(--border)', borderRadius: 12, background: 'var(--bg-elevated)' }}>
                  <div style={{ fontSize: 11, color: 'var(--text-muted)', marginBottom: 4 }}>{item.label}</div>
                  <div style={{ fontSize: 14, fontWeight: 700 }}>{item.value}</div>
                </div>
              ))}
            </div>
            
            <div style={{ marginTop: 24, padding: '16px 20px', border: '1px dashed var(--border)', borderRadius: 12, background: 'rgba(255,255,255,0.01)', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <div>
                <div style={{ fontSize: 14, fontWeight: 700, marginBottom: 4 }}>Резервная копия БД</div>
                <div style={{ fontSize: 12, color: 'var(--text-muted)' }}>Скачать текущий файл snet.db для восстановления или миграции</div>
              </div>
              <Button onClick={() => window.open('/api/settings/backup', '_blank')} variant="secondary" style={{ gap: 8 }}>
                 <RefreshCw size={14} /> Скачать .db
              </Button>
            </div>
          </Card>
        </motion.div>
      </div>
    </div>
  );
}
