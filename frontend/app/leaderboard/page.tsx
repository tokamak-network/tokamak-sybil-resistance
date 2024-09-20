import React from "react";
import { LeaderboardEntry } from "../utils/types";

const LeaderboardPage: React.FC = () => {
  const leaderboardData: LeaderboardEntry[] = [
    { rank: 1, address: "0x123...", score: 95, balance: 5.0 },
    { rank: 2, address: "0x456...", score: 90, balance: 4.5 },
  ];

  return (
    <div className="p-8">
      <h2 className="text-2xl font-bold mb-4">Leaderboard</h2>
      <div className="overflow-x-auto"></div>
    </div>
  );
};

export default LeaderboardPage;
