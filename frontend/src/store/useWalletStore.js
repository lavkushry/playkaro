import axios from 'axios';
import { create } from 'zustand';

const API_URL = 'http://localhost:8080/api/v1';

export const useWalletStore = create((set) => ({
  balance: 0,
  currency: 'INR',
  isLoading: false,
  error: null,

  fetchBalance: async (token) => {
    set({ isLoading: true });
    try {
      const response = await axios.get(`${API_URL}/wallet/`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      set({
        balance: response.data.balance,
        currency: response.data.currency,
        isLoading: false
      });
    } catch (error) {
      set({ error: error.message, isLoading: false });
    }
  },

  deposit: async (token, amount) => {
    set({ isLoading: true });
    try {
      // Optimistic Update
      set((state) => ({ balance: state.balance + parseFloat(amount) }));

      const response = await axios.post(
        `${API_URL}/wallet/deposit`,
        { amount: parseFloat(amount) },
        { headers: { Authorization: `Bearer ${token}` } }
      );

      // Sync with server response
      set({ balance: response.data.new_balance, isLoading: false });
    } catch (error) {
      // Rollback on error
      set((state) => ({
        balance: state.balance - parseFloat(amount),
        error: "Deposit failed",
        isLoading: false
      }));
    }
  },

  withdraw: async (token, amount) => {
    set({ isLoading: true });
    try {
      // Optimistic Update
      set((state) => ({ balance: state.balance - parseFloat(amount) }));

      const response = await axios.post(
        `${API_URL}/wallet/withdraw`,
        { amount: parseFloat(amount) },
        { headers: { Authorization: `Bearer ${token}` } }
      );

      set({ balance: response.data.new_balance, isLoading: false });
    } catch (error) {
      // Rollback
      set((state) => ({
        balance: state.balance + parseFloat(amount),
        error: "Withdraw failed",
        isLoading: false
      }));
    }
  },
}));
