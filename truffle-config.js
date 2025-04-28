// Configures Truffle to deploy to a local development network.
export const networks = {
  development: {
    host: "127.0.0.1", // Localhost (default: none)
    port: 8545, // Standard Ethereum port (default: none)
    network_id: "*", // Allow any network ID
  },
};
export const compilers = {
  solc: {
    version: "0.8.0", // Match Solidity version in contract
  },
};
