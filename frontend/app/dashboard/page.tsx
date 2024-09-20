import React from "react";

const Dashboard: React.FC = () => (
  <div className="p-8">
    <h2 className="text-2xl font-bold">Dashboard</h2>
    <div className="mt-4">
      <div>Your Address: 0x123...</div>
      <div>Balance: 0 ETH</div>
      <div>Vouches Received: 0</div>
    </div>
  </div>
);

export default Dashboard;
