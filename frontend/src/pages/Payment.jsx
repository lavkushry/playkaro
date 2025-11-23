import axios from "axios";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "../components/ui/Button";
import { Input } from "../components/ui/Input";
import { useWalletStore } from "../store/useWalletStore";

const API_URL = 'http://localhost:8080/api/v1';

export default function Payment() {
  const navigate = useNavigate();
  const { balance, fetchBalance } = useWalletStore();
  const [activeTab, setActiveTab] = useState("deposit");
  const [amount, setAmount] = useState("");
  const [method, setMethod] = useState("MOCK");
  const [loading, setLoading] = useState(false);
  const [kycLevel, setKycLevel] = useState(0);
  const token = localStorage.getItem("token");

  useEffect(() => {
    if (!token) {
      navigate("/login");
      return;
    }
    fetchBalance(token);
    fetchKYCStatus();
  }, [token, navigate, fetchBalance]);

  const fetchKYCStatus = async () => {
    try {
      const response = await axios.get(`${API_URL}/kyc/status`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      setKycLevel(response.data.kyc_level);
    } catch (err) {
      console.error(err);
    }
  };

  const handleDeposit = async (e) => {
    e.preventDefault();
    setLoading(true);
    try {
      const response = await axios.post(`${API_URL}/payment/deposit`, {
        amount: parseFloat(amount),
        method,
        gateway: "MOCK"
      }, {
        headers: { Authorization: `Bearer ${token}` }
      });

      if (response.data.status === "SUCCESS") {
        alert("Deposit successful!");
        fetchBalance(token);
        setAmount("");
      } else {
        alert("Deposit initiated. Please complete payment.");
      }
    } catch (err) {
      alert("Deposit failed: " + (err.response?.data?.error || err.message));
    } finally {
      setLoading(false);
    }
  };

  const handleWithdraw = async (e) => {
    e.preventDefault();
    if (kycLevel < 2) {
      alert("KYC verification required. Please complete KYC first.");
      navigate("/kyc");
      return;
    }
    setLoading(true);
    try {
      await axios.post(`${API_URL}/payment/withdraw`, {
        amount: parseFloat(amount)
      }, {
        headers: { Authorization: `Bearer ${token}` }
      });
      alert("Withdrawal request submitted!");
      fetchBalance(token);
      setAmount("");
    } catch (err) {
      alert("Withdrawal failed: " + (err.response?.data?.error || err.message));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-primary p-6">
      <div className="max-w-2xl mx-auto space-y-8">
        <div className="flex justify-between items-center">
          <h1 className="text-3xl font-bold text-accent-gold">Payment</h1>
          <Button variant="outline" onClick={() => navigate("/dashboard")}>Dashboard</Button>
        </div>

        {/* Balance Display */}
        <div className="bg-secondary p-6 rounded-xl border border-tertiary">
          <p className="text-text-secondary text-sm">Current Balance</p>
          <p className="text-4xl font-bold text-accent-gold">₹{balance.toFixed(2)}</p>
          {kycLevel < 2 && (
            <p className="text-status-error text-sm mt-2">
              ⚠️ Complete KYC to enable withdrawals
            </p>
          )}
        </div>

        {/* Tabs */}
        <div className="flex gap-4 border-b border-tertiary">
          <button
            onClick={() => setActiveTab("deposit")}
            className={`pb-3 px-4 font-medium transition-colors ${
              activeTab === "deposit"
                ? "text-accent-gold border-b-2 border-accent-gold"
                : "text-text-secondary hover:text-text-primary"
            }`}
          >
            Deposit
          </button>
          <button
            onClick={() => setActiveTab("withdraw")}
            className={`pb-3 px-4 font-medium transition-colors ${
              activeTab === "withdraw"
                ? "text-accent-gold border-b-2 border-accent-gold"
                : "text-text-secondary hover:text-text-primary"
            }`}
          >
            Withdraw
          </button>
        </div>

        {/* Deposit Form */}
        {activeTab === "deposit" && (
          <div className="bg-secondary p-8 rounded-xl border border-tertiary">
            <h2 className="text-xl font-semibold text-text-primary mb-6">Add Money</h2>
            <form onSubmit={handleDeposit} className="space-y-6">
              <Input
                label="Amount (₹)"
                type="number"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                placeholder="Enter amount"
                required
              />

              <div>
                <label className="block text-sm font-medium text-text-secondary mb-2">
                  Payment Method
                </label>
                <select
                  value={method}
                  onChange={(e) => setMethod(e.target.value)}
                  className="w-full px-4 py-3 bg-tertiary border border-tertiary rounded-lg text-text-primary focus:ring-2 focus:ring-accent-gold"
                >
                  <option value="MOCK">Mock Payment (Test)</option>
                  <option value="UPI">UPI</option>
                  <option value="CARD">Credit/Debit Card</option>
                  <option value="NETBANKING">Net Banking</option>
                </select>
              </div>

              <Button
                type="submit"
                className="w-full bg-status-success hover:bg-status-success/90 text-white"
                disabled={loading}
              >
                {loading ? "Processing..." : "Deposit Now"}
              </Button>
            </form>
          </div>
        )}

        {/* Withdraw Form */}
        {activeTab === "withdraw" && (
          <div className="bg-secondary p-8 rounded-xl border border-tertiary">
            <h2 className="text-xl font-semibold text-text-primary mb-6">Withdraw Money</h2>
            <form onSubmit={handleWithdraw} className="space-y-6">
              <Input
                label="Amount (₹)"
                type="number"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                placeholder="Enter amount"
                required
              />

              <div className="bg-tertiary p-4 rounded-lg">
                <p className="text-text-secondary text-sm">
                  ℹ️ KYC Level: <span className="text-accent-gold">{kycLevel}/2</span>
                </p>
                <p className="text-text-secondary text-sm mt-2">
                  Withdrawals are processed within 24-48 hours
                </p>
              </div>

              <Button
                type="submit"
                className="w-full bg-status-error hover:bg-status-error/90 text-white"
                disabled={loading || kycLevel < 2}
              >
                {loading ? "Processing..." : "Request Withdrawal"}
              </Button>
            </form>
          </div>
        )}
      </div>
    </div>
  );
}
