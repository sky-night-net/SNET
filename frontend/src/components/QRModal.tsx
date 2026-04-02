import Modal from './Modal';
import { QRCodeSVG } from 'qrcode.react';
import { Copy, Check } from 'lucide-react';
import { useState } from 'react';

interface QRModalProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  link: string;
}

export default function QRModal({ isOpen, onClose, title, link }: QRModalProps) {
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(link);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={title}
      width={400}
    >
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 24, padding: '10px 0' }}>
        <div style={{ background: 'white', padding: 16, borderRadius: 16 }}>
          <QRCodeSVG value={link} size={220} level="M" includeMargin={false} />
        </div>
        
        <div style={{ width: '100%', display: 'flex', flexDirection: 'column', gap: 8 }}>
          <div style={{ fontSize: 12, fontWeight: 700, color: 'var(--text-muted)', textTransform: 'uppercase' }}>
            URL Configuration
          </div>
          <div style={{ display: 'flex', gap: 8 }}>
            <div style={{ 
              flex: 1,
              background: 'rgba(255,255,255,0.05)', 
              border: '1px solid var(--border)',
              borderRadius: 8,
              padding: '10px 12px',
              fontSize: 12,
              color: 'var(--text-secondary)',
              fontFamily: 'monospace',
              whiteSpace: 'nowrap',
              overflow: 'hidden',
              textOverflow: 'ellipsis'
            }}>
              {link}
            </div>
            <button
              onClick={handleCopy}
              style={{
                display: 'flex', alignItems: 'center', gap: 6,
                padding: '0 16px', borderRadius: 8, border: 'none',
                background: 'var(--accent)', color: 'white',
                cursor: 'pointer', fontWeight: 600, fontSize: 13
              }}
            >
              {copied ? <Check size={16} /> : <Copy size={16} />}
            </button>
          </div>
        </div>
      </div>
    </Modal>
  );
}
