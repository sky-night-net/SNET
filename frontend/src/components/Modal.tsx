import React, { useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { X } from 'lucide-react';

interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  children: React.ReactNode;
  footer?: React.ReactNode;
  width?: number | string;
}

export default function Modal({ isOpen, onClose, title, children, footer, width = 500 }: ModalProps) {
  useEffect(() => {
    if (isOpen) document.body.style.overflow = 'hidden';
    else document.body.style.overflow = 'unset';
    return () => { document.body.style.overflow = 'unset'; };
  }, [isOpen]);

  return (
    <AnimatePresence>
      {isOpen && (
        <>
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={onClose}
            style={{
              position: 'fixed', inset: 0, zIndex: 100,
              background: 'rgba(0,0,0,0.6)', backdropFilter: 'blur(8px)'
            }}
          />
          <div style={{
            position: 'fixed', inset: 0, zIndex: 101,
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            padding: 20, pointerEvents: 'none'
          }}>
            <motion.div
              initial={{ opacity: 0, scale: 0.95, y: 20 }}
              animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.95, y: 20 }}
              transition={{ type: 'spring', damping: 25, stiffness: 300 }}
              style={{
                width: '100%', maxWidth: width,
                background: 'var(--bg-card)', border: '1px solid var(--border)',
                borderRadius: 24, boxShadow: 'var(--shadow-lifted)',
                display: 'flex', flexDirection: 'column',
                maxHeight: '90vh', pointerEvents: 'auto',
                overflow: 'hidden'
              }}
            >
              {/* Header */}
              <div style={{
                padding: '24px 28px 16px', display: 'flex', alignItems: 'center',
                justifyContent: 'space-between'
              }}>
                <h3 style={{ fontSize: 20, fontWeight: 800, letterSpacing: '-0.5px' }}>{title}</h3>
                <button
                  onClick={onClose}
                  style={{
                    background: 'var(--bg-elevated)', border: '1px solid var(--border)',
                    width: 32, height: 32, borderRadius: 10, cursor: 'pointer',
                    display: 'flex', alignItems: 'center', justifyContent: 'center',
                    color: 'var(--text-muted)', transition: 'all 0.2s'
                  }}
                  onMouseEnter={e => e.currentTarget.style.color = 'white'}
                  onMouseLeave={e => e.currentTarget.style.color = 'var(--text-muted)'}
                >
                  <X size={18} />
                </button>
              </div>

              {/* Body */}
              <div style={{ padding: '0 28px 28px', flex: 1, overflow: 'auto' }}>
                {children}
              </div>

              {/* Footer */}
              {footer && (
                <div style={{
                  padding: '16px 28px 24px', borderTop: '1px solid var(--border)',
                  display: 'flex', gap: 12, justifyContent: 'flex-end',
                  background: 'rgba(255,255,255,0.01)'
                }}>
                  {footer}
                </div>
              )}
            </motion.div>
          </div>
        </>
      )}
    </AnimatePresence>
  );
}
