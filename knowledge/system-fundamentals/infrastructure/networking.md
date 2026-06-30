# Networking

Network architecture and connectivity at Weave.

## Network Topology Diagram

---

## Questions to Ask

- [ ] How is our VPC structured?
- [ ] How do services in different clusters communicate?
- [ ] What's our DNS setup?
- [ ] How do we handle external API calls?
- [ ] What firewall rules do we have?
- [ ] How do we access production systems for debugging?
- [ ] What's our approach to network security?

---

## Key Concepts (System Design Interview)

### VPC Design
- **Subnets**: Public vs Private
- **CIDR blocks**: IP address allocation
- **Availability Zones**: Multi-AZ for redundancy
- **Weave's VPC design**: [Fill in]

### Service Discovery
| Method | Description | Pros | Cons |
|--------|-------------|------|------|
| DNS | Resolve service names | Simple, universal | TTL caching issues |
| Service Registry | Central registry (Consul) | Real-time updates | Additional component |
| K8s Services | Built-in K8s DNS | Native, automatic | K8s only |
| **Weave uses**: | [Fill in] | | |

### Network Security
- **Security Groups**: Instance-level firewall
- **Network ACLs**: Subnet-level firewall
- **WAF**: Web Application Firewall
- **Weave's approach**: [Fill in]

### DNS Architecture
- **External DNS**: Public-facing domains
- **Internal DNS**: Service discovery
- **Split-horizon**: Different responses internal vs external
- **Weave's DNS setup**: [Fill in]

---

## Weave Implementation

### VPC Structure
| VPC | Purpose | CIDR | Region |
|-----|---------|------|--------|
| [VPC 1] | [Purpose] | [CIDR] | [Region] |
| [VPC 2] | [Purpose] | [CIDR] | [Region] |

### Subnet Layout
| Subnet | Type | Purpose | CIDR |
|--------|------|---------|------|
| [Subnet 1] | Public/Private | [Purpose] | [CIDR] |
| [Subnet 2] | Public/Private | [Purpose] | [CIDR] |

### DNS Configuration
- **External domain**: [domain.com]
- **Internal domain**: [internal.domain]
- **DNS provider**: [Route53 / Cloud DNS / etc.]

### Firewall Rules
| Rule | Source | Destination | Port | Action |
|------|--------|-------------|------|--------|
| [Rule 1] | [Source] | [Dest] | [Port] | Allow/Deny |

### Cross-Cluster Communication
- **Method**: [VPC Peering / Service Mesh / etc.]
- **How it works**: [Fill in]

### External API Access
- **Egress method**: [NAT Gateway / Proxy / etc.]
- **Allowed destinations**: [Fill in]
- **Rate limiting**: [Fill in]

---

## Resources

- Network Architecture Docs: [Link]
- VPC Configuration: [Link]
- Firewall Rules: [Link]

---

## Notes

*Add your notes here as you learn:*
