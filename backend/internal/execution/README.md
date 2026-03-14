# Execution Package

This package will own job execution contracts, worker-side task handling, artifact capture, and retry-aware result reporting. The execution layer should remain explicit because it is where partial failures become operationally visible.
