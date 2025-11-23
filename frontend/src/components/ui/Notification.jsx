import { useEffect, useState } from "react";

export function useNotification() {
  const [notification, setNotification] = useState(null);

  const showNotification = (message, type = "info") => {
    setNotification({ message, type });
    setTimeout(() => setNotification(null), 5000);
  };

  return { notification, showNotification };
}

export function NotificationToast({ notification, onClose }) {
  useEffect(() => {
    if (notification) {
      const timer = setTimeout(() => {
        onClose();
      }, 5000);
      return () => clearTimeout(timer);
    }
  }, [notification, onClose]);

  if (!notification) return null;

  const styles = {
    success: "bg-status-success/20 border-status-success text-status-success",
    error: "bg-status-error/20 border-status-error text-status-error",
    info: "bg-accent-blue/20 border-accent-blue text-accent-blue",
  };

  return (
    <div className="fixed top-20 right-4 z-50 animate-slide-in">
      <div className={`px-6 py-4 rounded-lg border ${styles[notification.type]} max-w-md shadow-lg`}>
        <div className="flex items-center justify-between gap-4">
          <p className="font-medium">{notification.message}</p>
          <button
            onClick={onClose}
            className="text-text-secondary hover:text-text-primary"
          >
            ✕
          </button>
        </div>
      </div>
    </div>
  );
}
