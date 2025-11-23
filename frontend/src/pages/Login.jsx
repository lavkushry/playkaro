import { useState } from "react";
import { Link } from "react-router-dom";
import { Button } from "../components/ui/Button";
import { Input } from "../components/ui/Input";

export default function Login() {
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsLoading(true);
    try {
      const response = await fetch('http://localhost:8080/api/v1/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email: e.target.email.value, password: e.target.password.value }),
      });
      const data = await response.json();
      if (response.ok) {
        localStorage.setItem('token', data.token);
        window.location.href = '/dashboard';
      } else {
        alert(data.error);
      }
    } catch (err) {
      alert('Login failed');
    } finally {
      setIsLoading(false);
    }
  };


  return (
    <div className="min-h-screen flex items-center justify-center bg-primary p-4">
      <div className="w-full max-w-md space-y-8 bg-secondary p-8 rounded-2xl border border-tertiary shadow-2xl">
        <div className="text-center">
          <h2 className="text-3xl font-bold text-accent-gold">PlayKaro</h2>
          <p className="mt-2 text-text-secondary">Welcome back, champion!</p>
        </div>

        <form onSubmit={handleSubmit} className="mt-8 space-y-6">
          <Input
            label="Email"
            type="email"
            placeholder="you@example.com"
            required
          />
          <Input
            label="Password"
            type="password"
            placeholder="••••••••"
            required
          />

          <Button
            type="submit"
            className="w-full"
            disabled={isLoading}
          >
            {isLoading ? "Signing in..." : "Sign In"}
          </Button>
        </form>

        <div className="text-center text-sm text-text-secondary">
          Don't have an account?{" "}
          <Link to="/register" className="text-accent-blue hover:underline">
            Create one
          </Link>
        </div>
      </div>
    </div>
  );
}
