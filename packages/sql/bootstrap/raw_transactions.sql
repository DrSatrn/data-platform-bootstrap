-- This bootstrap SQL loads the landed transaction CSV into DuckDB and normalizes
-- the category and amount fields into a stable analytical shape.

create or replace table raw_transactions as
select
  transaction_id,
  cast(occurred_at as timestamp) as occurred_at,
  account_name,
  nullif(trim(category), '') as category,
  cast(amount as double) as amount
from read_csv_auto({{RAW_TRANSACTIONS_PATH}}, header = true);
