// SPDX-License-Identifier: GPL-3.0

pragma solidity >=0.8.2 <0.9.0;

contract Sybil { 
    uint number_nodes;
    mapping (uint => address) nodes;
    mapping (address => uint) deposits;
    mapping (address => uint) totalStaked;
    mapping (address => uint) scores;
    mapping (address => mapping (address => uint)) stakes;

    constructor()  {
      number_nodes = 0;  
    }


    function addDeposit() external payable {
        deposits[msg.sender] += msg.value;
        nodes[number_nodes] = msg.sender;
        number_nodes = number_nodes + 1;
    }

    function increaseStake(address target, uint amount) public {
        if (totalStaked[msg.sender] + amount <= deposits[msg.sender]) {
            stakes[msg.sender][target] += amount;
            totalStaked[msg.sender] += amount;
        }
    }

    function decreaseStake(address target, uint amount) public {
        stakes[msg.sender][target] -= amount;
        totalStaked[msg.sender] -= amount;
    }


    function calculateScores() public {

    }







    function getDeposit(address node) public view returns (uint) {
        return deposits[node];
    }

    function getTotalstaked(address node) public view returns (uint) {
        return totalStaked[node];
    }

    function getNumbernodes() public view returns (uint) {
        return number_nodes;
    }
    function getStake(address source, address target) public view returns (uint) {
        return stakes[source][target];
    }

    function getNodes() public view returns (address[] memory){
        address[] memory ret = new address[](number_nodes);
        for (uint i = 0; i < number_nodes; i++) {
            ret[i] = nodes[i];
        }
        return ret;
    }

    function getTotalstakeds() public view returns (uint[] memory){
        uint[] memory ret = new uint[](number_nodes);
        for (uint i = 0; i < number_nodes; i++) {
            address add = nodes[i];
            ret[i] = totalStaked[add];
        }
        return ret;
    }

    function getDeposits() public view returns (uint[] memory){
        uint[] memory ret = new uint[](number_nodes);
        for (uint i = 0; i < number_nodes; i++) {
            address add = nodes[i];
            ret[i] = deposits[add];
        }
        return ret;
    }

    function getStakes() public view returns (uint[] memory){
        uint[] memory ret = new uint[](number_nodes*number_nodes);
        for (uint i = 0; i < number_nodes; i++) {
            for (uint j = 0; j < number_nodes; j++) {
                ret[(i*number_nodes)+j]=stakes[nodes[i]][nodes[j]]; 
            }
        }
        return ret;
    }


    function getScore(address node) public view returns (uint) {
        return scores[node];
    }

    function getScores() public view returns (uint[] memory){
        uint[] memory ret = new uint[](number_nodes);
        for (uint i = 0; i < number_nodes; i++) {
            address add = nodes[i];
            ret[i] = scores[add];
        }
        return ret;
    }

}
