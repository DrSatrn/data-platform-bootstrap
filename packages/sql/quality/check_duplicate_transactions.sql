-- This quality query counts duplicated transaction identifiers in the raw
-- landing table. The result is used to surface operator-visible trust signals.

select
  count(*) as duplicate_count
from (
  select transaction_id
  from raw_transactions
  group by transaction_id
  having count(*) > 1
) duplicates;
