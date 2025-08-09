# Container Deployment

Deploy with Docker/Podman or into Kubernetes.

## Docker Compose
`docker-compose.yml` example:
```
version: '3.9'
services:
  sb29guard:
    image: ghcr.io/your-org/sb29-guard:latest # or build locally
    container_name: sb29guard
    ports:
      - "8080:8080"
    environment:
      SB29_LAW_URL: "https://search-prod.lis.state.oh.us/api/v2/general_assembly_135/legislation/sb29/05_EN/pdf/"
      # SB29_ALLOW_HOST_FALLBACK: "true"   # only for A-record redirect topology
    volumes:
      - ./policy:/app/policy:ro
    command: ["serve","--policy","/app/policy/domains.yaml"]
```

Reverse proxy (Traefik) labels example:
```
labels:
  - "traefik.enable=true"
  - "traefik.http.routers.sb29.rule=Host(`blocked.guard.local`)"
  - "traefik.http.routers.sb29.entrypoints=websecure"
  - "traefik.http.routers.sb29.tls.certresolver=letsencrypt"
  - "traefik.http.services.sb29.loadbalancer.server.port=8080"
  - "traefik.http.middlewares.sb29-headers.headers.customrequestheaders.X-Original-Host: blocked.guard.local"
  - "traefik.http.middlewares.sb29-headers.headers.customrequestheaders.X-Forwarded-Host: blocked.guard.local"
  - "traefik.http.routers.sb29.middlewares=sb29-headers"
```

## Kubernetes (Ingress)
Simple Deployment + Service:
```
apiVersion: apps/v1
kind: Deployment
metadata: { name: sb29guard }
spec:
  replicas: 1
  selector: { matchLabels: { app: sb29guard } }
  template:
    metadata: { labels: { app: sb29guard } }
    spec:
      containers:
        - name: sb29guard
          image: ghcr.io/your-org/sb29-guard:latest
          args: ["serve","--policy","/policy/domains.yaml"]
          ports: [{ containerPort: 8080 }]
          env:
            - name: SB29_LAW_URL
              value: "https://search-prod.lis.state.oh.us/api/v2/general_assembly_135/legislation/sb29/05_EN/pdf/"
          volumeMounts:
            - name: policy
              mountPath: /policy
      volumes:
        - name: policy
          configMap:
            name: sb29-policy
---
apiVersion: v1
kind: Service
metadata: { name: sb29guard }
spec:
  selector: { app: sb29guard }
  ports:
    - port: 8080
      targetPort: 8080
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: sb29guard
  annotations:
    nginx.ingress.kubernetes.io/configuration-snippet: |
      proxy_set_header X-Original-Host $host;
      proxy_set_header X-Forwarded-Host $host;
spec:
  tls:
    - hosts: [blocked.guard.local]
      secretName: sb29guard-tls
  rules:
    - host: blocked.guard.local
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: sb29guard
                port:
                  number: 8080
```
