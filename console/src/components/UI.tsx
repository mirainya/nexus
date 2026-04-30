import { useEffect, type ReactNode, type ButtonHTMLAttributes, type InputHTMLAttributes, type SelectHTMLAttributes, type TextareaHTMLAttributes } from 'react';

// ─── PageHeader ───

interface PageHeaderProps {
  title: string;
  description?: string;
  action?: ReactNode;
}

export function PageHeader({ title, description, action }: PageHeaderProps) {
  return (
    <div className="flex items-center justify-between mb-8">
      <div>
        <h2 className="text-2xl font-semibold text-gray-800 tracking-tight">{title}</h2>
        {description && <p className="text-sm text-gray-400 mt-1">{description}</p>}
      </div>
      {action}
    </div>
  );
}

// ─── Card ───

interface CardProps {
  children: ReactNode;
  className?: string;
  onClick?: () => void;
}

export function Card({ children, className = '', onClick }: CardProps) {
  return (
    <div
      onClick={onClick}
      className={`bg-surface-card rounded-2xl border border-border-soft p-6 transition-all duration-200 ${
        onClick ? 'cursor-pointer hover:shadow-md hover:border-nexus-200' : ''
      } ${className}`}
    >
      {children}
    </div>
  );
}

// ─── Badge ───

interface BadgeProps {
  children: ReactNode;
  variant?: 'default' | 'success' | 'warning' | 'error' | 'info';
}

const badgeStyles = {
  default: 'bg-gray-100 text-gray-600',
  success: 'bg-emerald-50 text-emerald-600',
  warning: 'bg-amber-50 text-amber-600',
  error: 'bg-red-50 text-red-500',
  info: 'bg-nexus-50 text-nexus-600',
};

