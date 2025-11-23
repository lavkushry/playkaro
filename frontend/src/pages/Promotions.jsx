import axios from "axios";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "../components/ui/Button";
import { Input } from "../components/ui/Input";

const API_URL = 'http://localhost:8080/api/v1';

export default function Promotions() {
  const navigate = useNavigate();
  const [bonuses, setBonuses] = useState([]);
  const [referralCode, setReferralCode] = useState("");
  const [applyCode, setApplyCode] = useState("");
  const [loading, setLoading] = useState(false);
  const token = localStorage.getItem("token");

  useEffect(() => {
    if (!token) {
      navigate("/login");
      return;
    }
    fetchBonuses();
  }, [token, navigate]);

  const fetchBonuses = async () => {
    try {
      const response = await axios.get(`${API_URL}/promotions/bonuses`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      setBonuses(response.data || []);
    } catch (err) {
      console.error(err);
    }
  };

  const handleClaimBonus = async (bonusType) => {
    setLoading(true);
    try {
      await axios.post(`${API_URL}/promotions/claim`, {
        bonus_type: bonusType
      }, {
        headers: { Authorization: `Bearer ${token}` }
      });
      alert(`${bonusType} bonus claimed successfully!`);
      fetchBonuses();
    } catch (err) {
      alert("Failed: " + (err.response?.data?.error || err.message));
    } finally {
      setLoading(false);
    }
  };

  const handleGenerateCode = async () => {
    try {
      const response = await axios.post(`${API_URL}/promotions/referral/generate`, {}, {
        headers: { Authorization: `Bearer ${token}` }
      });
      setReferralCode(response.data.referral_code);
    } catch (err) {
      alert("Failed: " + (err.response?.data?.error || err.message));
    }
  };

  const handleApplyCode = async () => {
    setLoading(true);
    try {
      await axios.post(`${API_URL}/promotions/referral/apply`, {
        referral_code: applyCode
      }, {
        headers: { Authorization: `Bearer ${token}` }
      });
      alert("Referral code applied! ₹50 bonus added.");
      setApplyCode("");
      fetchBonuses();
    } catch (err) {
      alert("Failed: " + (err.response?.data?.error || err.message));
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status) => {
    if (status === 'COMPLETED') return 'text-status-success';
    if (status === 'EXPIRED') return 'text-text-secondary';
    return 'text-accent-gold';
  };

  return (
    <div className="min-h-screen bg-primary p-6">
      <div className="max-w-5xl mx-auto space-y-8">
        <div className="flex justify-between items-center">
          <h1 className="text-3xl font-bold text-accent-gold">🎁 Promotions</h1>
          <Button variant="outline" onClick={() => navigate("/dashboard")}>Dashboard</Button>
        </div>

        {/* Available Bonuses */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="bg-gradient-to-br from-accent-gold/20 to-secondary p-6 rounded-xl border border-accent-gold">
            <h2 className="text-2xl font-bold text-accent-gold mb-2">Welcome Bonus</h2>
            <p className="text-text-secondary mb-4">Get ₹100 on your first bonus claim!</p>
            <div className="flex items-center justify-between mb-4">
              <div>
                <p className="text-sm text-text-secondary">Wagering Requirement</p>
                <p className="text-lg font-semibold text-text-primary">5x (₹500)</p>
              </div>
              <div>
                <p className="text-4xl font-bold text-accent-gold">₹100</p>
              </div>
            </div>
            <Button
              className="w-full"
              onClick={() => handleClaimBonus("WELCOME")}
              disabled={loading || bonuses.some(b => b.type === 'WELCOME')}
            >
              {bonuses.some(b => b.type === 'WELCOME') ? 'Already Claimed' : 'Claim Now'}
            </Button>
          </div>

          <div className="bg-gradient-to-br from-accent-blue/20 to-secondary p-6 rounded-xl border border-accent-blue">
            <h2 className="text-2xl font-bold text-accent-blue mb-2">Daily Bonus</h2>
            <p className="text-text-secondary mb-4">Get ₹20 every day!</p>
            <div className="flex items-center justify-between mb-4">
              <div>
                <p className="text-sm text-text-secondary">Wagering Requirement</p>
                <p className="text-lg font-semibold text-text-primary">3x (₹60)</p>
              </div>
              <div>
                <p className="text-4xl font-bold text-accent-blue">₹20</p>
              </div>
            </div>
            <Button
              className="w-full bg-accent-blue hover:bg-accent-blue/90"
              onClick={() => handleClaimBonus("DAILY")}
              disabled={loading}
            >
              Claim Daily Bonus
            </Button>
          </div>
        </div>

        {/* Referral Section */}
        <div className="bg-secondary p-8 rounded-xl border border-tertiary">
          <h2 className="text-2xl font-bold text-text-primary mb-6">Referral Program</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
            <div>
              <h3 className="text-lg font-semibold text-accent-gold mb-4">Your Referral Code</h3>
              {referralCode ? (
                <div className="flex gap-2">
                  <Input
                    value={referralCode}
                    readOnly
                    className="font-mono text-lg"
                  />
                  <Button
                    onClick={() => navigator.clipboard.writeText(referralCode)}
                    variant="outline"
                  >
                    Copy
                  </Button>
                </div>
              ) : (
                <Button onClick={handleGenerateCode}>Generate Code</Button>
              )}
              <p className="text-text-secondary text-sm mt-2">
                Share your code and get ₹50 when someone signs up!
              </p>
            </div>

            <div>
              <h3 className="text-lg font-semibold text-accent-blue mb-4">Apply Referral Code</h3>
              <div className="flex gap-2">
                <Input
                  value={applyCode}
                  onChange={(e) => setApplyCode(e.target.value)}
                  placeholder="Enter code"
                />
                <Button
                  onClick={handleApplyCode}
                  disabled={loading || !applyCode}
                >
                  Apply
                </Button>
              </div>
              <p className="text-text-secondary text-sm mt-2">
                Get ₹50 bonus when you apply a referral code!
              </p>
            </div>
          </div>
        </div>

        {/* Active Bonuses */}
        <div className="bg-secondary p-6 rounded-xl border border-tertiary">
          <h2 className="text-xl font-semibold text-text-primary mb-4">Your Bonuses</h2>
          {bonuses.length > 0 ? (
            <div className="space-y-3">
              {bonuses.map((bonus) => (
                <div key={bonus.id} className="bg-tertiary p-4 rounded-lg">
                  <div className="flex justify-between items-start mb-2">
                    <div>
                      <p className="font-medium text-text-primary">{bonus.type} BONUS</p>
                      <p className="text-sm text-text-secondary">
                        Expires: {new Date(bonus.expires_at).toLocaleDateString()}
                      </p>
                    </div>
                    <span className={`text-sm font-semibold ${getStatusColor(bonus.status)}`}>
                      {bonus.status}
                    </span>
                  </div>
                  <div className="grid grid-cols-3 gap-4 text-sm">
                    <div>
                      <p className="text-text-secondary">Amount</p>
                      <p className="text-text-primary font-medium">₹{bonus.amount.toFixed(2)}</p>
                    </div>
                    <div>
                      <p className="text-text-secondary">Wagered</p>
                      <p className="text-text-primary font-medium">
                        ₹{bonus.wagered.toFixed(2)} / ₹{bonus.wagering_requirement.toFixed(2)}
                      </p>
                    </div>
                    <div>
                      <p className="text-text-secondary">Progress</p>
                      <div className="w-full bg-secondary rounded-full h-2 mt-1">
                        <div
                          className="bg-accent-gold h-2 rounded-full transition-all"
                          style={{ width: `${Math.min((bonus.wagered / bonus.wagering_requirement) * 100, 100)}%` }}
                        ></div>
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-center text-text-secondary py-8">No bonuses yet. Claim one above!</p>
          )}
        </div>
      </div>
    </div>
  );
}
