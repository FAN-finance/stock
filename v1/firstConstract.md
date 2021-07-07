
```shell script
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/utils/cryptography/ECDSAUpgradeable.sol";

contract God is Initializable, OwnableUpgradeable {
    using ECDSAUpgradeable for bytes32;

    struct OP {
        address token;
        uint256 price;
        uint256 timestamp;
    }

    uint256 EXPIRY;
    uint256 SIGNATURENUM;

    mapping(address => bool) authorization;
    mapping(address => uint256) timestamps;
    mapping(address => uint256) prices;
    mapping(address => uint256) tax_mint;
    mapping(address => uint256) tax_burn;

    function initialize() public initializer {
        __Context_init_unchained();
        __Ownable_init_unchained();
    }

    function add_authorization(address addr) public onlyOwner {
        authorization[addr] = true;
    }

    function remove_authorization(address addr) public onlyOwner {
        delete authorization[addr];
    }

    function set_tax(address token, uint256[2] memory tax) public onlyOwner {
        set_tax_mint(token, tax[0]);
        set_tax_burn(token, tax[1]);
    }

    function set_tax_mint(address token, uint256 tax) public onlyOwner {
        tax_mint[token] = tax;
    }

    function set_tax_burn(address token, uint256 tax) public onlyOwner {
        tax_burn[token] = tax;
    }

    function set_expiry(uint256 data) public onlyOwner {
        EXPIRY = data;
    }

    function set_signature_num(uint256 data) public onlyOwner {
        SIGNATURENUM = data;
    }

    function swap(uint256[2] calldata ns, bytes[2] calldata oracle) public {
        OP memory x = sync(oracle[0]);
        OP memory y = sync(oracle[1]);
        require(x.timestamp == y.timestamp);
        uint256 nx = ns[0] * (1e18 - tax_burn[x.token]);
        uint256 ny = ns[1] * (1e18 + tax_mint[x.token]);
        require(nx * x.price >= ny * y.price);
        I(x.token).burn(msg.sender, ns[0]);
        I(y.token).mint(msg.sender, ns[1]);
    }

    function swap_limit_x(uint256[2] calldata ns, bytes[2] calldata oracle)
        public
    {
        OP memory x = sync(oracle[0]);
        OP memory y = sync(oracle[1]);
        require(x.timestamp == y.timestamp);
        uint256 ny = ns[1] * (1e18 + tax_mint[x.token]);
        uint256 nsx = (ny * y.price) / (x.price * 1e18 - tax_burn[x.token]) + 1;
        require(nsx <= ns[0]);
        I(x.token).burn(msg.sender, nsx);
        I(y.token).mint(msg.sender, ns[1]);
    }

    function swap_limit_y(uint256[2] calldata ns, bytes[2] calldata oracle)
        public
    {
        OP memory x = sync(oracle[0]);
        OP memory y = sync(oracle[1]);
        require(x.timestamp == y.timestamp);
        uint256 nx = ns[0] * (1e18 - tax_burn[x.token]);
        uint256 nsy = (nx * x.price) / (y.price * (1e18 + tax_mint[x.token]));
        require(nsy >= ns[1]);
        I(x.token).burn(msg.sender, ns[0]);
        I(y.token).mint(msg.sender, nsy);
    }

    function sync(bytes calldata data) internal returns (OP memory) {
        (bytes memory o, bytes[] memory s) = abi.decode(data, (bytes, bytes[]));
        OP memory op = abi.decode(o, (OP));
        if (op.timestamp <= timestamps[op.token]) {
            op.price = prices[op.token];
            op.timestamp = timestamps[op.token];
        } else {
            prices[op.token] = op.price;
            timestamps[op.token] = op.timestamp;
            require(s.length == SIGNATURENUM);
            bytes32 hash = keccak256(o).toEthSignedMessageHash();
            address auth = address(0);
            for (uint256 i = 0; i < s.length; i++) {
                address addr = hash.recover(s[i]);
                require(addr > auth);
                require(authorization[addr]);
                auth = addr;
            }
        }

        require(op.timestamp + EXPIRY > block.timestamp);
        return op;
    }
}

interface I {
    function mint(address, uint256) external;

    function burn(address, uint256) external;
}
```