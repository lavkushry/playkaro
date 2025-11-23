import axios from "axios";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "../components/ui/Button";
import { Input } from "../components/ui/Input";

const API_URL = 'http://localhost:8080/api/v1';

export default function KYC() {
  const navigate = useNavigate();
  const [kycData, setKycData] = useState({ kyc_level: 0, documents: [] });
  const [documentType, setDocumentType] = useState("PAN");
  const [documentURL, setDocumentURL] = useState("");
  const [loading, setLoading] = useState(false);
  const token = localStorage.getItem("token");

  useEffect(() => {
    if (!token) {
      navigate("/login");
      return;
    }
    fetchKYCStatus();
  }, [token, navigate]);

  const fetchKYCStatus = async () => {
    try {
      const response = await axios.get(`${API_URL}/kyc/status`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      setKycData(response.data);
    } catch (err) {
      console.error(err);
    }
  };

  const handleUpload = async (e) => {
    e.preventDefault();
    setLoading(true);
    try {
      await axios.post(`${API_URL}/kyc/upload`, {
        document_type: documentType,
        document_url: documentURL
      }, {
        headers: { Authorization: `Bearer ${token}` }
      });
      alert("Document uploaded successfully!");
      setDocumentURL("");
      fetchKYCStatus();
    } catch (err) {
      alert("Upload failed: " + (err.response?.data?.error || err.message));
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status) => {
    if (status === 'APPROVED') return 'text-status-success';
    if (status === 'REJECTED') return 'text-status-error';
    return 'text-accent-gold';
  };

  const getStatusBg = (status) => {
    if (status === 'APPROVED') return 'bg-status-success/20 border-status-success';
    if (status === 'REJECTED') return 'bg-status-error/20 border-status-error';
    return 'bg-accent-gold/20 border-accent-gold';
  };

  return (
    <div className="min-h-screen bg-primary p-6">
      <div className="max-w-3xl mx-auto space-y-8">
        <div className="flex justify-between items-center">
          <h1 className="text-3xl font-bold text-accent-gold">KYC Verification</h1>
          <Button variant="outline" onClick={() => navigate("/dashboard")}>Dashboard</Button>
        </div>

        {/* KYC Level Status */}
        <div className="bg-secondary p-6 rounded-xl border border-tertiary">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-text-secondary text-sm">Verification Level</p>
              <p className="text-3xl font-bold text-accent-gold">Level {kycData.kyc_level}/2</p>
            </div>
            <div className="text-right">
              <p className="text-text-secondary text-sm">Benefits</p>
              <p className="text-text-primary">
                {kycData.kyc_level >= 2 ? "✓ Full Access" : "Limited Withdrawals"}
              </p>
            </div>
          </div>

          {/* Progress Bar */}
          <div className="mt-4 h-2 bg-tertiary rounded-full overflow-hidden">
            <div
              className="h-full bg-accent-gold transition-all"
              style={{ width: `${(kycData.kyc_level / 2) * 100}%` }}
            ></div>
          </div>
        </div>

        {/* Upload Form */}
        <div className="bg-secondary p-8 rounded-xl border border-tertiary">
          <h2 className="text-xl font-semibold text-text-primary mb-6">Upload Documents</h2>
          <form onSubmit={handleUpload} className="space-y-6">
            <div>
              <label className="block text-sm font-medium text-text-secondary mb-2">
                Document Type
              </label>
              <select
                value={documentType}
                onChange={(e) => setDocumentType(e.target.value)}
                className="w-full px-4 py-3 bg-tertiary border border-tertiary rounded-lg text-text-primary focus:ring-2 focus:ring-accent-gold"
              >
                <option value="PAN">PAN Card</option>
                <option value="AADHAAR">Aadhaar Card</option>
                <option value="PASSPORT">Passport</option>
              </select>
            </div>

            <Input
              label="Document URL (Image Link)"
              type="url"
              value={documentURL}
              onChange={(e) => setDocumentURL(e.target.value)}
              placeholder="https://example.com/document.jpg"
              required
            />

            <div className="bg-tertiary p-4 rounded-lg">
              <p className="text-text-secondary text-sm">
                ℹ️ Supported formats: JPG, PNG, PDF (Max 5MB)
              </p>
              <p className="text-text-secondary text-sm mt-1">
                For demo: Use any image URL
              </p>
            </div>

            <Button
              type="submit"
              className="w-full"
              disabled={loading}
            >
              {loading ? "Uploading..." : "Upload Document"}
            </Button>
          </form>
        </div>

        {/* Document List */}
        <div className="bg-secondary p-6 rounded-xl border border-tertiary">
          <h2 className="text-xl font-semibold text-text-primary mb-4">Uploaded Documents</h2>

          {kycData.documents && kycData.documents.length > 0 ? (
            <div className="space-y-3">
              {kycData.documents.map((doc) => (
                <div key={doc.id} className={`p-4 border rounded-lg ${getStatusBg(doc.status)}`}>
                  <div className="flex justify-between items-start">
                    <div>
                      <p className="font-medium text-text-primary">{doc.document_type}</p>
                      <p className="text-sm text-text-secondary mt-1">
                        {new Date(doc.created_at).toLocaleDateString()}
                      </p>
                    </div>
                    <span className={`text-sm font-semibold ${getStatusColor(doc.status)}`}>
                      {doc.status}
                    </span>
                  </div>
                  {doc.remarks && (
                    <p className="text-sm text-text-secondary mt-2">
                      Remarks: {doc.remarks}
                    </p>
                  )}
                </div>
              ))}
            </div>
          ) : (
            <p className="text-center text-text-secondary py-8">
              No documents uploaded yet
            </p>
          )}
        </div>
      </div>
    </div>
  );
}
