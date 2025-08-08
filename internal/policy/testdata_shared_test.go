package policy

const samplePolicyYAML = `version: 0.1.0
updated: 2025-08-08
records:
  - domain: "exampletool.com"
    classification: NO_DPA
    rationale: "Vendor has not signed"
    last_review: 2025-08-01
    status: active
  - domain: "*.trackingwidgets.io"
    classification: EXPIRED_DPA
    rationale: "Expired DPA"
    last_review: 2025-07-15
    status: active
`
