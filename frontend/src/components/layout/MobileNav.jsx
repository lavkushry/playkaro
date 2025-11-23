import { useLocation, useNavigate } from "react-router-dom";

const navItems = [
  { path: "/dashboard", icon: "🏠", label: "Home" },
  { path: "/sportsbook", icon: "⚽", label: "Sports" },
  { path: "/casino", icon: "🎰", label: "Casino" },
  { path: "/promotions", icon: "🎁", label: "Promos" },
  { path: "/history", icon: "📊", label: "History" },
];

export default function MobileNav() {
  const navigate = useNavigate();
  const location = useLocation();

  return (
    <nav className="md:hidden fixed bottom-0 left-0 right-0 bg-secondary border-t border-tertiary z-50">
      <div className="flex justify-around items-center h-16">
        {navItems.map((item) => {
          const isActive = location.pathname === item.path;
          return (
            <button
              key={item.path}
              onClick={() => navigate(item.path)}
              className={`flex flex-col items-center justify-center flex-1 h-full transition-colors ${
                isActive ? "text-accent-gold" : "text-text-secondary"
              }`}
            >
              <span className="text-2xl mb-1">{item.icon}</span>
              <span className="text-xs font-medium">{item.label}</span>
            </button>
          );
        })}
      </div>
    </nav>
  );
}
