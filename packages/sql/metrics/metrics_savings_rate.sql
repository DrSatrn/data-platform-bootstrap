-- This metric SQL projects the reporting-friendly savings-rate series from the
-- curated mart. Keeping metrics in SQL makes the analytical contract easy to
-- reason about and reuse.

create or replace table metrics_savings_rate as
select
  month,
  savings_rate
from mart_monthly_cashflow
order by month;
