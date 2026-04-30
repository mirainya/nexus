import { useState } from 'react';
import { Outlet, NavLink, useNavigate } from 'react-router-dom';
import {
  LayoutDashboard, GitBranch, MessageSquareText, ListTodo,
  CheckCircle, Database, Settings, LogOut, Sparkles, FlaskConical,
  Search, Share2, Activity, Building2, PanelLeftClose, PanelLeftOpen, Menu, X,
} from 'lucide-react';

const navItems = [
  { to: '/', icon: LayoutDashboard, label: '仪表盘' },
  { to: '/pipelines', icon: GitBranch, label: '流水线' },
  { to: '/prompts', icon: MessageSquareText, label: '提示词' },
  { to: '/jobs', icon: ListTodo, label: '任务' },
  { to: '/reviews', icon: CheckCircle, label: '审核' },
  { to: '/entities', icon: Database, label: '实体' },
  { to: '/graph', icon: Share2, label: '图谱' },
  { to: '/search', icon: Search, label: '搜索推荐' },
  { to: '/observability', icon: Activity, label: '可观测性' },
  { to: '/tenants', icon: Building2, label: '租户' },
  { to: '/playground', icon: FlaskConical, label: '测试' },
  { to: '/settings', icon: Settings, label: '设置' },
];

export default function MainLayout() {
  const navigate = useNavigate();
  const [collapsed, setCollapsed] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);

  const handleLogout = () => {
    localStorage.removeItem('token');
    navigate('/login');
  };

  const sidebarContent = (
    <>
      <div className={`p-4 flex items-center ${collapsed ? 'justify-center' : 'gap-3 px-6'}`}>
        <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-nexus-400 to-sakura-400 flex items-center justify-center shrink-0">
          <Sparkles className="w-5 h-5 text-white" />
        </div>
        {!collapsed && (
          <div className="min-w-0">
            <h1 className="text-lg font-semibold text-gray-800 tracking-tight">Nexus</h1>
            <p className="text-xs text-gray-400">数据解析平台</p>
          </div>
        )}
      </div>

      <nav className="flex-1 px-3 py-2 space-y-1 overflow-y-auto">
        {navItems.map(({ to, icon: Icon, label }) => (
          <NavLink
            key={to}
            to={to}
            end={to === '/'}
            onClick={() => setMobileOpen(false)}
            title={collapsed ? label : undefined}
            className={({ isActive }) =>
              `flex items-center gap-3 ${collapsed ? 'justify-center px-2' : 'px-4'} py-2.5 rounded-xl text-sm transition-all duration-200 ${
                isActive
                  ? 'bg-nexus-50 text-nexus-600 font-medium shadow-sm'
                  : 'text-gray-500 hover:bg-surface-hover hover:text-gray-700'
              }`
            }
          >
            <Icon className="w-[18px] h-[18px] shrink-0" />
            {!collapsed && label}
          </NavLink>
        ))}
      </nav>

      <div className="p-3 border-t border-border-soft space-y-1">
        <button
          onClick={handleLogout}
          title={collapsed ? '退出登录' : undefined}
          className={`flex items-center gap-3 ${collapsed ? 'justify-center px-2' : 'px-4'} py-2.5 rounded-xl text-sm text-gray-400 hover:text-sakura-500 hover:bg-sakura-50 w-full transition-all duration-200`}
        >
          <LogOut className="w-[18px] h-[18px] shrink-0" />
          {!collapsed && '退出登录'}
        </button>
        <button
          onClick={() => setCollapsed(!collapsed)}
          className="hidden lg:flex items-center justify-center w-full py-2 rounded-xl text-gray-300 hover:text-gray-500 hover:bg-surface-hover transition-all"
        >
          {collapsed ? <PanelLeftOpen className="w-4 h-4" /> : <PanelLeftClose className="w-4 h-4" />}
        </button>
      </div>
    </>
  );

  return (
    <div className="flex h-screen bg-surface">
      {/* Desktop sidebar */}
      <aside className={`hidden lg:flex flex-col bg-surface-card border-r border-border-soft transition-all duration-300 ${collapsed ? 'w-[68px]' : 'w-64'}`}>
        {sidebarContent}
      </aside>

      {/* Mobile overlay */}
      {mobileOpen && (
        <div className="fixed inset-0 bg-black/30 z-40 lg:hidden" onClick={() => setMobileOpen(false)} />
      )}

      {/* Mobile sidebar */}
      <aside className={`fixed inset-y-0 left-0 z-50 w-64 bg-surface-card border-r border-border-soft flex flex-col transition-transform duration-300 lg:hidden ${mobileOpen ? 'translate-x-0' : '-translate-x-full'}`}>
        <button onClick={() => setMobileOpen(false)} className="absolute top-4 right-4 p-1 rounded-lg hover:bg-surface-hover text-gray-400">
          <X className="w-5 h-5" />
        </button>
        {sidebarContent}
      </aside>

      {/* Main */}
      <main className="flex-1 overflow-auto">
        <div className="lg:hidden sticky top-0 z-30 bg-surface-card/80 backdrop-blur-sm border-b border-border-soft px-4 py-3 flex items-center gap-3">
          <button onClick={() => setMobileOpen(true)} className="p-1.5 rounded-lg hover:bg-surface-hover text-gray-500">
            <Menu className="w-5 h-5" />
          </button>
          <div className="flex items-center gap-2">
            <div className="w-7 h-7 rounded-lg bg-gradient-to-br from-nexus-400 to-sakura-400 flex items-center justify-center">
              <Sparkles className="w-3.5 h-3.5 text-white" />
            </div>
            <span className="text-sm font-semibold text-gray-700">Nexus</span>
          </div>
        </div>
        <div className="p-4 sm:p-6 lg:p-8 max-w-7xl mx-auto">
          <Outlet />
        </div>
      </main>
    </div>
  );
}
