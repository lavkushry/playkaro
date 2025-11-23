import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "../components/ui/Button";
import { useAuthStore } from "../store/useAuthStore";
import { useWalletStore } from "../store/useWalletStore";

export default function Dashboard() {
  const navigate = useNavigate();
  const { balance, bonus, fetchBalance } = useWalletStore();
  const { user, logout } = useAuthStore();
  const [tickerIndex, setTickerIndex] = useState(0);

  useEffect(() => {
    fetchBalance();
    const interval = setInterval(() => {
      setTickerIndex(prev => (prev + 1) % winners.length);
    }, 3000);
    return () => clearInterval(interval);
  }, []);

  const winners = [
    { user: "Rahul_99", amount: 5000, game: "Roulette" },
    { user: "Priya_Win", amount: 1200, game: "Cricket" },
    { user: "Amit_King", amount: 10000, game: "Blackjack" },
    { user: "Sneha_S", amount: 3500, game: "Slots" },
  ];

  const promotions = [
    { title: "Welcome Bonus", desc: "Get ₹100 Free!", color: "from-accent-gold to-yellow-600", action: "/promotions" },
    { title: "IPL 2025", desc: "Bet on Cricket", color: "from-accent-blue to-blue-700", action: "/sportsbook" },
    { title: "Live Casino", desc: "Experience Real Dealers", color: "from-purple-600 to-purple-900", action: "/casino" },
  ];

  const [currentPromo, setCurrentPromo] = useState(0);

  return (
    <div className="min-h-screen bg-primary pb-20">
      {/* Navbar */}
      <nav className="bg-secondary/80 backdrop-blur-md border-b border-tertiary sticky top-0 z-30">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex items-center">
              <h1 className="text-2xl font-bold text-accent-gold tracking-wider">PlayKaro</h1>
            </div>
            <div className="flex items-center gap-4">
              <div className="hidden md:block text-right">
                <p className="text-sm text-text-secondary">Balance</p>
                <p className="font-bold text-accent-gold">₹{balance.toFixed(2)}</p>
              </div>
              <Button variant="outline" onClick={logout}>Logout</Button>
            </div>
          </div>
        </div>
      </nav>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6 space-y-8">

        {/* Live Ticker */}
        <div className="bg-secondary/50 rounded-lg p-2 flex items-center overflow-hidden border border-tertiary">
          <span className="bg-accent-gold text-primary text-xs font-bold px-2 py-1 rounded mr-3">LIVE WINS</span>
          <div className="flex-1 relative h-6">
            {winners.map((win, i) => (
              <div
                key={i}
                className={`absolute top-0 left-0 transition-all duration-500 flex gap-2 text-sm ${i === tickerIndex ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'
                  }`}
              >
                <span className="font-bold text-text-primary">{win.user}</span>
                <span className="text-text-secondary">won</span>
                <span className="font-bold text-status-success">₹{win.amount}</span>
                <span className="text-text-secondary">in {win.game}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Hero Carousel */}
        <div className="relative h-48 md:h-64 rounded-2xl overflow-hidden shadow-2xl group">
          <div
            className={`absolute inset-0 bg-gradient-to-r ${promotions[currentPromo].color} transition-colors duration-1000`}
          />
          <div className="absolute inset-0 flex flex-col justify-center items-start p-8 md:p-12">
            <h2 className="text-3xl md:text-5xl font-bold text-white mb-2 animate-slide-in">
              {promotions[currentPromo].title}
            </h2>
            <p className="text-white/90 text-lg md:text-xl mb-6 animate-slide-in">
              {promotions[currentPromo].desc}
            </p>
            <Button
              className="bg-white text-primary hover:bg-white/90 font-bold"
              onClick={() => navigate(promotions[currentPromo].action)}
            >
              Play Now
            </Button>
          </div>

          {/* Carousel Controls */}
          <div className="absolute bottom-4 left-1/2 -translate-x-1/2 flex gap-2">
            {promotions.map((_, i) => (
              <button
                key={i}
                onClick={() => setCurrentPromo(i)}
                className={`w-2 h-2 rounded-full transition-all ${i === currentPromo ? 'bg-white w-6' : 'bg-white/50'
                  }`}
              />
            ))}
          </div>
        </div>

        {/* Quick Actions */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div
            onClick={() => navigate("/payment")}
            className="bg-secondary p-4 rounded-xl border border-tertiary hover:border-accent-gold transition-all cursor-pointer group"
          >
            <div className="w-10 h-10 rounded-full bg-accent-gold/20 flex items-center justify-center mb-3 group-hover:bg-accent-gold/30">
              💰
            </div>
            <h3 className="font-semibold">Deposit</h3>
            <p className="text-xs text-text-secondary">Instant Add Money</p>
          </div>
          <div
            onClick={() => navigate("/kyc")}
            className="bg-secondary p-4 rounded-xl border border-tertiary hover:border-accent-blue transition-all cursor-pointer group"
          >
            <div className="w-10 h-10 rounded-full bg-accent-blue/20 flex items-center justify-center mb-3 group-hover:bg-accent-blue/30">
              🛡️
            </div>
            <h3 className="font-semibold">Verify KYC</h3>
            <p className="text-xs text-text-secondary">Secure Account</p>
          </div>
          <div
            onClick={() => navigate("/promotions")}
            className="bg-secondary p-4 rounded-xl border border-tertiary hover:border-purple-500 transition-all cursor-pointer group"
          >
            <div className="w-10 h-10 rounded-full bg-purple-500/20 flex items-center justify-center mb-3 group-hover:bg-purple-500/30">
              🎁
            </div>
            <h3 className="font-semibold">Bonuses</h3>
            <p className="text-xs text-text-secondary">Claim Rewards</p>
          </div>
          <div
            onClick={() => navigate("/analytics")}
            className="bg-secondary p-4 rounded-xl border border-tertiary hover:border-pink-500 transition-all cursor-pointer group"
          >
            <div className="w-10 h-10 rounded-full bg-pink-500/20 flex items-center justify-center mb-3 group-hover:bg-pink-500/30">
              📊
            </div>
            <h3 className="font-semibold">Analytics</h3>
            <p className="text-xs text-text-secondary">Track Stats</p>
          </div>
        </div>

        {/* Featured Games */}
        <div>
          <h2 className="text-xl font-bold text-text-primary mb-4">Featured Games</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            <div className="bg-secondary rounded-xl overflow-hidden border border-tertiary group cursor-pointer" onClick={() => navigate("/sportsbook")}>
              <div className="h-40 bg-gradient-to-br from-blue-900 to-blue-600 relative p-6 flex flex-col justify-end">
                <span className="absolute top-4 right-4 bg-red-500 text-white text-xs font-bold px-2 py-1 rounded animate-pulse">LIVE</span>
                <h3 className="text-2xl font-bold text-white">Cricket</h3>
                <p className="text-blue-100">India vs Australia</p>
              </div>
              <div className="p-4 flex justify-between items-center">
                <span className="text-text-secondary">Next Match: 2m</span>
                <span className="text-accent-gold font-bold">Odds: 1.85</span>
              </div>
            </div>

            <div className="bg-secondary rounded-xl overflow-hidden border border-tertiary group cursor-pointer" onClick={() => navigate("/casino")}>
              <div className="h-40 bg-gradient-to-br from-red-900 to-red-600 relative p-6 flex flex-col justify-end">
                <h3 className="text-2xl font-bold text-white">Roulette</h3>
                <p className="text-red-100">Live Dealer</p>
              </div>
              <div className="p-4 flex justify-between items-center">
                <span className="text-text-secondary">Evolution Gaming</span>
                <span className="text-accent-gold font-bold">RTP: 97.3%</span>
              </div>
            </div>

            <div className="bg-secondary rounded-xl overflow-hidden border border-tertiary group cursor-pointer" onClick={() => navigate("/casino")}>
              <div className="h-40 bg-gradient-to-br from-yellow-900 to-yellow-600 relative p-6 flex flex-col justify-end">
                <h3 className="text-2xl font-bold text-white">Blackjack</h3>
                <p className="text-yellow-100">High Stakes</p>
              </div>
              <div className="p-4 flex justify-between items-center">
                <span className="text-text-secondary">Pragmatic Play</span>
                <span className="text-accent-gold font-bold">Min: ₹500</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
