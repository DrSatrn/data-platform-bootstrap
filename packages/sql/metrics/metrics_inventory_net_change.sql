-- This metric table exposes warehouse-level net inventory change over time.
create or replace table metrics_inventory_net_change as
select
  month,
  warehouse,
  sum(net_quantity) as net_quantity
from mart_inventory_monthly_summary
group by 1, 2
order by month, warehouse;
