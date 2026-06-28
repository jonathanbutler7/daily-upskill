from typing import Protocol, runtime_checkable

# --- Exercise: Log Exporter Protocol ---
# Task: Create a Protocol that defines how a log exporter should behave.
# Then, implement two different exporters that meet this protocol without 
# explicitly inheriting from it.

@runtime_checkable
class LogExporter(Protocol):
    """
    Define a protocol that expects an 'export' method.
    The method should take a list of strings and return a boolean success.
    
    Keywords to search for:
    - "Python typing.Protocol"
    - "Structural subtyping Python"
    - "Go-style interfaces in Python"
    """
    def export(self, logs: list[str]) -> bool:
        ...

class S3Exporter:
    """
    Implement the LogExporter protocol for S3.
    Requirement: 
    - Print 'Uploading {len(logs)} logs to S3 bucket...'
    - Return True
    """
    def export(self, logs: list[str]) -> bool:
        print(f"Sending {len(logs)} to S3")
        return True

class DatadogExporter:
    """
    Implement the LogExporter protocol for Datadog.
    Requirement:
    - Print 'Sending {len(logs)} logs to Datadog via HTTP API...'
    - Return True
    """
    def export(self, logs: list[str]) -> bool:
        print(f"Sending {len(logs)} to Datadog")
        return True

def run_export(exporter: LogExporter, logs: list[str]):
    """
    A generic function that can take any exporter that 
    satisfies the LogExporter protocol.
    """
    print(f"DEBUG: Starting export with {exporter.__class__.__name__}...")
    
    success = exporter.export(logs)
    return success

if __name__ == "__main__":
    logs = ["INFO: System started", "ERROR: Connection failed"]
    
    # Test your implementation here:
    s3 = S3Exporter()
    success = run_export(s3, logs)
    print(f"Successfully exported S3 {success}")
    #
    dd = DatadogExporter()
    success = run_export(dd, logs)
    print(f"Successfully exported Datadog {success}")
