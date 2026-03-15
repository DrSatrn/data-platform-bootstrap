-- This transform materializes an intermediate monthly category rollup from the
-- staging-enriched transaction dataset. Keeping this layer explicit makes the
-- path from staging to marts easier to study and reuse.

create or replace table intermediate_category_monthly_rollup as
select
  strftime(date_trunc('month', occurred_at), '%Y-%m') as month,
  category,
  category_group,
  sum(case when amount < 0 then -amount else 0 end) as expense_total,
  count(*) as transaction_count
from staging_transactions_enriched
where amount < 0
group by 1, 2, 3
order by 1, 2;
