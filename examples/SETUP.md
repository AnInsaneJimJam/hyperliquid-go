# Hyperliquid Go SDK Examples Setup

## Getting Started

### 1. Get a Hyperliquid Testnet API Key

#### Create Account & Get Testnet Access
1. Visit [https://app.hyperliquid.xyz/](https://app.hyperliquid.xyz/)
2. Connect your wallet (MetaMask recommended)
3. Switch to **Testnet** using the network selector in the top-right corner
4. Get testnet funds from the faucet (look for "Faucet" or "Get Test Funds" button)

#### Generate API Key
1. Go to **Settings** → **API Keys**
2. Click **"Create API Key"** 
3. Enable trading permissions for testing
4. **Copy the private key** (starts with `0x...`) - this is your secret key

⚠️ **Security Note**: Never share your private key or commit it to version control!

### 2. Configure Your Environment

#### Update config.json
```bash
cp config.json.example config.json
```

Edit `config.json` and replace the placeholder:
```json
{
    "secret_key": "0xYOUR_ACTUAL_64_CHARACTER_PRIVATE_KEY_HERE",
    "account_address": ""
}
```

**Note**: Leave `account_address` empty - it will be derived automatically from your private key.

### 3. Test the Setup

#### Compile and Run Basic Order Example
```bash
go build basic_order.go example_utils.go config.go test_main.go
./a.out basic_order
```

#### Expected Output
```
Running with account address: 0x...
User state retrieved successfully for address: 0x...
no open positions
Placing order for ETH...
Order placed successfully: {...}
```

### 4. Available Examples

Run any example with:
```bash
go run basic_order.go example_utils.go config.go test_main.go <example_name>
```

Available examples:
- `basic_order` - Place and cancel a limit order
- `basic_market_order` - Place a market order  
- `basic_leverage` - Adjust account leverage
- `basic_ws` - WebSocket data streaming
- `basic_adding` - Liquidity provision strategy

### 5. Troubleshooting

#### "asset not found for name: ETH"
- This means metadata loading failed
- Check your internet connection and API key
- Ensure you're using testnet

#### "failed to get user state"
- Invalid API key or network issues
- Verify your private key is correct
- Check you're on testnet

#### "no equity" error
- Get testnet funds from the faucet first
- Ensure funds are deposited to your account

### 6. Next Steps

Once `basic_order` works, you can test other examples and modify them for your use case.

For production use:
1. Switch to mainnet API URL
2. Use a production private key
3. Implement proper key management (environment variables, key vaults, etc.)
