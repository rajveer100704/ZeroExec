import React, { useEffect, useState } from 'react';
import { Activity, XCircle, Clock, User, Play, HardDrive, History } from 'lucide-react';
import ReplayPlayer from './ReplayPlayer';

interface Session {
  id: string;
  user_id: string;
  role: string;
  status: string;
  start_time: string;
  last_activity: string;
  duration_sec: number;
}

interface DashboardProps {
  token: string;
}

const Dashboard: React.FC<DashboardProps> = ({ token }) => {
  const [activeTab, setActiveTab] = useState<'sessions' | 'recordings'>('sessions');
  const [sessions, setSessions] = useState<Session[]>([]);
  const [recordings, setRecordings] = useState<string[]>([]);
  const [selectedRecording, setSelectedRecording] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchSessions = async () => {
    try {
      const res = await fetch(`https://localhost:8081/admin/sessions?token=${token}`);
      if (res.ok) {
        const data = await res.json();
        setSessions(data);
      }
    } catch (err) {
      console.error('Failed to fetch sessions:', err);
    } finally {
      setLoading(false);
    }
  };

  const fetchRecordings = async () => {
    try {
      const res = await fetch(`https://localhost:8081/admin/recordings?token=${token}`);
      if (res.ok) {
        const data = await res.json();
        setRecordings(data);
      }
    } catch (err) {
      console.error('Failed to fetch recordings:', err);
    }
  };

  const terminateSession = async (id: string) => {
    if (!confirm(`Terminate session ${id}?`)) return;
    try {
      await fetch(`https://localhost:8081/admin/terminate?token=${token}&id=${id}`, { method: 'POST' });
      fetchSessions();
    } catch (err) {
      console.error('Failed to terminate session:', err);
    }
  };

  useEffect(() => {
    if (activeTab === 'sessions') {
      fetchSessions();
      const interval = setInterval(fetchSessions, 3000);
      return () => clearInterval(interval);
    } else {
      fetchRecordings();
    }
  }, [activeTab]);

  const formatDuration = (sec: number) => {
    const mins = Math.floor(sec / 60);
    const s = sec % 60;
    return `${mins}m ${s}s`;
  };

  return (
    <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-700">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h2 className="text-2xl font-bold text-white flex items-center gap-3">
            <Activity className="text-blue-500" /> Session Governance
          </h2>
          <p className="text-slate-400 text-sm mt-1">Real-time oversight and investigative playback.</p>
        </div>
        
        <div className="flex bg-white/5 p-1 rounded-xl border border-white/10">
           <button 
            onClick={() => setActiveTab('sessions')}
            className={`flex items-center gap-2 px-4 py-1.5 rounded-lg text-xs font-semibold transition-all ${activeTab === 'sessions' ? 'bg-blue-600 text-white shadow-lg shadow-blue-500/20' : 'text-slate-400 hover:text-white'}`}
           >
             <Activity size={14} /> Live
           </button>
           <button 
            onClick={() => setActiveTab('recordings')}
            className={`flex items-center gap-2 px-4 py-1.5 rounded-lg text-xs font-semibold transition-all ${activeTab === 'recordings' ? 'bg-blue-600 text-white shadow-lg shadow-blue-500/20' : 'text-slate-400 hover:text-white'}`}
           >
             <History size={14} /> Archive
           </button>
        </div>
      </div>

      {activeTab === 'sessions' ? (
        <div className="overflow-hidden rounded-2xl border border-white/10 bg-white/5 backdrop-blur-sm">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="border-b border-white/10 bg-white/5">
                <th className="px-6 py-4 text-xs font-semibold text-slate-400 uppercase tracking-wider">Session ID</th>
                <th className="px-6 py-4 text-xs font-semibold text-slate-400 uppercase tracking-wider">User Identity</th>
                <th className="px-6 py-4 text-xs font-semibold text-slate-400 uppercase tracking-wider">Role</th>
                <th className="px-6 py-4 text-xs font-semibold text-slate-400 uppercase tracking-wider">Status</th>
                <th className="px-6 py-4 text-xs font-semibold text-slate-400 uppercase tracking-wider">Duration</th>
                <th className="px-6 py-4 text-xs font-semibold text-slate-400 uppercase tracking-wider text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5">
              {loading ? (
                <tr><td colSpan={6} className="px-6 py-12 text-center text-slate-500">Loading registry...</td></tr>
              ) : sessions.length === 0 ? (
                <tr><td colSpan={6} className="px-6 py-12 text-center text-slate-500">No active sessions found.</td></tr>
              ) : (
                sessions.map(s => (
                  <tr key={s.id} className="hover:bg-white/[0.02] transition-colors group">
                    <td className="px-6 py-4 font-mono text-xs text-blue-400">{s.id.slice(0, 12)}...</td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-2">
                        <User size={14} className="text-slate-500" />
                        <span className="text-sm text-slate-200">{s.user_id}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <span className={`px-2 py-0.5 rounded-full text-[10px] font-bold uppercase ${
                        s.role === 'admin' ? 'bg-purple-500/20 text-purple-400 border border-purple-500/30' :
                        s.role === 'operator' ? 'bg-blue-500/20 text-blue-400 border border-blue-500/30' :
                        'bg-slate-500/20 text-slate-400 border border-slate-500/30'
                      }`}>
                        {s.role}
                      </span>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-2">
                        <div className={`w-1.5 h-1.5 rounded-full ${s.status === 'active' ? 'bg-green-500 animate-pulse' : 'bg-yellow-500'}`} />
                        <span className={`text-xs capitalize ${s.status === 'active' ? 'text-green-500' : 'text-yellow-500'}`}>{s.status}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4 text-xs text-slate-400">
                      <span className="flex items-center gap-1.5"><Clock size={12} /> {formatDuration(s.duration_sec)}</span>
                    </td>
                    <td className="px-6 py-4 text-right">
                      <button 
                        onClick={() => terminateSession(s.id)}
                        className="p-2 text-slate-500 hover:text-red-500 hover:bg-red-500/10 rounded-lg transition-all"
                        title="Terminate Session"
                      >
                        <XCircle size={18} />
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {recordings.length === 0 ? (
            <div className="col-span-full py-20 bg-white/5 rounded-3xl border border-white/10 border-dashed flex flex-col items-center justify-center text-slate-500">
               <HardDrive size={40} className="mb-4 opacity-20" />
               <p>No recordings found in archive.</p>
            </div>
          ) : (
            recordings.map(rec => (
              <div key={rec} className="p-6 bg-white/5 border border-white/10 rounded-2xl hover:border-blue-500/30 transition-all group relative overflow-hidden">
                 <div className="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity">
                    <History size={60} />
                 </div>
                 <div className="flex items-start justify-between mb-4">
                    <div className="w-10 h-10 bg-blue-500/10 rounded-xl flex items-center justify-center border border-blue-500/20">
                      <Activity size={20} className="text-blue-500" />
                    </div>
                 </div>
                 <h4 className="text-sm font-bold text-white mb-1 truncate">{rec}</h4>
                 <p className="text-[10px] text-slate-500 font-mono mb-6">FORMAT: .VTR (VT-RECORDING)</p>
                 
                 <button 
                  onClick={() => setSelectedRecording(rec)}
                  className="w-full py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-xl text-xs font-bold flex items-center justify-center gap-2 transition-all shadow-lg shadow-blue-500/10 group-hover:shadow-blue-500/20"
                 >
                    <Play size={14} fill="currentColor" /> Play Recording
                 </button>
              </div>
            ))
          )}
        </div>
      )}

      {selectedRecording && (
        <ReplayPlayer 
          token={token} 
          recordingId={selectedRecording} 
          onClose={() => setSelectedRecording(null)} 
        />
      )}
    </div>
  );
};

export default Dashboard;
