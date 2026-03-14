-- This bootstrap SQL loads the landed account balance JSON into DuckDB so
-- quality and downstream marts can reference the latest account snapshots.

create or replace table raw_account_balances as
select
  account_id,
  cast(captured_at as timestamp) as captured_at,
  cast(balance as double) as balance
from read_json_auto({{RAW_ACCOUNT_BALANCES_PATH}});
