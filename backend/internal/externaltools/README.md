# External Tool Execution

This package executes bounded optional pipeline tools through adapters.

Current status:

- generic runner is implemented
- `dbt` is the first adapter
- declared artifacts are captured after successful execution
- `dlt` and `pyspark` are reserved in the manifest contract but intentionally
  rejected during validation for now

The control plane remains authoritative. External tools are subprocess jobs
inside a normal pipeline run, not alternate schedulers or control planes.
