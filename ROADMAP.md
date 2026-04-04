# Skill Roadmap (2026)

## Table of Contents

- [Module 1: Python Foundations](#module-1-python-foundations-completed) ✅
- [Module 2: ML Engineering & Tools](#module-2-ml-engineering--tools)
- [Module 3: Data Engineering & Architecture](#module-3-data-engineering--architecture)
- [Module 4: System Design & Backend Patterns](#module-4-system-design--backend-patterns)
- [Module 5: Capstone Project](#module-5-capstone-project)

---

## Module 1: Python Foundations (Completed)
- [x] Using comprehensions & generators for memory-efficient data processing
- [x] Validating data boundaries with typing and Pydantic
- [x] Managing resources efficiently using context managers and file I/O
- [x] Writing cleaner data pipelines using map, filter, and reduce
- [x] Writing decorators for reusable logic (logging, metrics, retries)
- [x] Using Protocols for structural typing (Go-style interfaces in Python)
- [x] Testing service interactions using pytest fixtures and mocks
- [x] **Module 1 Quiz & Knowledge Check**

---

## Module 2: ML Engineering & Tools
> Builds on: Module 1 — Pydantic, Protocols, decorators, pytest

- [ ] Standardizing Hugging Face inference using Pydantic BaseModels
- [ ] Implementing a reusable ModelEvaluator class for common metrics
- [ ] Benchmarking model configurations using functional pipelines and Pydantic result models
- [ ] Building a multi-tenant vector DB retrieval interface using Protocols for RAG
- [ ] Versioning and templating prompts for A/B testing
- [ ] Wrapping quantized model inference (GGUF/AWQ) behind a Protocol-based interface with Pydantic I/O schemas
- [ ] Setting up CI/CD tests for model accuracy and latency regressions using pytest fixtures
- [ ] **Module 2 Quiz & Knowledge Check**

---

## Module 3: Data Engineering & Architecture
> Builds on: Modules 1–2 — Pydantic data contracts, Protocol-based model interfaces, functional pipelines

- [ ] Enforcing data contracts across services using Polars, DuckDB, and Pydantic
- [ ] Writing reusable Airflow operators using decorators and context managers
- [ ] Building a typed validation pipeline for ML model outputs using Pydantic and Polars
- [ ] Processing ML batch inference outputs stream-style using generators and Pydantic schemas
- [ ] Building resilient Kafka/RabbitMQ consumers with dead-letter queues using retry decorators
- [ ] Serving high-performance data APIs using FastAPI and Pydantic
- [ ] Implementing automated data retention and archival using context managers
- [ ] **Module 3 Quiz & Knowledge Check**

---

## Module 4: System Design & Backend Patterns
> Builds on: Modules 1–3 — decorators, context managers, Protocols, Pydantic, FastAPI, data pipelines

- [ ] Building a shared Redis wrapper with automatic serialization
- [ ] Implementing token-bucket rate limiting as decorator-based middleware
- [ ] Standardizing cross-service logging and tracing using decorators
- [ ] Designing a composable ML serving system using Protocols, Pydantic, and decorator-based middleware
- [ ] Adding circuit breakers and bulkheads to prevent cascading failures
- [ ] Profiling an ML inference and data pipeline end-to-end using decorator-based timing and functional metrics aggregation
- [ ] Implementing a safe secrets manager using the context manager pattern with Protocol-based credential providers
- [ ] **Module 4 Quiz & Knowledge Check**

---

## Module 5: Capstone Project
> Builds on: Modules 1–4 — full synthesis of all skills into one production-ready system

*Build a production-ready ML data pipeline end-to-end. Each part is a standalone exercise that requires skills from the specified prior modules.*

- [ ] **Part 1 — Typed ingestion layer**: read raw input data with context managers and validate with Pydantic data contracts *(M1)*
- [ ] **Part 2 — ML inference layer**: define a Protocol-based model interface and implement it with full Pydantic I/O schemas *(M1 + M2)*
- [ ] **Part 3 — Data pipeline**: process inference outputs stream-style using generators and functional patterns *(M1 + M2 + M3)*
- [ ] **Part 4 — Resilience layer**: wrap inference and pipeline steps with retry decorators and circuit breaker middleware *(M1 + M4)*
- [ ] **Part 5 — API layer**: serve the full pipeline via FastAPI with Pydantic request/response models *(M3 + M4)*
- [ ] **Part 6 — Test suite**: cover all layers end-to-end using pytest fixtures and mocks *(M1 through M4)*
- [ ] **Capstone Quiz & Final Review**

---

Mark done when you have a working exercise + concept file entry.
