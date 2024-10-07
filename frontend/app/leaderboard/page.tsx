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
      <div className="overflow-x-auto">
        <table className="min-w-full bg-white">
          <thead>
            <tr className="'w-full bg-gray-200 text-gray-600 uppercase text-sm leading-normal">
              <th className="py-3 px-6 text-left">Rank</th>
              <th className="py-3 px-6 text-left">Address</th>
              <th className="py-3 px-6 text-left">Score</th>
              <th className="py-3 px-6 text-left">Balance (ETH)</th>
            </tr>
          </thead>
          <tbody className="text-gray-600 text-sm font-light">
            {leaderboardData.map((entry) => (
              <tr
                key={entry.rank}
                className="border-b border-gray-200 hover:bg-gray-100"
              >
                <td className="py-3 px-5 text-left whitespace-nowrap">
                  {entry.rank}
                </td>
                <td className="py-3 px-5 text-left whitespace-nowrap">
                  {entry.address}
                </td>
                <td className="py-3 px-5 text-left whitespace-nowrap">
                  {entry.score}
                </td>
                <td className="py-3 px-5 text-left whitespace-nowrap">
                  {entry.balance}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default LeaderboardPage;
