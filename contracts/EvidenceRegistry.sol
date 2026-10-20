// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @title BanguninAja evidence registry
/// @notice Stores the Merkle root of each batch of BanguninAja records so anyone
/// can check when a batch was locked and that it has not been rewritten since.
/// The records themselves never leave the BanguninAja server; only their root does.
contract EvidenceRegistry {
    address public owner;

    mapping(address => bool) public anchorers;
    mapping(bytes32 => uint64) public anchoredAt;
    mapping(bytes32 => uint32) public recordCount;
    mapping(bytes32 => address) public anchoredBy;

    event Anchored(bytes32 indexed root, uint32 recordCount, address indexed anchorer, uint64 anchoredAt);
    event AnchorerChanged(address indexed account, bool allowed);
    event OwnershipTransferred(address indexed previousOwner, address indexed newOwner);

    error NotOwner();
    error NotAnchorer();
    error EmptyRoot();
    error AlreadyAnchored(bytes32 root);
    error ZeroAddress();

    modifier onlyOwner() {
        if (msg.sender != owner) revert NotOwner();
        _;
    }

    /// @param initialAnchorers wallets allowed to anchor besides the deployer,
    /// for example the BanguninAja server wallet.
    constructor(address[] memory initialAnchorers) {
        owner = msg.sender;
        emit OwnershipTransferred(address(0), msg.sender);
        _setAnchorer(msg.sender, true);
        for (uint256 i = 0; i < initialAnchorers.length; i++) {
            _setAnchorer(initialAnchorers[i], true);
        }
    }

    /// @notice Records a batch root. Each root can be anchored only once.
    function anchor(bytes32 root, uint32 count) external {
        if (!anchorers[msg.sender]) revert NotAnchorer();
        if (root == bytes32(0)) revert EmptyRoot();
        if (anchoredAt[root] != 0) revert AlreadyAnchored(root);
        uint64 timestamp = uint64(block.timestamp);
        anchoredAt[root] = timestamp;
        recordCount[root] = count;
        anchoredBy[root] = msg.sender;
        emit Anchored(root, count, msg.sender, timestamp);
    }

    /// @notice True when the root has been anchored.
    function isAnchored(bytes32 root) external view returns (bool) {
        return anchoredAt[root] != 0;
    }

    function setAnchorer(address account, bool allowed) external onlyOwner {
        _setAnchorer(account, allowed);
    }

    function transferOwnership(address newOwner) external onlyOwner {
        if (newOwner == address(0)) revert ZeroAddress();
        emit OwnershipTransferred(owner, newOwner);
        owner = newOwner;
    }

    function _setAnchorer(address account, bool allowed) private {
        if (account == address(0)) revert ZeroAddress();
        anchorers[account] = allowed;
        emit AnchorerChanged(account, allowed);
    }
}
