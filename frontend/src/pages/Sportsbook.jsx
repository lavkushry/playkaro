import { useQuery } from "@apollo/client";
import { useState } from "react";
import { GET_MATCHES } from "../graphql/queries";
import { useWalletStore } from "../store/useWalletStore";

export default function Sportsbook() {
  const { loading, error, data } = useQuery(GET_MATCHES);
  const [selectedMatch, setSelectedMatch] = useState(null);
  const [betAmount, setBetAmount] = useState("");
  const { balance, updateBalance } = useWalletStore();

  const matches = data?.matches || [];

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

  if (loading) return (
    <div className="min-h-screen bg-primary flex items-center justify-center">
      <div className="text-accent-gold text-xl">Loading matches...</div>
    </div>
  );

  if (error) return (
    <div className="min-h-screen bg-primary flex items-center justify-center">
      <div className="text-status-error text-xl">Error loading matches: {error.message}</div>
    </div>
  );

  return (
    <div className="min-h-screen bg-primary p-4 pb-20">
      <div className="max-w-7xl mx-auto">
        <h1 className="text-3xl font-bold text-accent-gold mb-4">Sportsbook</h1>
        <p className="text-text-secondary mb-8">Powered by GraphQL ⚡</p>

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
                <div className="flex gap-2">
                  <span className="text-accent-gold font-bold">{match.oddsA}</span>
                  <span className="text-text-secondary">-</span>
                  <span className="text-accent-gold font-bold">{match.oddsB}</span>
                </div>
              </div>

              {/* Expanded View */}
              {selectedMatch?.id === match.id && (
                <div className="p-4 space-y-4 animate-slide-in">
                  <div className="border-b border-tertiary pb-4">
                    <h4 className="text-sm text-text-secondary mb-3">Match Winner</h4>
                    <div className="grid grid-cols-2 gap-4">
                      <button
                        onClick={() => {
                          const amount = prompt(`Bet on ${match.teamA} @ ${match.oddsA}\\nEnter Amount:`);
                          if (amount) {
                            setBetAmount(amount);
                            placeBet(match.id, "Match Winner", match.teamA, match.oddsA);
                          }
                        }}
                        className="bg-primary hover:bg-tertiary border border-tertiary rounded-lg p-3 flex justify-between items-center transition-all group"
                      >
                        <span className="font-medium text-text-primary">{match.teamA}</span>
                        <span className="font-bold text-accent-gold group-hover:scale-110 transition-transform">{match.oddsA}</span>
                      </button>
                      <button
                        onClick={() => {
                          const amount = prompt(`Bet on ${match.teamB} @ ${match.oddsB}\\nEnter Amount:`);
                          if (amount) {
                            setBetAmount(amount);
                            placeBet(match.id, "Match Winner", match.teamB, match.oddsB);
                          }
                        }}
                        className="bg-primary hover:bg-tertiary border border-tertiary rounded-lg p-3 flex justify-between items-center transition-all group"
                      >
                        <span className="font-medium text-text-primary">{match.teamB}</span>
                        <span className="font-bold text-accent-gold group-hover:scale-110 transition-transform">{match.oddsB}</span>
                      </button>
                    </div>
                  </div>
                  {match.oddsDraw > 0 && (
                    <div>
                      <h4 className="text-sm text-text-secondary mb-3">Draw</h4>
                      <button
                        onClick={() => {
                          const amount = prompt(`Bet on Draw @ ${match.oddsDraw}\\nEnter Amount:`);
                          if (amount) {
                            setBetAmount(amount);
                            placeBet(match.id, "Draw", "Draw", match.oddsDraw);
                          }
                        }}
                        className="w-full bg-primary hover:bg-tertiary border border-tertiary rounded-lg p-3 flex justify-between items-center transition-all group"
                      >
                        <span className="font-medium text-text-primary">Draw</span>
                        <span className="font-bold text-accent-gold group-hover:scale-110 transition-transform">{match.oddsDraw}</span>
                      </button>
                    </div>
                  )}
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
