# Quality Package

This package owns data quality definitions and runtime result surfaces. Quality checks should stay explicit and operator-visible because silent data trust failures are costly.

The current implementation derives quality signals from the sample finance dataset so the first UI slice has meaningful, explainable status without requiring external tools.
