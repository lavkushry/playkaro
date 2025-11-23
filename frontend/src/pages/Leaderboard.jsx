import axios from "axios";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "../components/ui/Button";

const API_URL = 'http://localhost:8080/api/v1';

export default function Leaderboard() {
  const navigate = useNavigate();
  const [leaderboard, setLeaderboard] = useState([]);
  const [period, setPeriod] = useState("weekly");
  const token = localStorage.getItem("token");

  useEffect(() => {
    if (!token) {
      navigate("/login");
      return;
    }
    fetchLeaderboard();
  }, [token, navigate, period]);

  const fetchLeaderboard = async () => {
    try {
      const response = await axios.get(`${API_URL}/promotions/leaderboard`, {
        params: { period },
        headers: { Authorization: `Bearer ${token}` }
      });
      setLeaderboard(response.data || []);
    } catch (err) {
      console.error(err);
    }
  };

  const getMedalEmoji = (rank) => {
    if (rank === 1) return '🥇';
    if (rank === 2) return '🥈';
    if (rank === 3) return '🥉';
    return `#${rank}`;
  };

  return (
    <div className="min-h-screen bg-primary p-6">
      <div className="max-w-4xl mx-auto space-y-8">
        <div className="flex justify-between items-center">
          <h1 className="text-3xl font-bold text-accent-gold">🏆 Leaderboard</h1>
          <Button variant="outline" onClick={() => navigate("/dashboard")}>Dashboard</Button>
        </div>

        {/* Prize Pool */}
        <div className="bg-gradient-to-r from-accent-gold/20 to-accent-blue/20 p-8 rounded-xl border border-accent-gold">
          <div className="text-center">
            <p className="text-text-secondary text-sm uppercase tracking-wide mb-2">Total Prize Pool</p>
            <p className="text-5xl font-bold text-accent-gold mb-4">₹10,000</p>
            <p className="text-text-secondary">Top 10 players share the prize pool</p>
          </div>
        </div>

        {/* Period Selector */}
        <div className="flex gap-4 justify-center">
          {['daily', 'weekly', 'monthly'].map((p) => (
            <Button
              key={p}
              variant={period === p ? 'primary' : 'outline'}
              onClick={() => setPeriod(p)}
              className="capitalize"
            >
              {p}
            </Button>
          ))}
        </div>

        {/* Leaderboard Table */}
        <div className="bg-secondary rounded-xl border border-tertiary overflow-hidden">
          <table className="w-full">
            <thead className="bg-tertiary">
              <tr>
                <th className="text-left p-4 text-text-secondary font-medium">Rank</th>
                <th className="text-left p-4 text-text-secondary font-medium">Player</th>
                <th className="text-right p-4 text-text-secondary font-medium">Wagered</th>
                <th className="text-right p-4 text-text-secondary font-medium">Bets</th>
              </tr>
            </thead>
            <tbody>
              {leaderboard.map((entry) => (
                <tr key={entry.rank} className="border-b border-tertiary/50 hover:bg-tertiary/30 transition-colors">
                  <td className="p-4">
                    <span className="text-2xl">{getMedalEmoji(entry.rank)}</span>
                  </td>
                  <td className="p-4">
                    <span className="text-text-primary font-medium">{entry.username}</span>
                  </td>
                  <td className="p-4 text-right">
                    <span className="text-accent-gold font-semibold">₹{entry.total_wagered.toFixed(2)}</span>
                  </td>
                  <td className="p-4 text-right">
                    <span className="text-text-secondary">{entry.bet_count}</span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {leaderboard.length === 0 && (
            <div className="text-center py-16">
              <p className="text-text-secondary">No entries yet. Be the first to bet!</p>
            </div>
          )}
        </div>

        {/* How to Win */}
        <div className="bg-secondary p-6 rounded-xl border border-tertiary">
          <h2 className="text-xl font-semibold text-text-primary mb-4">How to Win</h2>
          <ul className="space-y-2 text-text-secondary">
            <li className="flex items-start gap-2">
              <span className="text-accent-gold">•</span>
              <span>Place bets on sports, casino, or slots to climb the leaderboard</span>
            </li>
            <li className="flex items-start gap-2">
              <span className="text-accent-gold">•</span>
              <span>Higher wagered amount = Higher rank</span>
            </li>
            <li className="flex items-start gap-2">
              <span className="text-accent-gold">•</span>
              <span>Top 10 players win cash prizes at the end of the period</span>
            </li>
            <li className="flex items-start gap-2">
              <span className="text-accent-gold">•</span>
              <span>Leaderboard resets at the end of each period</span>
            </li>
          </ul>
        </div>
      </div>
    </div>
  );
}
