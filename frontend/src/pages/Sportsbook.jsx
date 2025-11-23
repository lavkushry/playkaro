import { useEffect, useState } from "react";
import { useWalletStore } from "../store/useWalletStore";

export default function Sportsbook() {
  const [matches, setMatches] = useState([]);
  const [selectedMatch, setSelectedMatch] = useState(null);
  const [betAmount, setBetAmount] = useState("");
  const { balance, updateBalance } = useWalletStore();

  useEffect(() => {
    // Mock data for now, replace with API call
    setMatches([
      {
        id: "1",
        teamA: "India",
        teamB: "Australia",
        date: "2025-11-24T14:30:00Z",
        status: "LIVE",
        markets: [
          { name: "Match Winner", options: [{ label: "India", odds: 1.85 }, { label: "Australia", odds: 1.95 }] },
          { name: "Toss Winner", options: [{ label: "India", odds: 1.90 }, { label: "Australia", odds: 1.90 }] },
          { name: "Total Runs (1st Innings)", options: [{ label: "Over 300.5", odds: 1.85 }, { label: "Under 300.5", odds: 1.85 }] },
        ]
      },
      {
        id: "2",
        teamA: "England",
        teamB: "Pakistan",
        date: "2025-11-24T18:00:00Z",
        status: "UPCOMING",
        markets: [
          { name: "Match Winner", options: [{ label: "England", odds: 1.60 }, { label: "Pakistan", odds: 2.30 }] },
        ]
      }
    ]);
  }, []);

  const placeBet = (matchId, marketName, selection, odds) => {
    const amount = parseFloat(betAmount);
    if (!amount || amount <= 0) return alert("Enter valid amount");
    if (amount > balance) return alert("Insufficient balance");

    // Call API to place bet
    // For now, just update local state
    alert(`Bet Placed: ₹${amount} on ${selection} (${marketName}) @ ${odds}`);
    // updateBalance(-amount);
    setBetAmount("");
    setSelectedMatch(null);
  };

  return (
    <div className="min-h-screen bg-primary p-4 pb-20">
      <div className="max-w-7xl mx-auto">
        <h1 className="text-3xl font-bold text-accent-gold mb-8">Sportsbook</h1>

        {/* Match List */}
        <div className="space-y-6">
          {matches.map((match) => (
            <div key={match.id} className="bg-secondary rounded-xl border border-tertiary overflow-hidden">
              {/* Match Header */}
              <div className="bg-tertiary/50 p-4 flex justify-between items-center cursor-pointer" onClick={() => setSelectedMatch(selectedMatch?.id === match.id ? null : match)}>
                <div className="flex items-center gap-4">
                  <span className={`px-2 py-1 rounded text-xs font-bold ${match.status === 'LIVE' ? 'bg-red-500 text-white animate-pulse' : 'bg-blue-500 text-white'}`}>
                    {match.status}
                  </span>
                  <h3 className="text-xl font-bold text-text-primary">{match.teamA} vs {match.teamB}</h3>
                </div>
                <span className="text-text-secondary text-sm">{new Date(match.date).toLocaleString()}</span>
              </div>

              {/* Markets (Expanded View) */}
              {(selectedMatch?.id === match.id || match.status === 'LIVE') && (
                <div className="p-4 space-y-4 animate-slide-in">
                  {match.markets.map((market, idx) => (
                    <div key={idx} className="border-b border-tertiary last:border-0 pb-4 last:pb-0">
                      <h4 className="text-sm text-text-secondary mb-3">{market.name}</h4>
                      <div className="grid grid-cols-2 gap-4">
                        {market.options.map((option, optIdx) => (
                          <button
                            key={optIdx}
                            onClick={() => {
                              const amount = prompt(`Bet on ${option.label} @ ${option.odds}\nEnter Amount:`);
                              if (amount) {
                                setBetAmount(amount);
                                placeBet(match.id, market.name, option.label, option.odds);
                              }
                            }}
                            className="bg-primary hover:bg-tertiary border border-tertiary rounded-lg p-3 flex justify-between items-center transition-all group"
                          >
                            <span className="font-medium text-text-primary">{option.label}</span>
                            <span className="font-bold text-accent-gold group-hover:scale-110 transition-transform">{option.odds}</span>
                          </button>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
