import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Sparkles, ArrowRight } from 'lucide-react';
import { authApi } from '../../api';

export default function LoginPage() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      const res: any = await authApi.login({ username, password });
      localStorage.setItem('token', res.data.token);
      if (res.data.tenant_id) {
        localStorage.setItem('tenant_id', String(res.data.tenant_id));
      }
      navigate('/');
    } catch {
      setError('用户名或密码错误');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-nexus-50 via-lavender-50 to-sakura-50 flex items-center justify-center p-4">
      <div className="w-full max-w-sm">
        <div className="text-center mb-8">
          <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-nexus-400 to-sakura-400 flex items-center justify-center mx-auto mb-4 shadow-lg">
            <Sparkles className="w-8 h-8 text-white" />
          </div>
          <h1 className="text-2xl font-bold text-gray-800">Nexus</h1>
          <p className="text-sm text-gray-400 mt-1">数据解析控制台</p>
        </div>

        <form onSubmit={handleSubmit} className="bg-white rounded-2xl border border-border-soft p-8 shadow-sm">
          <div className="space-y-4">
            <div>
              <label className="block text-xs font-medium text-gray-500 mb-1.5">用户名</label>
              <input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="w-full px-4 py-3 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300 focus:ring-2 focus:ring-nexus-100 transition-all"
                placeholder="admin"
                required
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-500 mb-1.5">密码</label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full px-4 py-3 rounded-xl border border-border-soft bg-surface text-sm focus:outline-none focus:border-nexus-300 focus:ring-2 focus:ring-nexus-100 transition-all"
                placeholder="••••••••"
                required
              />
            </div>
          </div>

          {error && (
            <div className="text-xs text-red-500 bg-red-50 px-3 py-2 rounded-lg mt-3">{error}</div>
          )}

          <button
            type="submit"
            disabled={loading}
            className="w-full mt-6 py-3 rounded-xl bg-gradient-to-r from-nexus-500 to-nexus-600 text-white text-sm font-medium shadow-sm hover:shadow-lg hover:from-nexus-600 hover:to-nexus-700 transition-all disabled:opacity-50 flex items-center justify-center gap-2"
          >
            {loading ? (
              <span className="w-4 h-4 rounded-full border-2 border-white border-t-transparent animate-spin" />
            ) : (
              <>登录 <ArrowRight className="w-4 h-4" /></>
            )}
          </button>
        </form>

        <p className="text-center text-xs text-gray-300 mt-6">Nexus v0.1.0</p>
      </div>
    </div>
  );
}
