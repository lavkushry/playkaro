import { clsx } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs) {
  return twMerge(clsx(inputs));
}


export const Button = ({ className, variant = "primary", ...props }) => {
  const variants = {
    primary: "bg-accent-gold text-primary hover:bg-yellow-600",
    secondary: "bg-secondary text-text-primary border border-tertiary hover:bg-tertiary",
    outline: "border-2 border-accent-gold text-accent-gold hover:bg-accent-gold/10",
  };

  return (
    <button
      className={cn(
        "px-4 py-2 rounded-lg font-medium transition-all duration-200 active:scale-95 disabled:opacity-50 disabled:pointer-events-none",
        variants[variant],
        className
      )}
      {...props}
    />
  );
};
