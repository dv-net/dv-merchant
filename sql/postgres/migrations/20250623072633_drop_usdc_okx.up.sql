DELETE FROM exchange_chains
WHERE
	slug = 'okx'
	AND currency_id IN ('USDC.Ethereum', 'USDC.Polygon');