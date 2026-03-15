-- This bootstrap SQL loads landed inventory stock movements into DuckDB so
-- domain-specific transforms can generalize beyond the personal-finance slice.
create or replace table raw_stock_movements as
select
  sku,
  cast(movement_date as date) as movement_date,
  movement_type,
  warehouse,
  cast(quantity as integer) as quantity,
  cast(unit_cost as double) as unit_cost
from read_csv_auto({{RAW_STOCK_MOVEMENTS_PATH}}, header = true);
