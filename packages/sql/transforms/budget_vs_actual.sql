-- This transform joins monthly category spend with budget rules so the
-- reporting layer can surface overspend and underspend without embedding
-- business logic in the UI.

create or replace table mart_budget_vs_actual as
with months as (
  select distinct month
  from mart_category_spend
),
budget_grid as (
  select
    months.month,
    raw_budget_rules.category,
    raw_budget_rules.monthly_budget as budget_amount
  from months
  cross join raw_budget_rules
)
select
  budget_grid.month,
  budget_grid.category,
  budget_grid.budget_amount,
  coalesce(mart_category_spend.actual_spend, 0) as actual_spend,
  coalesce(mart_category_spend.actual_spend, 0) - budget_grid.budget_amount as variance_amount
from budget_grid
left join mart_category_spend
  on mart_category_spend.month = budget_grid.month
 and mart_category_spend.category = budget_grid.category
order by 1, 2;
