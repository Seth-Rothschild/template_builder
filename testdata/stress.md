---
template_version: slick
title: Quarterly Infrastructure Review
subtitle: Latency, Cost, and Reliability
author: Platform Team
date: 2026-09-17
header-left: Platform Team
header-right: Confidential
footer-left: Quarterly Infrastructure Review
footer-center: Draft
---

# Summary {#sec:summary}

Median request latency fell by 18% this quarter[^latency].

[^latency]: Measured at the load balancer, excluding health checks.

> [!NOTE]
> All figures are preliminary until finance closes the quarter.

## Goals

- Reduce p50 latency
- Migrate the billing service
- Retire the legacy queue

# Results {#sec:results}

The cost model is $C = \sum_{i=1}^{n} r_i \cdot h_i$, where $r_i$ is the
hourly rate and $h_i$ the hours consumed.

$$
A = \frac{U}{U + D}
$$

| Service  | p50 (ms) | p99 (ms) | Cost ($)  | Notes                          |
| :------- | -------: | -------: | --------: | :----------------------------- |
| Gateway  |       12 |       88 |    14,200 | Autoscaling enabled            |
| Billing  |       31 |      240 |     9,850 | Migrated to **new cluster**    |
| Search   |       47 |      610 |    22,400 | Index rebuild in `week 6`      |
| Auth     |        8 |       35 |     3,100 | Moved to a shared cache        |

| Team     | Budget | Used  |
| -------- | ------ | ----- |
| Payments | 43 min | 12 min |
| Identity | 43 min | 51 min |

## Cross References

Raw LaTeX style: see Figure \ref{fig:volume} on page \pageref{fig:volume}
and Section \ref{sec:results}.

GFM anchor style: see [the results section](#sec:results), [the goals](#goals),
and [the long table](#a-very-long-table).

## A Very Long Table

| # | Region          | Incident                        | Duration |
| - | --------------- | ------------------------------- | -------- |
| 1 | us-east-1       | DNS resolver saturation         | 14 min   |
| 2 | us-east-1       | Certificate expiry on edge      | 6 min    |
| 3 | eu-west-1       | Disk pressure on log shippers   | 22 min   |
| 4 | eu-west-1       | Config push rollback            | 9 min    |
| 5 | ap-southeast-2  | Cross-region replication lag    | 41 min   |
| 6 | us-west-2       | Noisy neighbor on shared host   | 17 min   |
| 7 | us-west-2       | Rate limiter misconfiguration   | 11 min   |
| 8 | us-east-1       | Upstream provider outage        | 63 min   |
| 9 | eu-central-1    | Kafka partition rebalance storm | 28 min   |
| 10 | sa-east-1      | Expired API token in cron job   | 4 min    |
| 11 | us-east-1      | Memory leak in sidecar          | 33 min   |
| 12 | ap-northeast-1 | Load balancer health check flap | 7 min    |
| 13 | us-west-2      | Schema migration lock           | 19 min   |
| 14 | eu-west-1      | TLS handshake regression        | 12 min   |
| 15 | us-east-1      | Queue consumer deadlock         | 26 min   |

# Special Characters

Costs rose 5% & margins held at 30% for team_a and team_b. The path
`C:\builds\#42` and the tilde ~ and caret ^ should survive, as should
{braces} and a literal \LaTeX string.

Raw LaTeX a user might try: \newpage and \textbf{bold via latex}.

<div style="page-break-after: always"></div>

~~Deprecated~~ items are struck through. Visit https://example.com or
[the dashboard](https://example.com/dash).

```python
def cost(rates, hours):
    return sum(r * h for r, h in zip(rates, hours))
```

![Request volume by hour \label{fig:volume}](image.png)

![Request volume by region](image.png){#fig:region width=50%}

# Long Content

A long unbroken identifier: supercalifragilisticexpialidocious_configuration_value_that_never_ends_and_keeps_going_forever.

A long inline code span: `kubectl get pods --all-namespaces --field-selector=status.phase!=Running -o wide --sort-by=.metadata.creationTimestamp`.

```python
result = compute_quarterly_infrastructure_cost(rates=hourly_rates_by_region, hours=consumed_hours_by_service, discounts=negotiated_discounts, currency='USD')
```

```
2026-09-17T04:00:12Z level=error service=billing msg="upstream request failed" upstream=https://payments.internal.example.com/v2/charges/authorize retry=3
```
