import React, { useState, useRef, useEffect } from 'react';
import { Sparkles, Send, X, Loader2, Bot, User, Trash2 } from 'lucide-react';

interface Message {
  role: 'user' | 'assistant';
  content: string;
  timestamp: Date;
}

interface AIAssistantProps {
  token: string;
  terminalContext: string;
  isOpen: boolean;
  onClose: () => void;
}

const AIAssistant: React.FC<AIAssistantProps> = ({ token, terminalContext, isOpen, onClose }) => {
  const [messages, setMessages] = useState<Message[]>([
    {
      role: 'assistant',
      content: 'Hello! I\'m **ZeroExec AI** — your terminal assistant. Ask me about commands, errors, or anything systems-related. I can see your recent terminal output for context.',
      timestamp: new Date(),
    }
  ]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  useEffect(() => {
    if (isOpen) {
      setTimeout(() => inputRef.current?.focus(), 300);
    }
  }, [isOpen]);

  const sendMessage = async () => {
    const trimmed = input.trim();
    if (!trimmed || loading) return;

    const userMsg: Message = { role: 'user', content: trimmed, timestamp: new Date() };
    setMessages(prev => [...prev, userMsg]);
    setInput('');
    setLoading(true);

    try {
      const res = await fetch(`http://127.0.0.1:8081/ai/chat?token=${token}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ message: trimmed, context: terminalContext }),
      });

      const data = await res.json();
      
      let replyError = data.error;
      if (replyError && replyError.includes("not configured")) {
        replyError = "AI unavailable — check API key configuration";
      }
      const reply = replyError || data.reply || 'No response received.';

      setMessages(prev => [...prev, {
        role: 'assistant',
        content: replyError ? `⚠️ ${reply}` : reply,
        timestamp: new Date(),
      }]);
    } catch (err) {
      setMessages(prev => [...prev, {
        role: 'assistant',
        content: '⚠️ Failed to connect to AI service. Is the gateway running?',
        timestamp: new Date(),
      }]);
    } finally {
      setLoading(false);
    }
  };

  const clearChat = () => {
    setMessages([{
      role: 'assistant',
      content: 'Chat cleared. How can I help?',
      timestamp: new Date(),
    }]);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  };

  // Simple markdown renderer for code blocks and bold
  const renderMarkdown = (text: string) => {
    const parts = text.split(/(```[\s\S]*?```|\*\*.*?\*\*|`[^`]+`)/g);
    return parts.map((part, i) => {
      if (part.startsWith('```') && part.endsWith('```')) {
        const code = part.slice(3, -3).replace(/^\w+\n/, '');
        return (
          <pre key={i} className="ai-code-block">
            <code>{code.trim()}</code>
          </pre>
        );
      }
      if (part.startsWith('**') && part.endsWith('**')) {
        return <strong key={i}>{part.slice(2, -2)}</strong>;
      }
      if (part.startsWith('`') && part.endsWith('`')) {
        return <code key={i} className="ai-inline-code">{part.slice(1, -1)}</code>;
      }
      return <span key={i}>{part}</span>;
    });
  };

  return (
    <div className={`ai-panel ${isOpen ? 'ai-panel-open' : 'ai-panel-closed'}`}>
      {/* Header */}
      <div className="ai-header">
        <div className="ai-header-title">
          <div className="ai-header-icon">
            <Sparkles size={16} />
          </div>
          <div>
            <h3>ZeroExec AI</h3>
            <p>Contextual Terminal Assistant</p>
          </div>
        </div>
        <div className="ai-header-actions">
          <button onClick={clearChat} className="ai-btn-icon" title="Clear chat">
            <Trash2 size={14} />
          </button>
          <button onClick={onClose} className="ai-btn-icon" title="Close">
            <X size={16} />
          </button>
        </div>
      </div>

      {/* Messages */}
      <div className="ai-messages">
        {messages.map((msg, i) => (
          <div key={i} className={`ai-message ai-message-${msg.role}`}>
            <div className="ai-message-avatar">
              {msg.role === 'assistant' ? (
                <Bot size={14} />
              ) : (
                <User size={14} />
              )}
            </div>
            <div className="ai-message-content">
              {renderMarkdown(msg.content)}
            </div>
          </div>
        ))}
        {loading && (
          <div className="ai-message ai-message-assistant">
            <div className="ai-message-avatar">
              <Bot size={14} />
            </div>
            <div className="ai-message-content ai-typing">
              <Loader2 size={14} className="ai-spinner" />
              <span>Thinking...</span>
            </div>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <div className="ai-input-bar">
        <input
          ref={inputRef}
          type="text"
          value={input}
          onChange={e => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Ask about commands, errors, or output..."
          className="ai-input"
          disabled={loading}
        />
        <button
          onClick={sendMessage}
          disabled={loading || !input.trim()}
          className="ai-send-btn"
        >
          <Send size={16} />
        </button>
      </div>
    </div>
  );
};

export default AIAssistant;
