import axios from "axios";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "../components/ui/Button";

const API_URL = 'http://localhost:8080/api/v1';

export default function Analytics() {
  const navigate = useNavigate();
  const [stats, setStats] = useState({
    totalBets: 0,
    totalWagered: 0,
    totalWon: 0,
    winRate: 0,
    profitLoss: 0,
  });
  const token = localStorage.getItem("token");

  useEffect(() => {
    if (!token) {
      navigate("/login");
      return;
    }
    fetchAnalytics();
  }, [token, navigate]);

  const fetchAnalytics = async () => {
    try {
      const betsResponse = await axios.get(`${API_URL}/bets`, {
        headers: { Authorization: `Bearer ${token}` }
      });

      const bets = betsResponse.data || [];

      const totalBets = bets.length;
      const totalWagered = bets.reduce((sum, bet) => sum + bet.amount, 0);
      const wonBets = bets.filter(b => b.status === 'WON');
      const totalWon = wonBets.reduce((sum, bet) => sum + bet.payout, 0);
      const winRate = totalBets > 0 ? (wonBets.length / totalBets) * 100 : 0;
      const profitLoss = totalWon - totalWagered;

      setStats({
        totalBets,
        totalWagered,
        totalWon,
        winRate,
        profitLoss,
      });
    } catch (err) {
      console.error(err);
    }
  };

  return (
    <div className="min-h-screen bg-primary p-6 pb-20">
      <div className="max-w-5xl mx-auto space-y-8">
        <div className="flex justify-between items-center">
          <h1 className="text-3xl font-bold text-accent-gold">📊 Analytics</h1>
          <Button variant="outline" onClick={() => navigate("/dashboard")}>Dashboard</Button>
        </div>

        {/* Key Metrics */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <div className="bg-gradient-to-br from-accent-blue/20 to-secondary p-6 rounded-xl border border-accent-blue">
            <p className="text-text-secondary text-sm mb-2">Total Bets</p>
            <p className="text-4xl font-bold text-accent-blue">{stats.totalBets}</p>
          </div>

          <div className="bg-gradient-to-br from-accent-gold/20 to-secondary p-6 rounded-xl border border-accent-gold">
            <p className="text-text-secondary text-sm mb-2">Total Wagered</p>
            <p className="text-4xl font-bold text-accent-gold">₹{stats.totalWagered.toFixed(2)}</p>
          </div>

          <div className="bg-gradient-to-br from-status-success/20 to-secondary p-6 rounded-xl border border-status-success">
            <p className="text-text-secondary text-sm mb-2">Win Rate</p>
            <p className="text-4xl font-bold text-status-success">{stats.winRate.toFixed(1)}%</p>
          </div>

          <div className="bg-gradient-to-br from-purple-500/20 to-secondary p-6 rounded-xl border border-purple-500">
            <p className="text-text-secondary text-sm mb-2">Total Won</p>
            <p className="text-4xl font-bold text-purple-400">₹{stats.totalWon.toFixed(2)}</p>
          </div>

          <div className={`bg-gradient-to-br ${stats.profitLoss >= 0 ? 'from-status-success/20 border-status-success' : 'from-status-error/20 border-status-error'} to-secondary p-6 rounded-xl border`}>
            <p className="text-text-secondary text-sm mb-2">Profit/Loss</p>
            <p className={`text-4xl font-bold ${stats.profitLoss >= 0 ? 'text-status-success' : 'text-status-error'}`}>
              {stats.profitLoss >= 0 ? '+' : ''}₹{stats.profitLoss.toFixed(2)}
            </p>
          </div>

          <div className="bg-gradient-to-br from-pink-500/20 to-secondary p-6 rounded-xl border border-pink-500">
            <p className="text-text-secondary text-sm mb-2">ROI</p>
            <p className="text-4xl font-bold text-pink-400">
              {stats.totalWagered > 0 ? ((stats.profitLoss / stats.totalWagered) * 100).toFixed(1) : 0}%
            </p>
          </div>
        </div>

        {/* Performance Overview */}
        <div className="bg-secondary p-6 rounded-xl border border-tertiary">
          <h2 className="text-xl font-semibold text-text-primary mb-4">Performance Overview</h2>
          <div className="space-y-4">
            <div>
              <div className="flex justify-between mb-2">
                <span className="text-text-secondary">Betting Activity</span>
                <span className="text-text-primary font-medium">{stats.totalBets} bets</span>
              </div>
              <div className="w-full bg-tertiary rounded-full h-2">
                <div
                  className="bg-accent-gold h-2 rounded-full transition-all"
                  style={{ width: `${Math.min((stats.totalBets / 100) * 100, 100)}%` }}
                ></div>
              </div>
            </div>

            <div>
              <div className="flex justify-between mb-2">
                <span className="text-text-secondary">Win Rate</span>
                <span className="text-text-primary font-medium">{stats.winRate.toFixed(1)}%</span>
              </div>
              <div className="w-full bg-tertiary rounded-full h-2">
                <div
                  className="bg-status-success h-2 rounded-full transition-all"
                  style={{ width: `${stats.winRate}%` }}
                ></div>
              </div>
            </div>
          </div>
        </div>

        {/* Insights */}
        <div className="bg-secondary p-6 rounded-xl border border-tertiary">
          <h2 className="text-xl font-semibold text-text-primary mb-4">Insights</h2>
          <div className="space-y-3 text-text-secondary">
            {stats.winRate > 50 && (
              <div className="flex items-start gap-2 text-status-success">
                <span>✓</span>
                <span>Great job! Your win rate is above 50%</span>
              </div>
            )}
            {stats.profitLoss > 0 && (
              <div className="flex items-start gap-2 text-status-success">
                <span>✓</span>
                <span>You're in profit! Keep up the good work</span>
              </div>
            )}
            {stats.totalBets < 10 && (
              <div className="flex items-start gap-2 text-accent-gold">
                <span>ℹ</span>
                <span>Place more bets to see detailed analytics</span>
              </div>
            )}
            {stats.profitLoss < 0 && (
              <div className="flex items-start gap-2 text-status-error">
                <span>⚠</span>
                <span>You're currently in loss. Bet responsibly</span>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
