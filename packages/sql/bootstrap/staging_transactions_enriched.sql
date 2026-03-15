-- This bootstrap SQL loads the Python-enriched staging transaction JSON so
-- downstream intermediate and mart models can consume standardized categories
-- without embedding enrichment rules in SQL.

create or replace table staging_transactions_enriched as
select
  transaction_id,
  cast(occurred_at as timestamp) as occurred_at,
  account_name,
  category,
  category_group,
  normalized_description,
  cast(inferred_category as boolean) as inferred_category,
  cast(amount as double) as amount
from read_json_auto({{STAGING_TRANSACTIONS_ENRICHED_PATH}});
