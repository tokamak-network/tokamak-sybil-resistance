require('dotenv').config();

const retrieveEnvVariable = (variableName) => {
    const variable = process.env[variableName] || undefined;
    if (!variable) {
      console.error(`${variableName} is not set`);
      // eslint-disable-next-line n/no-process-exit
      process.exit(1);
    }
    return variable;
};

// Wallet
const PRIVATE_KEY = retrieveEnvVariable('PRIVATE_KEY');

// Connection
const RPC_ENDPOINT = retrieveEnvVariable('RPC_ENDPOINT');

// Contract
const SYBIL_CONTRACT_ADDRESS = retrieveEnvVariable('SYBIL_CONTRACT_ADDRESS');

module.exports = { 
  PRIVATE_KEY, 
  RPC_ENDPOINT,
  SYBIL_CONTRACT_ADDRESS
};