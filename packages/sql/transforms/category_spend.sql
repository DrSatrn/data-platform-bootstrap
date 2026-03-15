-- This transform materializes category-level spend from the intermediate
-- rollup layer. It stays separate from the monthly cashflow mart because
-- reporting consumers often need category detail without coupling themselves
-- to raw or staging rows.

create or replace table mart_category_spend as
select
  month,
  category,
  expense_total as actual_spend
from intermediate_category_monthly_rollup
order by 1, 2;
