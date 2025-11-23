import confetti from "canvas-confetti";
import { useEffect, useRef, useState } from "react";
import { useWalletStore } from "../../store/useWalletStore";
import { Button } from "../ui/Button";

export default function SpinWheel() {
  const canvasRef = useRef(null);
  const [isSpinning, setIsSpinning] = useState(false);
  const [result, setResult] = useState(null);
  const { balance, updateBalance } = useWalletStore();

  const segments = [
    { label: "₹0", value: 0, color: "#EF4444" },
    { label: "₹50", value: 50, color: "#3B82F6" },
    { label: "₹100", value: 100, color: "#10B981" },
    { label: "₹20", value: 20, color: "#F59E0B" },
    { label: "₹500", value: 500, color: "#8B5CF6" },
    { label: "₹10", value: 10, color: "#EC4899" },
    { label: "JACKPOT", value: 1000, color: "#FCD34D" },
    { label: "TRY AGAIN", value: 0, color: "#6B7280" },
  ];

  const drawWheel = (rotation) => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    const centerX = canvas.width / 2;
    const centerY = canvas.height / 2;
    const radius = canvas.width / 2 - 10;
    const segmentAngle = (2 * Math.PI) / segments.length;

    ctx.clearRect(0, 0, canvas.width, canvas.height);

    segments.forEach((segment, i) => {
      ctx.beginPath();
      ctx.moveTo(centerX, centerY);
      ctx.arc(
        centerX,
        centerY,
        radius,
        i * segmentAngle + rotation,
        (i + 1) * segmentAngle + rotation
      );
      ctx.fillStyle = segment.color;
      ctx.fill();
      ctx.stroke();
      ctx.save();

      // Text
      ctx.translate(centerX, centerY);
      ctx.rotate(i * segmentAngle + rotation + segmentAngle / 2);
      ctx.textAlign = "right";
      ctx.fillStyle = "#fff";
      ctx.font = "bold 14px Arial";
      ctx.fillText(segment.label, radius - 20, 5);
      ctx.restore();
    });

    // Arrow
    ctx.beginPath();
    ctx.moveTo(centerX + 15, centerY - radius - 15);
    ctx.lineTo(centerX - 15, centerY - radius - 15);
    ctx.lineTo(centerX, centerY - radius + 10);
    ctx.fillStyle = "white";
    ctx.fill();
  };

  useEffect(() => {
    drawWheel(0);
  }, []);

  const spin = async () => {
    if (balance < 50) {
      alert("Insufficient balance! Cost to spin: ₹50");
      return;
    }

    // Deduct cost locally (in real app, call backend)
    // updateBalance(-50);
    setIsSpinning(true);
    setResult(null);

    let rotation = 0;
    let speed = 0.5;
    let deceleration = 0.005;
    const stopAngle = Math.random() * 2 * Math.PI; // Random stop

    const animate = () => {
      rotation += speed;
      speed -= deceleration;

      if (speed <= 0) {
        speed = 0;
        setIsSpinning(false);

        // Calculate result
        const segmentAngle = (2 * Math.PI) / segments.length;
        const normalizedRotation = rotation % (2 * Math.PI);
        const winningIndex = Math.floor(((2 * Math.PI) - normalizedRotation) / segmentAngle) % segments.length;
        const win = segments[winningIndex];

        setResult(win);
        if (win.value > 0) {
          confetti({ particleCount: 100, spread: 70, origin: { y: 0.6 } });
          // updateBalance(win.value);
        }
      } else {
        requestAnimationFrame(animate);
      }
      drawWheel(rotation);
    };

    animate();
  };

  return (
    <div className="flex flex-col items-center justify-center p-8 bg-secondary rounded-2xl border border-tertiary shadow-2xl max-w-md mx-auto">
      <h2 className="text-3xl font-bold text-accent-gold mb-4">Spin & Win</h2>
      <div className="relative mb-6">
        <canvas
          ref={canvasRef}
          width={300}
          height={300}
          className="rounded-full shadow-lg"
        />
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-16 h-16 bg-white rounded-full shadow-inner flex items-center justify-center border-4 border-tertiary">
          <span className="font-bold text-primary">PLAY</span>
        </div>
      </div>

      {result && (
        <div className="mb-4 text-center animate-bounce">
          <p className="text-text-secondary">You won</p>
          <p className="text-4xl font-bold text-status-success">{result.label}</p>
        </div>
      )}

      <Button
        onClick={spin}
        disabled={isSpinning}
        className="w-full bg-gradient-to-r from-accent-gold to-yellow-600 hover:from-yellow-500 hover:to-yellow-700 text-primary font-bold text-lg py-4"
      >
        {isSpinning ? "Spinning..." : "Spin Now (₹50)"}
      </Button>
    </div>
  );
}
