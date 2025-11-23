import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "../components/ui/Button";
import { Input } from "../components/ui/Input";
import { useBetStore } from "../store/useBetStore";
import { useWalletStore } from "../store/useWalletStore";

export default function Sportsbook() {
  const navigate = useNavigate();
  const { matches, betSlip, fetchMatches, addToBetSlip, updateBetAmount, placeBet, clearBetSlip, connectWebSocket, disconnectWebSocket } = useBetStore();
  const { balance, fetchBalance } = useWalletStore();
  const [showBetSlip, setShowBetSlip] = useState(false);
  const token = localStorage.getItem("token");

  useEffect(() => {
    if (!token) {
      navigate("/login");
      return;
    }
    fetchMatches();
    fetchBalance(token);
    connectWebSocket();

    return () => disconnectWebSocket();
  }, [token, fetchMatches, fetchBalance, connectWebSocket, disconnectWebSocket, navigate]);

  useEffect(() => {
    if (betSlip) setShowBetSlip(true);
  }, [betSlip]);

  const handlePlaceBet = async () => {
    try {
      await placeBet(token);
      await fetchBalance(token);
      alert("Bet placed successfully!");
      setShowBetSlip(false);
    } catch (err) {
      alert("Failed to place bet: " + (err.response?.data?.error || err.message));
    }
  };

  const potentialWin = betSlip ? (betSlip.amount * betSlip.odds).toFixed(2) : 0;

  return (
    <div className="min-h-screen bg-primary p-6">
      <div className="max-w-6xl mx-auto">
        {/* Header */}
        <div className="flex justify-between items-center mb-8">
          <div>
            <h1 className="text-3xl font-bold text-accent-gold">Sportsbook</h1>
            <p className="text-text-secondary">Live Matches</p>
          </div>
          <div className="flex gap-4 items-center">
            <div className="text-right">
              <p className="text-xs text-text-secondary">Balance</p>
              <p className="text-xl font-bold text-accent-gold">₹{balance.toFixed(2)}</p>
            </div>
            <Button variant="outline" onClick={() => navigate("/dashboard")}>
              Dashboard
            </Button>
          </div>
        </div>

        {/* Matches Grid */}
        <div className="grid grid-cols-1 gap-6">
          {matches.map((match) => (
            <div key={match.id} className="bg-secondary p-6 rounded-xl border border-tertiary">
              <div className="flex justify-between items-center">
                <div className="flex-1">
                  <h3 className="text-xl font-semibold text-text-primary">{match.team_a} vs {match.team_b}</h3>
                  <p className="text-sm text-text-secondary mt-1">
                    {new Date(match.start_time).toLocaleString()}
                  </p>
                </div>
                <div className="flex gap-4">
                  <button
                    onClick={() => addToBetSlip(match, 'TEAM_A')}
                    className="flex flex-col items-center bg-tertiary hover:bg-accent-gold/20 border border-accent-gold/30 rounded-lg px-6 py-3 transition-all"
                  >
                    <span className="text-xs text-text-secondary">{match.team_a}</span>
                    <span className="text-2xl font-bold text-accent-gold">{match.odds_a}</span>
                  </button>
                  <button
                    onClick={() => addToBetSlip(match, 'TEAM_B')}
                    className="flex flex-col items-center bg-tertiary hover:bg-accent-blue/20 border border-accent-blue/30 rounded-lg px-6 py-3 transition-all"
                  >
                    <span className="text-xs text-text-secondary">{match.team_b}</span>
                    <span className="text-2xl font-bold text-accent-blue">{match.odds_b}</span>
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>

        {/* Bet Slip Modal */}
        {showBetSlip && betSlip && (
          <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
            <div className="bg-secondary p-8 rounded-2xl border border-tertiary max-w-md w-full">
              <div className="flex justify-between items-center mb-6">
                <h2 className="text-2xl font-bold text-accent-gold">Bet Slip</h2>
                <button onClick={() => { clearBetSlip(); setShowBetSlip(false); }} className="text-text-secondary hover:text-text-primary">✕</button>
              </div>

              <div className="space-y-4">
                <div className="bg-tertiary p-4 rounded-lg">
                  <p className="text-sm text-text-secondary">Selection</p>
                  <p className="text-xl font-bold text-text-primary">{betSlip.team}</p>
                  <p className="text-sm text-accent-gold">Odds: {betSlip.odds}</p>
                </div>

                <Input
                  label="Bet Amount (₹)"
                  type="number"
                  value={betSlip.amount}
                  onChange={(e) => updateBetAmount(e.target.value)}
                />

                <div className="bg-tertiary p-4 rounded-lg">
                  <div className="flex justify-between">
                    <span className="text-text-secondary">Potential Win</span>
                    <span className="text-xl font-bold text-status-success">₹{potentialWin}</span>
                  </div>
                </div>

                <Button
                  className="w-full bg-status-success hover:bg-status-success/90 text-white"
                  onClick={handlePlaceBet}
                >
                  Place Bet
                </Button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
