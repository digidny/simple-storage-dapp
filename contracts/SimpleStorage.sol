// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract SimpleStorage {
    uint256 storedData;

    constructor(uint256 initVal) {
        storedData = initVal;
    }

    function set(uint256 x) public {
        storedData = x;
    }

    function get() public view returns (uint256) {
        return storedData;
    }

     function add(uint256 x) public returns (uint256) {
        storedData = storedData + x;
        return storedData;
    }
}
