import React from 'react';
import { motion } from 'framer-motion';

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label: string;
  error?: string;
  icon?: React.ReactNode;
}

export function Input({ label, error, icon, ...props }: InputProps) {
  const [focused, setFocused] = React.useState(false);

  return (
    <div style={{ marginBottom: 18 }}>
      <label style={{ 
        display: 'block', fontSize: 13, fontWeight: 600, 
        color: 'var(--text-secondary)', marginBottom: 8,
        letterSpacing: '0.2px'
      }}>
        {label}
      </label>
      <div style={{ position: 'relative' }}>
        {icon && (
          <div style={{
            position: 'absolute', left: 14, top: '50%', transform: 'translateY(-50%)',
            color: focused ? 'var(--accent)' : 'var(--text-muted)',
            transition: 'color 0.2s', display: 'flex'
          }}>
            {icon}
          </div>
        )}
        <input
          {...props}
          onFocus={(e) => { setFocused(true); props.onFocus?.(e); }}
          onBlur={(e) => { setFocused(false); props.onBlur?.(e); }}
          style={{
            width: '100%', padding: `12px 14px 12px ${icon ? 38 : 14}px`,
            background: 'rgba(255,255,255,0.04)', border: '1px solid var(--border)',
            borderColor: focused ? 'var(--accent)' : error ? 'var(--danger)' : 'var(--border)',
            borderRadius: 12, color: 'var(--text-primary)', fontSize: 14,
            outline: 'none', transition: 'all 0.2s cubic-bezier(0.16, 1, 0.3, 1)',
            boxShadow: focused ? '0 0 0 4px rgba(99,102,241,0.1)' : 'none',
            ...props.style
          }}
        />
      </div>
      {error && (
        <p style={{ color: 'var(--danger)', fontSize: 12, marginTop: 6, fontWeight: 500 }}>
          {error}
        </p>
      )}
    </div>
  );
}

interface SelectProps extends React.SelectHTMLAttributes<HTMLSelectElement> {
  label: string;
  options: { value: string; label: string }[];
}

export function Select({ label, options, ...props }: SelectProps) {
  return (
    <div style={{ marginBottom: 18 }}>
      <label style={{ 
        display: 'block', fontSize: 13, fontWeight: 600, 
        color: 'var(--text-secondary)', marginBottom: 8 
      }}>
        {label}
      </label>
      <select
        {...props}
        style={{
          width: '100%', padding: '12px 14px',
          background: 'var(--bg-elevated)', border: '1px solid var(--border)',
          borderRadius: 12, color: 'var(--text-primary)', fontSize: 14,
          outline: 'none', cursor: 'pointer',
          appearance: 'none', backgroundImage: 'url("data:image/svg+xml,%3Csvg xmlns=\'http://www.w3.org/2000/svg\' width=\'24\' height=\'24\' viewBox=\'0 0 24 24\' fill=\'none\' stroke=\'rgba(255,255,255,0.4)\' stroke-width=\'2\' stroke-linecap=\'round\' stroke-linejoin=\'round\'%3E%3Cpolyline points=\'6 9 12 15 18 9\'%3E%3C/polyline%3E%3C/svg%3E")',
          backgroundRepeat: 'no-repeat', backgroundPosition: 'right 12px center',
          backgroundSize: '16px',
          ...props.style
        }}
      >
        {options.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
      </select>
    </div>
  );
}

interface SwitchProps {
  label: string;
  checked: boolean;
  onChange: (val: boolean) => void;
}

export function Switch({ label, checked, onChange }: SwitchProps) {
  return (
    <div 
      style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 18, cursor: 'pointer' }}
      onClick={() => onChange(!checked)}
    >
      <span style={{ fontSize: 14, fontWeight: 600, color: 'var(--text-secondary)' }}>{label}</span>
      <div style={{
        width: 44, height: 24, borderRadius: 20,
        background: checked ? 'var(--accent)' : 'var(--bg-elevated)',
        border: '1px solid var(--border)',
        position: 'relative', transition: 'background 0.2s',
        display: 'flex', alignItems: 'center', padding: 2
      }}>
        <motion.div
          animate={{ x: checked ? 20 : 0 }}
          style={{
            width: 18, height: 18, borderRadius: '50%',
            background: 'white', boxShadow: '0 2px 4px rgba(0,0,0,0.2)'
          }}
        />
      </div>
    </div>
  );
}

export function Button({ children, loading, variant = 'primary', ...props }: any) {
  const isPrimary = variant === 'primary';
  return (
    <motion.button
      whileTap={{ scale: 0.97 }}
      {...props}
      disabled={loading || props.disabled}
      style={{
        padding: '12px 24px', borderRadius: 12, border: 'none',
        background: isPrimary ? 'var(--accent)' : 'var(--bg-elevated)',
        color: isPrimary ? 'white' : 'var(--text-secondary)',
        fontSize: 14, fontWeight: 700, cursor: loading ? 'not-allowed' : 'pointer',
        boxShadow: isPrimary ? '0 0 20px var(--accent-glow)' : 'none',
        display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8,
        transition: 'all 0.2s', opacity: loading ? 0.7 : 1,
        fontFamily: 'inherit',
        ...props.style
      }}
    >
      {loading && (
        <motion.div
          animate={{ rotate: 360 }}
          transition={{ duration: 1, repeat: Infinity, ease: 'linear' }}
          style={{ width: 16, height: 16, border: '2px solid rgba(255,255,255,0.3)', borderTopColor: 'white', borderRadius: '50%' }}
        />
      )}
      {children}
    </motion.button>
  );
}

interface CardProps {
  children: React.ReactNode;
  title?: string;
  icon?: any;
  footer?: React.ReactNode;
  style?: React.CSSProperties;
}

export function Card({ children, title, icon: Icon, footer, style }: CardProps) {
  return (
    <div style={{
      background: 'var(--bg-card)', border: '1px solid var(--border)',
      borderRadius: 16, overflow: 'hidden', display: 'flex', flexDirection: 'column',
      boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
      ...style
    }}>
      {(title || Icon) && (
        <div style={{
          padding: '16px 20px', borderBottom: '1px solid var(--border)',
          display: 'flex', alignItems: 'center', gap: 10,
          background: 'rgba(255,255,255,0.01)'
        }}>
          {Icon && <Icon size={18} color="var(--accent)" />}
          {title && <span style={{ fontSize: 14, fontWeight: 700, letterSpacing: '0.2px' }}>{title}</span>}
        </div>
      )}
      <div style={{ padding: 20, flex: 1 }}>
        {children}
      </div>
      {footer && (
        <div style={{
          padding: '14px 20px', background: 'rgba(255,255,255,0.02)',
          borderTop: '1px solid var(--border)'
        }}>
          {footer}
        </div>
      )}
    </div>
  );
}
