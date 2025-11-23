import { useState } from "react";
import { Link } from "react-router-dom";
import { Button } from "../components/ui/Button";
import { Input } from "../components/ui/Input";

export default function Register() {
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsLoading(true);
    // TODO: Integrate API
    setTimeout(() => setIsLoading(false), 1000);
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-primary p-4">
      <div className="w-full max-w-md space-y-8 bg-secondary p-8 rounded-2xl border border-tertiary shadow-2xl">
        <div className="text-center">
          <h2 className="text-3xl font-bold text-accent-gold">PlayKaro</h2>
          <p className="mt-2 text-text-secondary">Join the winning team!</p>
        </div>

        <form onSubmit={handleSubmit} className="mt-8 space-y-6">
          <Input
            label="Username"
            type="text"
            placeholder="CoolPunter99"
            required
          />
          <Input
            label="Email"
            type="email"
            placeholder="you@example.com"
            required
          />
          <Input
            label="Mobile Number"
            type="tel"
            placeholder="+91 98765 43210"
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
