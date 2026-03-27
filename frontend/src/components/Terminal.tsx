import React, { useEffect, useRef, useState } from 'react';
import { Terminal as Xterm } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import { WebglAddon } from '@xterm/addon-webgl';
import '@xterm/xterm/css/xterm.css';

interface TerminalProps {
  token: string;
  onClose?: () => void;
  onOutput?: (data: string) => void;
}

const Terminal: React.FC<TerminalProps> = ({ token, onClose, onOutput }) => {
  const terminalRef = useRef<HTMLDivElement>(null);
  const xtermRef = useRef<Xterm | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const [status, setStatus] = useState<'connecting' | 'connected' | 'error' | 'closed'>('connecting');

  useEffect(() => {
    if (!terminalRef.current) return;

    // 1. Initialize Xterm
    const term = new Xterm({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: '"JetBrains Mono", "Fira Code", monospace',
      theme: {
        background: '#0d1117',
        foreground: '#c9d1d9',
        cursor: '#58a6ff',
        selectionBackground: 'rgba(88, 166, 255, 0.3)',
      },
      allowProposedApi: true,
    });
    xtermRef.current = term;

    const fitAddon = new FitAddon();
    term.loadAddon(fitAddon);

    term.open(terminalRef.current);
    
    try {
        const webglAddon = new WebglAddon();
        term.loadAddon(webglAddon);
    } catch (e) {
        console.warn('WebGL addon failed to load, falling back to canvas', e);
    }

    fitAddon.fit();

    // 2. Setup WebSocket (Point to Gateway on 8081)
    const wsUrl = `ws://127.0.0.1:8081/ws?token=${token}`;
    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      setStatus('connected');
      term.focus();
    };

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        if (msg.type === 'output') {
          term.write(msg.data);
          onOutput?.(msg.data);
        } else if (msg.type === 'error') {
          term.write(`\r\n\x1b[31m[ERROR] ${msg.data}\x1b[0m\r\n`);
        }
      } catch (e) {
        console.error('Failed to parse WS message:', e);
      }
    };

    ws.onerror = () => {
      setStatus('error');
    };

    ws.onclose = () => {
      setStatus('closed');
      onClose?.();
    };

    term.onData((data) => {
      if (ws.readyState === WebSocket.OPEN) {
        // Send as JSON
        ws.send(JSON.stringify({
          type: 'input',
          data: data
        }));
      }
    });

    const handleResize = () => {
      fitAddon.fit();
      // We should also send resize to backend, but skipping for MVP
    };
    window.addEventListener('resize', handleResize);

    return () => {
      window.removeEventListener('resize', handleResize);
      term.dispose();
      ws.close();
    };
  }, [token, onClose]);

  return (
    <div className="relative w-full h-full bg-[#0d1117] rounded-lg overflow-hidden border border-white/10 shadow-2xl">
      <div className="absolute top-2 right-4 z-10 flex items-center gap-2">
        <div className={`w-2 h-2 rounded-full ${status === 'connected' ? 'bg-green-500 animate-pulse' : 'bg-red-500'}`} />
        <span className="text-[10px] uppercase font-bold text-white/40 tracking-widest">{status}</span>
      </div>
      <div ref={terminalRef} className="w-full h-full p-2" />
    </div>
  );
};

export default Terminal;
