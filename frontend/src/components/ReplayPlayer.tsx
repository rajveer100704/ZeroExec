import React, { useEffect, useRef, useState } from 'react';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import { Play, Pause, RotateCcw, X } from 'lucide-react';
import '@xterm/xterm/css/xterm.css';

interface RecordEntry {
  t: number;
  d: string;
}

interface ReplayPlayerProps {
  token: string;
  recordingId: string;
  onClose: () => void;
}

const ReplayPlayer: React.FC<ReplayPlayerProps> = ({ token, recordingId, onClose }) => {
  const terminalRef = useRef<HTMLDivElement>(null);
  const xtermRef = useRef<Terminal | null>(null);
  const [entries, setEntries] = useState<RecordEntry[]>([]);
  const [currentTime, setCurrentTime] = useState(0);
  const [isPlaying, setIsPlaying] = useState(false);
  const [playbackSpeed, setPlaybackSpeed] = useState(1);
  const [duration, setDuration] = useState(0);
  const timerRef = useRef<number | null>(null);

  useEffect(() => {
    const fetchRecording = async () => {
      try {
        const res = await fetch(`https://localhost:8081/admin/recording?token=${token}&id=${recordingId}`);
        const text = await res.text();
        const parsed = text.trim().split('\n').map(l => JSON.parse(l));
        setEntries(parsed);
        if (parsed.length > 0) {
          setDuration(parsed[parsed.length - 1].t);
        }
      } catch (err) {
        console.error('Failed to load recording:', err);
      }
    };
    fetchRecording();
  }, [token, recordingId]);

  useEffect(() => {
    if (!terminalRef.current) return;

    const term = new Terminal({
      theme: { background: '#0f172a' },
      cursorBlink: false,
      disableStdin: true,
      fontSize: 13,
      fontFamily: 'JetBrains Mono, monospace',
    });
    const fitAddon = new FitAddon();
    term.loadAddon(fitAddon);
    term.open(terminalRef.current);
    fitAddon.fit();
    xtermRef.current = term;

    return () => term.dispose();
  }, []);

  const renderToTime = (targetTime: number) => {
    if (!xtermRef.current) return;
    xtermRef.current.reset();
    
    for (const entry of entries) {
      if (entry.t <= targetTime) {
        xtermRef.current.write(entry.d);
      } else {
        break;
      }
    }
    setCurrentTime(targetTime);
  };

  useEffect(() => {
    if (isPlaying) {
      const interval = 50; // 20fps
      timerRef.current = window.setInterval(() => {
        setCurrentTime(prev => {
          const next = prev + (interval / 1000) * playbackSpeed;
          if (next >= duration) {
            setIsPlaying(false);
            return duration;
          }
          // Optimization: Only write new data
          const relevant = entries.filter(e => e.t > prev && e.t <= next);
          relevant.forEach(e => xtermRef.current?.write(e.d));
          return next;
        });
      }, interval);
    } else {
      if (timerRef.current) clearInterval(timerRef.current);
    }
    return () => { if (timerRef.current) clearInterval(timerRef.current); };
  }, [isPlaying, playbackSpeed, duration, entries]);

  const handleSeek = (e: React.ChangeEvent<HTMLInputElement>) => {
    const time = parseFloat(e.target.value);
    renderToTime(time);
  };

  return (
    <div className="fixed inset-0 z-[100] bg-black/80 backdrop-blur-md flex items-center justify-center p-8 animate-in fade-in duration-300">
      <div className="bg-[#020617] border border-white/10 rounded-3xl w-full max-w-5xl overflow-hidden shadow-2xl flex flex-col h-[80vh]">
        {/* Header */}
        <div className="px-6 py-4 border-b border-white/5 flex items-center justify-between bg-white/[0.02]">
           <div className="flex items-center gap-3">
              <div className="w-8 h-8 bg-blue-500/20 rounded-lg flex items-center justify-center">
                 <RotateCcw size={16} className="text-blue-500" />
              </div>
              <div>
                 <h3 className="text-sm font-bold text-white">Replay: {recordingId}</h3>
                 <p className="text-[10px] text-slate-500 uppercase tracking-widest font-mono">Investigative Playback Mode</p>
              </div>
           </div>
           <button onClick={onClose} className="p-2 hover:bg-white/5 rounded-xl transition-all text-slate-400 hover:text-white">
              <X size={20} />
           </button>
        </div>

        {/* Terminal Area */}
        <div className="flex-1 p-6 overflow-hidden">
           <div ref={terminalRef} className="w-full h-full rounded-xl overflow-hidden border border-white/5 bg-[#0f172a]" />
        </div>

        {/* Controls */}
        <div className="px-8 py-6 bg-white/[0.02] border-t border-white/5 space-y-4">
           {/* Timeline */}
           <div className="relative group">
              <input 
                type="range" 
                min={0} 
                max={duration} 
                step={0.1}
                value={currentTime}
                onChange={handleSeek}
                className="w-full h-1.5 bg-slate-800 rounded-lg appearance-none cursor-pointer accent-blue-500 hover:accent-blue-400 transition-all"
              />
              <div className="flex justify-between mt-2 text-[10px] font-mono text-slate-500">
                 <span>{currentTime.toFixed(1)}s</span>
                 <span>{duration.toFixed(1)}s</span>
              </div>
           </div>

           {/* Buttons */}
           <div className="flex items-center justify-between">
              <div className="flex items-center gap-4">
                 <button 
                  onClick={() => { renderToTime(0); setIsPlaying(false); }}
                  className="p-2 text-slate-400 hover:text-white hover:bg-white/5 rounded-lg transition-all"
                 >
                    <RotateCcw size={18} />
                 </button>
                 <button 
                  onClick={() => setIsPlaying(!isPlaying)}
                  className="w-12 h-12 bg-blue-600 hover:bg-blue-500 text-white rounded-full flex items-center justify-center shadow-lg shadow-blue-500/20 transition-all transform active:scale-95"
                 >
                    {isPlaying ? <Pause size={24} fill="currentColor" /> : <Play size={24} className="ml-1" fill="currentColor" />}
                 </button>
              </div>

              <div className="flex items-center gap-3 bg-white/5 px-4 py-2 rounded-xl border border-white/10">
                 <span className="text-[10px] uppercase font-bold text-slate-500 tracking-tighter">Speed</span>
                 {[0.5, 1, 2, 4].map(s => (
                   <button 
                    key={s}
                    onClick={() => setPlaybackSpeed(s)}
                    className={`text-xs font-mono px-2 py-0.5 rounded ${playbackSpeed === s ? 'bg-blue-500 text-white' : 'text-slate-400 hover:text-white'}`}
                   >
                     {s}x
                   </button>
                 ))}
              </div>
           </div>
        </div>
      </div>
    </div>
  );
};

export default ReplayPlayer;
