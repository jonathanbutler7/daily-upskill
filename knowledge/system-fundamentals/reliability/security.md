# Security

How Weave protects systems and data.

## Security Architecture Diagram

---

## Questions to Ask

- [ ] How do services authenticate with each other?
- [ ] How do we manage secrets?
- [ ] What's our approach to least privilege?
- [ ] How do we handle security vulnerabilities?
- [ ] What compliance requirements do we have?
- [ ] How do we handle user authentication?

---

## Key Concepts (System Design Interview)

### Authentication vs Authorization
| Concept | Question | Example |
|---------|----------|---------|
| Authentication | Who are you? | Login with username/password |
| Authorization | What can you do? | Check if user can delete resource |
| **Weave's approach**: | [Fill in] | |

### Service-to-Service Authentication
| Method | Description | Pros | Cons |
|--------|-------------|------|------|
| mTLS | Mutual TLS certificates | Strong, automatic | Certificate management |
| JWT | Signed tokens | Flexible, stateless | Token management |
| API Keys | Static keys | Simple | Key rotation |
| **Weave uses**: | [Fill in] | | |

### Secrets Management
| Approach | Description | Example |
|----------|-------------|---------|
| Environment variables | Injected at runtime | K8s secrets |
| Secrets manager | Centralized vault | HashiCorp Vault, AWS Secrets Manager |
| Sidecar | Injected by sidecar | Vault Agent |
| **Weave uses**: | [Fill in] | |

### Security Principles
- **Least Privilege**: Minimum necessary permissions
- **Defense in Depth**: Multiple security layers
- **Zero Trust**: Verify everything, trust nothing
- **Weave's principles**: [Fill in]

### Encryption
| Type | What It Protects | Example |
|------|-----------------|---------|
| In Transit | Data moving between systems | TLS/HTTPS |
| At Rest | Stored data | Encrypted disks, DB encryption |
| **Weave's encryption**: | [Fill in] | |

---

## Weave Implementation

### Authentication Flow
1. User authenticates: [How?]
2. Token issued: [What type?]
3. Token validated: [How?]
4. Authorization checked: [How?]

### Service Authentication
- **Method**: [mTLS / JWT / etc.]
- **Certificate management**: [How?]
- **Rotation policy**: [How often?]

### Secrets Management
| Secret Type | Storage | Rotation | Access |
|-------------|---------|----------|--------|
| API keys | [Where] | [Frequency] | [Who] |
| Database credentials | [Where] | [Frequency] | [Who] |
| TLS certificates | [Where] | [Frequency] | [Who] |

### RBAC Configuration
| Role | Permissions | Scope |
|------|-------------|-------|
| [Role 1] | [What they can do] | [Where] |
| [Role 2] | [What they can do] | [Where] |

### Security Scanning
| Type | Tool | Frequency |
|------|------|-----------|
| Dependency scanning | [Tool] | [When] |
| Container scanning | [Tool] | [When] |
| SAST | [Tool] | [When] |
| DAST | [Tool] | [When] |

### Vulnerability Response
1. **Detection**: [How we find vulnerabilities]
2. **Triage**: [How we assess severity]
3. **Remediation**: [How we fix]
4. **Verification**: [How we confirm fix]

### Compliance
| Requirement | Status | Evidence |
|-------------|--------|----------|
| [Requirement 1] | [Status] | [Link] |
| [Requirement 2] | [Status] | [Link] |

---

## Resources

- Security Documentation: [Link]
- Secrets Management Guide: [Link]
- Security Incident Response: [Link]

---

## Notes

*Add your notes here as you learn:*
