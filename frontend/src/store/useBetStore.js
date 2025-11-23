import axios from 'axios';
import { create } from 'zustand';

const API_URL = 'http://localhost:8080/api/v1';

export const useBetStore = create((set, get) => ({
  matches: [],
  betSlip: null,
  isLoading: false,
  ws: null,

  fetchMatches: async () => {
    set({ isLoading: true });
    try {
      const response = await axios.get(`${API_URL}/matches`);
      set({ matches: response.data || [], isLoading: false });
    } catch (error) {
      set({ isLoading: false });
    }
  },

  addToBetSlip: (match, selection) => {
    const odds = selection === 'TEAM_A' ? match.odds_a : match.odds_b;
    const team = selection === 'TEAM_A' ? match.team_a : match.team_b;
    set({
      betSlip: {
        matchId: match.id,
        selection,
        team,
        odds,
        amount: 100,
      },
    });
  },

  updateBetAmount: (amount) => {
    set((state) => ({
      betSlip: state.betSlip ? { ...state.betSlip, amount: parseFloat(amount) } : null,
    }));
  },

  placeBet: async (token) => {
    const { betSlip } = get();
    if (!betSlip) return;

    set({ isLoading: true });
    try {
      await axios.post(
        `${API_URL}/bet/`,
        {
          match_id: betSlip.matchId,
          selection: betSlip.selection,
          amount: betSlip.amount,
        },
        { headers: { Authorization: `Bearer ${token}` } }
      );
      set({ betSlip: null, isLoading: false });
      return true;
    } catch (error) {
      set({ isLoading: false });
      throw error;
    }
  },

  clearBetSlip: () => set({ betSlip: null }),

  connectWebSocket: () => {
    const ws = new WebSocket('ws://localhost:8080/ws');
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === 'ODDS_UPDATE') {
        set((state) => ({
          matches: state.matches.map((m) =>
            m.id === data.match_id
              ? { ...m, odds_a: data.odds_a, odds_b: data.odds_b }
              : m
          ),
        }));
      }
    };
    set({ ws });
  },

  disconnectWebSocket: () => {
    const { ws } = get();
    if (ws) {
      ws.close();
      set({ ws: null });
    }
  },
}));
