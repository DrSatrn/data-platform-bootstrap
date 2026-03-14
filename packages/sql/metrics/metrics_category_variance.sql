-- This metric SQL exposes category-level budget variance as a curated metric
-- table. Keeping it separate from the mart makes downstream reporting queries
-- explicit and easy to constrain.

create or replace table metrics_category_variance as
select
  month,
  category,
  variance_amount
from mart_budget_vs_actual
order by month, category;
