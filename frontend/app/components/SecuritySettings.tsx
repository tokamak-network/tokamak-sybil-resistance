import React from "react";
import * as CheckBox from "@radix-ui/react-checkbox";

const SecuritySettings: React.FC = () => {
  return (
    <div className="p-8">
      <h3 className="text-xl font-semibold">Security Settings</h3>
      <div className="flex items-center">
        <CheckBox.Root id="two-factor" className="mr-2">
          <CheckBox.Indicator className="bg-blue-500 w-4 h-4" />
        </CheckBox.Root>
        <label htmlFor="two-factor">Enable Two-Factor Authentication</label>
      </div>
    </div>
  );
};

export default SecuritySettings;
