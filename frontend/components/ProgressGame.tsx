"use client";

import { useState, useEffect } from "react";  
import confetti from "canvas-confetti";

type ProgressGameProps = {
  isProcessing: boolean;
  onComplete: () => void;
};

type Animal = {
  id: number;
  x: number;
  y: number;
  dx: number;
  dy: number;
  icon: string;
  size?: number;
};

const animalIcons = ["😈", "🦈", "🐸", "🐹", "🦊", "🤬", "🦹🏽" ]; 

export default function ProgressGame({ isProcessing, onComplete }: ProgressGameProps) {
  const [fakeProgress, setFakeProgress] = useState(0);
  const [animals, setAnimals] = useState<Animal[]>([]);
  const [score, setScore] = useState(0);
  const [finished, setFinished] = useState(false);

  // Simulate progress
  useEffect(() => {
    if (!isProcessing) return;
    const interval = window.setInterval(() => {
      setFakeProgress((prev) => Math.min(prev + Math.random() * 5, 95));
    }, 500);
    return () => clearInterval(interval);
  }, [isProcessing]);

  // Generate new animals
  useEffect(() => {
    if (!isProcessing) return;
    const interval = window.setInterval(() => {
      const size = 30 + Math.random() * 20;
      setAnimals((prev) => [
        ...prev,
        {
          id: Date.now() + Math.random(),
          x: Math.random() * 80 + 10,
          y: Math.random() * 60 + 10,
          dx: (Math.random() - 0.5) * 2,
          dy: (Math.random() - 0.5) * 2,
          icon: animalIcons[Math.floor(Math.random() * animalIcons.length)],
          size,
        },
      ]);
    }, Math.max(800 - score * 10, 200));
    return () => clearInterval(interval);
  }, [isProcessing, score]);

  // Move animals
  useEffect(() => {
    if (!isProcessing) return;
    const interval = window.setInterval(() => {
      setAnimals((prev) =>
        prev.map((a) => {
          let x = a.x + a.dx;
          let y = a.y + a.dy;

          if (x < 0) { x = 0; a.dx = Math.abs(a.dx); }
          if (x > 90) { x = 90; a.dx = -Math.abs(a.dx); }
          if (y < 0) { y = 0; a.dy = Math.abs(a.dy); }
          if (y > 60) { y = 60; a.dy = -Math.abs(a.dy); }

          return { ...a, x, y };
        })
      );
    }, 50);
    return () => clearInterval(interval);
  }, [isProcessing]);

  const handleAnimalClick = (id: number) => {
    setScore((prev) => prev + 1);
    setAnimals((prev) => prev.filter((a) => a.id !== id));

    confetti({
      particleCount: 10,
      spread: 50,
      origin: { y: 0.7 },
    });
  };

  // Finish confetti
  useEffect(() => {
    if (!isProcessing && !finished) {
      setFakeProgress(100);
      setFinished(true);

      confetti({
        particleCount: 150,
        spread: 120,
        origin: { y: 0.6 },
      });

      onComplete?.();
    }
  }, [isProcessing, finished, onComplete]);

  return (
    <div className="bg-white rounded-2xl shadow-xl p-6 mb-8 relative">
      {/* Progress Bar */}
      <div className="mb-4">
        <div className="flex justify-between text-sm text-gray-600 mb-1">
          <span>{finished ? "Completed!" : "Processing..."}</span>
          <span>{Math.round(fakeProgress)}%</span>
        </div>
        <div className="h-3 bg-gray-200 rounded-full overflow-hidden">
          <div
            className={`h-full transition-all duration-300 ${
              finished ? "bg-green-400" : "bg-gradient-to-r from-blue-500 to-indigo-500"
            }`}
            style={{ width: `${fakeProgress}%` }}
          />
        </div>
      </div>

      {/* Instruction */}
      {!finished && (
        <div className="mb-2 text-center text-gray-700 font-medium">
          🕹️ Tap the animals while you wait to score points!
        </div>
      )}

      {/* Mini-game */}
      {!finished && (
        <div className="relative w-full h-64 bg-indigo-50 rounded-xl overflow-hidden">
          {animals.map((a) => (
            <div
              key={a.id}
              onClick={() => handleAnimalClick(a.id)}
              onTouchStart={() => handleAnimalClick(a.id)}
              className="absolute cursor-pointer select-none"
              style={{
                left: `${a.x}%`,
                top: `${a.y}%`,
                fontSize: `${a.size}px`,
              }}
            >
              {a.icon}
            </div>
          ))}
        </div>
      )}

      {/* Score */}
      <div className="mt-4 text-center text-gray-700 font-semibold">
        {finished ? (
          <span>🎉 You scored {score} points! Great job! 🎊</span>
        ) : (
          <span>🎯 Score: {score} points</span>
        )}
      </div>
    </div>
  );
}
