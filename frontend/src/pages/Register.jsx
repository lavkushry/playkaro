import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Button } from "../components/ui/Button";
import { Input } from "../components/ui/Input";
import { useAuthStore } from "../store/useAuthStore";

export default function Register() {
  const navigate = useNavigate();
  const register = useAuthStore((state) => state.register);
  const isLoading = useAuthStore((state) => state.isLoading);
  const error = useAuthStore((state) => state.error);

  const [formData, setFormData] = useState({
    username: "",
    email: "",
    mobile: "",
    password: ""
  });

  const handleChange = (e) => {
    setFormData({ ...formData, [e.target.name]: e.target.value });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    const success = await register(formData);
    if (success) {
      navigate("/dashboard");
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-primary p-4">
      <div className="w-full max-w-md space-y-8 bg-secondary p-8 rounded-2xl border border-tertiary shadow-2xl">
        <div className="text-center">
          <h2 className="text-3xl font-bold text-accent-gold">PlayKaro</h2>
          <p className="mt-2 text-text-secondary">Join the winning team!</p>
        </div>

        {error && (
          <div className="bg-status-error/10 border border-status-error text-status-error p-3 rounded-lg text-sm text-center">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="mt-8 space-y-6">
          <Input
            label="Username"
            name="username"
            type="text"
            placeholder="CoolPunter99"
            value={formData.username}
            onChange={handleChange}
            required
          />
          <Input
            label="Email"
            name="email"
            type="email"
            placeholder="you@example.com"
            value={formData.email}
            onChange={handleChange}
            required
          />
          <Input
            label="Mobile Number"
            name="mobile"
            type="tel"
            placeholder="+91 98765 43210"
            value={formData.mobile}
            onChange={handleChange}
            required
          />
          <Input
            label="Password"
            name="password"
            type="password"
            placeholder="••••••••"
            value={formData.password}
            onChange={handleChange}
            required
          />

          <Button
            type="submit"
            className="w-full"
            disabled={isLoading}
          >
            {isLoading ? "Creating Account..." : "Create Account"}
          </Button>
        </form>

        <div className="text-center text-sm text-text-secondary">
          Already have an account?{" "}
          <Link to="/login" className="text-accent-blue hover:underline">
            Sign in
          </Link>
        </div>
      </div>
    </div>
  );
}
