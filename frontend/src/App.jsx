import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import Login from "./pages/Login";
import Register from "./pages/Register";

import MobileNav from "./components/layout/MobileNav";
import Admin from "./pages/Admin";
import Analytics from "./pages/Analytics";
import Casino from "./pages/Casino";
import Dashboard from "./pages/Dashboard";
import History from "./pages/History";
import KYC from "./pages/KYC";
import Leaderboard from "./pages/Leaderboard";
import Payment from "./pages/Payment";
import Promotions from "./pages/Promotions";
import Sportsbook from "./pages/Sportsbook";

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/sportsbook" element={<Sportsbook />} />
        <Route path="/admin" element={<Admin />} />
        <Route path="/history" element={<History />} />
        <Route path="/payment" element={<Payment />} />
        <Route path="/kyc" element={<KYC />} />
        <Route path="/casino" element={<Casino />} />
        <Route path="/promotions" element={<Promotions />} />
        <Route path="/leaderboard" element={<Leaderboard />} />
        <Route path="/analytics" element={<Analytics />} />
        <Route path="/" element={<Navigate to="/dashboard" replace />} />
      </Routes>
      <MobileNav />
    </BrowserRouter>
  );
}




export default App;