export function Badge({ children, variant = 'default' }: BadgeProps) {
  return (
    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${badgeStyles[variant]}`}>
      {children}
    </span>
  );
}

// ─── Button ───

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'danger' | 'ghost';
  size?: 'sm' | 'md';
  loading?: boolean;
}

export function Button({ variant = 'primary', size = 'md', loading, className = '', children, disabled, ...props }: ButtonProps) {
  const base = 'inline-flex items-center justify-center gap-2 rounded-xl font-medium transition-all duration-200 disabled:opacity-50';
  const sizes = { sm: 'px-3 py-1.5 text-xs', md: 'px-5 py-2.5 text-sm' };
  const variants = {
    primary: 'bg-gradient-to-r from-nexus-500 to-nexus-600 text-white shadow-sm hover:shadow-md hover:from-nexus-600 hover:to-nexus-700',
    secondary: 'bg-surface-hover text-gray-600 hover:bg-gray-100',
    danger: 'bg-red-50 text-red-500 hover:bg-red-100',
    ghost: 'text-gray-500 hover:text-gray-700 hover:bg-surface-hover',
  };
  return (
    <button className={`${base} ${sizes[size]} ${variants[variant]} ${className}`} disabled={disabled || loading} {...props}>
      {loading && <span className="w-3.5 h-3.5 rounded-full border-2 border-current border-t-transparent animate-spin" />}
      {children}
    </button>
  );
}

// ─── EmptyState ───

export function EmptyState({ message }: { message: string }) {
  return (
    <div className="flex flex-col items-center justify-center py-16 text-gray-400">
      <div className="w-16 h-16 rounded-full bg-lavender-50 flex items-center justify-center mb-4">
        <span className="text-2xl">✨</span>
      </div>
      <p className="text-sm">{message}</p>
    </div>
  );
}

// ─── Loading ───

export function Loading() {
  return (
    <div className="flex items-center justify-center py-16">
      <div className="w-8 h-8 rounded-full border-2 border-nexus-200 border-t-nexus-500 animate-spin" />
    </div>
  );
}

// ─── StatusBadge ───

const statusLabels: Record<string, string> = {
  pending: '待处理',
  running: '运行中',
  processing: '处理中',
  completed: '已完成',
  failed: '失败',
  approved: '已通过',
  rejected: '已拒绝',
  modified: '已修改',
};

export function StatusBadge({ status }: { status: string }) {
  const map: Record<string, BadgeProps['variant']> = {
    pending: 'warning',
    running: 'info',
    processing: 'info',
    completed: 'success',
    failed: 'error',
    approved: 'success',
    rejected: 'error',
    modified: 'info',
  };
  return <Badge variant={map[status] || 'default'}>{statusLabels[status] || status}</Badge>;
}

// ─── Modal ───

interface ModalProps {
  open: boolean;
  onClose: () => void;
  title: string;
  children: ReactNode;
  maxWidth?: string;
  footer?: ReactNode;
}

export function Modal({ open, onClose, title, children, maxWidth = 'max-w-md', footer }: ModalProps) {
  useEffect(() => {
    if (!open) return;
    const handler = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose(); };
    document.addEventListener('keydown', handler);
    document.body.style.overflow = 'hidden';
    return () => { document.removeEventListener('keydown', handler); document.body.style.overflow = ''; };
  }, [open, onClose]);

  if (!open) return null;

  return (
    <div className="fixed inset-0 bg-black/20 backdrop-blur-sm flex items-center justify-center z-50" onClick={onClose}>
      <div
        className={`bg-white rounded-2xl border border-border-soft p-6 w-full ${maxWidth} shadow-xl max-h-[85vh] overflow-y-auto animate-scale-in`}
        onClick={e => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-label={title}
      >
        <h3 className="text-lg font-semibold text-gray-800 mb-4">{title}</h3>
        {children}
        {footer && <div className="flex justify-end gap-2 mt-5">{footer}</div>}
      </div>
    </div>
  );
}

// ─── Tabs ───

interface TabItem {
  key: string;
  label: string;
  icon?: React.ElementType;
}

interface TabsProps {
  items: TabItem[];
  value: string;
  onChange: (key: string) => void;
}

export function Tabs({ items, value, onChange }: TabsProps) {
  return (
    <div className="flex gap-2 mb-6">
      {items.map(t => {
        const Icon = t.icon;
        return (
          <button
            key={t.key}
            onClick={() => onChange(t.key)}
            className={`inline-flex items-center gap-1.5 px-4 py-2 rounded-xl text-sm font-medium transition-all ${
              value === t.key
                ? 'bg-nexus-50 text-nexus-600 border border-nexus-200'
                : 'bg-surface-hover text-gray-500 border border-transparent hover:text-gray-700'
            }`}
          >
            {Icon && <Icon className="w-4 h-4" />}
            {t.label}
          </button>
        );
      })}
    </div>
  );
}

// ─── FilterTabs ───

interface FilterTabsProps {
  items: { key: string; label: string }[];
  value: string;
  onChange: (v: string) => void;
}

export function FilterTabs({ items, value, onChange }: FilterTabsProps) {
  return (
    <div className="flex gap-2 mb-6">
      {items.map(o => (
        <button
          key={o.key}
          onClick={() => onChange(o.key)}
          className={`px-3 py-1.5 rounded-lg text-xs font-medium transition-all ${
            value === o.key ? 'bg-nexus-50 text-nexus-600' : 'text-gray-400 hover:bg-surface-hover'
          }`}
        >
          {o.label}
        </button>
      ))}
    </div>
  );
}

// ─── Pagination ───

interface PaginationProps {
  page: number;
  total: number;
  pageSize?: number;
  onChange: (page: number) => void;
}

export function Pagination({ page, total, pageSize = 20, onChange }: PaginationProps) {
  const totalPages = Math.ceil(total / pageSize);
  if (totalPages <= 1) return null;

  return (
    <div className="flex items-center justify-center gap-2 mt-6">
      <button
        disabled={page <= 1}
        onClick={() => onChange(page - 1)}
        className="px-3 py-1.5 rounded-lg text-xs text-gray-500 hover:bg-surface-hover disabled:opacity-30 transition-colors"
      >
        上一页
      </button>
      <span className="px-3 py-1.5 text-xs text-gray-400">
        {page} / {totalPages}
      </span>
      <button
        disabled={page >= totalPages}
        onClick={() => onChange(page + 1)}
        className="px-3 py-1.5 rounded-lg text-xs text-gray-500 hover:bg-surface-hover disabled:opacity-30 transition-colors"
      >
        下一页
      </button>
    </div>
  );
}

// ─── Form Components ───

const inputBase = 'w-full px-4 py-2.5 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300 focus:ring-2 focus:ring-nexus-100 transition-all';

interface FormFieldProps {
  label: string;
  hint?: string;
  children: ReactNode;
}

export function FormField({ label, hint, children }: FormFieldProps) {
  return (
    <div>
      <label className="block text-xs font-medium text-gray-500 mb-1.5">
        {label}
        {hint && <span className="text-gray-300 ml-1">({hint})</span>}
      </label>
      {children}
    </div>
  );
}

export function Input({ className = '', ...props }: InputHTMLAttributes<HTMLInputElement>) {
  return <input className={`${inputBase} ${className}`} {...props} />;
}

export function Select({ className = '', children, ...props }: SelectHTMLAttributes<HTMLSelectElement>) {
  return <select className={`${inputBase} ${className}`} {...props}>{children}</select>;
}

export function Textarea({ className = '', ...props }: TextareaHTMLAttributes<HTMLTextAreaElement>) {
  return <textarea className={`${inputBase} resize-none ${className}`} {...props} />;
}
