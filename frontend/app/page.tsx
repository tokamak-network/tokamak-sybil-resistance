// app/page.tsx

import React from "react";
const LandingPage: React.FC = () => {
  return (
    <div>
      <section className="hero p-8 text-center">
        <h1 className="text-4xl font-bold">Welcome to SYB</h1>
        <p className="mt-4">Fight Sybil attacks with Ethereum.</p>
        <button className="mt-6 bg-blue-500 px-4 py-2 rounded text-white">
          Get Started
        </button>
      </section>
      <section className="how-it-works p-8">
        <h2 className="text-2xl font-bold">How It Works</h2>
        <div className="mt-4">
          Connect Wallet - Deposit ETH - Vouch/Receive Vouches - Compute Score
        </div>
      </section>
    </div>
  );
};

export default LandingPage;
