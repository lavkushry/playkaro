import axios from "axios";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "../components/ui/Button";

const API_URL = 'http://localhost:8080/api/v1';

export default function History() {
  const navigate = useNavigate();
  const token = localStorage.getItem("token");
  const [activeTab, setActiveTab] = useState("transactions");
  const [transactions, setTransactions] = useState([]);
  const [bets, setBets] = useState([]);

  useEffect(() => {
    if (!token) {
      navigate("/login");
      return;
    }
    fetchHistory();
  }, [token, navigate]);

  const fetchHistory = async () => {
    try {
      const [txRes, betsRes] = await Promise.all([
        axios.get(`${API_URL}/transactions`, { headers: { Authorization: `Bearer ${token}` } }),
        axios.get(`${API_URL}/bets`, { headers: { Authorization: `Bearer ${token}` } })
      ]);
      setTransactions(txRes.data || []);
      setBets(betsRes.data || []);
    } catch (err) {
      console.error(err);
    }
  };

  const getStatusColor = (status) => {
    if (status === 'COMPLETED' || status === 'WON') return 'bg-status-success text-white';
    if (status === 'LOST') return 'bg-status-error text-white';
    return 'bg-tertiary text-text-secondary';
  };

  return (
    <div className="min-h-screen bg-primary p-6">
      <div className="max-w-6xl mx-auto space-y-8">
        <div className="flex justify-between items-center">
          <h1 className="text-3xl font-bold text-accent-gold">History</h1>
          <Button variant="outline" onClick={() => navigate("/dashboard")}>Dashboard</Button>
        </div>

        {/* Tabs */}
        <div className="flex gap-4 border-b border-tertiary">
          <button
            onClick={() => setActiveTab("transactions")}
            className={`pb-3 px-4 font-medium transition-colors ${
              activeTab === "transactions"
                ? "text-accent-gold border-b-2 border-accent-gold"
                : "text-text-secondary hover:text-text-primary"
            }`}
          >
            Transactions
          </button>
          <button
            onClick={() => setActiveTab("bets")}
            className={`pb-3 px-4 font-medium transition-colors ${
              activeTab === "bets"
                ? "text-accent-gold border-b-2 border-accent-gold"
                : "text-text-secondary hover:text-text-primary"
            }`}
          >
            Bet History
          </button>
        </div>

        {/* Transactions Tab */}
        {activeTab === "transactions" && (
          <div className="bg-secondary p-6 rounded-xl border border-tertiary">
            <table className="w-full">
              <thead>
                <tr className="border-b border-tertiary">
                  <th className="text-left p-2 text-text-secondary">Type</th>
                  <th className="text-left p-2 text-text-secondary">Amount</th>
                  <th className="text-left p-2 text-text-secondary">Status</th>
                  <th className="text-left p-2 text-text-secondary">Date</th>
                </tr>
              </thead>
              <tbody>
                {transactions.map((tx) => (
                  <tr key={tx.id} className="border-b border-tertiary/50">
                    <td className="p-2 text-text-primary font-medium">{tx.type}</td>
                    <td className="p-2 text-text-primary">₹{tx.amount.toFixed(2)}</td>
                    <td className="p-2">
                      <span className={`px-2 py-1 rounded text-xs ${getStatusColor(tx.status)}`}>
                        {tx.status}
                      </span>
                    </td>
                    <td className="p-2 text-text-secondary text-sm">
                      {new Date(tx.created_at).toLocaleString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {/* Bets Tab */}
        {activeTab === "bets" && (
          <div className="bg-secondary p-6 rounded-xl border border-tertiary">
            <table className="w-full">
              <thead>
                <tr className="border-b border-tertiary">
                  <th className="text-left p-2 text-text-secondary">Match</th>
                  <th className="text-left p-2 text-text-secondary">Selection</th>
                  <th className="text-left p-2 text-text-secondary">Amount</th>
                  <th className="text-left p-2 text-text-secondary">Odds</th>
                  <th className="text-left p-2 text-text-secondary">Potential Win</th>
                  <th className="text-left p-2 text-text-secondary">Status</th>
                </tr>
              </thead>
              <tbody>
                {bets.map((bet) => (
                  <tr key={bet.id} className="border-b border-tertiary/50">
                    <td className="p-2 text-text-primary">{bet.team_a} vs {bet.team_b}</td>
                    <td className="p-2 text-text-secondary">{bet.selection === 'TEAM_A' ? bet.team_a : bet.team_b}</td>
                    <td className="p-2 text-text-primary">₹{bet.amount.toFixed(2)}</td>
                    <td className="p-2 text-accent-gold">{bet.odds}</td>
                    <td className="p-2 text-status-success">₹{bet.potential_win.toFixed(2)}</td>
                    <td className="p-2">
                      <span className={`px-2 py-1 rounded text-xs ${getStatusColor(bet.status)}`}>
                        {bet.status}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
