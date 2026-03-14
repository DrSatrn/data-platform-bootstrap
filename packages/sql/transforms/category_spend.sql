-- This transform materializes category-level spend from the raw transaction
-- landing table. It stays separate from the monthly cashflow mart because
-- reporting consumers often need category detail without coupling themselves
-- to raw signed transaction rows.

create or replace table mart_category_spend as
select
  strftime(date_trunc('month', occurred_at), '%Y-%m') as month,
  coalesce(category, 'Uncategorized') as category,
  sum(case when amount < 0 then -amount else 0 end) as actual_spend
from raw_transactions
where amount < 0
group by 1, 2
order by 1, 2;
