interface PageHeaderProps {
  title: string;
  description?: string;
  action?: React.ReactNode;
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

interface CardProps {
  children: React.ReactNode;
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

interface BadgeProps {
  children: React.ReactNode;
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

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
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

export function Loading() {
  return (
    <div className="flex items-center justify-center py-16">
      <div className="w-8 h-8 rounded-full border-2 border-nexus-200 border-t-nexus-500 animate-spin" />
    </div>
  );
}

interface StatusBadgeProps {
  status: string;
}

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

export function StatusBadge({ status }: StatusBadgeProps) {
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
