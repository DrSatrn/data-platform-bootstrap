-- This transform materializes a compact monthly inventory summary by SKU and
-- warehouse. It intentionally mirrors the finance marts: one curated table that
-- can drive both analytics and metric publication.
create or replace table mart_inventory_monthly_summary as
with monthly as (
  select
    strftime(date_trunc('month', movement_date), '%Y-%m') as month,
    sku,
    warehouse,
    sum(case when quantity >= 0 then quantity else 0 end) as receipts,
    sum(case when quantity < 0 then -quantity else 0 end) as issues,
    sum(quantity) as net_quantity,
    count(*) as movement_count
  from raw_stock_movements
  group by 1, 2, 3
)
select
  month,
  sku,
  warehouse,
  receipts,
  issues,
  net_quantity,
  sum(net_quantity) over (
    partition by sku, warehouse
    order by month
    rows between unbounded preceding and current row
  ) as closing_quantity,
  movement_count
from monthly
order by month, sku, warehouse;
