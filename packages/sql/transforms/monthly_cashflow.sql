-- This transform materializes the curated monthly cashflow mart used by the
-- reporting layer. The logic is intentionally explicit rather than relying on a
-- hidden semantic layer so the business rules remain easy to audit.

create or replace table mart_monthly_cashflow as
with monthly as (
  select
    strftime(date_trunc('month', occurred_at), '%Y-%m') as month,
    sum(case when amount >= 0 then amount else 0 end) as income,
    sum(case when amount < 0 then -amount else 0 end) as expenses
  from raw_transactions
  group by 1
)
select
  month,
  income,
  expenses,
  case
    when income > 0 then (income - expenses) / income
    else 0
  end as savings_rate
from monthly
order by month;
