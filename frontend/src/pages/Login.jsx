import { useState } from "react";
import { Link } from "react-router-dom";
import { Button } from "../components/ui/Button";
import { Input } from "../components/ui/Input";

export default function Login() {
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsLoading(true);
    const success = await login(formData.email, formData.password);
    if (success) {
      navigate("/dashboard");
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-primary p-4">
      <div className="w-full max-w-md space-y-8 bg-secondary p-8 rounded-2xl border border-tertiary shadow-2xl">
        <div className="text-center">
          <h2 className="text-3xl font-bold text-accent-gold">PlayKaro</h2>
          <p className="mt-2 text-text-secondary">Welcome back!</p>
        </div>

        {error && (
          <div className="bg-status-error/10 border border-status-error text-status-error p-3 rounded-lg text-sm text-center">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="mt-8 space-y-6">
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
            {isLoading ? "Signing In..." : "Sign In"}
          </Button>
        </form>

        <div className="text-center text-sm text-text-secondary">
          Don't have an account?{" "}
          <Link to="/register" className="text-accent-blue hover:underline">
            Create account
          </Link>
        </div>
      </div>
    </div>
  );
}
