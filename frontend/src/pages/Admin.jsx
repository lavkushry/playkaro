import axios from "axios";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "../components/ui/Button";
import { Input } from "../components/ui/Input";

const API_URL = 'http://localhost:8080/api/v1';

export default function Admin() {
  const navigate = useNavigate();
  const token = localStorage.getItem("token");
  const [matches, setMatches] = useState([]);
  const [showCreate, setShowCreate] = useState(false);
  const [newMatch, setNewMatch] = useState({
    team_a: "", team_b: "", odds_a: "1.80", odds_b: "2.10"
  });

  useEffect(() => {
    if (!token) {
      navigate("/login");
      return;
    }
    fetchMatches();
  }, [token, navigate]);

  const fetchMatches = async () => {
    try {
      const response = await axios.get(`${API_URL}/matches`);
      setMatches(response.data || []);
    } catch (err) {
      console.error(err);
    }
  };

  const handleCreateMatch = async (e) => {
    e.preventDefault();
    try {
      await axios.post(`${API_URL}/admin/matches`, newMatch, {
        headers: { Authorization: `Bearer ${token}` }
      });
      alert("Match created!");
      setShowCreate(false);
      fetchMatches();
    } catch (err) {
      alert("Failed: " + (err.response?.data?.error || err.message));
    }
  };

  const handleSettleMatch = async (matchId, winner) => {
    try {
      await axios.post(`${API_URL}/admin/matches/${matchId}/settle`, { winner }, {
        headers: { Authorization: `Bearer ${token}` }
      });
      alert("Match settled!");
      fetchMatches();
    } catch (err) {
      alert("Failed: " + (err.response?.data?.error || err.message));
    }
  };

  return (
    <div className="min-h-screen bg-primary p-6">
      <div className="max-w-6xl mx-auto space-y-8">
        <div className="flex justify-between items-center">
          <h1 className="text-3xl font-bold text-accent-gold">Admin Dashboard</h1>
          <div className="flex gap-4">
            <Button onClick={() => setShowCreate(true)}>Create Match</Button>
            <Button variant="outline" onClick={() => navigate("/dashboard")}>Dashboard</Button>
          </div>
        </div>

        {/* Matches Table */}
        <div className="bg-secondary p-6 rounded-xl border border-tertiary">
          <h2 className="text-xl font-semibold text-text-primary mb-4">Matches</h2>
          <table className="w-full">
            <thead>
              <tr className="border-b border-tertiary">
                <th className="text-left p-2 text-text-secondary">Teams</th>
                <th className="text-left p-2 text-text-secondary">Odds</th>
                <th className="text-left p-2 text-text-secondary">Status</th>
                <th className="text-left p-2 text-text-secondary">Actions</th>
              </tr>
            </thead>
            <tbody>
              {matches.map((match) => (
                <tr key={match.id} className="border-b border-tertiary/50">
                  <td className="p-2 text-text-primary">{match.team_a} vs {match.team_b}</td>
                  <td className="p-2 text-text-secondary">{match.odds_a} / {match.odds_b}</td>
                  <td className="p-2">
                    <span className={`px-2 py-1 rounded text-xs ${match.status === 'LIVE' ? 'bg-status-success text-white' : 'bg-tertiary text-text-secondary'}`}>
                      {match.status}
                    </span>
                  </td>
                  <td className="p-2">
                    {match.status === 'LIVE' && (
                      <div className="flex gap-2">
                        <Button className="text-xs" onClick={() => handleSettleMatch(match.id, 'TEAM_A')}>
                          {match.team_a} Wins
                        </Button>
                        <Button className="text-xs" onClick={() => handleSettleMatch(match.id, 'TEAM_B')}>
                          {match.team_b} Wins
                        </Button>
                      </div>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* Create Match Modal */}
        {showCreate && (
          <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
            <div className="bg-secondary p-8 rounded-2xl border border-tertiary max-w-md w-full">
              <h2 className="text-2xl font-bold text-accent-gold mb-6">Create Match</h2>
              <form onSubmit={handleCreateMatch} className="space-y-4">
                <Input label="Team A" value={newMatch.team_a} onChange={(e) => setNewMatch({...newMatch, team_a: e.target.value})} required />
                <Input label="Team B" value={newMatch.team_b} onChange={(e) => setNewMatch({...newMatch, team_b: e.target.value})} required />
                <Input label="Odds A" type="number" step="0.01" value={newMatch.odds_a} onChange={(e) => setNewMatch({...newMatch, odds_a: e.target.value})} required />
                <Input label="Odds B" type="number" step="0.01" value={newMatch.odds_b} onChange={(e) => setNewMatch({...newMatch, odds_b: e.target.value})} required />
                <div className="flex gap-4">
                  <Button type="submit" className="flex-1">Create</Button>
                  <Button type="button" variant="outline" className="flex-1" onClick={() => setShowCreate(false)}>Cancel</Button>
                </div>
              </form>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
