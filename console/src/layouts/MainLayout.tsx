import { Outlet, NavLink, useNavigate } from 'react-router-dom';
import {
  LayoutDashboard, GitBranch, MessageSquareText, ListTodo,
  CheckCircle, Database, Settings, LogOut, Sparkles, FlaskConical, Search, Share2, Activity, Building2
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

  const handleLogout = () => {
    localStorage.removeItem('token');
    navigate('/login');
  };

  return (
    <div className="flex h-screen bg-surface">
      {/* Sidebar */}
      <aside className="w-64 bg-surface-card border-r border-border-soft flex flex-col">
        {/* Logo */}
        <div className="p-6 flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-nexus-400 to-sakura-400 flex items-center justify-center">
            <Sparkles className="w-5 h-5 text-white" />
          </div>
          <div>
            <h1 className="text-lg font-semibold text-gray-800 tracking-tight">Nexus</h1>
            <p className="text-xs text-gray-400">数据解析平台</p>
          </div>
        </div>

        {/* Nav */}
        <nav className="flex-1 px-3 py-2 space-y-1">
          {navItems.map(({ to, icon: Icon, label }) => (
            <NavLink
              key={to}
              to={to}
              end={to === '/'}
              className={({ isActive }) =>
                `flex items-center gap-3 px-4 py-2.5 rounded-xl text-sm transition-all duration-200 ${
                  isActive
                    ? 'bg-nexus-50 text-nexus-600 font-medium shadow-sm'
                    : 'text-gray-500 hover:bg-surface-hover hover:text-gray-700'
                }`
              }
            >
              <Icon className="w-[18px] h-[18px]" />
              {label}
            </NavLink>
          ))}
        </nav>

        {/* Bottom */}
        <div className="p-3 border-t border-border-soft">
          <button
            onClick={handleLogout}
            className="flex items-center gap-3 px-4 py-2.5 rounded-xl text-sm text-gray-400 hover:text-sakura-500 hover:bg-sakura-50 w-full transition-all duration-200"
          >
            <LogOut className="w-[18px] h-[18px]" />
            退出登录
          </button>
        </div>
      </aside>

      {/* Main */}
      <main className="flex-1 overflow-auto">
        <div className="p-8 max-w-7xl mx-auto">
          <Outlet />
        </div>
      </main>
    </div>
  );
}
