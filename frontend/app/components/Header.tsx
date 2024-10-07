"use client";

import React from "react";
import Link from "next/link";

const Header: React.FC = () => {
  return (
    <header className="flex flex-col md:flex-row justify-between items-center p-4 bg-gray-800 text-white">
      <div className="logo mb-4 md:mb-0">SYB</div>
      <nav className="flex flex-col md:flex-row space-y-2 md:space-y-0 md:space-x-4 mb-4 md:mb-0">
        <Link href="/">Home</Link>
        <Link href="/register">Register</Link>
        <Link href="/dashboard">Dashboard</Link>
        <Link href="/leaderboard">Leaderboard</Link>
        <Link href="/settings">Settings</Link>
      </nav>
      <button>Connect Wallet</button>
    </header>
  );
};

export default Header;
