-- This quality query flags any SKU or warehouse slice that ends a month below
-- zero inventory.
select count(*) as count
from mart_inventory_monthly_summary
where closing_quantity < 0;
