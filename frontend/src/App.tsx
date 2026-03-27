import { useState, useEffect } from 'react';
import Terminal from './components/Terminal';
import Dashboard from './components/Dashboard';
import AIAssistant from './components/AIAssistant';
import { Terminal as TerminalIcon, ShieldCheck, Activity, LayoutDashboard, Globe, Check, Sparkles } from 'lucide-react';
import './App.css';

function App() {
  const [token, setToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [view, setView] = useState<'terminal' | 'dashboard'>('terminal');
  const [tunnelUrl, setTunnelUrl] = useState('');
  const [tunnelStatus, setTunnelStatus] = useState('off');
  const [copied, setCopied] = useState(false);
  const [aiOpen, setAiOpen] = useState(false);
  const [terminalContext, setTerminalContext] = useState('');

  useEffect(() => {
    fetch('http://127.0.0.1:8081/auth/token')
      .then(res => res.text())
      .then(t => { setToken(t); setLoading(false); })
      .catch(() => setLoading(false));
  }, []);

  // Poll /health for tunnel URL
  useEffect(() => {
    const poll = () => {
      fetch('http://127.0.0.1:8081/health')
        .then(r => r.json())
        .then(d => {
          setTunnelStatus(d.tunnel_status || 'off');
          setTunnelUrl(d.tunnel_url || '');
        }).catch(() => {});
    };
    poll();
    const id = setInterval(poll, 5000);
    return () => clearInterval(id);
  }, []);

  const copyTunnelUrl = () => {
    if (!tunnelUrl) return;
    navigator.clipboard.writeText(tunnelUrl);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="min-h-screen bg-[#020617] text-slate-200 selection:bg-blue-500/30 font-inter">
      {/* Header */}
      <header className="border-b border-white/5 bg-black/20 backdrop-blur-md sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
          <div className="flex items-center gap-8">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-gradient-to-br from-blue-600 to-indigo-700 rounded-xl flex items-center justify-center shadow-lg shadow-blue-500/20">
                <TerminalIcon className="text-white" size={20} />
              </div>
              <div>
                <h1 className="text-lg font-bold tracking-tight text-white">ZeroExec <span className="text-blue-500">v1.0</span></h1>
                <p className="text-[10px] text-slate-500 uppercase font-mono tracking-tighter">Secure Native Terminal Gateway</p>
              </div>
            </div>

            <nav className="flex items-center gap-1 bg-white/5 p-1 rounded-xl border border-white/10">
               <button 
                onClick={() => setView('terminal')}
                className={`flex items-center gap-2 px-4 py-1.5 rounded-lg text-xs font-medium transition-all ${view === 'terminal' ? 'bg-blue-600 text-white shadow-lg' : 'text-slate-400 hover:text-white hover:bg-white/5'}`}
               >
                 <TerminalIcon size={14} /> Terminal
               </button>
               <button 
                onClick={() => setView('dashboard')}
                className={`flex items-center gap-2 px-4 py-1.5 rounded-lg text-xs font-medium transition-all ${view === 'dashboard' ? 'bg-blue-600 text-white shadow-lg' : 'text-slate-400 hover:text-white hover:bg-white/5'}`}
               >
                 <LayoutDashboard size={14} /> Governance
               </button>
            </nav>
          </div>

          <div className="flex items-center gap-4">
              {/* AI Assistant Toggle */}
              <button
                onClick={() => setAiOpen(!aiOpen)}
                className={`flex items-center gap-2 px-3 py-1.5 rounded-xl text-xs font-semibold transition-all border ${
                  aiOpen
                    ? 'bg-violet-500/20 border-violet-500/40 text-violet-300 shadow-lg shadow-violet-500/10'
                    : 'bg-white/5 border-white/10 text-slate-400 hover:text-white hover:bg-white/10'
                }`}
              >
                <Sparkles size={14} className={aiOpen ? 'text-violet-400' : ''} />
                AI
              </button>
             {/* Tunnel Status Indicator */}
             {tunnelStatus !== 'off' && (
               <button
                 onClick={copyTunnelUrl}
                 title={tunnelUrl || 'Tunnel connecting...'}
                 className={`flex items-center gap-2 px-3 py-1 rounded-full border text-xs font-medium transition-all ${
                   tunnelStatus === 'active'
                     ? 'bg-violet-500/10 border-violet-500/30 text-violet-400 hover:bg-violet-500/20'
                     : 'bg-yellow-500/10 border-yellow-500/30 text-yellow-400'
                 }`}
               >
                 <Globe size={13} className={tunnelStatus === 'pending' ? 'animate-pulse' : ''} />
                 {tunnelStatus === 'active'
                   ? (copied ? <><Check size={12}/> Copied!</> : 'Remote Access')
                   : 'Connecting...'}
               </button>
             )}
             <div className="flex items-center gap-2 px-3 py-1 bg-green-500/10 rounded-full border border-green-500/20">
                <ShieldCheck className="text-green-500" size={14} />
                <span className="text-xs font-medium text-green-500">TLS Secured</span>
             </div>
             <div className="w-8 h-8 rounded-full bg-slate-800 border border-white/10 flex items-center justify-center">
                <div className="w-2 h-2 bg-blue-500 rounded-full animate-pulse" />
             </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className={`max-w-7xl mx-auto px-6 py-8 ${view === 'terminal' ? 'grid grid-cols-12 gap-6' : ''}`}>
        {view === 'terminal' ? (
          <>
            <aside className="col-span-3 space-y-6">
              <div className="p-5 rounded-2xl bg-white/5 border border-white/10 backdrop-blur-sm">
                  <h3 className="text-xs font-semibold text-slate-500 uppercase mb-4 flex items-center gap-2">
                      <Activity size={14} /> System Health
                  </h3>
                  <div className="space-y-4">
                      <div className="flex justify-between items-center">
                          <span className="text-sm">Gateway Latency</span>
                          <span className="text-xs font-mono text-green-400">12ms</span>
                      </div>
                      <div className="w-full bg-slate-800 h-1 rounded-full overflow-hidden">
                          <div className="bg-blue-500 h-full w-[10%]" />
                      </div>
                  </div>
              </div>
            </aside>

            <section className="col-span-9 h-[calc(100vh-12rem)] min-h-[500px]">
              {loading ? (
                <div className="w-full h-full flex flex-col items-center justify-center bg-white/5 rounded-2xl border border-white/10 border-dashed">
                    <div className="w-12 h-12 border-2 border-blue-500/30 border-t-blue-500 rounded-full animate-spin mb-4" />
                    <p className="text-slate-400 text-sm animate-pulse">Establishing Secure Handshake...</p>
                </div>
              ) : token ? (
                <Terminal token={token} onOutput={(data) => {
                  setTerminalContext(prev => (prev + data).slice(-2000));
                }} />
              ) : (
                <div className="w-full h-full flex flex-col items-center justify-center bg-red-500/5 rounded-2xl border border-red-500/20 border-dashed">
                    <p className="text-red-400 text-sm">Failed to acquire secure session token.</p>
                </div>
              )}
            </section>
          </>
        ) : (
          <Dashboard token={token || ''} />
        )}
      </main>

      {/* AI Assistant Panel */}
      <AIAssistant
        token={token || ''}
        terminalContext={terminalContext}
        isOpen={aiOpen}
        onClose={() => setAiOpen(false)}
      />
    </div>
  );
}

export default App;
