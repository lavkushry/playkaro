import axios from "axios";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "../components/ui/Button";

const API_URL = 'http://localhost:8080/api/v1';

export default function Casino() {
  const navigate = useNavigate();
  const [games, setGames] = useState([]);
  const [filteredGames, setFilteredGames] = useState([]);
  const [providerFilter, setProviderFilter] = useState("ALL");
  const [typeFilter, setTypeFilter] = useState("ALL");
  const [launchedGame, setLaunchedGame] = useState(null);
  const token = localStorage.getItem("token");

  useEffect(() => {
    if (!token) {
      navigate("/login");
      return;
    }
    fetchGames();
  }, [token, navigate]);

  useEffect(() => {
    applyFilters();
  }, [games, providerFilter, typeFilter]);

  const fetchGames = async () => {
    try {
      const response = await axios.get(`${API_URL}/casino/games`);
      setGames(response.data || []);
    } catch (err) {
      console.error(err);
    }
  };

  const applyFilters = () => {
    let filtered = games;

    if (providerFilter !== "ALL") {
      filtered = filtered.filter(g => g.provider_id === providerFilter);
    }

    if (typeFilter !== "ALL") {
      filtered = filtered.filter(g => g.type === typeFilter);
    }

    setFilteredGames(filtered);
  };

  const handleLaunchGame = async (gameId) => {
    try {
      const response = await axios.get(`${API_URL}/casino/launch`, {
        headers: { Authorization: `Bearer ${token}` },
        params: { game_id: gameId }
      });
      setLaunchedGame(response.data);
    } catch (err) {
      alert("Failed to launch game: " + (err.response?.data?.error || err.message));
    }
  };

  const providers = [...new Set(games.map(g => g.provider_id))];
  const types = [...new Set(games.map(g => g.type))];

  return (
    <div className="min-h-screen bg-primary p-6">
      <div className="max-w-7xl mx-auto space-y-8">
        {/* Header */}
        <div className="flex justify-between items-center">
          <h1 className="text-3xl font-bold text-accent-gold">🎰 Live Casino</h1>
          <Button variant="outline" onClick={() => navigate("/dashboard")}>Dashboard</Button>
        </div>

        {/* Filters */}
        <div className="bg-secondary p-6 rounded-xl border border-tertiary">
          <div className="flex flex-wrap gap-4">
            <div className="flex-1 min-w-[200px]">
              <label className="block text-sm font-medium text-text-secondary mb-2">Provider</label>
              <select
                value={providerFilter}
                onChange={(e) => setProviderFilter(e.target.value)}
                className="w-full px-4 py-3 bg-tertiary border border-tertiary rounded-lg text-text-primary focus:ring-2 focus:ring-accent-gold"
              >
                <option value="ALL">All Providers</option>
                {providers.map(p => (
                  <option key={p} value={p}>{p}</option>
                ))}
              </select>
            </div>

            <div className="flex-1 min-w-[200px]">
              <label className="block text-sm font-medium text-text-secondary mb-2">Game Type</label>
              <select
                value={typeFilter}
                onChange={(e) => setTypeFilter(e.target.value)}
                className="w-full px-4 py-3 bg-tertiary border border-tertiary rounded-lg text-text-primary focus:ring-2 focus:ring-accent-gold"
              >
                <option value="ALL">All Types</option>
                {types.map(t => (
                  <option key={t} value={t}>{t.replace('_', ' ')}</option>
                ))}
              </select>
            </div>
          </div>
        </div>

        {/* Game Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          {filteredGames.map((game) => (
            <div key={game.id} className="bg-secondary rounded-xl border border-tertiary overflow-hidden hover:border-accent-gold transition-colors group">
              <div className="relative aspect-video bg-tertiary">
                <img
                  src={game.thumbnail_url}
                  alt={game.name}
                  className="w-full h-full object-cover"
                />
                <div className="absolute inset-0 bg-black/50 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
                  <Button
                    onClick={() => handleLaunchGame(game.id)}
                    className="bg-accent-gold hover:bg-accent-gold/90 text-primary font-bold"
                  >
                    Play Now
                  </Button>
                </div>
              </div>

              <div className="p-4 space-y-2">
                <h3 className="font-semibold text-text-primary">{game.name}</h3>
                <div className="flex items-center justify-between text-sm">
                  <span className="text-text-secondary">{game.provider_id}</span>
                  <span className="text-accent-gold font-medium">RTP {game.rtp}%</span>
                </div>
                <div className="flex items-center justify-between text-xs text-text-secondary">
                  <span>Min: ₹{game.min_bet}</span>
                  <span>Max: ₹{game.max_bet}</span>
                </div>
              </div>
            </div>
          ))}
        </div>

        {filteredGames.length === 0 && (
          <div className="text-center py-16">
            <p className="text-text-secondary text-lg">No games found</p>
          </div>
        )}
      </div>

      {/* Game Launcher Modal */}
      {launchedGame && (
        <div className="fixed inset-0 bg-black/90 flex items-center justify-center p-4 z-50">
          <div className="bg-secondary rounded-2xl w-full max-w-6xl h-[90vh] flex flex-col">
            <div className="p-4 border-b border-tertiary flex justify-between items-center">
              <div>
                <p className="text-text-primary font-semibold">Game Session</p>
                <p className="text-text-secondary text-sm">Balance: ₹{launchedGame.balance.toFixed(2)}</p>
              </div>
              <Button
                variant="outline"
                onClick={() => setLaunchedGame(null)}
              >
                Close Game
              </Button>
            </div>
            <div className="flex-1 p-4">
              <iframe
                src={launchedGame.game_url}
                className="w-full h-full rounded-lg border border-tertiary"
                title="Game"
                allowFullScreen
              />
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
