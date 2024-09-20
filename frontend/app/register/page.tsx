"use client";

import React from "react";

const RegisterPage: React.FC = () => {
  return (
    <div className={"p-8"}>
      <h2 className="text-2xl font-bold mb-4">Register</h2>
      <div className="space-y-6">
        <div>
          <h3 className="text-xl font-semibold">Connect Your Wallet</h3>
        </div>
        <div>
          <h3 className="text-xl font-semibold">Deposit ETH</h3>
          <input
            type="number"
            placeholder="Amount"
            className="border p-2 rounded w-full md:w-1/2"
          />
          <button>Deposit</button>
        </div>

        <div>
          <h3 className="text-xl font-semibold">Status</h3>
          <div>No Transactions yet.</div>
        </div>
      </div>
    </div>
  );
};

export default RegisterPage;
