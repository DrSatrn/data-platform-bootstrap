-- This bootstrap SQL loads local budget rules into DuckDB so variance-oriented
-- curated marts can stay declarative and version-controlled.

create or replace table raw_budget_rules as
select
  category,
  cast(monthly_budget as double) as monthly_budget
from read_json_auto({{RAW_BUDGET_RULES_PATH}});
