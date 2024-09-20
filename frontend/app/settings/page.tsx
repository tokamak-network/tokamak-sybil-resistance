import React from "react";
import * as Checkbox from "@radix-ui/react-checkbox";
import AccountInfo from "../components/AccountInfo";
import SecuritySettings from "../components/SecuritySettings";

const SettingsPage: React.FC = () => {
  return (
    <div className="p-8">
      <h2 className="text-2xl font-bold mb-4">Settings</h2>
      <div className="space-y-6">
        <AccountInfo />
        <div>
          <SecuritySettings />
          <button className="bg-red-500 px-4 py-2 rounded text-white mt-2">
            Withdraw ETH to L1
          </button>
        </div>
      </div>
    </div>
  );
};

export default SettingsPage;
