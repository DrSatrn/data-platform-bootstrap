-- This quality query counts records whose category is missing or blank so the
-- platform can surface categorization coverage regressions.

select
  count(*) as uncategorized_count
from raw_transactions
where category is null or trim(category) = '';
