import { useState, useEffect } from 'react';
import { Plus, Server, ArrowUpRight, ArrowDownLeft, Zap, Shield, Trash2, Edit2 } from 'lucide-react';
import { api } from '../lib/api';
import { Card, Button, Badge } from '../components/UI';
import InboundModal from '../components/InboundModal';
import { useTranslation } from 'react-i18next';

function formatBytes(bytes: number) {
  if (!bytes) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

export default function InboundsPage() {
  const { t } = useTranslation();
  const [inbounds, setInbounds] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingInbound, setEditingInbound] = useState<any>(null);

  const fetchInbounds = async () => {
    try {
      const res = await api.get('/inbounds');
      if (res.data.success) {
        setInbounds(res.data.obj || []);
      }
    } catch (e) {
      console.error('Failed to fetch inbounds');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchInbounds();
  }, []);

  const handleCreate = () => {
    setEditingInbound(null);
    setIsModalOpen(true);
  };

  const handleEdit = (inbound: any) => {
    setEditingInbound(inbound);
    setIsModalOpen(true);
  };

  const handleDelete = async (id: number) => {
    if (!confirm(t('common.delete') + '?')) return;
    try {
      await api.delete(`/inbounds/${id}`);
      fetchInbounds();
    } catch (e) {
      alert('Ошибка при удалении');
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold">{t('nav.nodes')}</h1>
          <p className="text-text-muted text-sm">Управление входящими подключениями и протоколами</p>
        </div>
        <Button onClick={handleCreate} className="flex items-center gap-2">
          <Plus size={18} />
          {t('common.add')}
        </Button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {loading ? (
          <div className="col-span-full p-12 text-center text-text-muted">Загрузка нод...</div>
        ) : inbounds.length === 0 ? (
          <div className="col-span-full p-12 text-center border-2 border-dashed border-border rounded-2xl">
            <Server size={48} className="mx-auto mb-4 text-text-muted opacity-20" />
            <p className="text-text-muted">У вас пока нет настроенных нод. Нажмите "Добавить", чтобы начать.</p>
          </div>
        ) : (
          inbounds.map(ib => (
            <Card 
              key={ib.id}
              footer={
                <div className="flex justify-between items-center">
                  <div className="flex items-center gap-2 text-xs font-mono text-text-muted bg-black/20 px-2 py-1 rounded">
                    <Zap size={12} className="text-accent" />
                    {ib.tag}
                  </div>
                  <div className="flex gap-2">
                    <button 
                      onClick={() => handleEdit(ib)}
                      className="p-2 hover:bg-white/5 rounded-lg transition-colors text-text-secondary"
                    >
                      <Edit2 size={16} />
                    </button>
                    <button 
                      onClick={() => handleDelete(ib.id)}
                      className="p-2 hover:bg-red-500/10 rounded-lg transition-colors text-danger"
                    >
                      <Trash2 size={16} />
                    </button>
                  </div>
                </div>
              }
            >
              <div className="flex justify-between items-start mb-6">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-xl bg-accent/20 flex items-center justify-center text-accent">
                    <Shield size={20} />
                  </div>
                  <div>
                    <h3 className="font-bold text-lg leading-tight">{ib.remark}</h3>
                    <div className="flex items-center gap-2 mt-1">
                      <Badge variant="primary" className="uppercase">{ib.protocol}</Badge>
                      <span className="text-xs text-text-muted font-mono">{ib.port}</span>
                    </div>
                  </div>
                </div>
                {ib.enable ? (
                  <div className="w-2 h-2 rounded-full bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.5)]" />
                ) : (
                  <div className="w-2 h-2 rounded-full bg-text-muted" />
                )}
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="p-3 rounded-xl bg-white/5 border border-white/5">
                  <div className="flex items-center gap-2 text-xs text-text-muted mb-1">
                    <ArrowUpRight size={12} className="text-blue-400" />
                    Upload
                  </div>
                  <div className="font-bold font-mono text-sm">{formatBytes(ib.up)}</div>
                </div>
                <div className="p-3 rounded-xl bg-white/5 border border-white/5">
                  <div className="flex items-center gap-2 text-xs text-text-muted mb-1">
                    <ArrowDownLeft size={12} className="text-green-400" />
                    Download
                  </div>
                  <div className="font-bold font-mono text-sm">{formatBytes(ib.down)}</div>
                </div>
              </div>

              <div className="mt-4 pt-4 border-t border-white/5">
                <div className="flex justify-between text-xs text-text-muted mb-2">
                  <span>{t('common.traffic')}</span>
                  <span>{ib.total > 0 ? formatBytes(ib.total) : '∞'}</span>
                </div>
                {ib.total > 0 && (
                  <div className="w-full h-1.5 bg-white/5 rounded-full overflow-hidden">
                    <div 
                      className="h-full bg-accent" 
                      style={{ width: `${Math.min(100, ((ib.up + ib.down) / ib.total) * 100)}%` }} 
                    />
                  </div>
                )}
              </div>
            </Card>
          ))
        )}
      </div>

      <InboundModal 
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSuccess={fetchInbounds}
        inbound={editingInbound}
      />
    </div>
  );
}
