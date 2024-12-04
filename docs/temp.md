Go into tutotial: 

Below is an example Rust function to extract Solana instructions from a specific program ID and USDT transactions from an EVM blockchain. You can adapt this to your specific requirements.

{% tabs %}
{% tab title="Solana" %}
The following function extracts Solana instructions from a specific program ID.
```rust
fn map_filter_instructions(params: String, blk: Block) -> Result<Instructions, Vec<substreams::errors::Error>> {
    let filters = parse_filters_from_params(params)?;

    let mut instructions : Vec<Instruction> = Vec::new();

    blk.transactions.iter().for_each(|tx| {
        let msg = tx.transaction.clone().unwrap().message.unwrap();
        let acct_keys = msg.account_keys.as_slice();
        let insts : Vec<Instruction> = msg.instructions.iter()
            .filter(|inst| apply_filter(inst, &filters, acct_keys.to_vec()))
            .map(|inst| {
            Instruction {
                program_id: bs58::encode(acct_keys[inst.program_id_index as usize].to_vec()).into_string(),
                accounts: inst.accounts.iter().map(|acct| bs58::encode(acct_keys[*acct as usize].to_vec()).into_string()).collect(),
                data: bs58::encode(inst.data.clone()).into_string(),
            }
        }).collect();
        instructions.extend(insts);
    });

    Ok(Instructions { instructions })
}
```
{% endtab %}

{% tab title="EVM" %}
The following function extracts USDT transaction from EVM blockchains.
```rust
fn get_usdt_transaction(block: eth::Block) -> Result<Vec<Transaction>, substreams:error:Error> {
    let my_transactions = block.transactions().
        .filter(|transaction| transaction.to == USDT_CONTRACT_ADDRESS)
        .map(|transaction| MyTransaction(transaction.hash, transaction.from, transaction.to))
        .collect();
    Ok(my_transactions)
}
```
{% endtab %}
{% endtabs %}