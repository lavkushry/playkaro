import { useEffect, useRef, useState } from "react";

export default function Chat() {
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState("");
  const [isOpen, setIsOpen] = useState(false);
  const ws = useRef(null);
  const messagesEndRef = useRef(null);
  const username = localStorage.getItem("username") || "Guest"; // Simple username for now

  useEffect(() => {
    // Connect to WebSocket
    ws.current = new WebSocket("ws://localhost:8080/ws");

    ws.current.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === "chat_message") {
        setMessages((prev) => [...prev, data.payload]);
      }
    };

    return () => {
      if (ws.current) ws.current.close();
    };
  }, []);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  const sendMessage = (e) => {
    e.preventDefault();
    if (!input.trim()) return;

    const message = {
      type: "chat_message",
      payload: {
        username: username,
        message: input,
      },
    };

    ws.current.send(JSON.stringify(message));
    setInput("");
  };

  if (!isOpen) {
    return (
      <button
        onClick={() => setIsOpen(true)}
        className="fixed bottom-20 right-4 bg-accent-gold text-primary p-4 rounded-full shadow-lg hover:bg-accent-gold/90 transition-all z-40"
      >
        💬
      </button>
    );
  }

  return (
    <div className="fixed bottom-20 right-4 w-80 h-96 bg-secondary border border-tertiary rounded-xl shadow-2xl flex flex-col z-40 animate-slide-in">
      {/* Header */}
      <div className="p-3 border-b border-tertiary flex justify-between items-center bg-tertiary/50 rounded-t-xl">
        <h3 className="font-bold text-accent-gold">Global Chat</h3>
        <button onClick={() => setIsOpen(false)} className="text-text-secondary hover:text-text-primary">✕</button>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-4 space-y-3">
        {messages.map((msg, i) => (
          <div key={i} className={`flex flex-col ${msg.username === username ? 'items-end' : 'items-start'}`}>
            <span className="text-xs text-text-secondary mb-1">{msg.username}</span>
            <div className={`px-3 py-2 rounded-lg max-w-[80%] ${
              msg.username === username
                ? 'bg-accent-gold text-primary rounded-tr-none'
                : 'bg-tertiary text-text-primary rounded-tl-none'
            }`}>
              {msg.message}
            </div>
          </div>
        ))}
        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <form onSubmit={sendMessage} className="p-3 border-t border-tertiary flex gap-2">
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="Say something..."
          className="flex-1 bg-primary border border-tertiary rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-accent-gold"
        />
        <button
          type="submit"
          className="bg-accent-gold text-primary px-3 py-2 rounded-lg font-bold hover:bg-accent-gold/90"
        >
          ➤
        </button>
      </form>
    </div>
  );
}
