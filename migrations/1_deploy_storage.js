const SimpleStorage = artifacts.require("./SimpleStorage.sol");

export default function (deployer) {
  deployer.deploy(SimpleStorage, 5); // Initial value of 5
};

