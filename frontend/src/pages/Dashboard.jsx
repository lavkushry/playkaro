import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "../components/ui/Button";
import { useWalletStore } from "../store/useWalletStore";

export default function Dashboard() {
  const navigate = useNavigate();
  const { balance, currency, fetchBalance, deposit, withdraw } = useWalletStore();
  const [amount, setAmount] = useState("");

  // Mock Auth Token (In real app, get from auth store)
  const token = localStorage.getItem("token");

  useEffect(() => {
    if (!token) {
      navigate("/login");
      return;
    }
    fetchBalance(token);
  }, [token, fetchBalance, navigate]);

  const handleTransaction = (type) => {
    if (!amount) return;
    if (type === "DEPOSIT") {
      deposit(token, amount);
    } else {
      withdraw(token, amount);
    }
    setAmount("");
  };

  return (
    <div className="min-h-screen bg-primary p-6">
      <div className="max-w-4xl mx-auto space-y-8">
        {/* Header */}
        <div className="flex justify-between items-center">
          <h1 className="text-3xl font-bold text-accent-gold">Dashboard</h1>
          <div className="flex gap-4">
            <Button variant="secondary" onClick={() => navigate("/sportsbook")}>
              Sportsbook
            </Button>
            <Button variant="secondary" onClick={() => navigate("/history")}>
              History
            </Button>
            <Button variant="secondary" onClick={() => navigate("/admin")}>
              Admin
            </Button>
            <Button variant="outline" onClick={() => {
              localStorage.removeItem("token");
              navigate("/login");
            }}>Logout</Button>
          </div>
        </div>

        {/* Wallet Card */}
        <div className="bg-secondary p-8 rounded-2xl border border-tertiary shadow-xl">
          <h2 className="text-text-secondary text-sm uppercase tracking-wider">Total Balance</h2>
          <div className="text-5xl font-bold text-text-primary mt-2">
            {currency} {balance.toFixed(2)}
          </div>

          <div className="mt-8 grid grid-cols-1 md:grid-cols-2 gap-4">
            <Button
              className="flex-1 bg-status-success hover:bg-status-success/90 text-white"
              onClick={() => navigate("/payment")}
            >
              💰 Deposit Money
            </Button>
            <Button
              className="flex-1 bg-status-error hover:bg-status-error/90 text-white"
              onClick={() => navigate("/payment")}
            >
              💸 Withdraw Money
            </Button>
            <Button
              variant="outline"
              className="flex-1"
              onClick={() => navigate("/kyc")}
            >
              📄 Complete KYC
            </Button>
            <Button
              variant="outline"
              className="flex-1"
              onClick={() => navigate("/payment")}
            >
              📊 View Transactions
            </Button>
          </div>
        </div>

        {/* Recent Transactions Placeholder */}
        <div className="bg-secondary p-6 rounded-xl border border-tertiary">
          <h3 className="text-xl font-semibold text-text-primary mb-4">Recent Transactions</h3>
          <div className="text-text-secondary text-center py-8">
            No transactions yet. Start playing!
          </div>
        </div>
      </div>
    </div>
  );
}
