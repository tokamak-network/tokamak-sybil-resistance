"use client";

import React from "react";
import Link from "next/link";

const Header: React.FC = () => {
  return (
    <header className="flex justify-between items-center p-4 bg-gray-800 text-white">
      <div className="logo">SYB</div>
      <nav className="flex space-x-4">
        <Link href="/">Home</Link>
        <Link href="/registeer">Register</Link>
        <Link href="/dashboard">Dashboard</Link>
        <Link href="/leaderboard">Leaderboard</Link>
        <Link href="/settings">Settings</Link>
      </nav>
      <button>Connect Wallet</button>
    </header>
  );
};

export default Header;
