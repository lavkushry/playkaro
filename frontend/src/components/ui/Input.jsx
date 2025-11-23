import { cn } from "./Button";

export const Input = ({ className, label, error, ...props }) => {
  return (
    <div className="flex flex-col gap-1.5">
      {label && (
        <label className="text-sm font-medium text-text-secondary">
          {label}
        </label>
      )}
      <input
        className={cn(
          "flex h-10 w-full rounded-md border border-tertiary bg-secondary px-3 py-2 text-sm text-text-primary placeholder:text-text-secondary/50 focus:outline-none focus:ring-2 focus:ring-accent-gold/50 focus:border-accent-gold transition-all",
          error && "border-status-error focus:ring-status-error/50",
          className
        )}
        {...props}
      />
      {error && <span className="text-xs text-status-error">{error}</span>}
    </div>
  );
};
