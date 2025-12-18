// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";

// Wormhole 核心接口
interface IWormhole {
    struct Signature {
        bytes32 r;
        bytes32 s;
        uint8 v;
        uint8 guardianIndex;
    }

    struct VM {
        uint8 version;
        uint32 timestamp;
        uint32 nonce;
        uint16 emitterChainId;
        bytes32 emitterAddress;
        uint64 sequence;
        uint8 consistencyLevel;
        bytes payload;
        uint32 guardianSetIndex;
        Signature[] signatures;
        bytes32 hash;
    }
    
    function publishMessage(
        uint32 nonce,
        bytes memory payload,
        uint8 consistencyLevel
    ) external payable returns (uint64 sequence);
    
    function parseAndVerifyVM(
        bytes calldata encodedVM
    ) external view returns (VM memory vm, bool valid, string memory reason);
    
    function chainId() external view returns (uint16);
    function messageFee() external view returns (uint256);
}

/**
 * @title SimpleWormholeToken
 * @dev 简单 Wormhole 跨链代币合约
 */
contract SimpleWormholeToken is ERC20 {
    // Wormhole 核心合约
    IWormhole public immutable wormhole;
    uint16 public immutable chainId;
    
    // 已处理的 VAA
    mapping(bytes32 => bool) public processedVAAs;
    
    // 事件
    event CrossChainSent(
        uint64 sequence,
        address sender,
        uint16 targetChain,
        bytes32 targetContract,
        address recipient,
        uint256 amount
    );
    
    event CrossChainReceived(
        bytes32 vaaHash,
        uint16 sourceChain,
        bytes32 sourceContract,
        address recipient,
        uint256 amount
    );
    
    constructor(
        address wormholeAddress,
        string memory name,
        string memory symbol,
        uint256 initialSupply
    ) ERC20(name, symbol) {
        require(wormholeAddress != address(0), "Invalid wormhole address");
        
        wormhole = IWormhole(wormholeAddress);
        chainId = wormhole.chainId();
        
        if (initialSupply > 0) {
            _mint(msg.sender, initialSupply);
        }
    }
    
    // ============ 核心功能 ============
    
    /**
     * @dev 发送跨链转移
     */
    function sendCrossChain(
        uint16 targetChain,
        address targetContract,
        address recipient,
        uint256 amount,
        uint32 nonce
    ) external payable returns (uint64 sequence) {
        // 验证参数
        require(balanceOf(msg.sender) >= amount, "Insufficient balance");
        require(targetChain != chainId, "Cannot send to same chain");
        require(recipient != address(0), "Invalid recipient");
        require(amount > 0, "Amount must be greater than 0");
        
        // 检查费用
        uint256 fee = wormhole.messageFee();
        require(msg.value >= fee, "Insufficient fee");
        
        // 销毁代币
        _burn(msg.sender, amount);
        
        // 构建消息
        bytes memory payload = abi.encode(
            addressToBytes32(targetContract),                     // 目标合约
            keccak256("TOKEN_TRANSFER"),        // 消息类型
            recipient,                          // 接收地址
            amount,                             // 数量
            chainId,                            // 源链ID
            addressToBytes32(address(this)),    // 源代币地址
            block.timestamp                     // 时间戳
        );
        
        // 发送到 Wormhole
        sequence = wormhole.publishMessage{value: fee}(
            nonce,
            payload,
            200 // 高一致性级别
        );
        
        // 退还多余的费用
        if (msg.value > fee) {
            payable(msg.sender).transfer(msg.value - fee);
        }
        
        emit CrossChainSent(sequence, msg.sender, targetChain, addressToBytes32(targetContract), recipient, amount);
        
        return sequence;
    }
    
    /**
     * @dev 接收跨链转移
     */
    function receiveCrossChain(bytes calldata vaa) external returns (bool) {
        // 验证 VAA
        (IWormhole.VM memory vm, bool valid, string memory reason) = 
            wormhole.parseAndVerifyVM(vaa);
        
        require(valid, reason);
        
        // 检查是否已处理
        require(!processedVAAs[vm.hash], "VAA already processed");
        
        // 解析 payload
        (
            bytes32 targetContract,
            bytes32 messageType,
            address recipient,
            uint256 amount,
            uint16 sourceChain,
            bytes32 sourceContract,
            uint256 timestamp
        ) = abi.decode(vm.payload, (bytes32, bytes32, address, uint256, uint16, bytes32, uint256));
        
        // 验证消息类型
        require(messageType == keccak256("TOKEN_TRANSFER"), "Invalid message type");
        
        // 验证目标合约
        require(targetContract == addressToBytes32(address(this)), "Not target contract");
        
        // 时间戳验证（可选）
        require(block.timestamp < timestamp + 24 hours, "VAA expired");
        
        // 铸造代币
        _mint(recipient, amount);
        
        // 标记为已处理
        processedVAAs[vm.hash] = true;
        
        emit CrossChainReceived(vm.hash, sourceChain, sourceContract, recipient, amount);
        
        return true;
    }
    /**
     * @dev 地址转 bytes32
     */
    function addressToBytes32(address addr) internal pure returns (bytes32) {
        return bytes32(uint256(uint160(addr)));
    }
    
    /**
     * @dev 铸造代币
     */
    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }
    
    /**
     * @dev 销毁代币
     */
    function burn(uint256 amount) external {
        _burn(msg.sender, amount);
    }
    
    /**
     * @dev 获取 Wormhole 信息
     */
    function getWormholeInfo() external view returns (uint16, uint256) {
        return (wormhole.chainId(), wormhole.messageFee());
    }
    
    /**
     * @dev 验证 VAA
     */
    function verifyVAA(bytes calldata vaa) external view returns (bool valid, string memory reason) {
        (, valid, reason) = wormhole.parseAndVerifyVM(vaa);
    }
    
    /**
     * @dev 检查 VAA 是否已处理
     */
    function isVAAProcessed(bytes32 vaaHash) external view returns (bool) {
        return processedVAAs[vaaHash];
    }
    
    /**
     * @dev 估算费用
     */
    function estimateFee() external view returns (uint256) {
        return wormhole.messageFee();
    }
    
    // 接收 ETH
    receive() external payable {}
}
